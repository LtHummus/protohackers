package protocol

import "encoding/binary"

func serializeString(x string) []byte {
	lenHeader := binary.BigEndian.AppendUint32(nil, uint32(len(x)))
	lenHeader = append(lenHeader, x...)

	return lenHeader
}
