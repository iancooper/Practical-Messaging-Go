package main

import (
	"encoding/json"
	"github.com/iancooper/Practical-Messaging-Go/pipes-and-filters/datachannel"
)

type greeting struct {
	Message string
}

type enhancedGreeting struct {
	Message    string
	Salutation string
}

func main() {
	filter := datachannel.NewFilter(func(bytes []byte) (interface{}, error) {
		var greetings greeting
		err := json.Unmarshal(bytes, &greetings)
		return greetings, err
	},
		func(message interface{}) ([]byte, error) {
			return json.Marshal(message.(enhancedGreeting))
		})

	filter.Run(
		func(msg interface{}) interface{} {
			greeting := msg.(greeting)
			return enhancedGreeting{Message: greeting.Message, Salutation: "Clarissa Harlowe"}
		})
}
