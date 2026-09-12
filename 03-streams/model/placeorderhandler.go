package model

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/iancooper/Practical-Messaging-Go/streams/simplemessaging"
)

// dualWriteWindow is how long to pause between publishing the event and returning to the
// pump -- which is to say, between the Kafka write and the RabbitMQ acknowledgement.
//
// In a real service that gap is microseconds wide. It is still a gap, and a service that
// handles a million orders will fall into it. Widening it to fifteen seconds does not create
// the problem; it just means you can aim at it.
//
//	DUAL_WRITE_WINDOW=15 go run ./cmd/receiver
var dualWriteWindow = readWindow()

func readWindow() time.Duration {
	seconds, err := strconv.Atoi(os.Getenv("DUAL_WRITE_WINDOW"))
	if err != nil {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// PlaceOrderHandler is application code: price the order, place it, and tell the world it
// happened.
//
// A command came in on a queue; a fact goes out on a stream. That is an entirely ordinary
// shape and you have probably written it -- which is the point.
//
// ---------------------------------------------------------------------------------------
//
//	Nothing in this type is wrong, and that is what makes exercise 3 worth doing. The
//	handler is clean, the domain has no broker in it, the ordering is the sensible one.
//	Read PROBE.md before you run it, and predict what a crash costs you.
//
// ---------------------------------------------------------------------------------------
type PlaceOrderHandler struct {
	catalogue *Catalogue
	events    simplemessaging.IPublishEvents[OrderPlaced]
}

func NewPlaceOrderHandler(
	catalogue *Catalogue, events simplemessaging.IPublishEvents[OrderPlaced],
) *PlaceOrderHandler {
	return &PlaceOrderHandler{catalogue: catalogue, events: events}
}

func (h *PlaceOrderHandler) Handle(order PlaceOrder) error {
	price, err := h.catalogue.PriceOf(order.Sku)
	if err != nil {
		return err
	}
	total := price * float64(order.Quantity)

	fmt.Printf("  placed order %s: %d x %s for %.2f\n", order.Id, order.Quantity, order.Sku, total)

	if err := h.events.Publish(NewOrderPlaced(order, total)); err != nil {
		return err
	}

	if dualWriteWindow > 0 {
		fmt.Println("  [the event is on the stream. RabbitMQ has NOT been acked yet.]")
		fmt.Printf("  [you have %.0f seconds. kill -9 %d]\n", dualWriteWindow.Seconds(), os.Getpid())
		time.Sleep(dualWriteWindow)
	}

	return nil
}
