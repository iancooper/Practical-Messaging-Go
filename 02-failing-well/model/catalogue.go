package model

import (
	"errors"
	"fmt"
	"time"
)

// ErrUnknownSku is permanent. This SKU does not exist and asking again will not change that.
var ErrUnknownSku = errors.New("not in the catalogue")

// ErrCatalogueUnavailable is transient. The lookup is unwell; the order is fine. Try again
// shortly.
var ErrCatalogueUnavailable = errors.New("catalogue is unavailable")

// SlowLookup is how long the slow lookup takes.
const SlowLookup = 30 * time.Second

// FlakyFailures is how many times FLAKY-1 fails before it starts working.
const FlakyFailures = 2

// Catalogue is reference data the handler needs: what does this SKU cost?
//
// It fails in the three ways a real lookup fails, and telling them apart is the exercise:
//
//	WIDGET-1, GIZMO-2   fine
//	GIZMO-SLOW          in the catalogue, but the lookup takes 30 seconds
//	FLAKY-1             fails twice, then works -- a service that was restarting
//	anything else       not in the catalogue, and never will be
type Catalogue struct {
	flakyAttempts int
}

var prices = map[string]float64{
	"WIDGET-1":   9.99,
	"GIZMO-2":    24.50,
	"GIZMO-SLOW": 24.50,
	"FLAKY-1":    12.00,
}

func NewCatalogue() *Catalogue {
	return &Catalogue{}
}

func (c *Catalogue) PriceOf(sku string) (float64, error) {
	if sku == "GIZMO-SLOW" {
		fmt.Printf("  catalogue: looking up %s (this one takes %.0fs)\n", sku, SlowLookup.Seconds())
		time.Sleep(SlowLookup)
	}

	if sku == "FLAKY-1" {
		c.flakyAttempts++
		if c.flakyAttempts <= FlakyFailures {
			return 0, fmt.Errorf("%w (attempt %d for '%s')", ErrCatalogueUnavailable, c.flakyAttempts, sku)
		}
		fmt.Printf("  catalogue: %s worked on attempt %d\n", sku, c.flakyAttempts)
	}

	price, found := prices[sku]
	if !found {
		return 0, fmt.Errorf("'%s' is %w", sku, ErrUnknownSku)
	}

	return price, nil
}
