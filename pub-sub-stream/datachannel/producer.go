package datachannel

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"time"
)

type Serializer func(message Record) (string, error)

const kafkaBrokers = "localhost:9092" // Update with your Kafka broker(s)

type Producer struct {
	topic     string
	serialize Serializer
	producer  *kafka.Producer
}

func NewProducer(topic string, serializer func(message Record) (string, error)) *Producer {

	producer := new(Producer)
	producer.topic = topic
	producer.serialize = serializer

	configMap := &kafka.ConfigMap{
		"bootstrap.servers": kafkaBrokers,
	}

	// TODO Create the Kafka Producer instance

	return producer
}

func (p *Producer) produceMessage(topic, key, value string) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
		Timestamp:      time.Now(),
	}

	deliveryChan := make(chan kafka.Event)

	// TODO Produce the message to the topic

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	close(deliveryChan)

	return nil
}

func (p *Producer) Send(message Record) {
	b, err := p.serialize(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing message: %v\n", err)
	}

	err = p.produceMessage(p.topic, message.getId(), b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error producing message: %v\n", err)
	}
}

func (p *Producer) Close() {
	// TODO Flush and close the producer
}
