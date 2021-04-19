package main

import (
	p2p "github.com/iancooper/Practical-Messaging-Go/publish-subscribe/pubsub"
	"log"
)

func main() {
	channel := p2p.NewChannel("hello-pubsub")
	defer channel.Close()

	ok, message := channel.Receive()
	if ok {
		log.Println("Received Message", message)
	}
}
