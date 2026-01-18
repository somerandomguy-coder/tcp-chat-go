package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	// "github.com/gorilla/websocket"
)

//	var upgrader = websocket.Upgrader{
//		ReadBufferSize:  1024,
//		WriteBufferSize: 1024,
//	}
func handlePacket(conn net.Conn) string {
	header := make([]byte, 4)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		handleError(err, "", conn)
	}
	headerSize := binary.BigEndian.Uint32(header)
	payload := make([]byte, headerSize)
	_, err = io.ReadFull(conn, payload)
	result := string(payload)
	return result
}

func read_server_msg(conn net.Conn, err error) {
	for {
		line := handlePacket(conn)
		if err != nil {
			break
		}
		fmt.Println(line)
	}
}

func sendMsg(msg string, conn net.Conn) error {
	headerSize := len(msg)
	header := uint32(headerSize)
	payload := []byte(msg)

	err := binary.Write(conn, binary.BigEndian, header)
	err = binary.Write(conn, binary.BigEndian, payload)
	if err != nil {
		fmt.Printf("failed to send msg")
	}
	return err

}

func handleError(err error, msg string, conn net.Conn) {
	if msg == "" {
		msg = "something went wrong"
	}
	if conn != nil {
		sendMsg(msg, conn)
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
		sendMsg(input, conn)
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
