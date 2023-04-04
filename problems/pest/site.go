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
	hub  *Hub
	Site uint32

	Targets map[string]*protocol.PopulationTarget

	Conn net.Conn
}

func NewSite(site uint32, hub *Hub) (*Site, error) {
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
		hub:     hub,
		Site:    site,
		Conn:    conn,
		Targets: targets,
	}, nil
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
