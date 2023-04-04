package protocol

import (
	"fmt"
	"io"
)

type Error struct {
	Message string
}

func (e *Error) String() string {
	return fmt.Sprintf("Error{Message: %s}", e.Message)
}

func (e *Error) Serialize() ([]byte, error) {
	return wrapPayload(MagicNumberError, serializeString(e.Message)), nil
}

func deserializeError(r io.Reader) (Message, error) {
	message, err := readString(r)
	if err != nil {
		return nil, err
	}

	return &Error{
		Message: message,
	}, nil
}
