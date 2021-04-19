package main

import (
	"log"
	"time"
)

type greeting struct {
	Message string
}

func (g greeting) greet(log log.Logger) {
	log.Println(g.Message)
}

type enhancedGreeting struct {
	Message    string
	Salutation string
}

func (g *enhancedGreeting) greet() {
	log.Println("Received message: To - ", g.Salutation, "Message: ", g.Message, "at ", time.Now())
}
