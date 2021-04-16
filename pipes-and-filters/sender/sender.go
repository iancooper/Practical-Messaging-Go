package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/pipes-and-filters/datachannel"
	"log"
)

type greeting struct {
	Message string
}

func (g greeting) greet(log log.Logger) {
	log.Println(g.Message)
}

func main() {
	producer := dc.NewProducer("source-p2p", func(message interface{}) ([]byte, error) {
		return json.Marshal(message.(greeting))
	})
	defer producer.Close()

	producer.Send(greeting{
		Message: "Hello World",
	})
	log.Println("Sent Message")
}
