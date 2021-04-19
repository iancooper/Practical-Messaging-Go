package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/datatype-channel/datachannel"
)

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
