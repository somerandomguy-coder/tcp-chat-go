package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	// "github.com/gorilla/websocket"
)

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

func read_server_msg(conn net.Conn, err error) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if err != nil {
			break
		}
		fmt.Println(line)
	}
}

func handleError(err error, msg string, conn net.Conn) {
	if msg == "" {
		msg = "something went wrong"
	}
	if conn != nil {
		fmt.Fprintln(conn, msg, err)
	}
	fmt.Fprintln(os.Stderr, msg, err)
}

func send_msg(conn net.Conn, err error) {
	if err != nil {
		handleError(err, "", conn)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		fmt.Fprintln(conn, input)
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Something is wrong, but I don't know what, err: ", err)
	}
	go read_server_msg(conn, err)
	send_msg(conn, err)
}
