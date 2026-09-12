// A command in on a queue, a fact out on a stream. This process is both a RabbitMQ consumer
// and a Kafka producer, which is an extremely common shape and the reason exercise 3 exists.
//
//	go run ./cmd/receiver
//	DUAL_WRITE_WINDOW=15 go run ./cmd/receiver     # for Probe A
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/iancooper/Practical-Messaging-Go/streams/model"
	"github.com/iancooper/Practical-Messaging-Go/streams/simpleeventing"
	"github.com/iancooper/Practical-Messaging-Go/streams/simplemessaging"
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

	pump := simplemessaging.NewMessagePump[model.PlaceOrder](
		model.NewPlaceOrderMapper(),
		model.NewPlaceOrderHandler(model.NewCatalogue(), publisher),
		simplemessaging.HostName)

	if err := pump.Run(ctx); err != nil {
		log.Fatalln("pump stopped:", err)
	}

	fmt.Println("Receiver stopped.")
}
