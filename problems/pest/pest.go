package pest

import (
	"fmt"
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
)

func RunPest(port int) {
	log.Warn().Msg("to be implemented")

	h := &protocol.SiteVisit{
		Site: 12345,
		Observations: []protocol.Observation{
			{
				Species: "dog",
				Count:   1,
			},
			{
				Species: "rat",
				Count:   5,
			},
		},
	}

	s, err := h.Serialize()
	if err != nil {
		log.Panic().Err(err).Msg("could not serialize")
	}

	log.Info().Hex("bytes", s).Msg("serialized")

	fmt.Printf("%s\n", h.String())
}
