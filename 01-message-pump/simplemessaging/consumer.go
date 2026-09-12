package simplemessaging

import (
	amqp091 "github.com/rabbitmq/amqp091-go"
)

// DataTypeChannelConsumer is the consumer half of the Messaging Gateway. Again, the only
// type that knows this is RabbitMQ.
//
// Under RMQ, to receive, we:
//
//  1. open a socket connection to the broker
//  2. create a channel on that socket
//  3. declare the same direct exchange the producer publishes to
//  4. declare a queue to hold our messages
//  5. bind the queue to the routing key on that exchange
//
// Both ends declare the exchange, so it does not matter which starts first. Only we declare
// the queue -- which does mean that anything published before the first run of the consumer
// went nowhere. Start the consumer once before you send anything.
//
// This is a Polling Consumer: Receive asks the broker whether there is anything there. It
// costs us a goroutine while idle and buys us not having to hold a connection open for the
// broker to call back on.
//
// This type is given to you and it is correct. Read it for the AMQP vocabulary; the
// exercise is not here.
type DataTypeChannelConsumer[T IAmAMessage] struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queueName  string
}

func NewDataTypeChannelConsumer[T IAmAMessage](hostName string) (*DataTypeChannelConsumer[T], error) {
	connection, err := amqp091.Dial(amqpURL(hostName))
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, err
	}

	routingKey := RoutingKeyFor[T]()
	queueName := QueueNameFor[T]()

	if err := channel.ExchangeDeclare(
		ExchangeName, amqp091.ExchangeDirect, true, false, false, false, nil); err != nil {
		channel.Close()
		connection.Close()
		return nil, err
	}

	// Durable queue to go with the persistent messages: no point writing a message to disk
	// and then keeping it in a queue that evaporates on restart.
	if _, err := channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		channel.Close()
		connection.Close()
		return nil, err
	}

	if err := channel.QueueBind(queueName, routingKey, ExchangeName, false, nil); err != nil {
		channel.Close()
		connection.Close()
		return nil, err
	}

	return &DataTypeChannelConsumer[T]{connection: connection, channel: channel, queueName: queueName}, nil
}

// Receive asks the broker for one message. ok is false when the queue was empty.
//
// autoAck is false, so what comes back is *locked to us* and not yet removed from the
// queue. The broker is now waiting to be told what happened. Until we tell it, this message
// shows in the management console as "unacked".
func (c *DataTypeChannelConsumer[T]) Receive() (delivery amqp091.Delivery, ok bool, err error) {
	return c.channel.Get(c.queueName, false)
}

// Acknowledge says: I am done with this message. The broker may forget it.
func (c *DataTypeChannelConsumer[T]) Acknowledge(deliveryTag uint64) error {
	return c.channel.Ack(deliveryTag, false)
}

// Reject says: I am not done with this message.
//
//	requeue true  -- put it back, someone will try again (possibly us, immediately).
//	requeue false -- reject it. On a plain queue that deletes it.
func (c *DataTypeChannelConsumer[T]) Reject(deliveryTag uint64, requeue bool) error {
	return c.channel.Nack(deliveryTag, false, requeue)
}

func (c *DataTypeChannelConsumer[T]) Close() {
	c.channel.Close()
	c.connection.Close()
}
