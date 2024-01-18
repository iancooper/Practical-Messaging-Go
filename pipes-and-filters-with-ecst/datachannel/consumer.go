package datachannel

import "log"

const invalid_exchange = "practical-messaging-invalid"

type Handler func(message interface{})
type Deserializer func(bytes []byte) (interface{}, error)

type Consumer struct {
	*channel
	deserialize Deserializer
	handle      Handler
}

// Consumer
func NewConsumer(qName string, deserializer Deserializer, handler Handler) *Consumer {
	consumer := new(Consumer)
	consumer.channel = newChannel(qName, true)
	consumer.deserialize = deserializer
	consumer.handle = handler
	return consumer
}

func (c *Consumer) Receive() {
	ch, err := c.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", c.channel)
	defer ch.Close()

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS", c.channel)

	msgs, err := ch.Consume(
		c.queueName,      // queue
		"event-consumer", // consumer
		false,            // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	failOnError(err, "Failed to receive from RabbitMQ", c.channel)

	forever := make(chan bool)

	go func(c *Consumer) {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)
			message, err := c.deserialize(msg.Body)
			if err == nil {
				c.handle(message)
				ch.Ack(msg.DeliveryTag, false)
			} else {
				ch.Nack(msg.DeliveryTag, false, false) //requeue true will push to DLQ
				log.Println("Error receiving message", err.Error())
			}

		}
	}(c)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
