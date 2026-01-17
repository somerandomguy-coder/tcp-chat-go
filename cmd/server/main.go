package main

import (
	"bufio"
	"fmt"
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
		fmt.Fprintln(conn, msg, err)
	}
	fmt.Fprintln(os.Stderr, msg, err)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Fprintln(conn, "200/OK")
	var name string
	fmt.Fprintln(conn, "Please put in your name")
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		name = scanner.Text()
		for name == "" {
			fmt.Fprintln(conn, "Please put in your name, your name can't be empty")
			if !scanner.Scan() {
				return
			}
			name = scanner.Text()

		}
	}
	defer handleDisconnect(conn)
	fmt.Fprintf(conn, "Hello %s \n", name)
	broadcast(name + " joined the chat")
	newClient := Client{name, conn}
	clientMu.Lock()
	clients = append(clients, newClient)
	clientMu.Unlock()
	for scanner.Scan() {
		msg := name + ": " + scanner.Text()
		fmt.Fprintln(os.Stdout, msg)
		broadcast(msg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error with %s: %v\n", name, err)
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
		_, err := fmt.Fprintln(client.connection, msg)
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
		fmt.Fprintln(os.Stderr, "something went wrong")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			handleError(err, "", conn)
		}
		go handleConnection(conn)

	}
}
