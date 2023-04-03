package protocol

import (
	"encoding/binary"
	"io"
)

func serializeString(x string) []byte {
	lenHeader := binary.BigEndian.AppendUint32(nil, uint32(len(x)))
	lenHeader = append(lenHeader, x...)

	return lenHeader
}

func readString(r io.Reader) (string, error) {
	var length uint32
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}

	strBytes := make([]byte, length)
	_, err = io.ReadFull(r, strBytes)
	if err != nil {
		return "", err
	}

	return string(strBytes), nil
}
