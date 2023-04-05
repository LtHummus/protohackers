package pest

import (
	"github.com/rs/zerolog/log"
)

type Visit struct {
	Site         uint32
	Observations map[string]uint32
}

type Hub struct {
	connections map[uint32]*Site
	visitChan   chan *Visit
}

func NewHub() *Hub {
	h := &Hub{
		connections: map[uint32]*Site{},
		visitChan:   make(chan *Visit, 1000),
	}

	go h.handleSiteVisit()

	return h
}

func (h *Hub) getSite(site uint32) (*Site, error) {
	s := h.connections[site]

	if s != nil {
		return s, nil
	}

	var err error
	s, err = NewSite(site, h)
	if err != nil {
		log.Error().Err(err).Uint32("site", site).Msg("could not contact authority server")
		return nil, err
	}

	log.Info().Uint32("site", site).Msg("created site")
	h.connections[site] = s

	return s, nil
}

func (h *Hub) handleSiteVisit() {
	for visit := range h.visitChan {
		site, err := h.getSite(visit.Site)
		if err != nil {
			log.Fatal().Err(err).Uint32("site", visit.Site).Msg("could not validate authority")
		}

		site.Visits <- visit
	}
}
