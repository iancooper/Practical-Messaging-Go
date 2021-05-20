package main

import (
	"log"

	p2p "github.com/iancooper/Practical-Messaging-Go/point-to-point/p2pchannel"
)

func main() {
	channel := p2p.NewChannel("hello-p2p")
	defer channel.Close()

	channel.Send("Hello World")
	log.Println("Sent Message")
}
