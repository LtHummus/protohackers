package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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

	return wrapPayload(MagicNumberCreatePolicy, buf.Bytes()), nil
}

func deserializeCreatePolicy(r io.Reader) (Message, error) {
	species, err := readString(r)
	if err != nil {
		return nil, err
	}

	var action PolicyAction
	err = binary.Read(r, binary.BigEndian, &action)
	if err != nil {
		return nil, err
	}

	if action != Cull && action != Conserve {
		return nil, errors.New("invalid policy action")
	}

	return &CreatePolicy{
		Species: species,
		Action:  action,
	}, nil
}
