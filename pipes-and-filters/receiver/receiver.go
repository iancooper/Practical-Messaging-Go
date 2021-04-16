package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/pipes-and-filters/datachannel"
	"log"
	"time"
)

type enhancedGreeting struct {
	Message    string
	Salutation string
}

func (g *enhancedGreeting) greet() {
	log.Println("Received message: To - ", g.Salutation, "Message: ", g.Message, "at ", time.Now())
}

func main() {
	consumer := dc.NewConsumer("sink-p2p",
		func(bytes []byte) (interface{}, error) {
			var greetings enhancedGreeting
			err := json.Unmarshal(bytes, &greetings)
			return greetings, err
		},
		func(message interface{}) {
			greeting := message.(enhancedGreeting)
			greeting.greet()
		},
	)
	defer consumer.Close()

	consumer.Receive()

}
