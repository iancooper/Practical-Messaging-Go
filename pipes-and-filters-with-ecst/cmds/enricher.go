package main

import (
	"database/sql"
	"encoding/json"
	"github.com/iancooper/Practical-Messaging-Go/pipes-and-filters/biography"
	"github.com/iancooper/Practical-Messaging-Go/pipes-and-filters/datachannel"
	"log"
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

			db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/Lookup")
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			name := "Clarissa Harlow"
			bio, err := biography.GetBiography(db, name)
			return datachannel.EnhancedGreeting{Message: greeting.Message, Salutation: name, Bio: bio}
		})
}
