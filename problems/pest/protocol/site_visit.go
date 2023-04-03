package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type Observation struct {
	Species string
	Count   uint32
}

type SiteVisit struct {
	Site         uint32
	Observations []Observation
}

func (o *Observation) String() string {
	return fmt.Sprintf("Observation{Species: %s, Count: %d}", o.Species, o.Count)
}

func (o *Observation) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.Write(serializeString(o.Species))
	binary.Write(&buf, binary.BigEndian, o.Count)

	return buf.Bytes(), nil
}

func (s *SiteVisit) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("SiteVisit{Site: %d, Populations: [", s.Site))
	for _, curr := range s.Observations {
		st := curr.String()
		sb.WriteString(st)
		sb.WriteString(", ")
	}
	sb.WriteString("]}")

	return sb.String()
}

func (s *SiteVisit) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, s.Site)
	binary.Write(&buf, binary.BigEndian, uint32(len(s.Observations)))
	for _, curr := range s.Observations {
		b, _ := curr.Serialize()
		buf.Write(b)
	}

	return wrapPayload(0x58, buf.Bytes()), nil
}
