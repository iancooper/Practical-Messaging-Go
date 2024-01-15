package datachannel

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"os/signal"
	"syscall"
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
		"bootstrap.servers":        kafkaBrokers,
		"group.id":                 "SimpleEventing",
		"session.timeout.ms":       6000,
		"auto.offset.reset":        "earliest",
		"enable.auto.offset.store": false,
	}

	consumer, err := kafka.NewConsumer(configMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kafka consumer: %v\n", err)
	}

	err = consumer.SubscribeTopics([]string{c.topic}, nil)

	defer consumer.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			msg, err := consumer.ReadMessage(-1)
			if err == nil {
				message, err := c.deserialize(msg.Value)
				if err == nil {
					err := c.handle(message)
					if err == nil {
						partition, err := consumer.StoreMessage(msg)
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

}
