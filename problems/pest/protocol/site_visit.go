package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

	return wrapPayload(MagicNumberSiteVisit, buf.Bytes()), nil
}

func deserializeSiteVisit(r io.Reader) (Message, error) {
	var header struct {
		Site     uint32
		PopCount uint32
	}

	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	obs := make([]Observation, header.PopCount)
	for i := range obs {
		species, err := readString(r)
		if err != nil {
			return nil, err
		}
		var count uint32
		err = binary.Read(r, binary.BigEndian, &count)
		if err != nil {
			return nil, err
		}
		obs[i] = Observation{
			Species: species,
			Count:   count,
		}
	}

	return &SiteVisit{
		Site:         header.Site,
		Observations: obs,
	}, nil
}

func (s *SiteVisit) ValidateObservations() error {
	m := map[string]uint32{}
	for _, curr := range s.Observations {
		count, exists := m[curr.Species]
		if exists && count != curr.Count {
			return fmt.Errorf("conflicting observations: site %d, %s has both %d and %d", s.Site, curr.Species, count, curr.Count)
		}

		m[curr.Species] = curr.Count
	}

	return nil
}
