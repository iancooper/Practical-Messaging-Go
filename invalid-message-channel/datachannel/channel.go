package datachannel

import (
	"github.com/streadway/amqp"
	"log"
)

type channel struct {
	xchng      string
	queueName  string
	routingKey string
	conn       *amqp.Connection
}

type Producer struct {
	*channel
	serialize Serializer
}

type Consumer struct {
	*channel
	deserialize Deserializer
}
type Serializer func(message interface{}) ([]byte, error)
type Deserializer func(bytes []byte) (interface{}, error)

//We just use a contant here for convenience, in reality you configure this
const exchange = "practical-messaging-go"
const invalid_exchange = "practical-messaging-invalid"

func NewProducer(qName string, serializer Serializer) *Producer {
	producer := new(Producer)
	producer.channel = newChannel(qName)
	producer.serialize = serializer
	return producer
}

func NewConsumer(qName string, deserializer Deserializer) *Consumer {
	consumer := new(Consumer)
	consumer.channel = newChannel(qName)
	consumer.deserialize = deserializer
	return consumer
}

func newChannel(qName string) *channel {

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

	return channel
}

//Channel
func (channel *channel) Close() {
	if channel.conn != nil {
		channel.conn.Close()
	}
}

//Consumer
func (c *Consumer) Receive() (bool, interface{}) {
	ch, err := c.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", c.channel)
	defer ch.Close()

	msg, ok, err := ch.Get(
		c.queueName, //queue name
		false,       //we don't auto-ack as we may reject
	)
	failOnError(err, "Failed to receive from RabbitMQ", c.channel)

	if ok {
		message, err := c.deserialize(msg.Body)
		if err == nil {
			ch.Ack(msg.DeliveryTag, false)
			return true, message
		} else {
			ch.Nack(msg.DeliveryTag, false, false) //requeue true will push to DLQ
			log.Println("Error receiving message", err.Error())
		}

	}
	return false, nil
}

//Producer
func (p *Producer) Send(message interface{}) {
	ch, err := p.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", p.channel)
	defer ch.Close()

	b, err := p.serialize(message)
	failOnError(err, "Failed to serialize message", p.channel)

	err = ch.Publish(
		p.xchng,      //exchange
		p.routingKey, //routing key
		false,        //mandatory
		false,        //immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})
	failOnError(err, "Error sending message to RabbitMQ", p.channel)
}

func failOnError(err error, msg string, channel *channel) {
	if err != nil {
		channel.Close()
		log.Fatalf("%s: %s", msg, err)
	}

}
