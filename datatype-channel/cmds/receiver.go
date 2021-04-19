package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/datatype-channel/datachannel"
)

func main() {
	consumer := dc.NewConsumer("data-p2p", func(bytes []byte) (interface{}, error) {
		//TODO: deserialize from message
	})
	defer consumer.Close()

	ok, message := consumer.Receive()
	if ok {
		greeting := message.(greeting)
		greeting.greet()
	}
}
