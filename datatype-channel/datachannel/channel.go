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

type producer struct {
	*channel
	serialize Serializer
}

type consumer struct {
	*channel
	deserialize Deserializer
}
type Serializer func(message interface{}) ([]byte, error)
type Deserializer func(bytes []byte) (interface{}, error)

//We just use a contant here for convenience, in reality you configure this
const exchange = "practical-messaging-go"

func NewProducer(qName string, serializer Serializer) *producer {
	producer := new(producer)
	producer.channel = newChannel(qName)
	producer.serialize = serializer
	return producer
}

func NewConsumer(qName string, deserializer Deserializer) *consumer {
	consumer := new(consumer)
	consumer.channel = newChannel(qName)
	consumer.deserialize = deserializer
	return consumer
}

func newChannel(qName string) *channel {

	channel := new(channel)
	channel.xchng = exchange
	channel.queueName = qName
	channel.routingKey = qName

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ", channel)
	channel.conn = conn

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel", channel)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange, 			// name
		"direct", 		// type
		false,    	// durable
		false,    // auto-deleted
		false,    	// internal
		false,    	// no-wait
		nil,     		// arguments
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

//Channel
func (channel *channel) Close() {
	if channel.conn != nil {
		channel.conn.Close()
	}
}

//Consumer
func (c *consumer) Receive() (bool, interface{}) {
	ch, err := c.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", c.channel)
	defer ch.Close()

	msg, ok, err := ch.Get(
		c.queueName, 		 //queue name
		true,        //auto ack when we read
	)
	failOnError(err, "Failed to receive from RabbitMQ", c.channel)

	if ok {
		message, err := c.deserialize(msg.Body)
		if err == nil {
			return true, message
		} else{
			log.Println("Error receiving message", err.Error())
		}

	}
	return false, nil
}

//Producer
func (p *producer) Send(message interface{}) {
	ch, err := p.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", p.channel)
	defer ch.Close()

	b, err := p.serialize(message)
	failOnError(err, "Failed to serialize message", p.channel)

	err = ch.Publish(
		p.xchng,      			//exchange
		p.routingKey, 			//routing key
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
