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
	"time"
)

const (
	Row = 10
	Col = 10
)

const maxNameLen = 64

const SERVER_IP = "127.0.0.1"
const SERVER_PORT = "30000"

var timer *time.Timer

const turnLimit = time.Second * 10

var gamePlayingStatus = PLAYING

type gameStatus int
type command int

const (
	MOVE command = iota
	EXIT
	GG
	CHAT
)

const (
	PLAYING gameStatus = iota
	WIN
	LOSE
	DRAW
)

type Board [][]int

type User struct {
	nickname string
	ip       string
	port     int
	isFirst  bool
}

func convertToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

// ========= Network =========

func match(nickname string) (me User, opponent User) {
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

	// recv opponent info <nickname>:<ip>:<port>:<my turn>:<opponent turn>
	recv := string(buffer[:n])
	splitted := strings.Split(recv, ":")
	if len(splitted) != 5 {
		panic("invalid opponent info")
	}

	myTurn := convertToInt(splitted[3])
	opponentTurn := convertToInt(splitted[4])

	return User{
			nickname: nickname,
			ip:       SERVER_IP,
			port:     clientPort,
			isFirst:  myTurn == 0,
		}, User{
			nickname: splitted[0],
			ip:       splitted[1],
			port:     convertToInt(splitted[2]),
			isFirst:  opponentTurn == 0,
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

func checkUserName(name string) bool {
	if len(name) > maxNameLen {
		return false
	}

	for _, c := range name {
		if c < 'a' || c > 'z' {
			return false
		}
	}

	return true
}

func runNextTurn(turn *int, count *int, win *int, board Board, x, y int) {
	board[x][y] = *turn%2 + 1
	*count++
	printBoard(board)

	if *count == Row*Col {
		fmt.Println("draw")
		gamePlayingStatus = DRAW
		return
	}

	*win = checkWin(board, x, y)
	if *win != 0 {
		if *win == 1 {
			fmt.Println("you win")
			gamePlayingStatus = WIN
		} else {
			fmt.Println("you lose")
			gamePlayingStatus = LOSE
		}
		return
	}

	*turn++
	timer = time.AfterFunc(turnLimit, func() {
		fmt.Println("time over")
		gamePlayingStatus = LOSE
	})

	clear()
}

func runTimer(conn *net.UDPConn) {
	timer = time.AfterFunc(turnLimit, func() {
		gamePlayingStatus = LOSE
		fmt.Println("time over")
		sendMessage(conn, fmt.Sprintf("%d", int(GG)))
		gamePlayingStatus = LOSE
	})
}

func main() {
	// @STEP 1: get command
	// example: go run P2POmokClient.go <nickname>
	if len(os.Args) != 2 {
		panic("invalid command, must enter nickname!")
	}

	nickname := os.Args[1]
	if !checkUserName(nickname) {
		panic("invalid nickname")
	}

	me, opponent := match(nickname)
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

	if me.isFirst {
		fmt.Println("you are first")
	} else {
		fmt.Println("you are second")
	}

	time.Sleep(1 * time.Second)

	// set timer
	if me.isFirst {
		go runTimer(sendConn)
	}

	// init board
	board := Board{}
	turn, count, win := 0, 0, 0
	for i := 0; i < Row; i++ {
		var tempRow []int
		for j := 0; j < Col; j++ {
			tempRow = append(tempRow, 0)
		}
		board = append(board, tempRow)
	}

	// recv message
	go func(c *net.UDPConn) {
		for {
			// message format: <command> ...etc
			msg := recvMessage(c)

			splitted := strings.Split(strings.TrimSpace(msg), " ")
			if len(splitted) == 0 {
				continue
			}

			cmd, _ := strconv.Atoi(splitted[0])

			switch command(cmd) {
			case MOVE:
				// run game logic
				x := convertToInt(splitted[1])
				y := convertToInt(splitted[2])

				runNextTurn(&me.port, &me.port, &me.port, board, x, y)
			case EXIT:
				fmt.Println("opponent exited")
				gamePlayingStatus = WIN
				return
			case GG:
				fmt.Println("opponent gave up")
				gamePlayingStatus = WIN
				return
			case CHAT:
				fmt.Printf("%s> %s\n", opponent.nickname, strings.Join(splitted[1:], " "))
			}

		}
	}(recvConn)

	// game logic
	clear()

	printBoard(board)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("error reading: ", err.Error())
			return
		}

		splittedInput := strings.Split(strings.TrimSpace(input), " ")
		if len(splittedInput) == 0 {
			fmt.Println("invalid command")
			continue
		}

		// command list (1): "\gg", "\exit", "\\ <x> <y>"
		cmd := strings.TrimSpace(splittedInput[0])

		if cmd == "\\exit" {
			sendMessage(sendConn, fmt.Sprintf("%d", int(EXIT)))
			if timer != nil {
				timer.Stop()
			}
			return
		}

		if gamePlayingStatus != PLAYING {
			fmt.Println("game is already finished")
			continue
		}

		switch cmd {
		case "\\gg":
			sendMessage(sendConn, fmt.Sprintf("%d", int(GG)))
			gamePlayingStatus = LOSE

			if timer != nil {
				timer.Stop()
			}
			continue
		case "\\\\":
			// move command
			if len(splittedInput) != 3 {
				fmt.Println("invalid command")
				continue
			}

			if (opponent.isFirst && turn%2 == 0) || (!opponent.isFirst && turn%2 == 1) {
				fmt.Println("not your turn")
				continue
			}

			x := convertToInt(splittedInput[1])
			y := convertToInt(splittedInput[2])

			// check valid move
			if x < 0 || y < 0 || x >= Row || y >= Col {
				fmt.Println("error, out of bound!")
				time.Sleep(1 * time.Second)
				continue
			} else if board[x][y] != 0 {
				fmt.Println("error, already used!")
				time.Sleep(1 * time.Second)
				continue
			}

			// stop timer
			if timer != nil {
				timer.Stop()
			}

			// send move command
			sendMessage(sendConn, fmt.Sprintf("%d %d %d", int(MOVE), x, y))

			runNextTurn(&turn, &count, &win, board, x, y)

		default:
			// chat command
			sendMessage(sendConn, fmt.Sprintf("%d %s", int(CHAT), input))
		}
	}
}
