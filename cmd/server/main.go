package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"slices"
	"sync"

	// "net/http"
	"os"
	// "github.com/gorilla/websocket"
)

type Client struct {
	name       string
	connection net.Conn
}

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

//	func handler(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK) // Set the status code
//		fmt.Fprintln(w, "Hello, World!")
//		fmt.Println(r.RemoteAddr)
//		fmt.Println(r.Header.Get("X-Forwarded-For"))
//	}
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
	}
	headerSize := binary.BigEndian.Uint32(header)
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
		fmt.Printf("failed to send msg")
	}
	return err

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var name string
	sendMsg("Please put in your name", conn)
	name = handlePacket(conn)
	for name == "" {
		sendMsg("Please put in your name, your name can't be empty", conn)
		name = handlePacket(conn)
	}

	defer handleDisconnect(conn)
	sendMsg("Hello "+name+"\n", conn)
	broadcast(name + " joined the chat")
	newClient := Client{name, conn}
	clientMu.Lock()
	clients = append(clients, newClient)
	clientMu.Unlock()

	for {
		msg := name + ": " + handlePacket(conn)
		fmt.Fprintln(os.Stdout, msg)
		broadcast(msg)
	}
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
	// http.HandleFunc("/", handler)
	fmt.Println("server is serving...")
	// http.ListenAndServe(":8080", nil)
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
