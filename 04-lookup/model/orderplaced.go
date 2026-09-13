package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrNotAnOrderPlaced means the record is not an OrderPlaced and never will be.
var ErrNotAnOrderPlaced = errors.New("record is not an OrderPlaced")

// OrderPlaced is an Event Message: this happened. Many readers may care, none of them may
// reply, and it is a statement about the past rather than a request.
//
// Contrast PlaceOrder, which was a Command: one recipient, allowed to fail. Same system, ten
// milliseconds apart, and the difference in intent is the whole reason one goes on a queue
// and the other on a stream.
type OrderPlaced struct {
	Id       string    `json:"Id"`
	OrderId  string    `json:"OrderId"`
	Sku      string    `json:"Sku"`
	Quantity int       `json:"Quantity"`
	Total    float64   `json:"Total"`
	PlacedAt time.Time `json:"PlacedAt"`
}

func (e OrderPlaced) ID() string { return e.Id }

func NewOrderPlaced(order PlaceOrder, total float64) OrderPlaced {
	return OrderPlaced{
		Id:       newId(),
		OrderId:  order.Id,
		Sku:      order.Sku,
		Quantity: order.Quantity,
		Total:    total,
		PlacedAt: time.Now().UTC(),
	}
}

func SerializeOrderPlaced(event OrderPlaced) (string, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func DeserializeOrderPlaced(body string) (OrderPlaced, error) {
	decoder := json.NewDecoder(bytes.NewReader([]byte(body)))
	decoder.DisallowUnknownFields()

	var event OrderPlaced
	if err := decoder.Decode(&event); err != nil {
		return OrderPlaced{}, fmt.Errorf("%w: %v", ErrNotAnOrderPlaced, err)
	}
	if event.Id == "" || event.OrderId == "" {
		return OrderPlaced{}, fmt.Errorf("%w: Id and OrderId are required", ErrNotAnOrderPlaced)
	}
	return event, nil
}
