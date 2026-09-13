package simpleeventing

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/iancooper/Practical-Messaging-Go/lookup/simplemessaging"
)

// EventStreamProducer appends records to a Kafka topic. The producer half of the eventing
// gateway.
//
// Note what is *not* here, compared with the RabbitMQ producer: no exchange, no binding, no
// routing key. You append to a log, and the key decides which partition the record lands in.
// Same key, same partition, so records that must stay in order must share a key.
type EventStreamProducer[T simplemessaging.IAmAMessage] struct {
	serialize    func(T) (string, error)
	partitionKey func(T) string
	producer     *kafka.Producer
	topic        string
}

// NewEventStreamProducer creates the topic if it is not there and connects a producer.
//
// partitionKey decides which partition a record belongs in -- and therefore what it is
// ordered with respect to. Records sharing a key share a partition and stay in order;
// records with different keys have no order between them at all.
//
// This is a design decision and there is no safe default, which is why you have to pass it.
// Key an order's events by the order and they arrive in sequence. Key them by the event's
// own id and every event is independent -- which is fine right up until two events about the
// same thing are processed out of order by different consumers.
func NewEventStreamProducer[T simplemessaging.IAmAMessage](
	serialize func(T) (string, error), partitionKey func(T) string, bootstrapServers string,
) (*EventStreamProducer[T], error) {
	topic := TopicFor[T]()
	if err := EnsureTopicExists(topic, bootstrapServers); err != nil {
		return nil, err
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		// Wait for the leader and all in-sync replicas before calling a write done. This is
		// the producer-side half of guaranteed delivery, and it is the cheap half -- exactly
		// as it was on RabbitMQ, where it was one 'persistent' flag.
		"acks":               "all",
		"enable.idempotence": true,
	})
	if err != nil {
		return nil, err
	}

	return &EventStreamProducer[T]{
		serialize: serialize, partitionKey: partitionKey, producer: producer, topic: topic,
	}, nil
}

// Send appends a record and waits until the broker has acknowledged it.
//
// We wait on the delivery channel rather than firing and forgetting: fire-and-forget would
// return before the record was durable, and then "I produced the event" would be a claim
// about a buffer in this process rather than about anything the broker has.
func (p *EventStreamProducer[T]) Send(message T) error {
	body, err := p.serialize(message)
	if err != nil {
		return err
	}
	return p.append(p.partitionKey(message), body)
}

// SendRaw appends a record we did not serialize -- used to put something unreadable on the
// stream.
func (p *EventStreamProducer[T]) SendRaw(key string, body string) error {
	return p.append(key, body)
}

func (p *EventStreamProducer[T]) append(key string, body string) error {
	delivered := make(chan kafka.Event, 1)

	err := p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(body),
	}, delivered)
	if err != nil {
		return err
	}

	message := (<-delivered).(*kafka.Message)
	if message.TopicPartition.Error != nil {
		return message.TopicPartition.Error
	}

	fmt.Printf("  -> %s partition %d offset %d\n",
		*message.TopicPartition.Topic, message.TopicPartition.Partition, message.TopicPartition.Offset)
	return nil
}

func (p *EventStreamProducer[T]) Close() {
	p.producer.Flush(int((5 * time.Second).Milliseconds()))
	p.producer.Close()
}
