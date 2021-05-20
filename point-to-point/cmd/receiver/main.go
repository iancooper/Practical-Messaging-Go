package main

import (
	"log"

	p2p "github.com/iancooper/Practical-Messaging-Go/point-to-point/p2pchannel"
)

func main() {
	channel := p2p.NewChannel("hello-p2p")
	defer channel.Close()

	ok, message := channel.Receive()
	if ok {
		log.Printf("Received Message: %v", message)
	}
}
