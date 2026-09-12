package simplemessaging

import (
	"fmt"
	"reflect"
)

// The names both ends of the channel have to agree on, in one place.
//
// A Datatype Channel carries one type of message, so we derive the routing key from the
// type. Producer and consumer both compute it, which is how they find each other without a
// shared config file.
//
// Go erases nothing, but a type parameter is not a value, so the name has to come from
// reflection: reflect.TypeFor[T]().String() gives "model.PlaceOrder". That is the same
// string Java's getName() produces, so a Go sender and a Java receiver would find the same
// queue -- which is the point of deriving the name rather than configuring it.
const (
	// HostName and Port are one of the two places the broker's address appears. The other
	// is simpleeventing, in exercise 3.
	HostName = "localhost"
	Port     = 5672

	ExchangeName = "practical-messaging-pump"
)

// amqpURL is the address we dial. Username and password are RabbitMQ's defaults.
func amqpURL(hostName string) string {
	return fmt.Sprintf("amqp://guest:guest@%s:%d/", hostName, Port)
}

func RoutingKeyFor[T IAmAMessage]() string {
	return "message-pump." + reflect.TypeFor[T]().String()
}

func QueueNameFor[T IAmAMessage]() string {
	return RoutingKeyFor[T]()
}
