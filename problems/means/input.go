package means

type inputMessage struct {
	Kind uint8
	A    int32
	B    int32
}

func (im *inputMessage) IsValid() bool {
	return im.Kind == 'I' || im.Kind == 'Q'
}

func (im *inputMessage) IsInsert() bool {
	return im.Kind == 'I'
}

func (im *inputMessage) Timestamp() int32 {
	return im.A
}

func (im *inputMessage) Price() int32 {
	return im.B
}

func (im *inputMessage) IsQuery() bool {
	return im.Kind == 'Q'
}

func (im *inputMessage) MinTime() int32 {
	return im.A
}

func (im *inputMessage) MaxTime() int32 {
	return im.B
}
