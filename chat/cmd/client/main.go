package main

import (
	"encoding/binary"
	"log"
	"net"
	"os"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"google.golang.org/protobuf/proto"
	"rustedskyline.io/tcpchat/internal/common"
	chat "rustedskyline.io/tcpchat/internal/proto"
)

const InputPrompt = "> "

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	chatPane := initChatPane()
	chatInput := initInputPane()

	ui.Render(chatPane)
	ui.Render(chatInput)

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.8,
			ui.NewCol(1.0, chatPane),
		),
		ui.NewRow(0.2,
			ui.NewCol(1.0, chatInput),
		),
	)

	ui.Render(grid)

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
	go func() {
		for e := range ui.PollEvents() {
			if e.Type == ui.KeyboardEvent {
				if e.ID == "<Enter>" {
					sendMessage(conn, chatInput.Text)
					chatPane.Rows = append(chatPane.Rows, chatInput.Text)
					chatInput.Text = InputPrompt
				} else if e.ID == "<C-c>" {
					ui.Close()
					wg.Done()
					os.Exit(0)
				} else if e.ID == "<Space>" {
					chatInput.Text += " "
				} else if e.ID == "<Backspace>" {
					chatInput.Text = chatInput.Text[:len(chatInput.Text)-1]
				} else {
					chatInput.Text += e.ID
				}
				ui.Render(chatPane)
				ui.Render(chatInput)
			}
		}
	}()

	go readServer(conn, chatPane)
	wg.Wait()
}

func sendMessage(conn net.Conn, input string) {
	sizeHeader, messageBody := marshalMessage(input, conn.LocalAddr().String())
	data := append(sizeHeader[:], messageBody...)

	if _, err := conn.Write(data); err != nil {
		log.Printf("error on writing to connection %v\n", err)
	}
}

// Read from connection and print contents indefinitely
func readServer(conn net.Conn, chatPane *widgets.List) {
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

		chatPane.Rows = append(chatPane.Rows, "["+protoMessage.FromIp+"] "+protoMessage.Text)
		ui.Render(chatPane)
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

func initInputPane() *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Text = InputPrompt
	return p
}

func initChatPane() *widgets.List {
	l := widgets.NewList()
	l.Title = "Chat"
	l.Rows = []string{}
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	return l
}
