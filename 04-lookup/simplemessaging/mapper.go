package simplemessaging

import "errors"

// IAmAMessageMapper is the Translate stage: a message body on the wire becomes a domain object.
//
// This is the seam. Everything on the broker side of it is the gateway's business;
// everything on the other side is yours.
//
// It fails when the body is not the type this channel carries, and it says so by returning
// an error that wraps ErrUnmappableMessage. That is a *different kind of failure* from one
// returned by a handler, and the pump has to treat it differently -- which is most of this
// exercise.
type IAmAMessageMapper[T IAmAMessage] interface {
	// MapToRequest returns an error wrapping ErrUnmappableMessage when the body is not a T
	// and never will be.
	MapToRequest(body string) (T, error)
}

// ErrUnmappableMessage means "I was handed this message and I cannot read it."
//
// The gateway defines it so the pump does not have to know whether the body was JSON, XML
// or protobuf. Retrying it is pointless: the bytes will not change.
//
// Go has no exceptions, so there is no type to catch. What there is instead is a sentinel
// error and errors.Is, and the distinction the exercise is about survives the change of
// mechanism intact: the pump still asks "which kind of failure is this?" and still gets a
// yes or a no. Wrap this with %w and errors.Is will find it however deeply it is buried.
var ErrUnmappableMessage = errors.New("unmappable message")
