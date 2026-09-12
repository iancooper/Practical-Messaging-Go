package simplemessaging

import (
	"fmt"
	"reflect"
	"time"
)

// The topology, in one place -- and there is more of it than there was in exercise 1.
//
// Four queues now, because a message that is never going to be handled has more than one
// place it can end up, and *which* place is how you find out what went wrong.
//
//	practical-messaging-streams              (direct)
//	  `-- streams.<T>                  the work
//	         x-dead-letter-exchange:    ...dlx
//	         x-dead-letter-routing-key: retry.streams.<T>
//
//	practical-messaging-streams.dlx          (direct)
//	  |-- retry.streams.<T>            work waiting to be tried again
//	  |      x-message-ttl:             5000
//	  |      x-dead-letter-exchange:    practical-messaging-streams
//	  |      x-dead-letter-routing-key: streams.<T>
//	  |
//	  |-- invalid.streams.<T>          a body we could not read
//	  `-- dead.streams.<T>             work we retried and gave up on
//
// The retry queue is the part worth understanding, because RabbitMQ does the work and you
// do not. Nothing consumes it, and every message in it has five seconds to live. So each
// message expires -- and an expired message is dead-lettered, and this queue's dead-letter
// exchange points back at the main exchange. It comes home on a timer nobody wrote.
//
// That is *Requeue with Delay*, built out of a TTL and a dead-letter exchange. Stock
// RabbitMQ: no plugin, no scheduler, no code.
//
// And notice which way round the two hops go. Rejecting a message from the work queue sends
// it to *retry*, not to the dead letter queue -- because the round trip work -> retry ->
// work is a cycle RabbitMQ manages end to end, and a cycle it manages is a cycle it will
// count for you in the x-death header. The other two destinations are terminal, so nothing
// needs counting and we can publish to them directly.
//
// The two terminal queues should be empty. When they are not, that is the alert.
const (
	// HostName and Port are one of the two places the broker's address appears. The other
	// is simpleeventing, in exercise 3.
	HostName = "localhost"
	Port     = 5672

	ExchangeName           = "practical-messaging-streams"
	DeadLetterExchangeName = ExchangeName + ".dlx"
)

// RetryDelay is how long a message waits in the retry queue before it comes back.
const RetryDelay = 5 * time.Second

func amqpURL(hostName string) string {
	return fmt.Sprintf("amqp://guest:guest@%s:%d/", hostName, Port)
}

func RoutingKeyFor[T IAmAMessage]() string {
	return "streams." + reflect.TypeFor[T]().String()
}

func QueueNameFor[T IAmAMessage]() string {
	return RoutingKeyFor[T]()
}

// RetryQueueNameFor names work waiting for its next attempt. Should always be nearly empty.
func RetryQueueNameFor[T IAmAMessage]() string {
	return "retry." + QueueNameFor[T]()
}

// InvalidQueueNameFor names bodies we could not read. Terminal: nothing here is ever retried.
func InvalidQueueNameFor[T IAmAMessage]() string {
	return "invalid." + QueueNameFor[T]()
}

// DeadLetterQueueNameFor names work we retried and gave up on. Terminal: an operator's in-tray.
func DeadLetterQueueNameFor[T IAmAMessage]() string {
	return "dead." + QueueNameFor[T]()
}
