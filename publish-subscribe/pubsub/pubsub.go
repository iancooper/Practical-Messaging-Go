package pubsub

import (
	"github.com/streadway/amqp"
	"log"
)

type Pubsub struct {
	xchng     string
	queueName string
	conn      *amqp.Connection
}

//We just use a contant here for convenience, in reality you configure this
const exchange = "practical-messaging-go-fanout"

func NewChannel(qName string) *Pubsub {

	channel := new(Pubsub)
	channel.xchng = exchange
	channel.queueName = qName

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	channel.conn = conn

	// # TODO; Declare a fanout exchange, non-durable, and not auto-deleting
	//
	//  Optional: Only declare if a consumer (workls without this step byt not realistic)
	//  TODO: Declare a non-durable queue, exclusive and auto-deleting with no name
	//  TODO: Get the random queue name from the result of the queue creation operation
	//  TODO: Bind the queue to the exchante with the returned queue name
	//
	//

	return channel
}

func (channel *Pubsub) Receive() (bool, string) {
	ch, err := channel.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	defer ch.Close()

	msg, ok, err := ch.Get(
		channel.queueName, //queue name
		true,              //auto ack when we read
	)
	failOnError(err, "Failed to receive from RabbitMQ", channel)

	if ok {
		return true, string(msg.Body[:])
	} else {
		return false, ""
	}
}

func (channel *Pubsub) Send(message string) {
	ch, err := channel.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	defer ch.Close()

	//TODO: Publish the message with an empty routing key

	failOnError(err, "Error sending message to RabbitMQ", channel)
}

func (channel *Pubsub) Close() {
	if channel.conn != nil {
		channel.conn.Close()
	}
}

func failOnError(err error, msg string, channel *Pubsub) {
	if err != nil {
		if channel.conn != nil {
			channel.conn.Close()

			log.Fatalf("%s: %s", msg, err)
		}
	}
}
