package main

import (
	p2p "github.com/iancooper/Practical-Messaging-Go/publish-subscribe/pubsub"
	"log"
)

func main() {
	channel := p2p.NewChannel("hello-pubsub")
	defer channel.Close()

	channel.Send("Hello World")
	log.Println("Sent Message")
}
