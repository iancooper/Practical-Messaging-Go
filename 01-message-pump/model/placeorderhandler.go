package model

import (
	"fmt"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

// PlaceOrderHandler is application code. What the business actually wanted: price the order
// and place it.
//
// ---------------------------------------------------------------------------------------
//
//	THIS HANDLER IS PART OF THE EXERCISE. See PROBE.md.
//
//	Ask yourself one question before you read any further: how much of this method is
//	about placing an order?
//
// ---------------------------------------------------------------------------------------
type PlaceOrderHandler struct {
	catalogue *Catalogue
}

func NewPlaceOrderHandler(catalogue *Catalogue) *PlaceOrderHandler {
	return &PlaceOrderHandler{catalogue: catalogue}
}

func (h *PlaceOrderHandler) Handle(delivery amqp091.Delivery) error {
	order, err := DeserializePlaceOrder(string(delivery.Body))
	if err != nil {
		return err
	}

	price, err := h.catalogue.PriceOf(order.Sku)
	if err != nil {
		return err
	}
	total := price * float64(order.Quantity)

	fmt.Printf("  placed order %s: %d x %s for %.2f (customer %s)\n",
		order.Id, order.Quantity, order.Sku, total, order.CustomerId)
	return nil
}
