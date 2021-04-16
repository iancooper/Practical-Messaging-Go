package datachannel

import (
	"github.com/streadway/amqp"
	"log"
)

//We just use a contant here for convenience, in reality you configure this
const exchange = "practical-messaging-go"

type channel struct {
	xchng      string
	queueName  string
	routingKey string
	conn       *amqp.Connection
}

//Channel
func newChannel(qName string, declareQueue bool) *channel {

	channel := new(channel)
	channel.xchng = exchange
	channel.queueName = qName
	channel.routingKey = qName
	invalid_routing_key := qName + ".invalid"

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	channel.conn = conn

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel", channel)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange", channel)

	if declareQueue {
		newQueue(qName, err, ch, invalid_routing_key, channel)
	}

	return channel
}

func newQueue(qName string, err error, ch *amqp.Channel, invalid_routing_key string, channel *channel) {
	_, err = ch.QueueDeclare(
		qName, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{"x-dead-letter-exchange": invalid_exchange, "x-dead-letter-routing-key": invalid_routing_key},
	)
	failOnError(err, "Failed to declare a queue", channel)

	err = ch.QueueBind(
		channel.queueName,  // queue name
		channel.routingKey, // routing key
		exchange,           // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue", channel)

	//We don't need a second exchange, but it's one option to segregate them
	err = ch.ExchangeDeclare(
		invalid_exchange, // name
		"direct",         // type
		false,            // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a invalid message exchange", channel)

	_, err = ch.QueueDeclare(
		invalid_routing_key, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 //arguments
	)
	failOnError(err, "Failed to declare a queue", channel)

	err = ch.QueueBind(
		invalid_routing_key, // queue name
		invalid_routing_key, // routing key
		invalid_exchange,    // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue", channel)
}

func (channel *channel) Close() {
	if channel.conn != nil {
		channel.conn.Close()
	}
}

func failOnError(err error, msg string, channel *channel) {
	if err != nil {
		channel.Close()
		log.Fatalf("%s: %s", msg, err)
	}

}
