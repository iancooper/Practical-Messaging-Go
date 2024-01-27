package main

import (
	"encoding/json"
	"log"
	. "github.com/iancooper/Practical-Messaging-Go/datatype-channel/cmd"
	dc "github.com/iancooper/Practical-Messaging-Go/datatype-channel/datachannel"
)

func main() {
	producer := dc.NewProducer("data-p2p", func(message interface{}) ([]byte, error) {
		// TODO: serialize a greeting
	})
	defer producer.Close()

	producer.Send(Greeting{
		Message: "Hello World",
	})
	log.Println("Sent Message")
}
