package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type DialAuthority struct {
	Site uint32
}

func (d *DialAuthority) String() string {
	return fmt.Sprintf("DialAuthority{Site: %d}", d.Site)
}

func (d *DialAuthority) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, d)

	return wrapPayload(0x53, buf.Bytes()), nil
}
