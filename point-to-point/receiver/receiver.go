package main

import (
	p2p "github.com/iancooper/Practical-Messaging-Go/point-to-point/p2pchannel"
	"log"
)

func main() {
	channel := p2p.NewChannel("hello-p2p")
	defer channel.Close()

	ok, message := channel.Receive()
	if ok {
		log.Println("Received Message", message)
	}

	forever := make(chan bool)
	log.Printf("To exit press CTRL+C")
	<-forever

}
