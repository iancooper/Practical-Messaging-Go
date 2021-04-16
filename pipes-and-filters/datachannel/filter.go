package datachannel

type filter struct {
	deserialize Deserializer
	serialize   Serializer
}

type Transform func(msg interface{}) interface{}

func NewFilter(deserialiser Deserializer, serializer Serializer) *filter {
	filter := new(filter)
	filter.deserialize = deserialiser
	filter.serialize = serializer
	return filter
}

func (f *filter) Run(transform Transform) {
	producer := NewProducer("sink-p2p", f.serialize)
	defer producer.Close()

	msgs := make(chan interface{})
	consumer := NewConsumer("source-p2p", f.deserialize, func(message interface{}) {
		msgs <- message
	})
	defer consumer.Close()

	go func(p *Producer) {
		for msg := range msgs {
			newMsg := transform(msg)
			p.Send(newMsg)
		}
	}(producer)

	consumer.Receive()
}
