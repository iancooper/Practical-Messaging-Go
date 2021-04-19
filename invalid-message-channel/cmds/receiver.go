package main

import (
	"errors"
	dc "github.com/iancooper/Practical-Messaging-Go/datatype-channel/datachannel"
)

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
