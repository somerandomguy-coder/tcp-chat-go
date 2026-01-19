package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"sync"
)

type Client struct {
	name       string
	connection net.Conn
}

var (
	clients  []Client
	clientMu sync.Mutex
)

func handleError(err error, msg string, conn net.Conn) {
	if msg == "" {
		msg = "something went wrong"
	}
	if conn != nil {
		sendMsg(msg, conn)
	}
	fmt.Fprintln(os.Stderr, msg, err)
}

func handlePacket(conn net.Conn) string {
	header := make([]byte, 4)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		handleError(err, "", conn)
		return ""
	}
	headerSize := binary.BigEndian.Uint32(header)
	if headerSize > 4*1024*1024 {
		sendMsg("File too big, please reconnect", conn)
		return ""
	}
	payload := make([]byte, headerSize)
	_, err = io.ReadFull(conn, payload)
	result := string(payload)
	return result
}

func sendMsg(msg string, conn net.Conn) error {
	headerSize := len(msg)
	header := uint32(headerSize)
	payload := []byte(msg)

	err := binary.Write(conn, binary.BigEndian, header)
	err = binary.Write(conn, binary.BigEndian, payload)
	if err != nil {
		fmt.Println("failed to send msg")
	}
	return err

}

func handleConnection(conn net.Conn) string {
	defer conn.Close()

	var name string
	sendMsg("Please put in your name", conn)
	name = handlePacket(conn)
	if name == "" {
		sendMsg("Please put in your name next time, your name can't be empty", conn)
		return ""
	}

	defer handleDisconnect(conn)
	sendMsg("Hello "+name+"\n", conn)
	broadcast(name + " joined the chat")
	newClient := Client{name, conn}
	clientMu.Lock()
	clients = append(clients, newClient)
	clientMu.Unlock()

	for {
		msg := handlePacket(conn)
		if msg == "" {
			break
		}
		msg = name + ": " + msg
		_, err := fmt.Fprintln(os.Stdout, msg)
		if err != nil {
			return ""
		}
		broadcast(msg)
	}
	return ""
}

func handleDisconnect(conn net.Conn) {
	clientMu.Lock()
	for index, client := range clients {
		if client.connection == conn {
			clients = slices.Delete(clients, index, index+1)
			clientMu.Unlock()
			broadcast(client.name + " has left the chat")
			break
		}
	}
}

func broadcast(msg string) {
	clientMu.Lock()
	defer clientMu.Unlock()
	for _, client := range clients {
		err := sendMsg(msg, client.connection)
		if err != nil {
			fmt.Printf("Broadcast failed for %s, removing them.\n", client.name)
		}
	}

}

func main() {
	fmt.Println("server is serving...")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("something went wrong")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			handleError(err, "", conn)
		}
		go handleConnection(conn)

	}
}
