package common

import (
	"log"

	"google.golang.org/protobuf/proto"
	chat "rustedskyline.io/tcpchat/internal/proto"
)

const MessageDelim = '\r'

func UnmarshalMessage(messageBody []byte, protoMessage *chat.Message) {
	err := proto.Unmarshal(messageBody, protoMessage)
	if err != nil {
		log.Printf("error unmarshalling proto message %v", err)
	}
}
