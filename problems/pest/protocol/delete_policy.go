package protocol

import (
	"encoding/binary"
	"fmt"
)

type DeletePolicy struct {
	Policy uint32
}

func (d *DeletePolicy) String() string {
	return fmt.Sprintf("DeletePolicy{Policy: %d}", d.Policy)
}

func (d *DeletePolicy) Serialize() ([]byte, error) {
	p := make([]byte, 4)
	binary.BigEndian.PutUint32(p, d.Policy)

	return wrapPayload(0x56, p), nil
}
