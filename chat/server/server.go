package main

import (
	"net"
	"strings"
)

const ip = "127.0.0.1"
const port = 30000

func main() {
	// create udp connection
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// recv message
	buffer := make([]byte, 1024)
	for {
		n, recvConn, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		message := strings.TrimSpace(string(buffer[:n]))
		println("recv message:", message)

		// send message back to client
		_, err = conn.WriteToUDP([]byte(message), recvConn)
		if err != nil {
			panic(err)
		}
	}
}
