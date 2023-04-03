package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type TargetPopulations struct {
	Site    uint32
	Targets []PopulationTarget
}

type PopulationTarget struct {
	Species string
	Min     uint32
	Max     uint32
}

func (t *TargetPopulations) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("TargetPopulations{Site: %d, Populations: [", t.Site))
	for _, curr := range t.Targets {
		sb.WriteString(curr.String())
		sb.WriteString(", ")
	}
	sb.WriteString("]")

	return sb.String()
}

func (t *TargetPopulations) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, t.Site)
	binary.Write(&buf, binary.BigEndian, uint32(len(t.Targets)))
	for _, curr := range t.Targets {
		s, _ := curr.Serialize()
		buf.Write(s)
	}

	return wrapPayload(0x54, buf.Bytes()), nil
}

func (p *PopulationTarget) String() string {
	return fmt.Sprintf("PopulationTarget{Species: %s, Min: %d, Max: %d}", p.Species, p.Min, p.Max)
}

func (p *PopulationTarget) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.Write(serializeString(p.Species))
	binary.Write(&buf, binary.BigEndian, p.Min)
	binary.Write(&buf, binary.BigEndian, p.Max)

	return buf.Bytes(), nil
}
