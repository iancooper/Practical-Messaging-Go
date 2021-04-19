package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/event-consumer/datachannel"
	"log"
)

func main() {
	producer := dc.NewProducer("invalid-p2p", func(message interface{}) ([]byte, error) {
		return json.Marshal(message.(greeting))
	})
	defer producer.Close()

	producer.Send(greeting{
		Message: "Hello World",
	})
	log.Println("Sent Message")
}
