package datachannel

type Biography struct {
	Id          string
	Description string
}

func (b Biography) getId() string {
	return b.Id
}
