package protocol

import (
	"encoding/binary"
	"fmt"
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

	return wrapPayload(0x57, b), nil
}
