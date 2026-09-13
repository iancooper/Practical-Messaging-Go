// Puts things on channels, including things the other end will not like.
//
//	go run ./cmd/sender                 a good order, on the queue
//	go run ./cmd/sender burst 20        twenty good orders
//	go run ./cmd/sender poison          a SKU that is not in the catalogue (queue-side failure)
//	go run ./cmd/sender unmappable      a body that is not a PlaceOrder (queue-side failure)
//	go run ./cmd/sender bad-event       A RECORD THE STREAM CONSUMER CANNOT READ -- straight onto Kafka
//
// 'flaky' and 'slow' are gone. They sent GIZMO-SLOW and FLAKY-1, which were failures of a
// lookup that was called on demand -- and there is no call any more. Removing the thing that
// could fail is not the same as fixing it, and Probe B is the bill.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iancooper/Practical-Messaging-Go/lookup/model"
	"github.com/iancooper/Practical-Messaging-Go/lookup/simpleeventing"
	"github.com/iancooper/Practical-Messaging-Go/lookup/simplemessaging"
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

	case "poison":
		// Well-formed, maps perfectly, and no price was ever published for it -- so the
		// handler will fail on it every single time.
		publish(producer, model.PlaceOrderFor("NOPE-404"))

	case "unmappable":
		// Valid JSON, wrong shape. No number of retries will make this a PlaceOrder.
		body := `{"Id":"c0ffee","ProductCode":"WIDGET-1","Qty":1}`
		if err := producer.SendRaw(body); err != nil {
			fmt.Fprintln(os.Stderr, "could not send:", err)
			os.Exit(1)
		}
		fmt.Println("Sent an unmappable body:", body)

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
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Try: good, poison, unmappable, burst, bad-event\n", command)
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
