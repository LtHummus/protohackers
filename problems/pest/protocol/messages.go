package protocol

import (
	"fmt"
)

type Message interface {
	fmt.Stringer
	Serialize() []byte
}

type Hello struct {
	Protocol string
	Version  uint32
}

func (h *Hello) String() string {
	return fmt.Sprintf("Hello{Protol: %s, Version: %d}", h.Protocol, h.Version)
}
