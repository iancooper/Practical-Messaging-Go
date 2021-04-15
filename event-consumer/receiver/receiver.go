package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/event-consumer/datachannel"
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
	consumer := dc.NewConsumer("invalid-p2p",
		func(bytes []byte) (interface{}, error) {
			var greetings greeting
			err := json.Unmarshal(bytes, &greetings)
			return greetings, err
		},
		func(message interface{}) {
			greeting := message.(greeting)
			greeting.greet()
		},
	)
	defer consumer.Close()

	consumer.Receive()

}
