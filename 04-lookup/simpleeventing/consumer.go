package simpleeventing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/iancooper/Practical-Messaging-Go/lookup/simplemessaging"
)

const retryDelay = 2 * time.Second

// EventStreamConsumer reads records from a Kafka topic and hands them to application code.
//
// It is the same four stages as the message pump -- Get, Translate, Dispatch, Handle -- and
// then the fifth thing, the one that decides everything about failure, is different:
//
//	On a queue:  acknowledge this message. The broker holds the others.
//	On a stream: commit this offset. It means "I am past everything up to here."
//
// An offset is a bookmark, not a lock. There is no per-record acknowledgement, so there is
// nothing to withhold for one record and grant for another. You are either past a point in
// the log or you are not.
//
// Which means every mechanism exercise 2 relied on is simply absent:
//
//	requeue                 -- nothing to hand back
//	requeue with delay      -- nothing holding it, so nothing to hold it longer
//	reject                  -- nothing to route it away
//	dead letter queue       -- nothing to move it there
//	redelivery count        -- nothing counting
//
// This consumer takes the default answer, which is the one most frameworks take for you:
// retry in place. Read what that does to a partition, then read PROBE.md.
type EventStreamConsumer[T simplemessaging.IAmAMessage] struct {
	mapper  func(string) (T, error)
	handler func(T) error
	topic   string
	group   string

	consumer *kafka.Consumer
}

func NewEventStreamConsumer[T simplemessaging.IAmAMessage](
	mapper func(string) (T, error), handler func(T) error,
	consumerGroup string, bootstrapServers string,
) (*EventStreamConsumer[T], error) {
	topic := TopicFor[T]()

	// Either end may create the topic, so it does not matter which you start first.
	if err := EnsureTopicExists(topic, bootstrapServers); err != nil {
		return nil, err
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          consumerGroup,
		// Start at the beginning of the log the first time this group ever reads it. A queue
		// has no equivalent of this setting, because a queue has no past.
		"auto.offset.reset": "earliest",
		// Commit when we say so. Auto-commit on a timer would move the bookmark past records
		// we have not finished with, which is the stream's version of acking early.
		"enable.auto.commit": false,
	})
	if err != nil {
		return nil, err
	}

	if err := consumer.SubscribeTopics([]string{topic}, nil); err != nil {
		consumer.Close()
		return nil, err
	}

	return &EventStreamConsumer[T]{
		mapper: mapper, handler: handler, topic: topic, group: consumerGroup, consumer: consumer,
	}, nil
}

func (c *EventStreamConsumer[T]) Run(ctx context.Context) error {
	fmt.Printf("Following %s as group '%s'\n", c.topic, c.group)

	// Leave the group tidily so the next run does not wait for a session timeout. That wait
	// is 45 seconds by default, and a consumer that joins while the broker still believes
	// the last one is alive is given no partitions and prints nothing at all.
	defer c.consumer.Close()

	for ctx.Err() == nil {
		// GET
		record, err := c.consumer.ReadMessage(time.Second)
		if err != nil {
			var kafkaError kafka.Error
			if errors.As(err, &kafkaError) && kafkaError.Code() == kafka.ErrTimedOut {
				continue
			}
			// Broker-level, not record-level: the topic is not there yet, a rebalance is in
			// progress, the broker is restarting. None of those are this record's fault,
			// because there is no record.
			fmt.Printf("  consume failed: %v -- retrying\n", err)
			sleep(ctx, retryDelay)
			continue
		}

		where := fmt.Sprintf("p%d@%d", record.TopicPartition.Partition, record.TopicPartition.Offset)

		// TRANSLATE
		message, err := c.mapper(string(record.Value))
		if err == nil {
			// DISPATCH and HANDLE
			err = c.handler(message)
		}

		if err != nil {
			fmt.Printf("  FAILED %s: %v\n", where, err)
			fmt.Println("  there is no nack, no requeue and no dead letter topic, so: retry in place")

			// Wind the bookmark back to this record and read it again. The partition stops
			// here until this record succeeds -- which, if it never can, is forever. The
			// other partitions carry on, perfectly happily.
			if _, err := c.consumer.SeekPartitions([]kafka.TopicPartition{record.TopicPartition}); err != nil {
				return err
			}
			sleep(ctx, retryDelay)
			continue
		}

		// Move the bookmark. Everything up to and including this offset is done.
		if _, err := c.consumer.CommitMessage(record); err != nil {
			return err
		}
		fmt.Printf("  committed %s\n", where)
	}

	return nil
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
