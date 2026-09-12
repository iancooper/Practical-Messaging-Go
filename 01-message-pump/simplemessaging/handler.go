package simplemessaging

import amqp091 "github.com/rabbitmq/amqp091-go"

// IAmAHandler is the contract your application code implements so the pump can dispatch to it.
type IAmAHandler[T IAmAMessage] interface {
	Handle(delivery amqp091.Delivery) error
}
