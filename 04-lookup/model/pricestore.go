package model

import "time"

// Price is one row of the local copy: a price, when it became true, and when we heard.
//
//	Sku        what it is a price for
//	Amount     the price itself
//	ChangedAt  when the catalogue service says it changed. Comes off the event
//	AppliedAt  when *we* wrote it down. The gap between the two is Probe A
type Price struct {
	Sku       string
	Amount    float64
	ChangedAt time.Time
	AppliedAt time.Time
}

// IPriceStore is the local copy of somebody else's reference data, as the domain sees it.
//
// ---------------------------------------------------------------------------------------
//
//	THIS INTERFACE IS DECLARED BY model, AND THAT IS THE WHOLE POINT OF IT.
//
//	model imports simplemessaging and nothing else. It does not know that the copy is
//	SQLite, that it is a file, or that a separate process fills it -- localcopy knows all
//	three, cmd/receiver puts the two together, and the domain names none of it.
//
//	It is exercise 1's fix arriving a second time, against a storage technology instead
//	of a broker, and the seam did not have to change to cope. Add
//	github.com/mattn/go-sqlite3 to an import block under model/ and you have undone it --
//	and, as in exercise 1, the compiler is the check rather than a code-review opinion.
//
// ---------------------------------------------------------------------------------------
//
// Note that it answers three different questions, not one. "Have you a price for this?" is
// the obvious one. "Have you any prices at all?" separates *the SKU is unknown* from *we
// are not ready yet*, which is Probe C. "How old is the newest thing you have?" is the only
// question that can tell a current copy from a stale one, and it is the one nothing asks
// often enough.
//
// A lookup that can only say yes or no cannot be operated. That is a design decision you
// make when you write the interface, long before anybody needs the answer.
//
// The shape of each method is the gateway's (value, ok, error), as in
// simplemessaging.DataTypeChannelConsumer.Receive: "I have not got one" and "I could not go
// and look" are two different answers and a single nil would merge them.
type IPriceStore interface {
	// Lookup returns the price for this SKU. ok is false when the local copy has no row
	// for it -- which, on its own, does not tell you whether the SKU exists.
	Lookup(sku string) (Price, bool, error)

	// Count is how many prices the copy holds. ZERO MEANS "I HAVE NEVER BEEN FILLED",
	// which is not the same fact as "that SKU is not a thing" and must not produce the
	// same behaviour.
	Count() (int, error)

	// NewestChangedAt is when the catalogue last changed something we know about. ok is
	// false when the copy is empty.
	//
	// This is the number Probe C is really about. A copy that cannot say how old it is
	// cannot be monitored, and a copy that cannot be monitored is one you find out about
	// from a customer.
	NewestChangedAt() (time.Time, bool, error)
}
