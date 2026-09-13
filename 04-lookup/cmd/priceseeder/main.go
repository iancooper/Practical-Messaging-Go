// Stands in for the catalogue service: it owns prices, and it tells the world when one
// changes.
//
//	go run ./cmd/priceseeder seed                 a starting price for every SKU
//	go run ./cmd/priceseeder set WIDGET-1 11.99   change one, on demand
//
// It publishes and exits. It keeps no state, because the stream is the state -- which is the
// whole argument for ECST, and the reason a new price consumer with an empty copy can catch
// up by reading the log from the beginning.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iancooper/Practical-Messaging-Go/lookup/model"
	"github.com/iancooper/Practical-Messaging-Go/lookup/simpleeventing"
)

// The two SKUs that survived exercise 4. GIZMO-SLOW and FLAKY-1 are gone: they were failures
// of an on-demand lookup, and there is no longer a lookup to fail. See model/catalogue.go.
var starting = []struct {
	Sku   string
	Price float64
}{
	{"WIDGET-1", 9.99},
	{"GIZMO-2", 24.50},
}

func main() {
	command := "seed"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	producer, err := simpleeventing.NewEventStreamProducer[model.PriceChanged](
		model.SerializePriceChanged,
		// Keyed by SKU, so two changes to one price stay in order. Key it by the event's
		// own id instead and they land on different partitions, and yesterday's price can
		// be applied on top of today's -- which is a bug you will not see until the day it
		// costs money.
		func(event model.PriceChanged) string { return event.Sku },
		simpleeventing.BootstrapServers)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not connect to Kafka:", err)
		os.Exit(1)
	}
	defer producer.Close()

	switch command {
	case "seed":
		for _, price := range starting {
			publish(producer, model.PriceChangedFor(price.Sku, price.Price))
		}
		fmt.Printf("Seeded %d prices.\n", len(starting))

	case "set":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: set <SKU> <PRICE>   e.g. set WIDGET-1 11.99")
			os.Exit(1)
		}
		newPrice, err := strconv.ParseFloat(os.Args[3], 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Usage: set <SKU> <PRICE>   e.g. set WIDGET-1 11.99")
			os.Exit(1)
		}
		publish(producer, model.PriceChangedFor(os.Args[2], newPrice))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Try: seed, set\n", command)
		os.Exit(1)
	}
}

func publish(producer *simpleeventing.EventStreamProducer[model.PriceChanged], event model.PriceChanged) {
	if err := producer.Send(event); err != nil {
		fmt.Fprintln(os.Stderr, "could not send:", err)
		os.Exit(1)
	}
	fmt.Printf("Published %s = %.2f at %s\n",
		event.Sku, event.Price, event.ChangedAt.Format("15:04:05.000"))
}
