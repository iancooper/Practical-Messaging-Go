package model

import "fmt"

// PlaceOrderHandler is application code, and this is what it should look like: a domain type
// in, a nil or an error out. No delivery, no headers, no acknowledgement, no broker.
//
// A test can call this. So can an HTTP endpoint. That is a consequence of the separation
// rather than the reason for it, but it is a good smoke alarm: if you cannot call your
// handler from a test without a broker running, the mapper has not finished its job.
type PlaceOrderHandler struct {
	catalogue *Catalogue
}

func NewPlaceOrderHandler(catalogue *Catalogue) *PlaceOrderHandler {
	return &PlaceOrderHandler{catalogue: catalogue}
}

func (h *PlaceOrderHandler) Handle(order PlaceOrder) error {
	price, err := h.catalogue.PriceOf(order.Sku)
	if err != nil {
		return err
	}
	total := price * float64(order.Quantity)

	fmt.Printf("  placed order %s: %d x %s for %.2f (customer %s)\n",
		order.Id, order.Quantity, order.Sku, total, order.CustomerId)
	return nil
}
