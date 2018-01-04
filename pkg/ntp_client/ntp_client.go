package ntp_client

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
)

type NTPClient struct {
	pairPackets  []*PairPacketNTP
	QueryOptions *QueryOptions
}

type PacketNTP struct {
	LiVnMode  uint8 // Leap indicator(2), Version number(3), Mode(3)
	Stratum   uint8 // Stratum level of the local clock (8)
	Poll      uint8 // Maximum interval between messages(8)
	Precision uint8 // Precision of local clock (8)

	RootDelay      uint32 // Total round trip delay time (32)
	RootDispersion uint32 // Max error aloud from primary clock source (32)
	ReferenceID    uint32 // Reference clock identifier (32)

	ReferenceTime uint64 // Reference time stamp (64)
	OriginTime    uint64 // Origin time stamp (64)
	ReceivedTime  uint64 // Received time stamp (64)
	TransmitTime  uint64 // Transmitted time stamp (64)
}

type PairPacketNTP struct {
	queryPacket      *PacketNTP
	recvPacket       *PacketNTP
	localAddress     *net.UDPAddr
	remoteAdderss    *net.UDPAddr
	serverHost       string
	transmissionTime time.Duration
}

type QueryOptions struct {
	Timeout      time.Duration // defaults to 5 seconds
	Version      int           // NTP protocol version, defaults to 4
	TTL          int           // IP TTL to use, defaults to system default
	Port         int           // Server port, defaults to 123
	LocalAddress string        // IP address to use for the client address, defaults to golang net
}

var (
	Epoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC) //Epoch of NTP
)

func NewNTPClient(hosts []string) *NTPClient {
	ntpClient := new(NTPClient)

	ntpClient.QueryOptions = NewQueryOptions()

	for _, host := range hosts {
		ntpClient.pairPackets = append(ntpClient.pairPackets, NewPairPacket(host))
	}

	return ntpClient
}

func (n *NTPClient) GetOffset() (offset time.Duration, err error) {
	var result *multierror.Error
	var offsets []time.Duration

	for _, pair := range n.pairPackets {
		off, err := pair.requestTime(n.QueryOptions)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting time from pair: %v", err))
		} else {
			offsets = append(offsets, off)
		}
	}

	offset, err = n.averageOffSet(offsets)
	if err != nil {
		result = multierror.Append(result, err)
	}

	return offset, result.ErrorOrNil()
}

func NewPairPacket(host string) *PairPacketNTP {
	pairPacket := &PairPacketNTP{
		recvPacket:   new(PacketNTP),
		localAddress: new(net.UDPAddr),
		serverHost:   host,
	}

	pairPacket.queryPacket = pairPacket.NewQueryPacket()

	return pairPacket
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		Timeout: (time.Second * 5),
		Version: 4,
		Port:    123,
	}
}

func (p *PairPacketNTP) NewQueryPacket() *PacketNTP {
	packet := &PacketNTP{
		LiVnMode: 0x1b, // 00,100,011 leep indicator=0, version=4, mode=3 (client)
	}

	return packet
}

