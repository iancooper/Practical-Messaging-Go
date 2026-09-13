// Package localcopy is the only package in this exercise that names SQLite.
//
// That is the boundary, and in Go the boundary is the import block: nothing under model/
// imports github.com/mattn/go-sqlite3, this package does, and the compiler will not let a
// package name something it did not import. Same rule and same reason as simplemessaging's
// amqp091-go and simpleeventing's confluent-kafka-go -- the thing that knows which
// technology this is gets a package drawn round it.
package localcopy

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	// The driver registers itself with database/sql, which is why it is a blank import:
	// even here, the only line that names SQLite is a string.
	_ "github.com/mattn/go-sqlite3"

	"github.com/iancooper/Practical-Messaging-Go/lookup/model"
)

// DefaultPath sits next to whatever you ran, so all four processes share one copy.
const DefaultPath = "prices.db"

// timeLayout is RFC 3339 with a FIXED nine digits of fraction, and that fixed width is the
// point of writing a layout out rather than using time.RFC3339Nano.
//
// RFC3339Nano trims trailing zeros, so "...:00Z" and "...:00.5Z" come out different lengths --
// and SQLite's MAX() over a TEXT column is a string comparison, so NewestChangedAt would
// quietly order them wrong. Same width, always UTC, and lexicographic order is chronological
// order again.
const timeLayout = "2006-01-02T15:04:05.000000000Z07:00"

// SqlitePriceStore is the local copy of the catalogue's prices, in a SQLite file.
//
// A file is the cheapest durable store there is, and durable is the only property the
// exercise actually needs: the price consumer that fills this runs in its own process, and
// Probes B, C and D all turn on what is still here after that process dies.
//
// It implements model.IPriceStore, which model declares, so the domain reads prices without
// knowing any of this exists. It also exposes Apply, which model does *not* know about:
// reading the copy is the domain's business, and maintaining it is the price consumer's.
type SqlitePriceStore struct {
	db *sql.DB
}

// Open opens (and if need be creates) the copy at path.
func Open(path string) (*SqlitePriceStore, error) {
	// PRAGMA journal_mode=WAL and PRAGMA busy_timeout=5000, said in the DSN.
	//
	// WAL, because a reader and a writer are two different processes here and the default
	// journal would have them locking each other out. The busy timeout is the other half of
	// the same thought: a writer that finds the file busy waits five seconds instead of
	// failing at once. This is a real decision and not boilerplate -- "my local copy is a
	// file" stops being free the moment two processes want it at once.
	//
	// In the DSN rather than an Exec because busy_timeout is a property of a CONNECTION and
	// database/sql hands out a pool: run it once as a statement and the next connection the
	// pool opens for you has the default again. Put it in the DSN and every connection gets
	// it. (journal_mode is written into the file, so that one would have stuck either way.)
	db, err := sql.Open("sqlite3",
		fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", path))
	if err != nil {
		return nil, err
	}

	// One connection. Nothing here is concurrent -- the receiver only reads and the price
	// consumer only writes, each on one goroutine -- and a pool of writers onto one file
	// buys nothing but lock contention we would then have to explain.
	db.SetMaxOpenConns(1)

	// The price is TEXT rather than REAL on purpose. A price is a decimal and SQLite's REAL
	// is a double, and 9.99 is not a double. Go has no decimal type either, so the float64
	// we carry in the event is already an approximation -- which is exactly why it must not
	// be rounded a second time by the store. Storing money in a float is a bug that takes
	// months to surface and this is the line that keeps it to one place.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS prices (
			sku        TEXT PRIMARY KEY,
			price      TEXT NOT NULL,
			changed_at TEXT NOT NULL,
			applied_at TEXT NOT NULL
		)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SqlitePriceStore{db: db}, nil
}

func (s *SqlitePriceStore) Lookup(sku string) (model.Price, bool, error) {
	row := s.db.QueryRow(
		"SELECT price, changed_at, applied_at FROM prices WHERE sku = ?", sku)

	var priceText, changedAt, appliedAt string
	if err := row.Scan(&priceText, &changedAt, &appliedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Not an error: we simply have no row for this SKU. Whether that is because
			// the SKU does not exist or because we have never been filled is the
			// Catalogue's question to ask, and Count is how it asks it.
			return model.Price{}, false, nil
		}
		return model.Price{}, false, err
	}

	price, err := parsePrice(priceText)
	if err != nil {
		return model.Price{}, false, err
	}
	changed, err := time.Parse(timeLayout, changedAt)
	if err != nil {
		return model.Price{}, false, err
	}
	applied, err := time.Parse(timeLayout, appliedAt)
	if err != nil {
		return model.Price{}, false, err
	}

	return model.Price{Sku: sku, Amount: price, ChangedAt: changed, AppliedAt: applied}, true, nil
}

func (s *SqlitePriceStore) Count() (int, error) {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM prices").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *SqlitePriceStore) NewestChangedAt() (time.Time, bool, error) {
	// MAX over an empty table is one row containing NULL rather than no rows at all, so
	// the emptiness arrives as a nil string and not as sql.ErrNoRows.
	var newest sql.NullString
	if err := s.db.QueryRow("SELECT MAX(changed_at) FROM prices").Scan(&newest); err != nil {
		return time.Time{}, false, err
	}
	if !newest.Valid {
		return time.Time{}, false, nil
	}

	when, err := time.Parse(timeLayout, newest.String)
	if err != nil {
		return time.Time{}, false, err
	}
	return when, true, nil
}

// Apply writes a price change into the copy, and returns when the row was written, for
// Probe A's arithmetic.
//
// LAST WRITER WINS, which is only safe because model.PriceChanged is a snapshot -- run this
// twice with the same event and the row ends up the same. That is Probe D's whole payout,
// and it was decided in step 1.
func (s *SqlitePriceStore) Apply(event model.PriceChanged) (time.Time, error) {
	appliedAt := time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO prices (sku, price, changed_at, applied_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(sku) DO UPDATE SET
			price      = excluded.price,
			changed_at = excluded.changed_at,
			applied_at = excluded.applied_at`,
		event.Sku,
		formatPrice(event.Price),
		event.ChangedAt.UTC().Format(timeLayout),
		appliedAt.Format(timeLayout))
	if err != nil {
		return time.Time{}, err
	}

	return appliedAt, nil
}

func (s *SqlitePriceStore) Close() error { return s.db.Close() }

// formatPrice writes the number in the shortest decimal that parses back to the same
// float64 -- 'f' with a precision of -1 -- so that 9.99 comes out of the file as 9.99 and
// not as 9.9900000000000002. Round-tripping exactly is the only thing TEXT buys us over
// REAL, so it is worth being explicit about which formatting gives it.
func formatPrice(amount float64) string {
	return strconv.FormatFloat(amount, 'f', -1, 64)
}

func parsePrice(text string) (float64, error) {
	return strconv.ParseFloat(text, 64)
}

// Compile-time proof that the store satisfies the interface the domain declares. If Apply
// or a rename ever breaks that, it breaks here rather than in cmd/receiver.
var _ model.IPriceStore = (*SqlitePriceStore)(nil)
