package main

import (
	"encoding/json"
	dc "github.com/iancooper/Practical-Messaging-Go/pub-sub-stream/datachannel"
	"log"
)

func main() {
	producer := dc.NewProducer("Pub-Sub-Stream-Biography", func(message dc.Record) (string, error) {
		b, err := json.Marshal(message)
		return string(b), err
	})

	defer producer.Close()

	biographies := []dc.Biography{
		{"Clarissa Harlow", "A young woman whose quest for virtue is continually thwarted by her family"},
		{"Pamela Andrews", "A young woman whose virtue is rewarded."},
		{"Harriet Byron", "An orphan, and heir to a considerable fortune of fifteen thousand pounds."},
		{"Charles Grandison", "A man of feeling who truly cannot be said to feel"},
	}

	for _, biography := range biographies {
		producer.Send(biography)
		log.Println("Sent Message " + biography.Id + " " + biography.Description)
	}
}
