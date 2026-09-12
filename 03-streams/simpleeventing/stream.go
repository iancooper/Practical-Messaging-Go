package simpleeventing

import (
	"context"
	"fmt"
	"reflect"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/iancooper/Practical-Messaging-Go/streams/simplemessaging"
)

// The stream's names and shape, in one place -- and notice how much shorter this file is
// than simplemessaging/channel.go.
//
// There is one topic. There is no retry topic, no invalid record topic and no dead letter
// topic, because nothing in Kafka will move a record to one for you. If you want any of
// those, you write the producer, the consumer, the scheduler and the state yourself.
//
// Three partitions, because a partition is the unit of both ordering and parallelism: one
// consumer in a group holds a partition at a time, records within a partition are ordered,
// and records in different partitions are not ordered relative to each other at all.
const (
	// BootstrapServers is one of the two places a broker's address appears. The other is
	// simplemessaging.
	BootstrapServers = "localhost:9092"

	ConsumerGroup = "practical-messaging-streams"
	Partitions    = 3
)

// TopicFor derives the topic from the type, the way the queue's routing key is derived from
// it -- except that this one uses the bare type name, so the topic is streams.OrderPlaced in
// every language this course ships in.
func TopicFor[T simplemessaging.IAmAMessage]() string {
	return "streams." + reflect.TypeFor[T]().Name()
}

// EnsureTopicExists creates the topic if it is not there.
//
// Both the producer and the consumer call this, so it does not matter which you start first.
// (Compare RabbitMQ, where only the consumer declares the queue -- so anything published
// before the consumer's first ever run went nowhere. Kafka's topic is shared state that
// either end can create, which is a small but real difference in how the two feel to
// operate.)
//
// It is here rather than left to the broker's auto-create so that the partition count is
// ours to choose, and so the exercises do not depend on a broker setting.
func EnsureTopicExists(topic string, bootstrapServers string) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	if err != nil {
		return err
	}
	defer admin.Close()

	results, err := admin.CreateTopics(context.Background(), []kafka.TopicSpecification{{
		Topic:             topic,
		NumPartitions:     Partitions,
		ReplicationFactor: 1,
	}})
	if err != nil {
		return err
	}

	for _, result := range results {
		switch result.Error.Code() {
		case kafka.ErrNoError:
			fmt.Printf("Created topic %s with %d partitions\n", topic, Partitions)
		case kafka.ErrTopicAlreadyExists:
			// Somebody got there first, which is the normal case after the first run.
		default:
			return fmt.Errorf("could not create topic %s: %w", topic, result.Error)
		}
	}

	return nil
}
