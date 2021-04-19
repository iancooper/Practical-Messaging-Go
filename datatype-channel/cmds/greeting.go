package main

import (
	"log"
	"time"
)

//We don't want to share code between sender and receiver, but its convenient for this exercise to do so

type greeting struct{
	Message string
}

func (g greeting) greet(){
	log.Println("Received message: ", g.Message, "at", time.Now())
}

