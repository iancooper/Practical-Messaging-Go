package simplemessaging

// IAmAHandler is the contract your application code implements so the pump can dispatch to it.
//
// A domain type in, and an error out. No delivery, no channel, no headers, no ack. The
// handler does not know the pump exists, which is exactly why a test can call it too.
type IAmAHandler[T IAmAMessage] interface {
	Handle(message T) error
}
