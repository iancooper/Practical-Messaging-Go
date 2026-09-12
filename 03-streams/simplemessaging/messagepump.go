package simplemessaging

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const pollInterval = time.Second

// maxRetries is how many times we retry before giving up. This is n.
const maxRetries = 3

// MessagePump runs Get -> Translate -> Dispatch -> Handle, in a loop, until cancelled.
//
// ---------------------------------------------------------------------------------------
//
//	THIS IS EXERCISE 2's ANSWER, AND IT IS CORRECT. Nothing here needs fixing.
//
//	  - Translate goes through a Message Mapper; the handler takes a domain type
//	  - the acknowledgement happens after the work, not before it
//	  - a body we cannot read goes to the invalid message queue, and is never retried
//	  - work that failed is retried with a delay, up to a limit, then dead-lettered
//	  - the retry count comes from RabbitMQ's x-death header; we count nothing ourselves
//
//	Every one of those five is something the broker does for you. Exercise 3 is about what
//	happens to this list when the channel is a stream instead of a queue.
//
//	There is one line in here that is now a problem, and it is not a problem with the pump.
//	Read PROBE.md.
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

		switch {
		case errors.Is(err, ErrUnmappableMessage):
			// A failure to UNDERSTAND. The bytes are not going to change, so there is
			// nothing to retry -- retrying this is the definition of a poison pill.
			// Publish it to the invalid message queue, where someone can go and look at it.
			// A reject would send it to the *retry* queue, which is the one thing this
			// message must never go to.
			fmt.Printf("  INVALID: %v\n", err)
			fmt.Printf("  -> %s\n", InvalidQueueNameFor[T]())
			if err := consumer.SendToInvalidMessageQueue(delivery); err != nil {
				return err
			}
			if err := consumer.Acknowledge(delivery.DeliveryTag); err != nil {
				return err
			}

		case err != nil:
			// A failure to PROCESS. The message was perfectly readable; the work failed.
			// That may have been bad luck, so it is worth trying again -- but not forever.
			retries := consumer.RetriesSoFar(delivery)

			if retries < maxRetries {
				fmt.Printf("  FAILED on attempt %d: %v\n", retries+1, err)
				fmt.Printf("  -> retrying in %.0fs\n", RetryDelay.Seconds())
				if err := consumer.RejectForRetry(delivery.DeliveryTag); err != nil {
					return err
				}
			} else {
				fmt.Printf("  GIVING UP after %d attempts: %v\n", retries+1, err)
				fmt.Printf("  -> %s\n", DeadLetterQueueNameFor[T]())
				if err := consumer.SendToDeadLetter(delivery); err != nil {
					return err
				}
				if err := consumer.Acknowledge(delivery.DeliveryTag); err != nil {
					return err
				}
			}

		default:
			// Only now are we done with it.
			if err := consumer.Acknowledge(delivery.DeliveryTag); err != nil {
				return err
			}
		}
	}

	return nil
}
