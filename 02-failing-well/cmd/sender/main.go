// The producer, and the source of every failure you are asked to survive.
//
//	go run ./cmd/sender                 a good order
//	go run ./cmd/sender flaky           TRANSIENT: the lookup fails twice, then works
//	go run ./cmd/sender poison          PERMANENT: a SKU that is not in the catalogue, ever
//	go run ./cmd/sender unmappable      INVALID:   a body that is not a PlaceOrder at all
//	go run ./cmd/sender slow            an order whose lookup takes 30 seconds
//	go run ./cmd/sender burst 20        twenty good orders
//
// Three of those are three different failures. They should not all end up in the same place.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iancooper/Practical-Messaging-Go/failing-well/model"
	"github.com/iancooper/Practical-Messaging-Go/failing-well/simplemessaging"
)

func main() {
	command := "good"
	if len(os.Args) > 1 {
		command = os.Args[1]
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
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Try: good, flaky, poison, unmappable, slow, burst\n", command)
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
