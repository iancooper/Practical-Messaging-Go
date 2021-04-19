package main

import (
	"log"
	"time"
)

type greeting struct {
	Message string
}

func (g greeting) greet() {
	log.Println("Received message: ", g.Message, "at", time.Now())
}

