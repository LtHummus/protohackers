package pest

import (
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
	"sync"
)

type policy struct {
	id uint32
}

type Hub struct {
	lock    sync.Mutex
	targets map[uint32]map[string]*protocol.PopulationTarget

	policies map[uint32]map[string]*policy

	connections map[uint32]*Site
}

func NewHub() *Hub {
	return &Hub{
		targets:     map[uint32]map[string]*protocol.PopulationTarget{},
		connections: map[uint32]*Site{},
		policies:    map[uint32]map[string]*policy{},
	}
}

func (h *Hub) checkConnection(site uint32) error {
	if h.connections[site] == nil {
		log.Debug().Uint32("site", site).Msg("no connection, exists...connecting")
		c, err := NewSite(site, h)
		if err != nil {
			log.Error().Err(err).Msg("could not contact authority server")
			if err != nil {
				return err
			}
		}
		log.Debug().Uint32("site", site).Msg("connection made")
		h.connections[site] = c
	}

	if h.targets[site] == nil {
		log.Debug().Uint32("site", site).Msg("updated site targets")
		h.targets[site] = h.connections[site].Targets
	}

	if h.policies[site] == nil {
		h.policies[site] = map[string]*policy{}
	}

	return nil
}

func (h *Hub) querySite(site uint32) error {
	log.Info().Uint32("site", site).Msg("querying for info")
	err := h.checkConnection(site)
	if err != nil {
		return err
	}

	return nil
}

func (h *Hub) DeregisterSite(site uint32) {
	h.lock.Lock()
	defer h.lock.Unlock()
}

func (h *Hub) HandleSiteVisit(site uint32, obs []protocol.Observation) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	err := h.querySite(site)
	if err != nil {
		return err
	}

	obsMap := map[string]*protocol.Observation{}
	for _, curr := range obs {
		x := curr
		obsMap[curr.Species] = &x
	}
	for species, target := range h.targets[site] {
		log.Info().Uint32("site", site).Str("species", species).Msg("checking species")
		curr := obsMap[species]
		if curr == nil {
			curr = &protocol.Observation{
				Species: species,
				Count:   0,
			}
		}

		// delete old policy if it exists
		if p := h.policies[site][species]; p != nil {
			log.Info().Uint32("site", site).Uint32("policy_id", p.id).Msg("deleting policy id")
			err = h.connections[site].DeletePolicy(p.id)
			if err != nil {
				log.Error().Err(err).Msg("could not delete old policy")
			}
		}

		var action protocol.PolicyAction

		if curr.Count < target.Min {
			log.Info().Str("species", curr.Species).Uint32("count", curr.Count).Uint32("min", target.Min).Uint32("max", target.Max).Uint32("site", site).Msg("should create conserve policy")
			action = protocol.Conserve
		} else if curr.Count > target.Max {
			log.Info().Str("species", curr.Species).Uint32("count", curr.Count).Uint32("max", target.Max).Uint32("max", target.Max).Uint32("site", site).Msg("should create cull policy")
			action = protocol.Cull
		} else {
			log.Info().Str("species", curr.Species).Uint32("count", curr.Count).Uint32("max", target.Max).Uint32("max", target.Max).Uint32("site", site).Msg("should delete existing polciy if there is one")
		}

		if action != 0 {
			policyId, err := h.connections[site].CreatePolicy(curr.Species, action)
			if err != nil {
				log.Error().Err(err).Msg("could not create policy")
				continue
			}

			h.policies[site][curr.Species] = &policy{id: policyId}
			log.Info().Str("species", curr.Species).Uint32("site", site).Uint32("policy_id", policyId).Msg("created policy")
		}
	}

	return nil
}
