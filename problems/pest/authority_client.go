package pest

import (
	"errors"
	"fmt"
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
	"net"
)

const (
	AuthorityServer     = "pestcontrol.protohackers.com"
	AuthorityServerPort = 20547
)

var (
	authorityServerAddress = fmt.Sprintf("%s:%d", AuthorityServer, AuthorityServerPort)
	helloBytes, _          = (&protocol.Hello{Protocol: "pestcontrol", Version: 1}).Serialize()
)

type Site struct {
	Site    uint32
	Targets map[string]*protocol.PopulationTarget

	Conn net.Conn
}

func NewSite(site uint32) (*Site, error) {
	log.Info().Uint32("site", site).Msg("starting site client")
	conn, err := net.Dial("tcp", authorityServerAddress)
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
		return nil, errors.New("could not get target populations")
	}

	log.Info().Uint32("site", site).Int("target_population_count", len(targetPopulations.Targets)).Msg("targets gotten")

	targets := map[string]*protocol.PopulationTarget{}
	for _, curr := range targetPopulations.Targets {
		log.Info().Uint32("site", site).Str("species", curr.Species).Uint32("min", curr.Min).Uint32("max", curr.Max).Msg("got population target")
		targets[curr.Species] = &curr
	}

	return &Site{
		Site:    site,
		Targets: targets,
		Conn:    conn,
	}, nil
}

func (s *Site) HandleSiteVisit(obs []protocol.Observation) {
	for _, curr := range obs {
		target := s.Targets[curr.Species]
		if target == nil {
			continue
		}

		if curr.Count < target.Min {
			log.Info().Str("species", curr.Species).Uint32("count", curr.Count).Uint32("min", target.Min).Uint32("max", target.Max).Msg("should create conserve policy")
		} else if curr.Count > target.Max {
			log.Info().Str("species", curr.Species).Uint32("count", curr.Count).Uint32("max", target.Max).Uint32("max", target.Max).Msg("should create cull policy")
		} else {
			log.Info().Str("species", curr.Species).Uint32("count", curr.Count).Uint32("max", target.Max).Uint32("max", target.Max).Msg("should delete existing polciy if there is one")
		}
	}
}
