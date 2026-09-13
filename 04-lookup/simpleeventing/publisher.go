package simpleeventing

import "github.com/iancooper/Practical-Messaging-Go/lookup/simplemessaging"

// KafkaEventPublisher implements simplemessaging.IPublishEvents over a Kafka topic. This is
// the only place the handler's "tell the world" becomes "append to a log".
type KafkaEventPublisher[T simplemessaging.IAmAMessage] struct {
	producer *EventStreamProducer[T]
}

func NewKafkaEventPublisher[T simplemessaging.IAmAMessage](
	producer *EventStreamProducer[T],
) *KafkaEventPublisher[T] {
	return &KafkaEventPublisher[T]{producer: producer}
}

func (p *KafkaEventPublisher[T]) Publish(event T) error { return p.producer.Send(event) }

func (p *KafkaEventPublisher[T]) Close() { p.producer.Close() }
