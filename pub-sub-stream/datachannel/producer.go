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

	p, err := kafka.NewProducer(configMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kafka producer: %v\n", err)
	}

	producer.producer = p
	return producer
}

func (p *Producer) produceMessage(topic, key, value string) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
		Timestamp:      time.Now(),
	}

	err := p.producer.Produce(message, nil)
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) Send(message Record) {
	b, err := p.serialize(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing message: %v\n", err)
	}

	err = p.produceMessage("biographies", message.getId(), b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error producing message: %v\n", err)
	} else {
		fmt.Printf("Produced message for %s\n", message.getId())
	}
}

func (p *Producer) Close() {
	p.producer.Flush(115 * 1000)
	p.producer.Close()
}
