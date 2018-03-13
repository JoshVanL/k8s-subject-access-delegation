package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		handle(fmt.Errorf("Please provide one tcp service address and one pod name."))
	}

	for {

		conn, err := net.Dial("tcp", os.Args[1])
		if err != nil {
			handle(fmt.Errorf("failed to dial into conn: %v", err))
		}

		if _, err := conn.Write([]byte(os.Args[2])); err != nil {
			handle(fmt.Errorf("failed to write pod name to server: %v", err))
		}

		buf := make([]byte, 2048)

		n, err := conn.Read(buf)
		if err != nil {
			handle(fmt.Errorf("failed to read from server: %v", err))
		}

		fmt.Printf("\n------------\nGot from server:\n%s------------\n", string(buf[:n]))

		conn.Close()

		time.Sleep(time.Second * 2)
	}

}

func handle(err error) {
	fmt.Printf("%s\n", err.Error())
	os.Exit(1)
}
