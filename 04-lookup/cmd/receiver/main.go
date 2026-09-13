// A command in on a queue, a fact out on a stream -- and now the price comes from a local
// copy of somebody else's data rather than from a map.
//
//	go run ./cmd/receiver
//	DUAL_WRITE_WINDOW=10 go run ./cmd/receiver     # exercise 3's Probe A, still here
//
// This is the only file that knows all three things at once: that prices live in SQLite,
// that the domain wants a model.IPriceStore, and that the two fit together. Composition is
// the application's job. model declares the interface and names none of the rest of it.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/iancooper/Practical-Messaging-Go/lookup/localcopy"
	"github.com/iancooper/Practical-Messaging-Go/lookup/model"
	"github.com/iancooper/Practical-Messaging-Go/lookup/simpleeventing"
	"github.com/iancooper/Practical-Messaging-Go/lookup/simplemessaging"
)

func main() {
	pid := os.Getpid()
	fmt.Println("Receiver starting. PID", pid)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		<-interrupts
		fmt.Println("\nStopping after the current message...")
		cancel()
	}()

	producer, err := simpleeventing.NewEventStreamProducer[model.OrderPlaced](
		model.SerializeOrderPlaced,
		func(event model.OrderPlaced) string { return event.OrderId },
		simpleeventing.BootstrapServers)
	if err != nil {
		log.Fatalln("could not connect to Kafka:", err)
	}

	publisher := simpleeventing.NewKafkaEventPublisher(producer)
	defer publisher.Close()

	prices, err := localcopy.Open(localcopy.DefaultPath)
	if err != nil {
		log.Fatalln("could not open the local copy:", err)
	}
	defer prices.Close()

	// Say how old the copy is, once, at startup. It is the only line in the system that
	// knows -- and watch what Probe C makes of that. Knowing at startup is not the same as
	// noticing, and a receiver that prices ten thousand orders from a four-day-old copy
	// will say this once.
	count, err := prices.Count()
	if err != nil {
		log.Fatalln("could not read the local copy:", err)
	}
	newest, any, err := prices.NewestChangedAt()
	if err != nil {
		log.Fatalln("could not read the local copy:", err)
	}
	how := "empty"
	if any {
		how = fmt.Sprintf("newest change %s old", age(newest))
	}
	fmt.Printf("Local copy is %s, holding %d prices -- %s.\n", localcopy.DefaultPath, count, how)

	pump := simplemessaging.NewMessagePump[model.PlaceOrder](
		model.NewPlaceOrderMapper(),
		// PlaceOrderHandler has not changed since exercise 3, and nothing in this line
		// asks it to. The Catalogue takes an interface; this file chose what implements it.
		model.NewPlaceOrderHandler(model.NewCatalogue(prices), publisher),
		simplemessaging.HostName)

	if err := pump.Run(ctx); err != nil {
		log.Fatalln("pump stopped:", err)
	}

	fmt.Println("Receiver stopped.")
}

// age says how old, in words, without saying "1 minutes". A copy's age is the one number
// that tells a current local copy from a stale one, so it is worth printing in a shape a
// human reads.
func age(when time.Time) string {
	d := time.Since(when)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%.0fs", d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%.0fm", d.Minutes())
	case d < 24*time.Hour:
		return fmt.Sprintf("%.0fh", d.Hours())
	default:
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
}
