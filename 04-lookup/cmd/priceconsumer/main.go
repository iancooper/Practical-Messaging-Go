// Follows streams.PriceChanged and maintains the local copy of the catalogue's prices.
//
//	go run ./cmd/priceconsumer
//	PRICE_WRITE_WINDOW=15 go run ./cmd/priceconsumer     # for Probe D
//
// It is a process of its own, and that is deliberate. A goroutine inside the receiver would
// have been less code; it would also have made Probe B "set a flag" instead of "kill -9 a
// real process", and the whole of exercise 4 is about what happens to the people downstream
// of a thing that stops.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/iancooper/Practical-Messaging-Go/lookup/localcopy"
	"github.com/iancooper/Practical-Messaging-Go/lookup/model"
	"github.com/iancooper/Practical-Messaging-Go/lookup/simpleeventing"
)

// priceWriteWindow is how long to pause BETWEEN the two writes below, so that you can aim a
// kill at the gap.
//
// In a real service the gap is microseconds wide. It is still a gap. This is the same
// instrument as DUAL_WRITE_WINDOW in exercise 3, one layer down and in your own code.
func priceWriteWindow() time.Duration {
	seconds, err := strconv.Atoi(os.Getenv("PRICE_WRITE_WINDOW"))
	if err != nil {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func main() {
	pid := os.Getpid()
	fmt.Println("PriceConsumer starting. PID", pid)

	window := priceWriteWindow()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		<-interrupts
		fmt.Println("\nStopping...")
		cancel()
	}()

	store, err := localcopy.Open(localcopy.DefaultPath)
	if err != nil {
		log.Fatalln("could not open the local copy:", err)
	}
	defer store.Close()

	reader, err := simpleeventing.NewEventStreamReader[model.PriceChanged](
		model.DeserializePriceChanged,
		simpleeventing.PriceConsumerGroup,
		simpleeventing.BootstrapServers)
	if err != nil {
		log.Fatalln("could not connect to Kafka:", err)
	}
	defer reader.Close()

	fmt.Printf("Following %s as group '%s'\n", reader.Topic(), reader.Group())

	count, err := store.Count()
	if err != nil {
		log.Fatalln("could not read the local copy:", err)
	}
	newest, any, err := store.NewestChangedAt()
	if err != nil {
		log.Fatalln("could not read the local copy:", err)
	}
	if any {
		fmt.Printf("Local copy is %s, holding %d prices -- newest change %s old.\n",
			localcopy.DefaultPath, count, age(newest))
	} else {
		fmt.Printf("Local copy is %s, holding %d prices -- empty.\n", localcopy.DefaultPath, count)
	}

	if window > 0 {
		fmt.Printf("PRICE_WRITE_WINDOW is %.0fs -- there is a gap between the two writes.\n",
			window.Seconds())
	}

	for ctx.Err() == nil {
		record, err := reader.Read(ctx)
		if err != nil {
			log.Fatalln("price consumer stopped:", err)
		}
		if record == nil {
			continue
		}

		event := record.Message

		// ---------------------------------------------------------------------------
		//  TWO WRITES, TWO STORES, NO TRANSACTION. PROBE D IS THE ORDER OF THESE LINES.
		//
		//  The price goes into SQLite. The offset goes into Kafka. Nothing on this
		//  machine can make those two happen together, which is exactly what exercise 3
		//  showed you in the receiver -- except that this time it is a loop you wrote,
		//  and it looks like one step.
		//
		//  As written: apply, then commit. Die in between and the record is read again
		//  on restart and applied twice, which is harmless *because PriceChanged is a
		//  snapshot*. Swap the two lines and die in between and the price is lost for
		//  ever, because Kafka will never offer it again.
		// ---------------------------------------------------------------------------
		appliedAt, err := store.Apply(event)
		if err != nil {
			log.Fatalln("could not apply the price:", err)
		}

		fmt.Printf("  %s = %.2f  (%s)  published-to-applied %.0f ms\n",
			event.Sku, event.Price, record.Where,
			appliedAt.Sub(event.ChangedAt).Seconds()*1000)

		pauseInTheWindow(ctx, window, pid)

		if err := reader.Commit(record); err != nil {
			log.Fatalln("could not commit the offset:", err)
		}
	}

	fmt.Println("PriceConsumer stopped.")
}

func pauseInTheWindow(ctx context.Context, window time.Duration, pid int) {
	if window <= 0 {
		return
	}

	fmt.Println("  [one of the two writes has happened and the other has not.]")
	fmt.Printf("  [you have %.0f seconds. kill -9 %d]\n", window.Seconds(), pid)
	select {
	case <-ctx.Done():
	case <-time.After(window):
	}
}

// age says how old, in words, without saying "1 minutes". A copy's age is the one number
// that tells a current local copy from a stale one, so it is worth printing in a shape a
// human reads.
func age(when time.Time) string {
	d := time.Since(when)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%.0fs", d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%.0fm", d.Minutes())
	case d < 24*time.Hour:
		return fmt.Sprintf("%.0fh", d.Hours())
	default:
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
}
