package main

import (
	"bufio"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
)

// TODO: use logger

const portDefault = "9090"
const hostDefault = "localhost"
const welcomeSentinel = "WELCOME"

func main() {

	host := hostDefault
	port := portDefault

	log.Printf("Starting chat server on %v:%v\n", host, port)

	l, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("Error starting chat server %v\n", err)
		return
	}

	defer func() {
		err := l.Close()
		if err != nil {
			log.Fatalf("error shutting down server!")
		}
	}()

	// maintain a map of all connected clients
	var connMap = &sync.Map{}

	// accept new connections and spawn handler routine per client
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("error accepting connection %v", err)
		}

		if _, err := conn.Write([]byte("Welcome " + conn.RemoteAddr().String() + "\n")); err != nil {
			log.Printf("error on writing to connection %v", err)
		}

		id := uuid.New().String()
		connMap.Store(id, conn)

		// start a new goroutine for each client connection
		go handleUserConnection(id, conn, connMap)
	}
}

// Handle communication with a single client
func handleUserConnection(id string, clientConn net.Conn, connMap *sync.Map) {
	log.Printf("Client connected %s\n", clientConn.RemoteAddr().String())

	defer func() {
		clientAddr := clientConn.RemoteAddr().String()
		err := clientConn.Close()
		if err != nil {
			log.Printf("error closing client connection from %v", clientAddr)
		}
		connMap.Delete(id)
	}()

	broadcast(connMap, clientConn, welcomeSentinel)

	for {
		userInput, err := bufio.NewReader(clientConn).ReadString('\n')
		if err != nil {
			log.Printf("error reading from client %v", err)
			return
		}

		// Don't write empty space to clients
		if userInput == "\n" {
			continue
		}

		broadcast(connMap, clientConn, userInput)
	}
}

// Broadcast a message to all connected clients
func broadcast(connMap *sync.Map, clientConn net.Conn, userInput string) {
	//TODO: pass connection by address
	//TODO: find better pattern for welcomeSentinel

	// Fan-out write
	connMap.Range(func(key, value interface{}) bool {
			if conn, ok := value.(net.Conn); ok {
				clientAddr := clientConn.RemoteAddr().String()
				// don't send a response to the client that just talked
				if conn.RemoteAddr().String() == clientAddr {
					// Within sync.Map.Range() acts as a continue
					return true
				}

				message := "\n"
				if userInput == welcomeSentinel {
					message = clientAddr + " joined the chat"
				} else {
					message = "[" + clientAddr + "] " + userInput
				}

				if _, err := conn.Write([]byte(message + "\n")); err != nil {
					log.Printf("error on writing to connection %v\n", err)
				}
			}
			return true
		})
}
