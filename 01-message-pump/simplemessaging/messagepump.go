package simplemessaging

import (
	"context"
	"fmt"
	"time"
)

const pollInterval = time.Second

// MessagePump takes a message off a channel, gets it to application code, and repeats
// until cancelled.
//
//	Get -> Translate -> Dispatch -> Handle
//
// Each of those four stages fails in its own way, which is why a message that is never
// going to be handled has four different places it can end up.
//
// ---------------------------------------------------------------------------------------
//
//	THIS PUMP IS THE EXERCISE.
//
//	It compiles, it runs, and messages flow through it. It is also wrong, in more than one
//	way, and every way it is wrong is something that has shipped to production somewhere.
//
//	Read it before you run it. Then read PROBE.md. Do not copy this file into anything.
//
// ---------------------------------------------------------------------------------------
type MessagePump[T IAmAMessage] struct {
	handler  IAmAHandler[T]
	hostName string
}

func NewMessagePump[T IAmAMessage](handler IAmAHandler[T], hostName string) *MessagePump[T] {
	return &MessagePump[T]{handler: handler, hostName: hostName}
}

// Run pumps until ctx is cancelled. Cancelling is the polite ending: it stops the loop
// between messages rather than in the middle of one.
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
			// Nothing there. Yield, so a Polling Consumer does not spin the CPU.
			select {
			case <-ctx.Done():
			case <-time.After(pollInterval):
			}
			continue
		}

		fmt.Println("Got delivery", delivery.DeliveryTag)

		// We have the message in our hands, so the broker does not need to hold it for us
		// any more. Tell it we are done and let it free the slot.
		if err := consumer.Acknowledge(delivery.DeliveryTag); err != nil {
			return err
		}

		// TRANSLATE, DISPATCH and HANDLE
		if err := p.handler.Handle(delivery); err != nil {
			return err
		}
	}

	return nil
}
