package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const (
	Row = 10
	Col = 10
)

const SERVER_IP = "127.0.0.1"
const SERVER_PORT = "30000"

type Board [][]int

type User struct {
	nickname string
	ip       string
	port     int
}

func convertToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

// ========= Network =========

func enterQueue(nickname string) (me User, opponent User) {
	// @STEP 2: connect to server
	conn, err := net.Dial("tcp", SERVER_IP+":"+SERVER_PORT)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	connInfo := conn.LocalAddr().(*net.TCPAddr)
	clientPort := connInfo.Port
	fmt.Printf("rendezvous connection is established at %s:%d\n", SERVER_IP, clientPort)

	// send <nickname>:<port>
	_, err = conn.Write([]byte(
		fmt.Sprintf("%s:%d", nickname, clientPort),
	))
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}

	// recv opponent info <nickname>:<ip>:<port>
	recv := string(buffer[:n])
	splitted := strings.Split(recv, ":")
	if len(splitted) != 3 {
		panic("invalid opponent info")
	}

	return User{
			nickname: nickname,
			ip:       SERVER_IP,
			port:     clientPort,
		}, User{
			nickname: splitted[0],
			ip:       splitted[1],
			port:     convertToInt(splitted[2]),
		}
}

func sendMessage(c *net.UDPConn, msg string) {
	_, err := c.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}

func recvMessage(c *net.UDPConn) string {
	buffer := make([]byte, 1024)
	n, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(buffer[:n]))
}

// ========= Game Logic =========

func printBoard(b Board) {
	fmt.Print("   ")
	for j := 0; j < Col; j++ {
		fmt.Printf("%2d", j)
	}

	fmt.Println()
	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()

	for i := 0; i < Row; i++ {
		fmt.Printf("%d |", i)
		for j := 0; j < Col; j++ {
			c := b[i][j]
			if c == 0 {
				fmt.Print(" +")
			} else if c == 1 {
				fmt.Print(" 0")
			} else if c == 2 {
				fmt.Print(" @")
			} else {
				fmt.Print(" |")
			}
		}

		fmt.Println(" |")
	}

	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()
}

func checkWin(b Board, x, y int) int {
	lastStone := b[x][y]
	startX, startY, endX, endY := x, y, x, y

	// Check X
	for startX-1 >= 0 && b[startX-1][y] == lastStone {
		startX--
	}
	for endX+1 < Row && b[endX+1][y] == lastStone {
		endX++
	}

	// Check Y
	startX, startY, endX, endY = x, y, x, y
	for startY-1 >= 0 && b[x][startY-1] == lastStone {
		startY--
	}
	for endY+1 < Row && b[x][endY+1] == lastStone {
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 1
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && startY-1 >= 0 && b[startX-1][startY-1] == lastStone {
		startX--
		startY--
	}
	for endX+1 < Row && endY+1 < Col && b[endX+1][endY+1] == lastStone {
		endX++
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 2
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && endY+1 < Col && b[startX-1][endY+1] == lastStone {
		startX--
		endY++
	}
	for endX+1 < Row && startY-1 >= 0 && b[endX+1][startY-1] == lastStone {
		endX++
		startY--
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	return 0
}

func clear() {
	fmt.Printf("%s", runtime.GOOS)

	clearMap := make(map[string]func()) //Initialize it
	clearMap["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearMap["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clearMap["darwin"] = func() {
		cmd := exec.Command(("clear"))
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	value, ok := clearMap[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                             //if we defined a clearMap func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clearMap terminal screen :(")
	}
}

func main() {
	// @STEP 1: get command
	// example: go run P2POmokClient.go <nickname>
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run P2POmokClient.go <nickname>")
		return
	}

	nickname := os.Args[1]
	me, opponent := enterQueue(nickname)
	// create udp connection
	recvConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(me.ip),
		Port: me.port,
	})
	if err != nil {
		panic(err)
	}
	defer recvConn.Close()

	sendConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(opponent.ip),
		Port: opponent.port,
	})
	if err != nil {
		panic(err)
	}
	defer sendConn.Close()

	// recv message
	go func(c *net.UDPConn) {
		for {
			msg := recvMessage(c)
			fmt.Println("recv message:", msg)
		}
	}(recvConn)

	clear()
	board := Board{}
	// x, y, turn, count, win := -1, -1, 0, 0, 0
	for i := 0; i < Row; i++ {
		var tempRow []int
		for j := 0; j < Col; j++ {
			tempRow = append(tempRow, 0)
		}
		board = append(board, tempRow)
	}

	printBoard(board)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("error reading: ", err.Error())
			return
		}

		splittedInput := strings.Split(input, " ")
		if len(splittedInput) == 0 {
			fmt.Println("invalid command")
			continue
		}

		// command list (1): "\gg", "\exit", "\\ <x> <y>"
		command := splittedInput[0]
		if strings.HasPrefix(command, "\\") {
			// @TODO: check valise command
		} else {
			// think it as chat message and send it to opponent
			sendMessage(sendConn, input)
		}

		// cnt, _ := fmt.Scanf("%d %d ", &x, &y)

		// if cnt != 2 {
		// 	fmt.Println("error, must enter x y!")
		// 	time.Sleep(1 * time.Second)
		// 	continue
		// } else if x < 0 || y < 0 || x >= Row || y >= Col {
		// 	fmt.Println("error, out of bound!")
		// 	time.Sleep(1 * time.Second)
		// 	continue
		// } else if board[x][y] != 0 {
		// 	fmt.Println("error, already used!")
		// 	time.Sleep(1 * time.Second)
		// 	continue
		// }

		// if turn == 0 {
		// 	board[x][y] = 1
		// } else {
		// 	board[x][y] = 2
		// }

		// clear()
		// printBoard(board)

		// win = checkWin(board, x, y)
		// if win != 0 {
		// 	fmt.Printf("player %d wins!\n", win)
		// 	break
		// }

		// count += 1
		// if count == Row*Col {
		// 	fmt.Printf("draw!\n")
		// 	break
		// }

		// turn = (turn + 1) % 2
	}
}
