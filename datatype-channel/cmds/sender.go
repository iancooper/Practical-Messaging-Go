package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/datatype-channel/datachannel"
	"log"
)

func main() {
	producer := dc.NewProducer("data-p2p", func(message interface{}) ([]byte, error) {
		// TODO: serialize a greeting
	})
	defer producer.Close()

	producer.Send(greeting{
		Message: "Hello World",
	})
	log.Println("Sent Message")
}
