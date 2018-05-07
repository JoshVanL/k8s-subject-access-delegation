package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const nameSpace = "default"
const addr = ":8800"
const kind = "tcp"
const service = "pod-server-svc"

func main() {
	// Setup listener
	//listener, err := net.Listen(kind, addr)
	//if err != nil {
	//	fmt.Println("Error listening:", err.Error())
	//	os.Exit(1)
	//}
	//defer listener.Close()

	//fmt.Println("Listening on " + addr)

	for {
		//// Acc connection
		//conn, err := listener.Accept()
		//if err != nil {
		//	fmt.Println("Error accepting: ", err.Error())
		//}

		//fmt.Printf("Connection opened: %v\n", conn.LocalAddr().String())

		//// Handle connection -- concurrent
		//go handleReq(conn)

		go get("pod-server-pod")
		//if _, _, err := get(os.Args[1]); err != nil {
		//panic(err)
		time.Sleep(time.Second)
	}
}

func handleReq(conn net.Conn) {
	// Read buffer from conn
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Skipping request... error reading message: %v", err)

	} else {
		// Get pod ane svc using input as pod name
		resp := string(buf[:n])

		podName := strings.TrimSpace(resp)
		pod, svc, err := get(podName)
		respond(conn, podName, pod, svc, err)
	}

	if err := conn.Close(); err != nil {
		fmt.Printf("Error closing connection: %v", err)
	} else {
		fmt.Printf("Connection closed: %v\n", conn.LocalAddr().String())
	}
}

func get(podName string) (podP *corev1.Pod, svcP *corev1.Service, err error) {
	var errs error

	// Read this Kubernetes config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	// Get pod using input as pod name
	options := metav1.GetOptions{}
	pod, err := c.Core().Pods(nameSpace).Get(podName, options)
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Error getting pod by name: %v", err))
	}

	// Get service
	svc, err := c.Core().Services(nameSpace).Get(service, options)
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Error getting service by name: %v", err))
	}

	return pod, svc, errs
}

// Send back response
func respond(conn net.Conn, podName string, pod *corev1.Pod, svc *corev1.Service, err error) {

	buff := []byte(fmt.Sprintf("(%02d:%02d:%02d) Getting names from pod: %s and service: %s\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second(), podName, service))

	if err != nil || pod == nil || svc == nil {
		//conn.Write([]byte(fmt.Sprintf("Errors: %s\n", err)))
		buff = append(buff, []byte(fmt.Sprintf("Failed to get pod and service name. (Forbidden)\n"))...)

	} else {

		if pod != nil {
			buff = append(buff, []byte(fmt.Sprintf("Names: %s, %s, %s\n", pod.GetName(), pod.GetNamespace(), pod.GetCreationTimestamp().String()))...)
		}

		if svc != nil {
			buff = append(buff, []byte(fmt.Sprintf("Service: %v\n", svc.GetName()))...)
		}

	}
	conn.Write(buff)
}
