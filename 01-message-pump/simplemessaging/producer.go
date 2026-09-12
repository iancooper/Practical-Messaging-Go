package simplemessaging

import (
	"context"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

// DataTypeChannelProducer is the producer half of the Messaging Gateway: the only type here
// that knows this is RabbitMQ.
//
// Under RMQ, to send, we:
//
//  1. open a socket connection to the broker
//  2. create a channel (a lightweight logical connection) on that socket
//  3. declare a direct exchange to publish to
//
// We do not declare the queue. The consumer does that, and binds it to our routing key.
// That is the asymmetry AMQP has and a queue API does not: we publish to an exchange and
// have no idea who, if anyone, is listening.
//
// This type is given to you and it is correct. Read it for the AMQP vocabulary; the
// exercise is not here.
type DataTypeChannelProducer[T IAmAMessage] struct {
	serialize  func(T) (string, error)
	connection *amqp091.Connection
	channel    *amqp091.Channel
	routingKey string
}

// NewDataTypeChannelProducer connects to the broker and declares the exchange.
//
// Connecting is I/O and I/O fails, so this returns an error rather than panicking. That is
// not a style preference: exercise 2 is entirely about a pump that tells failures apart and
// chooses a policy for each, and a gateway that calls log.Fatal on your behalf has taken
// that decision away from you.
//
// serialize turns a T into the string we put in the body.
func NewDataTypeChannelProducer[T IAmAMessage](
	serialize func(T) (string, error), hostName string,
) (*DataTypeChannelProducer[T], error) {
	connection, err := amqp091.Dial(amqpURL(hostName))
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, err
	}

	// Durable, so the exchange survives a broker restart.
	err = channel.ExchangeDeclare(ExchangeName, amqp091.ExchangeDirect, true, false, false, false, nil)
	if err != nil {
		channel.Close()
		connection.Close()
		return nil, err
	}

	return &DataTypeChannelProducer[T]{
		serialize:  serialize,
		connection: connection,
		channel:    channel,
		routingKey: RoutingKeyFor[T](),
	}, nil
}

// Send a message. The routing key is derived from T, so sender and receiver match up
// without either knowing about the other.
func (p *DataTypeChannelProducer[T]) Send(message T) error {
	body, err := p.serialize(message)
	if err != nil {
		return err
	}
	return p.publish(body)
}

// SendRaw sends a body we did not serialize -- used to put something unmappable on the channel.
func (p *DataTypeChannelProducer[T]) SendRaw(body string) error {
	return p.publish(body)
}

func (p *DataTypeChannelProducer[T]) publish(body string) error {
	// DeliveryMode 2 is persistent: the broker writes the message to its message store, so
	// it survives a broker restart. This is the producer-side half of guaranteed delivery,
	// and it is the cheap half.
	return p.channel.PublishWithContext(
		context.Background(),
		ExchangeName,
		p.routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
}

func (p *DataTypeChannelProducer[T]) Close() {
	p.channel.Close()
	p.connection.Close()
}
