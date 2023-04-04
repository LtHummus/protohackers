package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

	return wrapPayload(MagicNumberTargetPopulations, buf.Bytes()), nil
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

func deserializeTargetPopulations(r io.Reader) (Message, error) {
	var site uint32
	err := binary.Read(r, binary.BigEndian, &site)
	if err != nil {
		return nil, err
	}

	var ruleCount uint32
	err = binary.Read(r, binary.BigEndian, &ruleCount)

	targets := make([]PopulationTarget, ruleCount)
	for i := range targets {
		species, err := readString(r)
		if err != nil {
			return nil, err
		}
		var minMax struct {
			Min uint32
			Max uint32
		}
		err = binary.Read(r, binary.BigEndian, &minMax)
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
