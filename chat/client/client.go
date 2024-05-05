package main

import (
	"bufio"
	"net"
	"os"
)

const ip = "127.0.0.1"
const port = 30000

func sendMessage(conn *net.UDPConn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}

func recvMessage(conn *net.UDPConn) string {
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		panic(err)
	}
	return string(buffer[:n])
}

func main() {
	// dial udp connection
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	// recv message
	go func(c *net.UDPConn) {
		for {
			msg := recvMessage(c)
			println("recv message:", msg)
		}
	}(conn)

	// send message
	for {
		msg, _ := reader.ReadString('\n')
		sendMessage(conn, msg)
	}

}
