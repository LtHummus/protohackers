package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Hello struct {
	Protocol string
	Version  uint32
}

func (h *Hello) String() string {
	return fmt.Sprintf("Hello{Protol: %s, Version: %d}", h.Protocol, h.Version)
}

func (h *Hello) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.Write(serializeString(h.Protocol))
	binary.Write(&buf, binary.BigEndian, h.Version)

	return wrapPayload(0x50, buf.Bytes()), nil
}
