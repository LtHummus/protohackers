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

type deserializationFunc func(header messageHeader, r io.Reader) (Message, error)

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

	log.Debug().Uint8("magic_number", header.MagicNumber).Uint32("length", header.Length).Msg("read header")

	f := deserializationMap[header.MagicNumber]
	if f == nil {
		return nil, errors.New("invalid magic number")
	}

	return f(header, r)
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

func deserializeHello(header messageHeader, r io.Reader) (Message, error) {
	packetWrapper, err := makeWrapperAndComputeChecksum(header, r)
	if err != nil {
		return nil, err
	}

	protocol, err := readString(packetWrapper)
	if err != nil {
		return nil, err
	}

	var version uint32
	err = binary.Read(packetWrapper, binary.BigEndian, &version)
	if err != nil {
		return nil, err
	}

	return &Hello{
		Protocol: protocol,
		Version:  version,
	}, nil
}

func deserializeError(header messageHeader, r io.Reader) (Message, error) {
	packetWrapper, err := makeWrapperAndComputeChecksum(header, r)
	if err != nil {
		return nil, err
	}

	message, err := readString(packetWrapper)
	if err != nil {
		return nil, err
	}

	return &Error{
		Message: message,
	}, nil
}

func deserializeOk(header messageHeader, r io.Reader) (Message, error) {
	_, err := makeWrapperAndComputeChecksum(header, r)
	if err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

func deserializeDialAuthority(header messageHeader, r io.Reader) (Message, error) {
	w, err := makeWrapperAndComputeChecksum(header, r)
	if err != nil {
		return nil, err
	}

	var site uint32
	err = binary.Read(w, binary.BigEndian, &site)
	if err != nil {
		return nil, err
	}

	return &DialAuthority{
		Site: site,
	}, nil

}

func deserializeTargetPopulations(header messageHeader, r io.Reader) (Message, error) {
	w, err := makeWrapperAndComputeChecksum(header, r)
	if err != nil {
		return nil, err
	}

	var site uint32
	err = binary.Read(w, binary.BigEndian, &site)
	if err != nil {
		return nil, err
	}

	var ruleCount uint32
	err = binary.Read(w, binary.BigEndian, &ruleCount)

	targets := make([]PopulationTarget, ruleCount)
	for i := range targets {
		species, err := readString(w)
		if err != nil {
			return nil, err
		}
		var minMax struct {
			Min uint32
			Max uint32
		}
		err = binary.Read(w, binary.BigEndian, &minMax)
		if err != nil {
			return nil, err
		}
		targets[i] = PopulationTarget{
			Species: species,
			Min:     minMax.Min,
			Max:     minMax.Max,
		}
	}

	return &TargetPopulations{
		Site:    site,
		Targets: targets,
	}, nil
}
