package main

import (
	"encoding/json"
	"github.com/iancooper/Practical-Messaging-Go/pipes-and-filters/datachannel"
)

func main() {
	filter := datachannel.NewFilter(func(bytes []byte) (interface{}, error) {
		var greetings datachannel.Greeting
		err := json.Unmarshal(bytes, &greetings)
		return greetings, err
	},
		func(message interface{}) ([]byte, error) {
			return json.Marshal(message.(datachannel.EnhancedGreeting))
		})

	filter.Run(
		func(msg interface{}) interface{} {
			greeting := msg.(datachannel.Greeting)
			return datachannel.EnhancedGreeting{Message: greeting.Message, Salutation: "Clarissa Harlowe"}
		})
}
