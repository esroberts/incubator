package main

import (
	"encoding/binary"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"rustedskyline.io/tcpchat/internal/common"
	chat "rustedskyline.io/tcpchat/internal/proto"
)

const portDefault = "9090"
const hostDefault = "localhost"

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

		message := "Welcome " + conn.RemoteAddr().String() + "\n"
		sizeHeader, messageBody := marshalMessage(message, conn.LocalAddr().String())
		data := append(sizeHeader[:], messageBody...)

		// log.Printf("data: %v", data)
		if _, err := conn.Write(data); err != nil {
			log.Printf("error on writing to connection %v", err)
		}

		id := uuid.New().String()
		connMap.Store(id, conn)

		// start a new goroutine for each client connection
		go handleUserConnection(id, conn, connMap)
	}
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

	broadcast(connMap, clientConn, nil)

	for {
		data := make([]byte, 1024)
		n, err := clientConn.Read(data)
		if err != nil {
			log.Printf("error reading from client %v", err)
			return
		}
		messageBody := data[4:n]

		protoMessage := &chat.Message{}
		common.UnmarshalMessage(messageBody, protoMessage)

		// Don't write empty space to clients
		if protoMessage.Text == "\n" {
			continue
		}

		broadcast(connMap, clientConn, protoMessage)
	}
}

// Broadcast a message to all connected clients
func broadcast(connMap *sync.Map, clientConn net.Conn, protoMessage *chat.Message) {

	var clientAddr string
	if protoMessage != nil {
		clientAddr = protoMessage.FromIp
		log.Printf("Client %v sent a message of size %v\n", protoMessage.FromIp, len(protoMessage.Text))
	}

	// Fan-out write
	connMap.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {

			// skip client that sent the message
			if conn.RemoteAddr().String() == clientAddr {
				// Within sync.Map.Range() acts as a continue
				return true
			}

			message := "\n"
			if protoMessage == nil {
				message = clientAddr + " joined the chat"
			} else {
				message = protoMessage.Text
			}

			sizeHeader, messageBody := marshalMessage(message, clientAddr)
			data := append(sizeHeader[:], messageBody...)

			if _, err := conn.Write(data); err != nil {
				log.Printf("error on writing to connection %v\n", err)
			}
		}
		return true
	})
}
