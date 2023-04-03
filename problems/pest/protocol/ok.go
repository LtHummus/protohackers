package protocol

type Ok struct{}

func (o *Ok) String() string {
	return "Ok{}"
}

func (o *Ok) Serialize() ([]byte, error) {
	return wrapPayload(0x52, []byte{}), nil
}
