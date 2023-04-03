package protocol

import (
	"bytes"
	"fmt"
)

type PolicyAction byte

const (
	Cull     PolicyAction = 0x90
	Conserve PolicyAction = 0xa0
)

type CreatePolicy struct {
	Species string
	Action  PolicyAction
}

func (c *CreatePolicy) String() string {
	var actionName string
	if c.Action == Cull {
		actionName = "cull"
	} else if c.Action == Conserve {
		actionName = "conserve"
	} else {
		actionName = "????"
	}

	return fmt.Sprintf("CreatePolicy{Species: %s, Action: %s}", c.Species, actionName)
}

func (c *CreatePolicy) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.Write(serializeString(c.Species))
	buf.WriteByte(byte(c.Action))

	return wrapPayload(0x55, buf.Bytes()), nil
}
