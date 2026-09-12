// The consumer. Runs a Message Pump until you stop it.
//
//	go run ./cmd/receiver
//
// Ctrl-C stops it *between* messages, which is the polite ending.
//
// Several probes want the rude ending instead -- a process that dies with a message in its
// hands. Use the PID printed below, from another terminal:
//
//	kill -9 <pid>
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/iancooper/Practical-Messaging-Go/message-pump/model"
	"github.com/iancooper/Practical-Messaging-Go/message-pump/simplemessaging"
)

func main() {
	// Go's standard output is unbuffered, so every line is on the terminal as it is
	// printed. Several probes end with this process being killed, and a buffered line you
	// never see is a probe you cannot read.
	pid := os.Getpid()
	fmt.Println("Receiver starting. PID", pid)
	fmt.Printf("Ctrl-C to stop between messages; 'kill -9 %d' to stop mid-message.\n", pid)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl-C asks the pump to stop, and the pump notices between messages.
	//
	// Note that this is a goroutine and not a defer. A deferred "stopping politely" would
	// print on *every* exit, including a panic on the way out of the pump -- so it would
	// appear directly underneath the stack trace that had just killed the process, which is
	// a lie told at exactly the moment you are reading most carefully.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		<-interrupts
		fmt.Println("\nStopping after the current message...")
		cancel()
	}()

	pump := simplemessaging.NewMessagePump[model.PlaceOrder](
		model.NewPlaceOrderHandler(model.NewCatalogue()), simplemessaging.HostName)

	if err := pump.Run(ctx); err != nil {
		log.Fatalln("pump stopped:", err)
	}

	fmt.Println("Receiver stopped.")
}
