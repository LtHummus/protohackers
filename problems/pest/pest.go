package pest

import (
	"bytes"
	"fmt"
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
)

func RunPest(port int) {
	log.Warn().Msg("to be implemented")

	h := &protocol.TargetPopulations{
		Site: 12345,
		Targets: []protocol.PopulationTarget{
			{
				Species: "cat",
				Min:     1,
				Max:     4,
			},
			{
				Species: "dog",
				Min:     5,
				Max:     9,
			},
		},
	}

	payload, _ := h.Serialize()

	r := bytes.NewReader(payload)

	p, err := protocol.Deserialize(r)
	if err != nil {
		log.Panic().Err(err).Msg("could not deserialize")
	}

	fmt.Printf("%v\n", p)
}
