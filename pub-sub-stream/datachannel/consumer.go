package datachannel

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Handler func(message interface{}) error
type Deserializer func(bytes []byte) (interface{}, error)

type Consumer struct {
	deserialize Deserializer
	handle      Handler
	topic       string
}

// Consumer
func NewConsumer(topic string, deserializer Deserializer, handler Handler) *Consumer {
	consumer := new(Consumer)
	consumer.deserialize = deserializer
	consumer.handle = handler
	consumer.topic = topic
	return consumer
}

func (c *Consumer) Receive() {
	configMap := &kafka.ConfigMap{
		//TODO Configure the consumer
		// bootstrap.servers
		// group.id
		// auto.offset.reset should be earliest
		// enable.auto.offset.store should be false
	}

	// TODO Create the Kafka Consumer instance
	// Suscibe the consumer to the topic

	// defer consumer.Close()

	for {
		// TODO Read the next message from the topic
		if err == nil {
			message, err := c.deserialize(msg.Value)
			if err == nil {
				err := c.handle(message)
				if err == nil {
					// TODO Store the message (offset)
					if err != nil {
						fmt.Printf("Error storing message: %v\n", err)
					} else {
						fmt.Printf("Stored message in partition %v\n", partition)
					}
				} else {
					fmt.Printf("Error handling message: %v\n", err)
				}
			} else {
				fmt.Println("Error receiving message", err.Error())
			}
		} else {
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}
