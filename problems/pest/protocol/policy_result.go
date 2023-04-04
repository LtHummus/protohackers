package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

type PolicyResult struct {
	Policy uint32
}

func (p *PolicyResult) String() string {
	return fmt.Sprintf("PolicyResult{Policy: %d}", p.Policy)
}

func (p *PolicyResult) Serialize() ([]byte, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, p.Policy)

	return wrapPayload(MagicNumberPolicyResult, b), nil
}

func deserializePolicyResult(r io.Reader) (Message, error) {
	var policy uint32
	err := binary.Read(r, binary.BigEndian, &policy)
	if err != nil {
		return nil, err
	}

	return &PolicyResult{
		Policy: policy,
	}, nil
}
