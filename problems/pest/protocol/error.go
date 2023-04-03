package protocol

import (
	"fmt"
)

type Error struct {
	Message string
}

func (e *Error) String() string {
	return fmt.Sprintf("Error{Message: %s}", e.Message)
}

func (e *Error) Serialize() ([]byte, error) {
	return wrapPayload(0x51, serializeString(e.Message)), nil
}
