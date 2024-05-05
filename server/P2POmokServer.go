package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// save connection list
// {
// 	"nickname": string
// 	"connection": net.Conn
// 	"ip": string
// 	"port": string
// }

type Connection struct {
	nickname   string
	connection net.Conn
	ip         string
	port       string
}

var connections = make(map[string]Connection)

// @TODO: Make it const
const PORT = 30000

// @TODO: make utility
func convertToString(value int) string {
	return strconv.Itoa(value)
}

func main() {
	listener, err := net.Listen("tcp", ":"+convertToString(PORT))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server is running at port", PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}

		// meesage format: <nickname>:<ip>:<port>
		recv := string(buffer[:n])
		splitted := strings.Split(recv, ":")

		if len(splitted) != 3 {
			fmt.Println("Invalid message format: ", recv)
			return
		}

		nickname := splitted[0]
		connections[nickname] = Connection{
			nickname:   splitted[0],
			connection: conn,
			ip:         splitted[1],
			port:       splitted[2],
		}

		// run matching logic

		fmt.Println("Received message: ", recv)
	}
}
