package model

import (
	"errors"
	"fmt"
)

// ErrUnknownSku is permanent. This SKU does not exist and asking again will not change that.
var ErrUnknownSku = errors.New("not in the catalogue")

// ErrLocalCopyEmpty is transient. We have no copy of the catalogue yet -- the price consumer
// has not started, or has not caught up. The SKU may be perfectly good; we are simply not
// ready to price it.
//
// This is a different fact from ErrUnknownSku and the difference is the point of Probe C.
// One of them is about the order and one of them is about us.
var ErrLocalCopyEmpty = errors.New("the local copy has no prices in it yet")

// Catalogue is reference data the handler needs: what does this SKU cost?
//
// Everything about how this answers has changed, and its signature has not. In exercises 1
// to 3 it was a map that pretended to be a service call, and it failed in three ways: slow
// (GIZMO-SLOW), briefly unwell (FLAKY-1) and unknown. Two of those three are gone, and
// their going is the lesson -- they were *on-demand* failures, and there is no longer a call
// to be slow or unwell. Get It In Advance did not fix them. It removed the thing that could
// fail, and bought you Probe B instead.
//
// The handler did not change. It still asks the catalogue for a price, and the catalogue
// still decides where prices come from. That is what the seam was for.
type Catalogue struct {
	prices IPriceStore
}

// NewCatalogue takes the store as an interface the domain declares. Whatever is passed in
// here knows about SQLite; the Catalogue does not, and cannot.
func NewCatalogue(prices IPriceStore) *Catalogue {
	return &Catalogue{prices: prices}
}

func (c *Catalogue) PriceOf(sku string) (float64, error) {
	price, found, err := c.prices.Lookup(sku)
	if err != nil {
		return 0, err
	}
	if found {
		return price.Amount, nil
	}

	// Two different failures wear the same shape -- a lookup that returned nothing -- and
	// exercise 2 spent forty minutes on why that matters. "I have no copy yet" is about us
	// and will fix itself; "that SKU is not a thing" is about the order and never will.
	//
	// The domain's job is to say which. What to do about each is the pump's policy and not
	// ours, and Probe C is about the fact that the pump currently does the same thing with
	// both: it asks errors.Is(err, ErrUnmappableMessage), gets false for each of these, and
	// retries them three times and dead-letters them. Two sentinels, one answer.
	count, err := c.prices.Count()
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, fmt.Errorf("cannot price '%s': %w", sku, ErrLocalCopyEmpty)
	}

	return 0, fmt.Errorf("'%s' is %w", sku, ErrUnknownSku)
}
