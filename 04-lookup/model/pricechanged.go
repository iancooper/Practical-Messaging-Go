package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrNotAPriceChanged means the record is not a PriceChanged and never will be.
var ErrNotAPriceChanged = errors.New("record is not a PriceChanged")

// PriceChanged is the ECST event: the catalogue service, which owns SKUs and prices, says
// what one costs now.
//
// It is a SNAPSHOT, NOT A DELTA, and that is the decision the whole exercise turns on.
// "WIDGET-1 is now 11.99" can be applied twice with no harm; "WIDGET-1 went up by 2.00"
// cannot. Probe D is where that stops being a matter of taste -- see README.md, step 1.
//
// ChangedAt is here for two reasons and both are probes. Probe A subtracts it from the
// moment the consumer applies the record, and that difference is your staleness. Probe C
// asks how old your local copy is, and a copy that does not carry a date cannot answer.
//
// The field names on the wire are capitalised, and the same in every language this course
// ships in, so a C# seeder and a Go price consumer understand each other.
type PriceChanged struct {
	Id        string    `json:"Id"`
	Sku       string    `json:"Sku"`
	Price     float64   `json:"Price"`
	ChangedAt time.Time `json:"ChangedAt"`
}

// ID satisfies simplemessaging.IAmAMessage without the domain having to import it: Go's
// interfaces are structural, so nothing here names the gateway.
func (e PriceChanged) ID() string { return e.Id }

// PriceChangedFor is what the catalogue service would publish: a SKU, its new price, and
// the moment the change became true. The clock is the publisher's, which is worth knowing
// when you read Probe A's number.
func PriceChangedFor(sku string, price float64) PriceChanged {
	return PriceChanged{
		Id:        newId(),
		Sku:       sku,
		Price:     price,
		ChangedAt: time.Now().UTC(),
	}
}

func SerializePriceChanged(event PriceChanged) (string, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// DeserializePriceChanged turns a record's body into a PriceChanged, or returns an error
// wrapping ErrNotAPriceChanged.
//
// DisallowUnknownFields for the same reason as everywhere else in this course: a body we
// have not understood should fail here rather than turn into a zero-valued event that
// quietly sets a price to nothing.
func DeserializePriceChanged(body string) (PriceChanged, error) {
	decoder := json.NewDecoder(bytes.NewReader([]byte(body)))
	decoder.DisallowUnknownFields()

	var event PriceChanged
	if err := decoder.Decode(&event); err != nil {
		return PriceChanged{}, fmt.Errorf("%w: %v", ErrNotAPriceChanged, err)
	}
	if event.Id == "" || event.Sku == "" {
		return PriceChanged{}, fmt.Errorf("%w: Id and Sku are required", ErrNotAPriceChanged)
	}
	return event, nil
}
