package datachannel

import (
	"github.com/streadway/amqp"
)

type Serializer func(message interface{}) ([]byte, error)

type Producer struct {
	*Channel
	serialize Serializer
}

//Producer
func NewProducer(qName string, serializer Serializer) *Producer {
	producer := new(Producer)
	producer.Channel = newChannel(qName, false)
	producer.serialize = serializer
	return producer
}

func (p *Producer) Send(message interface{}) {
	ch, err := p.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", p.Channel)
	defer ch.Close()

	b, err := p.serialize(message)
	failOnError(err, "Failed to serialize message", p.Channel)

	err = ch.Publish(
		p.xchng,      //exchange
		p.routingKey, //routing key
		false,        //mandatory
		false,        //immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})
	failOnError(err, "Error sending message to RabbitMQ", p.Channel)
}
