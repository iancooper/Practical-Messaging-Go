// Puts things on channels, including things the other end will not like.
//
//	go run ./cmd/sender                 a good order, on the queue
//	go run ./cmd/sender burst 20        twenty good orders
//	go run ./cmd/sender poison          a SKU that is not in the catalogue (queue-side failure)
//	go run ./cmd/sender unmappable      a body that is not a PlaceOrder (queue-side failure)
//	go run ./cmd/sender flaky           a lookup that fails twice then works
//	go run ./cmd/sender bad-event       A RECORD THE STREAM CONSUMER CANNOT READ -- straight onto Kafka
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iancooper/Practical-Messaging-Go/streams/model"
	"github.com/iancooper/Practical-Messaging-Go/streams/simpleeventing"
	"github.com/iancooper/Practical-Messaging-Go/streams/simplemessaging"
)

func main() {
	command := "good"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if command == "bad-event" {
		// Appended to the stream directly, because we need a poison *record* rather than a
		// poison message. There is no such thing as putting it on an invalid record topic
		// for us.
		stream, err := simpleeventing.NewEventStreamProducer[model.OrderPlaced](
			model.SerializeOrderPlaced,
			func(event model.OrderPlaced) string { return event.OrderId },
			simpleeventing.BootstrapServers)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not connect to Kafka:", err)
			os.Exit(1)
		}
		defer stream.Close()

		record := `{"Id":"c0ffee","OrderRef":"not-a-field","Sku":"WIDGET-1"}`
		fmt.Println("Appending a record the consumer cannot map:")
		fmt.Println(" ", record)
		if err := stream.SendRaw("poison", record); err != nil {
			fmt.Fprintln(os.Stderr, "could not send:", err)
			os.Exit(1)
		}
		return
	}

	producer, err := simplemessaging.NewDataTypeChannelProducer[model.PlaceOrder](
		model.SerializePlaceOrder, simplemessaging.HostName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not connect to RabbitMQ:", err)
		os.Exit(1)
	}
	defer producer.Close()

	switch command {
	case "good":
		publish(producer, model.PlaceOrderFor("WIDGET-1"))

	case "flaky":
		// The order is fine. The catalogue is having a bad minute and will recover.
		publish(producer, model.PlaceOrderFor("FLAKY-1"))

	case "poison":
		// Well-formed, maps perfectly, and the handler will fail on it every single time.
		publish(producer, model.PlaceOrderFor("NOPE-404"))

	case "unmappable":
		// Valid JSON, wrong shape. No number of retries will make this a PlaceOrder.
		body := `{"Id":"c0ffee","ProductCode":"WIDGET-1","Qty":1}`
		if err := producer.SendRaw(body); err != nil {
			fmt.Fprintln(os.Stderr, "could not send:", err)
			os.Exit(1)
		}
		fmt.Println("Sent an unmappable body:", body)

	case "slow":
		publish(producer, model.PlaceOrderFor("GIZMO-SLOW"))

	case "burst":
		count := 20
		if len(os.Args) > 2 {
			if n, err := strconv.Atoi(os.Args[2]); err == nil {
				count = n
			}
		}
		for i := 0; i < count; i++ {
			publish(producer, model.PlaceOrderForQuantity("WIDGET-1", i+1, "CUST-001"))
		}
		fmt.Println("Sent", count, "orders")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Try: good, flaky, poison, unmappable, slow, burst, bad-event\n", command)
		os.Exit(1)
	}
}

func publish(producer *simplemessaging.DataTypeChannelProducer[model.PlaceOrder], order model.PlaceOrder) {
	if err := producer.Send(order); err != nil {
		fmt.Fprintln(os.Stderr, "could not send:", err)
		os.Exit(1)
	}
	fmt.Printf("Sent order %s: %d x %s\n", order.Id, order.Quantity, order.Sku)
}