func (p *PairPacketNTP) requestTime(options *QueryOptions) (offset time.Duration, err error) {
	var result *multierror.Error

	if options.LocalAddress != "" {
		p.localAddress, err = net.ResolveUDPAddr("udp", net.JoinHostPort(options.LocalAddress, "0"))
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to resolve local address from options: %v", err))
		}
	}

	if options.Version < 2 || options.Version > 4 {
		result = multierror.Append(result, fmt.Errorf("requested ntp version number is not supported: %d", options.Version))
	}

	if result != nil {
		return 0, result
	}

	p.setVersionNumber(options.Version)

	p.remoteAdderss, err = net.ResolveUDPAddr("udp", net.JoinHostPort(p.serverHost, strconv.Itoa(options.Port)))
	if err != nil {
		return 0, fmt.Errorf("failed to resolve host address: %v", err)
	}

	con, err := net.DialUDP("udp", p.localAddress, p.remoteAdderss)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to remote server: %v", err)
	}
	defer con.Close()

	con.SetDeadline(time.Now().Add(options.Timeout))

	// Use a random transmit time in message to increase privacy and prevent spoofing
	randomBits := make([]byte, 8)
	if _, err := rand.Read(randomBits); err != nil {
		return 0, fmt.Errorf("failed to generate random bits for transmit time: %v", err)
	}

	p.queryPacket.TransmitTime = uint64(binary.BigEndian.Uint64(randomBits))
	realTransmitTime := time.Now()

	// Send query
	if err := binary.Write(con, binary.BigEndian, p.queryPacket); err != nil {
		return 0, fmt.Errorf("failed to send ntp query to server: %v", err)
	}

	// Receive response
	if err := binary.Read(con, binary.BigEndian, p.recvPacket); err != nil {
		return 0, fmt.Errorf("responding network error: %v", err)
	}

	p.transmissionTime = time.Since(realTransmitTime)
	p.queryPacket.OriginTime = p.ntpTime(realTransmitTime)

	return p.calculateOffset(), nil
}

func (p *PairPacketNTP) setVersionNumber(version int) {
	p.queryPacket.LiVnMode = (p.queryPacket.LiVnMode & 0xc7) | (uint8(version) << 3) // ( LiVnMode & 11000111) | 00VER111
}

func (p *PairPacketNTP) calculateOffset() time.Duration {
	//t1 = local clock, time request sent by client;
	//t2 = server clock, time request received by server;
	//t3 = server clock, time reply sent by server;
	//t4 = local clock, time reply received by client
	//o = ((t2 - t1) + (t3 - t4)) / 2
	t2t1 := p.epochTime(p.recvPacket.ReceivedTime).Sub(p.epochTime(p.queryPacket.OriginTime))
	t3t4 := p.epochTime(p.recvPacket.TransmitTime).Sub(p.epochTime(p.recvPacket.ReceivedTime))
	return (t2t1 + t3t4) // / time.Duration(2)
}

//golang time to 64 bit conv
func (p *PairPacketNTP) ntpTime(goTime time.Time) uint64 {
	nsec := uint64(goTime.Sub(Epoch))
	nanoPerSec := uint64(1e9)
	sec := nsec / nanoPerSec
	frac := (((nsec - sec*nanoPerSec) << 32) + nanoPerSec - 1) / nanoPerSec
	return uint64(sec<<32 | frac)
}

func (p *PairPacketNTP) duration(ntptime uint64) time.Duration {
	nanoPerSec := uint64(1e9)
	sec := (ntptime >> 32) * nanoPerSec
	frac := (ntptime & 0xffffffff) * nanoPerSec >> 32
	return time.Duration(sec + frac)
}

func (p *PairPacketNTP) epochTime(atime uint64) time.Time {
	return Epoch.Add(p.duration(atime))
}

func (p *PairPacketNTP) verifyResponsePacket() error {
	var result *multierror.Error

	//verify server mode in response
	if (p.recvPacket.LiVnMode & 0x07) != 0x04 {
		result = multierror.Append(result, errors.New("response packet did not respond with server mode"))
	}

	if p.recvPacket.OriginTime != p.queryPacket.TransmitTime {
		result = multierror.Append(result, errors.New("response transmit time did not match query's"))
	}

	if p.recvPacket.ReceivedTime > p.recvPacket.TransmitTime {
		result = multierror.Append(result, errors.New("server clock has ticked backwards"))
	}

	return result.ErrorOrNil()
}

func (n *NTPClient) averageOffSet(offsets []time.Duration) (offset time.Duration, err error) {
	var avg int64

	if len(offsets) == 0 {
		return time.Duration(0), errors.New("no offsets to average - check internet connection?")
	}

	for _, offset := range offsets {
		avg += offset.Nanoseconds()
	}

	return time.Duration(avg / int64(len(offsets))), nil
}
