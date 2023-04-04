package protocol

import "io"

type Ok struct{}

func (o *Ok) String() string {
	return "Ok{}"
}

func (o *Ok) Serialize() ([]byte, error) {
	return wrapPayload(MagicNumberOK, []byte{}), nil
}

func deserializeOk(r io.Reader) (Message, error) {
	return &Ok{}, nil
}
