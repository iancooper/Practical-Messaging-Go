package simplemessaging

import (
	"context"
	"fmt"
	"time"
)

const pollInterval = time.Second

// MessagePump runs Get -> Translate -> Dispatch -> Handle, in a loop, until cancelled.
//
// ---------------------------------------------------------------------------------------
//
//	THIS IS EXERCISE 1's ANSWER, AND EXERCISE 2's PROBLEM.
//
//	Everything exercise 1 asked for is here and correct:
//	  - the Translate stage is back, and it goes through a Message Mapper
//	  - the handler takes a domain type; nothing in model/ imports a broker library
//	  - the acknowledgement happens after the work, not before it
//	  - a failure no longer kills the loop
//
//	So it does not lose messages any more, and it does not fall over. It is still wrong, and
//	the way it is wrong is worse than falling over, because it will not show up in your logs
//	as a crash. Read PROBE.md.
//
// ---------------------------------------------------------------------------------------
type MessagePump[T IAmAMessage] struct {
	mapper   IAmAMessageMapper[T]
	handler  IAmAHandler[T]
	hostName string
}

func NewMessagePump[T IAmAMessage](
	mapper IAmAMessageMapper[T], handler IAmAHandler[T], hostName string,
) *MessagePump[T] {
	return &MessagePump[T]{mapper: mapper, handler: handler, hostName: hostName}
}

func (p *MessagePump[T]) Run(ctx context.Context) error {
	consumer, err := NewDataTypeChannelConsumer[T](p.hostName)
	if err != nil {
		return err
	}
	defer consumer.Close()

	fmt.Println("Pump running on", QueueNameFor[T]())

	for ctx.Err() == nil {
		// GET
		delivery, ok, err := consumer.Receive()
		if err != nil {
			return err
		}

		if !ok {
			select {
			case <-ctx.Done():
			case <-time.After(pollInterval):
			}
			continue
		}

		// TRANSLATE
		message, err := p.mapper.MapToRequest(string(delivery.Body))
		if err == nil {
			// DISPATCH and HANDLE
			err = p.handler.Handle(message)
		}

		if err != nil {
			// Something went wrong, and we must not lose the message. Put it back on the
			// queue so it gets tried again.
			fmt.Printf("  FAILED: %v -- putting it back\n", err)
			if err := consumer.Requeue(delivery.DeliveryTag); err != nil {
				return err
			}
			continue
		}

		// Only now are we done with it.
		if err := consumer.Acknowledge(delivery.DeliveryTag); err != nil {
			return err
		}
	}

	return nil
}
