// Follows the OrderPlaced stream and counts what it sees.
//
//	go run ./cmd/streamconsumer
//
// The count is the point. An order placed once should appear once.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/iancooper/Practical-Messaging-Go/lookup/model"
	"github.com/iancooper/Practical-Messaging-Go/lookup/simpleeventing"
)

func main() {
	fmt.Println("StreamConsumer starting. PID", os.Getpid())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		<-interrupts
		cancel()
	}()

	// How many times have we seen an event for each order? Anything above one is a
	// duplicate, and duplicates are what exercise 3 is about.
	seen := map[string]int{}

	consumer, err := simpleeventing.NewEventStreamConsumer[model.OrderPlaced](
		model.DeserializeOrderPlaced,
		func(event model.OrderPlaced) error {
			seen[event.OrderId]++
			flag := ""
			if count := seen[event.OrderId]; count > 1 {
				flag = fmt.Sprintf("  <-- DUPLICATE, seen %d times", count)
			}
			fmt.Printf("  order %s: %d x %s for %.2f%s\n",
				event.OrderId, event.Quantity, event.Sku, event.Total, flag)
			return nil
		},
		simpleeventing.ConsumerGroup, simpleeventing.BootstrapServers)
	if err != nil {
		log.Fatalln("could not connect to Kafka:", err)
	}

	if err := consumer.Run(ctx); err != nil {
		log.Fatalln("stream consumer stopped:", err)
	}

	events := 0
	for _, count := range seen {
		events += count
	}
	fmt.Println()
	fmt.Printf("Distinct orders seen: %d. Events read: %d.\n", len(seen), events)
	for orderId, count := range seen {
		if count > 1 {
			fmt.Printf("  %s arrived %d times\n", orderId, count)
		}
	}
}
