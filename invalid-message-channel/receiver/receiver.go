package main

import (
	"errors"
	dc "github.com/iancooper/Practical-Messaging-Go/datatype-channel/datachannel"
	"log"
	"time"
)

type greeting struct {
	Message string
}

func (g greeting) greet() {
	log.Println("Received message: ", g.Message, "at", time.Now())
}

func main() {
	consumer := dc.NewConsumer("invalid-p2p", func(bytes []byte) (interface{}, error) {
		//
		return nil, errors.New("Failed to deserialize message")
	})
	defer consumer.Close()

	ok, message := consumer.Receive()
	if ok {
		greeting := message.(greeting)
		greeting.greet()
	}
}
