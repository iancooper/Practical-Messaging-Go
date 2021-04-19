package datachannel

type Filter struct {
	deserialize Deserializer
	serialize   Serializer
}

type Transform func(msg interface{}) interface{}

func NewFilter(deserialiser Deserializer, serializer Serializer) *Filter {
	filter := new(Filter)
	filter.deserialize = deserialiser
	filter.serialize = serializer
	return filter
}

func (f *Filter) Run(transform Transform) {
	producer := NewProducer("sink-p2p", f.serialize)
	defer producer.Close()

	//TODO: Add a consumer, and put messages consumed onto a channel msgs in the handler

	//TODO:
	//For each message in msgs
	//Transform the message
	//Send the message on

	consumer.Receive()
}
