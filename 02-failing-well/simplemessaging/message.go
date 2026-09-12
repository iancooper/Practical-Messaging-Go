package simplemessaging

// IAmAMessage is anything we are willing to put on a channel.
//
// The Id is the message's identity, not the entity's -- we need it to key a stream, and
// later to answer "have I seen this before?".
//
// Go would normally call this interface Message, and would normally not put an I on the
// front of anything. We keep the course's names so that the C#, Java and Go versions of
// these exercises are talking about the same things; it is the only place we do that.
type IAmAMessage interface {
	ID() string
}
