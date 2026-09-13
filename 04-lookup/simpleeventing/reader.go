package simpleeventing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/iancooper/Practical-Messaging-Go/lookup/simplemessaging"
)

// StreamRecord is a record read off the stream, and where it was. The position is ours, not
// yours: Where is a string for printing, and the two kafka.TopicPartition fields are
// unexported, so nothing outside this package can name a Kafka type.
type StreamRecord[T simplemessaging.IAmAMessage] struct {
	// Message is the translated record.
	Message T

	// Where is "p1@42" -- the partition and offset, for printing.
	Where string

	// position is where this record is, for Seek.
	position kafka.TopicPartition

	// next is where the NEXT record is, which is what a commit actually means.
	//
	// A COMMITTED OFFSET IS "THE NEXT ONE I HAVE NOT READ", NOT "THE LAST ONE I DID READ".
	// Commit this record's own offset and you have told the broker to start here again, so
	// every restart replays the record you just finished -- and a fully drained group sits
	// at LAG 1 per partition for ever. That looks exactly like the duplicate Probe D is
	// about, and is not it. Off by one, and the symptom is somebody else's bug.
	//
	// confluent-kafka-go gives you both halves of this and only one of them adds the one:
	// Consumer.CommitMessage(msg) does `offsets[0].Offset++` for you, and
	// Consumer.CommitOffsets(offsets) commits exactly what you hand it. Commit below uses
	// the second, so the +1 is here, once, where it can be read.
	next kafka.TopicPartition
}

// EventStreamReader is the same gateway as EventStreamConsumer, with one difference that is
// the whole reason it exists: THE CALLER COMMITS.
//
// EventStreamConsumer does Get, Translate, Dispatch, Handle and then commits for you, which
// is the right shape when the ordering is not what you are studying. In exercise 4 the
// ordering *is* what you are studying -- Probe D is "apply the record, then commit the
// offset, and die in between" -- so the commit has to be a line in the application that you
// can move.
//
// Notice that this is a *gateway* decision, not an application one. The application still
// names no Kafka type: it gets a StreamRecord, and it says Commit or Seek.
type EventStreamReader[T simplemessaging.IAmAMessage] struct {
	mapper func(string) (T, error)
	topic  string
	group  string

	consumer *kafka.Consumer
}

func NewEventStreamReader[T simplemessaging.IAmAMessage](
	mapper func(string) (T, error), consumerGroup string, bootstrapServers string,
) (*EventStreamReader[T], error) {
	topic := TopicFor[T]()

	// Either end may create the topic, so it does not matter which you start first.
	if err := EnsureTopicExists(topic, bootstrapServers); err != nil {
		return nil, err
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          consumerGroup,
		// The local copy is built by replaying the whole log, which is the thing a stream
		// can do and a queue cannot. A new consumer with an empty database reads from the
		// start and catches up; that is Archive and Replay from exercise 3, earning its
		// keep.
		"auto.offset.reset": "earliest",
		// Commit when we say so -- which here means when the application says so.
		"enable.auto.commit": false,
	})
	if err != nil {
		return nil, err
	}

	if err := consumer.SubscribeTopics([]string{topic}, nil); err != nil {
		consumer.Close()
		return nil, err
	}

	return &EventStreamReader[T]{
		mapper: mapper, topic: topic, group: consumerGroup, consumer: consumer,
	}, nil
}

func (r *EventStreamReader[T]) Topic() string { return r.topic }

func (r *EventStreamReader[T]) Group() string { return r.group }

// Read is Get and Translate. It returns a nil record and a nil error when there was nothing
// to read, or when the broker was unhappy in a way that is not this record's fault -- a
// rebalance, a topic that does not exist yet. Both are normal and the caller should just
// come round again.
//
// An error means the record could not be mapped, and it is fatal. There is no invalid record
// topic on a stream and nothing will move it to one for you, which is exercise 3's finding
// and not this exercise's problem to solve -- so the caller stops loudly rather than
// skipping quietly.
func (r *EventStreamReader[T]) Read(ctx context.Context) (*StreamRecord[T], error) {
	record, err := r.consumer.ReadMessage(time.Second)
	if err != nil {
		var kafkaError kafka.Error
		if errors.As(err, &kafkaError) && kafkaError.Code() == kafka.ErrTimedOut {
			return nil, nil
		}
		fmt.Printf("  consume failed: %v -- retrying\n", err)
		sleep(ctx, retryDelay)
		return nil, nil
	}

	where := fmt.Sprintf("p%d@%d", record.TopicPartition.Partition, record.TopicPartition.Offset)

	message, err := r.mapper(string(record.Value))
	if err != nil {
		return nil, fmt.Errorf(
			"cannot read the record at %s: %w. There is no invalid record topic on a "+
				"stream -- see exercise 3", where, err)
	}

	next := record.TopicPartition
	next.Offset++ // the one that CommitOffsets will not add for you

	return &StreamRecord[T]{
		Message:  message,
		Where:    where,
		position: record.TopicPartition,
		next:     next,
	}, nil
}

// Commit moves the bookmark. Everything up to and including this record is done.
func (r *EventStreamReader[T]) Commit(record *StreamRecord[T]) error {
	_, err := r.consumer.CommitOffsets([]kafka.TopicPartition{record.next})
	return err
}

// Seek winds the bookmark back to this record and reads it again.
func (r *EventStreamReader[T]) Seek(record *StreamRecord[T]) error {
	_, err := r.consumer.SeekPartitions([]kafka.TopicPartition{record.position})
	return err
}

// Close leaves the group tidily so the next run does not wait for a session timeout. That
// wait is 45 seconds by default, and a consumer that joins while the broker still believes
// the last one is alive is given no partitions and prints nothing at all.
func (r *EventStreamReader[T]) Close() { r.consumer.Close() }
