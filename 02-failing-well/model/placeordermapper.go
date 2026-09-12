package model

import (
	"fmt"

	"github.com/iancooper/Practical-Messaging-Go/failing-well/simplemessaging"
)

// PlaceOrderMapper is the Translate stage for this channel: a body becomes a PlaceOrder, or
// it does not and we say so clearly.
//
// Note what it does *not* do: it does not log, it does not decide anything, and it does not
// know a broker exists. It converts, or it fails.
type PlaceOrderMapper struct{}

func NewPlaceOrderMapper() *PlaceOrderMapper {
	return &PlaceOrderMapper{}
}

func (m *PlaceOrderMapper) MapToRequest(body string) (PlaceOrder, error) {
	order, err := DeserializePlaceOrder(body)
	if err != nil {
		// Translate the decoder's complaint into the gateway's vocabulary. The pump should
		// not have to know we chose JSON.
		//
		// %w on the gateway's sentinel is what makes errors.Is work in the pump; the rest of
		// the string is for the human reading the log.
		return PlaceOrder{}, fmt.Errorf("%w: %v", simplemessaging.ErrUnmappableMessage, err)
	}
	return order, nil
}
