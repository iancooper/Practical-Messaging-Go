package simplemessaging

import (
	"context"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

// DataTypeChannelConsumer is the consumer half of the Messaging Gateway. It declares the
// topology described in channel.go and hands the pump five things it can do with a message.
//
// The plumbing is given to you and it is correct. Declaring exchanges and binding queues is
// AMQP vocabulary, not judgement, and you can read it here at your leisure. The exercise is
// deciding *which of these five to call, and when* -- and that lives in the pump.
type DataTypeChannelConsumer[T IAmAMessage] struct {
	connection          *amqp091.Connection
	channel             *amqp091.Channel
	queueName           string
	retryQueueName      string
	invalidQueueName    string
	deadLetterQueueName string
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

	consumer := &DataTypeChannelConsumer[T]{
		connection:          connection,
		channel:             channel,
		queueName:           QueueNameFor[T](),
		retryQueueName:      RetryQueueNameFor[T](),
		invalidQueueName:    InvalidQueueNameFor[T](),
		deadLetterQueueName: DeadLetterQueueNameFor[T](),
	}

	if err := consumer.declare(RoutingKeyFor[T]()); err != nil {
		consumer.Close()
		return nil, err
	}

	return consumer, nil
}

func (c *DataTypeChannelConsumer[T]) declare(routingKey string) error {
	if err := c.channel.ExchangeDeclare(
		ExchangeName, amqp091.ExchangeDirect, true, false, false, false, nil); err != nil {
		return err
	}
	if err := c.channel.ExchangeDeclare(
		DeadLetterExchangeName, amqp091.ExchangeDirect, true, false, false, false, nil); err != nil {
		return err
	}

	// The work queue. Rejecting a message from here (nack, requeue false) sends it to the
	// dead-letter exchange with the *retry* routing key -- so a rejection lands in the retry
	// queue without us publishing anything.
	//
	// Note what this means: the queue's dead-letter routing key is fixed at declare time.
	// One reject, one destination. Anything else you want to do with a message, you do by
	// publishing it somewhere yourself.
	if _, err := c.channel.QueueDeclare(c.queueName, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    DeadLetterExchangeName,
		"x-dead-letter-routing-key": c.retryQueueName,
	}); err != nil {
		return err
	}
	if err := c.channel.QueueBind(c.queueName, routingKey, ExchangeName, false, nil); err != nil {
		return err
	}

	// Bodies we could not read. A terminal destination: no TTL, no dead-letter exchange.
	if _, err := c.channel.QueueDeclare(c.invalidQueueName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := c.channel.QueueBind(
		c.invalidQueueName, c.invalidQueueName, DeadLetterExchangeName, false, nil); err != nil {
		return err
	}

	// Work we gave up on. Also terminal. This is the one an operator looks in.
	if _, err := c.channel.QueueDeclare(c.deadLetterQueueName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := c.channel.QueueBind(
		c.deadLetterQueueName, c.deadLetterQueueName, DeadLetterExchangeName, false, nil); err != nil {
		return err
	}

	// The retry queue: a waiting room with a clock on the door.
	// Nothing consumes it. Every message in it expires after RetryDelay, and an expired
	// message is dead-lettered -- back to the main exchange, and so back to the work queue.
	// RabbitMQ stamps an x-death header on the way through, which is how we count attempts.
	if _, err := c.channel.QueueDeclare(c.retryQueueName, true, false, false, false, amqp091.Table{
		"x-message-ttl":             int32(RetryDelay.Milliseconds()),
		"x-dead-letter-exchange":    ExchangeName,
		"x-dead-letter-routing-key": routingKey,
	}); err != nil {
		return err
	}
	return c.channel.QueueBind(c.retryQueueName, c.retryQueueName, DeadLetterExchangeName, false, nil)
}

// Receive asks the broker for one message. ok is false when the queue was empty.
func (c *DataTypeChannelConsumer[T]) Receive() (delivery amqp091.Delivery, ok bool, err error) {
	return c.channel.Get(c.queueName, false)
}

// Acknowledge says: done. The broker may forget it.
func (c *DataTypeChannelConsumer[T]) Acknowledge(deliveryTag uint64) error {
	return c.channel.Ack(deliveryTag, false)
}

// Requeue puts it back on the queue, right now, for someone to try again immediately.
// There is no limit on this and no delay. Think about what that means before you use it.
func (c *DataTypeChannelConsumer[T]) Requeue(deliveryTag uint64) error {
	return c.channel.Nack(deliveryTag, false, true)
}

// RejectForRetry rejects it. Because of the work queue's arguments, the broker routes it to
// the **retry queue**, where it waits and then comes back on its own. One call, and RabbitMQ
// does the moving -- and because RabbitMQ owns both hops, RabbitMQ counts them for you.
func (c *DataTypeChannelConsumer[T]) RejectForRetry(deliveryTag uint64) error {
	return c.channel.Nack(deliveryTag, false, false)
}

// SendToInvalidMessageQueue publishes it to the invalid message queue. Terminal: a body
// nobody can read.
//
// This is a publish, not a reject -- so the original delivery is still outstanding and it is
// still your problem. Headers are carried forward, because x-death is the attempt count and
// losing it resets the clock.
func (c *DataTypeChannelConsumer[T]) SendToInvalidMessageQueue(delivery amqp091.Delivery) error {
	return c.republish(delivery, c.invalidQueueName)
}

// SendToDeadLetter sends it to the dead letter queue. Terminal: somebody has to come and look.
func (c *DataTypeChannelConsumer[T]) SendToDeadLetter(delivery amqp091.Delivery) error {
	return c.republish(delivery, c.deadLetterQueueName)
}

func (c *DataTypeChannelConsumer[T]) republish(delivery amqp091.Delivery, routingKey string) error {
	return c.channel.PublishWithContext(
		context.Background(), DeadLetterExchangeName, routingKey, false, false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  delivery.ContentType,
			Headers:      delivery.Headers,
			Body:         delivery.Body,
		})
}

// RetriesSoFar answers: how many times has this message been round the retry loop?
//
// RabbitMQ records every dead-lettering in an x-death header: an array of entries, one per
// (queue, reason) pair, each with a count. A message that has expired out of the retry queue
// twice has an entry for that queue with count 2. A message arriving for the first time has
// no x-death header at all, so it has had no attempts yet.
//
// Look at this header in the management console. It is the most useful thing RabbitMQ will
// tell you about a message's history and almost nobody knows it is there.
func (c *DataTypeChannelConsumer[T]) RetriesSoFar(delivery amqp091.Delivery) int {
	deaths, ok := delivery.Headers["x-death"].([]interface{})
	if !ok {
		return 0
	}

	for _, death := range deaths {
		entry, ok := death.(amqp091.Table)
		if !ok {
			continue
		}
		if queue, _ := entry["queue"].(string); queue != c.retryQueueName {
			continue
		}
		switch count := entry["count"].(type) {
		case int64:
			return int(count)
		case int32:
			return int(count)
		case int:
			return count
		}
	}

	return 0
}

func (c *DataTypeChannelConsumer[T]) Close() {
	c.channel.Close()
	c.connection.Close()
}
