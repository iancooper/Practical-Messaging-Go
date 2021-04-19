package greeting

import (
	"log"
	"time"
)

//We don't want to share code between sender and receiver, but it's convenient for this exercise to do so

type Greeting struct {
	Message string
}

func (g Greeting) Greet() {
	log.Println("Received message: ", g.Message, "at", time.Now())
}
