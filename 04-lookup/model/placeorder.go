package model

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrNotAPlaceOrder means the body is not a PlaceOrder and no amount of retrying will
// change that. Deserialize wraps it, so errors.Is(err, ErrNotAPlaceOrder) is the test.
var ErrNotAPlaceOrder = errors.New("body is not a PlaceOrder")

// PlaceOrder is a Command Message: go and do this. One recipient, and it is allowed to fail.
//
// Every member is required and unmapped members are disallowed, so a body that is not
// exactly this shape will not deserialize. That is deliberate -- you need a message the
// receiver cannot understand, and "nearly the right JSON" is the realistic version of one.
//
// The field names on the wire are capitalised, and the same in every language this course
// ships in, so a Python sender and a Go receiver understand each other.
type PlaceOrder struct {
	Id         string `json:"Id"`
	Sku        string `json:"Sku"`
	Quantity   int    `json:"Quantity"`
	CustomerId string `json:"CustomerId"`
}

// ID satisfies simplemessaging.IAmAMessage without the domain having to import it: Go's
// interfaces are structural, so nothing here names the gateway.
func (o PlaceOrder) ID() string { return o.Id }

func SerializePlaceOrder(order PlaceOrder) (string, error) {
	body, err := json.Marshal(order)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// DeserializePlaceOrder turns a body into a PlaceOrder, or returns an error wrapping
// ErrNotAPlaceOrder.
//
// DisallowUnknownFields is the strictness that makes an unmappable body possible. Leaving
// it off is how most services quietly accept a message they have not understood -- and
// encoding/json leaves it off by default.
func DeserializePlaceOrder(body string) (PlaceOrder, error) {
	decoder := json.NewDecoder(bytes.NewReader([]byte(body)))
	decoder.DisallowUnknownFields()

	var order PlaceOrder
	if err := decoder.Decode(&order); err != nil {
		return PlaceOrder{}, fmt.Errorf("%w: %v", ErrNotAPlaceOrder, err)
	}

	// encoding/json will happily leave a missing field at its zero value, so "required" is
	// enforced here. Without this, a body with no Sku at all becomes a PlaceOrder whose Sku
	// is "" and the failure moves from the mapper into the handler -- the wrong place for it.
	if order.Id == "" || order.Sku == "" || order.CustomerId == "" || order.Quantity <= 0 {
		return PlaceOrder{}, fmt.Errorf("%w: Id, Sku, Quantity and CustomerId are all required", ErrNotAPlaceOrder)
	}

	return order, nil
}

// newId is a random identifier in the shape of a UUID. Go has no UUID in its standard
// library and the domain has no business taking a dependency on one for this.
func newId() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("no randomness available: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func PlaceOrderFor(sku string) PlaceOrder {
	return PlaceOrderForQuantity(sku, 1, "CUST-001")
}

func PlaceOrderForQuantity(sku string, quantity int, customerId string) PlaceOrder {
	return PlaceOrder{Id: newId(), Sku: sku, Quantity: quantity, CustomerId: customerId}
}
