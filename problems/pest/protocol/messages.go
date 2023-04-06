package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
)

type Message interface {
	fmt.Stringer
	Serialize() ([]byte, error)
}

type deserializationFunc func(r io.Reader) (Message, error)

type messageHeader struct {
	MagicNumber uint8
	Length      uint32
}

func (mh *messageHeader) Sum() uint8 {
	n := mh.MagicNumber
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, mh.Length)
	for _, c := range b {
		n += c
	}
	return n
}

const (
	MaxMessageLength = 1 * 1024 * 1024 // 1 megabyte
)

const (
	MagicNumberHello             uint8 = 0x50
	MagicNumberError             uint8 = 0x51
	MagicNumberOK                uint8 = 0x52
	MagicNumberDialAuthority     uint8 = 0x53
	MagicNumberTargetPopulations uint8 = 0x54
	MagicNumberCreatePolicy      uint8 = 0x55
	MagicNumberDeletePolicy      uint8 = 0x56
	MagicNumberPolicyResult      uint8 = 0x57
	MagicNumberSiteVisit         uint8 = 0x58
)

var deserializationMap = map[uint8]deserializationFunc{
	MagicNumberHello:             deserializeHello,
	MagicNumberError:             deserializeError,
	MagicNumberOK:                deserializeOk,
	MagicNumberDialAuthority:     deserializeDialAuthority,
	MagicNumberTargetPopulations: deserializeTargetPopulations,
	MagicNumberCreatePolicy:      deserializeCreatePolicy,
	MagicNumberDeletePolicy:      deserializeDeletePolicy,
	MagicNumberPolicyResult:      deserializePolicyResult,
	MagicNumberSiteVisit:         deserializeSiteVisit,
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

func Deserialize(r io.Reader) (Message, error) {
	var header messageHeader
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	if header.Length > MaxMessageLength {
		log.Warn().Uint32("length", header.Length).Msg("message is too long")
		return nil, errors.New("message too long")
	}

	payloadReader, err := makeWrapperAndComputeChecksum(header, r)
	if err != nil {
		return nil, err
	}

	f := deserializationMap[header.MagicNumber]
	if f == nil {
		return nil, errors.New("invalid magic number")
	}

	m, err := f(payloadReader)
	if err != nil {
		log.Error().Err(err).Msg("error decoding message")
		return nil, err
	}

	return m, nil
}

func makeWrapperAndComputeChecksum(header messageHeader, r io.Reader) (*bytes.Reader, error) {
	buf := make([]byte, header.Length-5)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	s := header.Sum()
	for _, c := range buf {
		s += c
	}

	if s != 0x00 {
		log.Warn().Uint8("computed_checksum", s).Msg("invalid checksum")
		return nil, errors.New("invalid checksum")
	}

	return bytes.NewReader(buf), nil
}
