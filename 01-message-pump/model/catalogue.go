package model

import (
	"errors"
	"fmt"
	"time"
)

// ErrUnknownSku is permanent. This SKU does not exist and asking again will not change that.
var ErrUnknownSku = errors.New("not in the catalogue")

// SlowLookup is how long a lookup takes. GIZMO-SLOW is the one that hurts.
const SlowLookup = 30 * time.Second

// Catalogue is reference data the handler needs in order to do its job: what does this SKU cost?
//
// Here it is a map, because exercises 1 to 3 are not about lookups. It behaves like the
// real thing in the two ways that matter to us: it can be slow, and it can not know.
//
// (Exercise 4, if you get to it, replaces this with a local copy filled from a stream.)
type Catalogue struct{}

var prices = map[string]float64{
	"WIDGET-1":   9.99,
	"GIZMO-2":    24.50,
	"GIZMO-SLOW": 24.50, // in the catalogue, but the lookup crawls
}

func NewCatalogue() *Catalogue {
	return &Catalogue{}
}

func (c *Catalogue) PriceOf(sku string) (float64, error) {
	if sku == "GIZMO-SLOW" {
		fmt.Printf("  catalogue: looking up %s (this one takes %.0fs)\n", sku, SlowLookup.Seconds())
		time.Sleep(SlowLookup)
	}

	price, found := prices[sku]
	if !found {
		return 0, fmt.Errorf("'%s' is %w", sku, ErrUnknownSku)
	}

	return price, nil
}
