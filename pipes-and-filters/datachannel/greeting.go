package datachannel

import (
	"log"
	"time"
)

type Greeting struct {
	Message string
}

func (g Greeting) Greet(log log.Logger) {
	log.Println(g.Message)
}

type EnhancedGreeting struct {
	Message    string
	Salutation string
}

func (g *EnhancedGreeting) Greet() {
	log.Println("Received message: To - ", g.Salutation, "Message: ", g.Message, "at ", time.Now())
}
