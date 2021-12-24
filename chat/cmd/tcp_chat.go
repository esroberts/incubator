package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
)

const WELCOME_SENTINEL = "WELCOME"

func main() {
	l, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		return
	}

	defer l.Close()

	var connMap = &sync.Map{}

	for {
		// Accept client connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("error accepting connection %v", err)
		}

		if _, err := conn.Write([]byte("Welcome " + conn.RemoteAddr().String() + "\n")); err != nil {
			fmt.Println("error on writing to connection %v", err)
		}

		id := uuid.New().String()
		connMap.Store(id, conn)

		// start a new goroutine for each client connection
		go handleUserConnection(id, conn, connMap)
	}
}

func handleUserConnection(id string, c net.Conn, connMap *sync.Map) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	defer func() {
		c.Close()
		connMap.Delete(id)
	}()

	fanOutWrite(connMap, c, WELCOME_SENTINEL)

	for {
		userInput, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Printf("error reading from client %v", err)
			return
		}

		// Don't write empty space to clients
		if userInput == "\n" {
			continue
		}

		// // Fan-out write
		// connMap.Range((func(key, value interface{}) bool {
		// 	if conn, ok := value.(net.Conn); ok {
		// 		// don't send a response to the client that just talked
		// 		if conn.RemoteAddr().String() == c.RemoteAddr().String() {
		// 			// Within sync.Map.Range() acts as a continue
		// 			return true
		// 		}
		// 		if _, err := conn.Write([]byte(conn.RemoteAddr().String() + ": " + userInput)); err != nil {
		// 			fmt.Println("error on writing to connection %v", err)
		// 		}
		// 	}
		// 	return true
		// }))
		fanOutWrite(connMap, c, userInput)
	}
}

//TODO: pass connection by address
func fanOutWrite(connMap *sync.Map, clientConn net.Conn, userInput string) {
	// Fan-out write
	connMap.Range((func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			// don't send a response to the client that just talked
			if conn.RemoteAddr().String() == clientConn.RemoteAddr().String() {
				// Within sync.Map.Range() acts as a continue
				return true
			}

			clientAddr := conn.RemoteAddr().String()

			var message string = "\n"
			if userInput == WELCOME_SENTINEL {
				message = clientAddr + " joined the chat\n"
			} else {
				message = "[" + clientAddr + "] " + userInput
			}

			if _, err := conn.Write([]byte(message)); err != nil {
				fmt.Println("error on writing to connection %v", err)
			}
		}
		return true
	}))
}
