package datachannel

import (
	"github.com/streadway/amqp"
)

type Serializer func(message interface{}) ([]byte, error)

type Producer struct {
	*channel
	serialize Serializer
}

//Producer
func NewProducer(qName string, serializer Serializer) *Producer {
	producer := new(Producer)
	producer.channel = newChannel(qName, false)
	producer.serialize = serializer
	return producer
}

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
