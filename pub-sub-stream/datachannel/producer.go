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
}

func NewProducer(topic string, serializer func(message Record) (string, error)) *Producer {
	producer := new(Producer)
	producer.topic = topic
	producer.serialize = serializer
	return producer
}

func produceMessage(producer *kafka.Producer, topic, key, value string) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
		Timestamp:      time.Now(),
	}

	err := producer.Produce(message, nil)
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) Send(message Record) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": kafkaBrokers,
	}

	producer, err := kafka.NewProducer(configMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kafka producer: %v\n", err)
	}

	defer producer.Close()

	b, err := p.serialize(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing message: %v\n", err)
	}

	err = produceMessage(producer, "biographies", message.getId(), b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error producing message: %v\n", err)
	} else {
		fmt.Printf("Produced message for %s\n", message.getId())
	}
}
