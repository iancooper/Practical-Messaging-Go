package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/datatype-channel/datachannel"
	"log"
	"time"
)

type greeting struct{
	Message string
}

func (g greeting) greet(){
	log.Println("Received message: ", g.Message, "at", time.Now())
}

func main() {
	consumer := dc.NewConsumer("data-p2p", func(bytes []byte) (interface{}, error) {
		var greetings greeting
		err := json.Unmarshal(bytes, &greetings)
		return greetings, err
	})
	defer consumer.Close()

	ok, message := consumer.Receive()
	if ok {
		greeting := message.(greeting)
		greeting.greet()
	}
}
