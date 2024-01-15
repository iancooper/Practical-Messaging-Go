package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/pub-sub-stream/datachannel"
)

func main() {
	consumer := dc.NewConsumer("sink-p2p",
		func(bytes []byte) (interface{}, error) {
			var bio = dc.Biography{}
			err := json.Unmarshal(bytes, &bio)
			return bio, err
		},
		func(message interface{}) error {
			greeting := message.(dc.Biography)
			//add to MySQL
		},
	)

	consumer.Receive()

}
