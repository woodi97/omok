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

// array of connections Connection[]
var connections = []Connection{}

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
			fmt.Println("Connection closed: ", err.Error())
			return
		}

		// meesage format: <nickname>:<port>
		recv := string(buffer[:n])
		fmt.Println("recv message: ", recv)
		splitted := strings.Split(recv, ":")

		if len(splitted) != 2 {
			fmt.Println("Invalid message format: ", recv)
			return
		}

		// search name is already used
		for _, connection := range connections {
			if connection.nickname == splitted[0] {
				fmt.Println("Nickname is already used: ", splitted[0])
				_, err = conn.Write([]byte("Nickname is already used"))
				if err != nil {
					fmt.Println("Error sending message: ", err.Error())
				}
				return
			}
		}

		// save connection
		connections = append(connections, Connection{
			nickname:   splitted[0],
			connection: conn,
			ip:         conn.RemoteAddr().(*net.TCPAddr).IP.String(),
			port:       splitted[1],
		})

		// run matching logic
		if len(connections) == 2 {
			firstPlayer := connections[0]
			secondPlayer := connections[1]

			// send opponent info
			// <nickname>:<ip>:<port>
			_, err = firstPlayer.connection.Write([]byte(
				fmt.Sprintf("%s:%s:%s", secondPlayer.nickname, secondPlayer.ip, secondPlayer.port),
			))

			if err != nil {
				fmt.Println("Error sending message: ", err.Error())
				return
			}

			_, err = secondPlayer.connection.Write([]byte(
				fmt.Sprintf("%s:%s:%s", firstPlayer.nickname, firstPlayer.ip, firstPlayer.port),
			))

			if err != nil {
				fmt.Println("Error sending message: ", err.Error())
				return
			}

			// clear connections
			connections = []Connection{}
		}
	}
}
