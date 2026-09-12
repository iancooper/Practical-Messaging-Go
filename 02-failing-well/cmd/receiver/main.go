// The consumer. Runs a Message Pump until you stop it.
//
//	go run ./cmd/receiver
//
// Ctrl-C stops it between messages. 'kill -9 <pid>' stops it mid-message.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/iancooper/Practical-Messaging-Go/failing-well/model"
	"github.com/iancooper/Practical-Messaging-Go/failing-well/simplemessaging"
)

func main() {
	pid := os.Getpid()
	fmt.Println("Receiver starting. PID", pid)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A goroutine and not a defer: a deferred "stopping politely" would print on every exit,
	// including a panic, and so would appear underneath the stack trace that killed it.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		<-interrupts
		fmt.Println("\nStopping after the current message...")
		cancel()
	}()

	pump := simplemessaging.NewMessagePump[model.PlaceOrder](
		model.NewPlaceOrderMapper(),
		model.NewPlaceOrderHandler(model.NewCatalogue()),
		simplemessaging.HostName)

	if err := pump.Run(ctx); err != nil {
		log.Fatalln("pump stopped:", err)
	}

	fmt.Println("Receiver stopped.")
}
