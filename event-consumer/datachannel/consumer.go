package datachannel

import "log"

const invalid_exchange = "practical-messaging-invalid"

type Handler func(message interface{})
type Deserializer func(bytes []byte) (interface{}, error)

type Consumer struct {
	*Channel
	deserialize Deserializer
	handle      Handler
}

//Consumer
func NewConsumer(qName string, deserializer Deserializer, handler Handler) *Consumer {
	consumer := new(Consumer)
	consumer.Channel = newChannel(qName, true)
	consumer.deserialize = deserializer
	consumer.handle = handler
	return consumer
}

func (c *Consumer) Receive() {
	ch, err := c.conn.Channel()
	failOnError(err, "Failed to connect to RabbitMQ", c.Channel)
	defer ch.Close()

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS", c.Channel)

	//TODO: Consume the queue - will give you back a channel
	//Note that we move on from Get which polls to Consume here

	forever := make(chan bool)

	//TODO:
	//For each message on the channel
	//Deserialize (translate) the nessage
	//if it can be deserialized
		//Dispatch to the handler
		//Ack the message
	//else
		//Nack the message

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
