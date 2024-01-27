package main

import (
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var sampleBiographies = map[string]string{
	"Clarissa Harlow":   "A young woman whose quest for virtue is continually thwarted by her family.",
	"Pamela Andrews":    "A young woman whose virtue is rewarded.",
	"Harriet Byron":     "An orphan, and heir to a considerable fortune of fifteen thousand pounds.",
	"Charles Grandison": "A man of feeling who truly cannot be said to feel.",
}

const kafkaBrokers = "localhost:9092" // Update with your Kafka broker(s)

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

func main() {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": kafkaBrokers,
	}

	producer, err := kafka.NewProducer(configMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kafka producer: %v\n", err)
		os.Exit(1)
	}

	defer producer.Close()

	for name, biography := range sampleBiographies {
		err := produceMessage(producer, "biographies", name, biography)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error producing message: %v\n", err)
		} else {
			fmt.Printf("Produced message for %s\n", name)
		}
	}

	// Wait for any outstanding messages to be delivered and delivery reports received
	i := producer.Flush(15 * 1000)
	for i > 0 {
		fmt.Printf("Waiting for %d deliveries\n", i)
		time.Sleep(1 * time.Second)
		i = producer.Flush(15 * 1000)
	}

}
