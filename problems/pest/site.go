package pest

import (
	"errors"
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
	"net"
)

const (
	AuthorityServer = "pestcontrol.protohackers.com:20547"
)

var (
	helloBytes, _ = (&protocol.Hello{Protocol: "pestcontrol", Version: 1}).Serialize()
)

type policy struct {
	id     uint32
	action protocol.PolicyAction
}

type Site struct {
	hub  *Hub
	Site uint32

	Targets  map[string]*protocol.PopulationTarget
	Policies map[string]*policy

	Visits chan *protocol.SiteVisit

	Conn net.Conn
}

func NewSite(site uint32, hub *Hub) (*Site, error) {
	log.Info().Uint32("site", site).Msg("starting site client")
	conn, err := net.Dial("tcp", AuthorityServer)
	if err != nil {
		return nil, err
	}

	conn.Write(helloBytes)
	helloResp, err := protocol.Deserialize(conn)
	if err != nil {
		return nil, err
	}

	hi, ok := helloResp.(*protocol.Hello)
	if !ok {
		return nil, errors.New("authority server did not send hello")
	}

	if hi.Version != 1 || hi.Protocol != "pestcontrol" {
		log.Warn().Uint32("version", hi.Version).Str("protocol", hi.Protocol).Msg("invalid protocol or version")
		return nil, errors.New("invalid protocol or version")
	}

	da, err := (&protocol.DialAuthority{Site: site}).Serialize()
	if err != nil {
		return nil, err
	}

	conn.Write(da)

	targetBytes, err := protocol.Deserialize(conn)
	if err != nil {
		return nil, err
	}

	targetPopulations, ok := targetBytes.(*protocol.TargetPopulations)
	if !ok {
		log.Warn().Type("message_type", targetBytes).Msg("invalid target message gotten")
		errorMessge, ok := targetBytes.(*protocol.Error)
		if ok {
			log.Error().Str("error_message", errorMessge.Message).Msg("message from authority")
		}

		return nil, errors.New("could not get target populations")
	}

	log.Info().Uint32("site", site).Int("target_population_count", len(targetPopulations.Targets)).Msg("targets gotten")

	targets := map[string]*protocol.PopulationTarget{}
	for _, curr := range targetPopulations.Targets {
		log.Info().Uint32("site", site).Str("species", curr.Species).Uint32("min", curr.Min).Uint32("max", curr.Max).Msg("got population target")
		targets[curr.Species] = &curr
	}

	s := &Site{
		hub:      hub,
		Site:     site,
		Conn:     conn,
		Targets:  targets,
		Policies: map[string]*policy{},
		Visits:   make(chan *protocol.SiteVisit, 100),
	}

	go s.visitHandler()

	return s, nil
}

func (s *Site) visitHandler() {
	for visit := range s.Visits {
		if visit.Site != s.Site {
			log.Panic().Uint32("site_visit_id", visit.Site).Uint32("site_id", s.Site).Msg("site id mismatch!")
		}

		obsMap := map[string]*protocol.Observation{}
		for _, curr := range visit.Observations {
			x := curr
			obsMap[curr.Species] = &x
		}

		for species, target := range s.Targets {
			log.Debug().Uint32("site", s.Site).Str("species", species).Msg("checking species")
			curr := obsMap[species]
			if curr == nil {
				curr = &protocol.Observation{
					Species: species,
					Count:   0,
				}
			}

			currPolicy := s.Policies[species]

			var action protocol.PolicyAction
			var actionStr string
			if curr.Count < target.Min {
				action = protocol.Conserve
				actionStr = "conserve"
			} else if curr.Count > target.Max {
				action = protocol.Cull
				actionStr = "cull"
			}

			if currPolicy != nil && currPolicy.action == action {
				// same policy already exists, do nothing
				continue
			}

			// our policy is different (or new), so delete the old one
			if currPolicy != nil {
				err := s.DeletePolicy(currPolicy.id)
				if err != nil {
					log.Fatal().Err(err).Msg("could not delete policy")
				}
				delete(s.Policies, species)
			}

			if action != 0 {
				policyId, err := s.CreatePolicy(species, action)
				if err != nil {
					log.Fatal().Err(err).Str("species", species).Uint32("site", s.Site).Msg("could not create policy")
					continue
				}

				s.Policies[species] = &policy{
					id:     policyId,
					action: action,
				}
				log.Info().Uint32("site", s.Site).Str("species", species).Str("action", actionStr).Uint32("policy_id", policyId).Msg("policy created")
			}
		}
	}
}

func (s *Site) CreatePolicy(species string, action protocol.PolicyAction) (uint32, error) {
	p := &protocol.CreatePolicy{
		Species: species,
		Action:  action,
	}

	payload, err := p.Serialize()
	if err != nil {
		return 0, err
	}

	_, err = s.Conn.Write(payload)
	if err != nil {
		return 0, err
	}

	resp, err := protocol.Deserialize(s.Conn)
	if err != nil {
		return 0, err
	}

	cPolicy, ok := resp.(*protocol.PolicyResult)
	if !ok {
		log.Warn().Type("response_type", resp).Msg("no policy result")
		return 0, err
	}

	log.Debug().Str("policy", cPolicy.String()).Msg("policy created")

	return cPolicy.Policy, nil

}

func (s *Site) DeletePolicy(id uint32) error {
	payload, err := (&protocol.DeletePolicy{Policy: id}).Serialize()
	if err != nil {
		return err
	}

	_, err = s.Conn.Write(payload)
	if err != nil {
		return err
	}

	resp, err := protocol.Deserialize(s.Conn)
	if err != nil {
		return err
	}

	_, ok := resp.(*protocol.Ok)
	if !ok {
		log.Warn().Type("response_type", resp).Msg("could not confirm delete")
		return err
	}

	return nil
}
