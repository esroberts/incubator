package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"rustedskyline.io/tcpchat/internal/common"
	chat "rustedskyline.io/tcpchat/internal/proto"
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

		sizeHeader, messageBody := marshalMessage(input, conn.LocalAddr().String())
		data := append(sizeHeader[:], messageBody...)

		if _, err := conn.Write(data); err != nil {
			log.Printf("error on writing to connection %v\n", err)
		}
	}
}

// Read from connection and print contents indefinitely
func readServer(conn net.Conn) {
	//TODO: error cleanly exits program
	for {
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		if err != nil {
			log.Printf("error reading from connection %v", err)
		}
		messageBody := data[4:n]

		protoMessage := &chat.Message{}
		err = proto.Unmarshal(messageBody, protoMessage)
		if err != nil {
			log.Printf("error unmarshalling proto message %v", err)
		}

		fmt.Printf(protoMessage.Text)
		prompt()
	}
}

func prompt() {
	fmt.Print(">> ")
}

func marshalMessage(text string, fromIp string) ([]byte, []byte) {
	text = text + string(common.MessageDelim)
	message := &chat.Message{Text: text,
		FromIp:       fromIp,
		UtcTimestamp: time.Now().UTC().Unix()}
	data, err := proto.Marshal(message)
	if err != nil {
		log.Printf("error marshalling proto message %v", err)
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(data)))
	return b, data
}
