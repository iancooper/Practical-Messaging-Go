package p2pchannel

import (
	"github.com/streadway/amqp"
	"log"
)

type P2p struct {
	xchng      string
	queueName  string
	routingKey string
	conn       *amqp.Connection
}

//We just use a contant here for convenience, in reality you configure this
const exchange = "practical-messaging-go"

func NewChannel(qName string) *P2p {

	channel := new(P2p)
	channel.xchng = exchange
	channel.queueName = qName
	channel.routingKey = qName

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	channel.conn = conn

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel", channel)
	defer ch.Close()

	// TODO: Declare an exchange, with exchange_name, type direct, non-durable, and not auto-delete on the channel
	// TODO: Declare a queue with _queue_name, non-durable, non-exclusive, and not auto-delete

	return channel
}

func (channel *P2p) Receive() (bool, string) {
	ch, err := channel.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	defer ch.Close()

	//# TODO: use basic_get on the channel to retrieve a message from the queue not using auto_ack
}

func (channel *P2p) Send(message string) {
	ch, err := channel.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	defer ch.Close()

	// TODO: publish the message to the exchange using _routing_key n the channel
}

func (channel *P2p) Close() {
	if channel.conn != nil {
		channel.conn.Close()
	}
}

func failOnError(err error, msg string, channel *P2p) {
	if err != nil {
		if channel.conn != nil {
			channel.conn.Close()

		log.Fatalf("%s: %s", msg, err)
		}
	}
}
