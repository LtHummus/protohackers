package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
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

	return wrapPayload(MagicNumberDeletePolicy, p), nil
}

func deserializeDeletePolicy(r io.Reader) (Message, error) {
	var policy uint32
	err := binary.Read(r, binary.BigEndian, &policy)
	if err != nil {
		return nil, err
	}

	return &DeletePolicy{
		Policy: policy,
	}, nil
}
