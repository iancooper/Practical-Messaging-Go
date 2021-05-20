package p2pchannel

import (
	"log"

	"github.com/streadway/amqp"
)

type P2P struct {
	xchng      string
	queueName  string
	routingKey string
	conn       *amqp.Connection
}

//We just use a constant here for convenience, in reality you'd configure this
const exchange = "practical-messaging-go"

func NewChannel(qName string) *P2P {
	channel := &P2P{
		xchng:      exchange,
		queueName:  qName,
		routingKey: qName,
	}

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

	_, err = ch.QueueDeclare(
		qName, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue", channel)

	err = ch.QueueBind(
		channel.queueName,  // queue name
		channel.routingKey, // routing key
		exchange,           // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue", channel)

	return channel
}

func (channel *P2P) Receive() (bool, string) {
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
	}

	return false, ""
}

func (channel *P2P) Send(message string) {
	ch, err := channel.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	defer ch.Close()

	err = ch.Publish(
		channel.xchng,      //exchange
		channel.routingKey, //routing key
		false,              //mandatory
		false,              //immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	failOnError(err, "Error sending message to RabbitMQ", channel)
}

func (channel *P2P) Close() {
	if channel.conn != nil {
		channel.conn.Close()
	}
}

func failOnError(err error, msg string, channel *P2P) {
	if err != nil {
		if channel.conn != nil {
			channel.conn.Close()

			log.Fatalf("%s: %s", msg, err)
		}
	}
}
