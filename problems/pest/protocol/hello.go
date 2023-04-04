package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

	return wrapPayload(MagicNumberHello, buf.Bytes()), nil
}

func deserializeHello(r io.Reader) (Message, error) {
	protocol, err := readString(r)
	if err != nil {
		return nil, err
	}

	var version uint32
	err = binary.Read(r, binary.BigEndian, &version)
	if err != nil {
		return nil, err
	}

	return &Hello{
		Protocol: protocol,
		Version:  version,
	}, nil
}
