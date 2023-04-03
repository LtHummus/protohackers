package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Message interface {
	fmt.Stringer
	Serialize() ([]byte, error)
}

func wrapPayload(magicNumber byte, payload []byte) []byte {
	sum := magicNumber

	length := uint32(len(payload) + 1 + 4 + 1) // payload + magic number + length + checksum
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	for _, curr := range lengthBytes {
		sum += curr
	}

	for _, curr := range payload {
		sum += curr
	}

	checksum := 0x00 - sum

	buf := bytes.Buffer{}
	buf.WriteByte(magicNumber)
	buf.Write(lengthBytes)
	buf.Write(payload)
	buf.WriteByte(checksum)

	return buf.Bytes()
}
