package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		log.Fatalf("client failed to connect %v", err)
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("error shutting down server!")
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go readStdin(conn)
	go readServer(conn)
	wg.Wait()
}

// Read from stdin and print input to connection indefinitely
func readStdin(conn net.Conn) {
	//TODO: 'quit' command exits loop and cleanly exits program
	for {
		reader := bufio.NewReader(os.Stdin)
		prompt()
		input, _ := reader.ReadString('\n')

		// TODO: use common terminal printer
		_, err := fmt.Fprintf(conn, input+"\n")
		if err != nil {
			fmt.Printf("error printing message to connection")
		}
	}
}

// Read from connection and print contents indefinitely
func readServer(conn net.Conn) {
	//TODO: error cleanly exits program
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("Error %v\n", err)
			continue
		}
		fmt.Print(message)
		prompt()
	}
}

func prompt() {
	fmt.Print(">> ")
}
