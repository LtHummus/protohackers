package pest

import (
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
	"net"
)

type client struct {
	conn net.Conn
	hub  *Hub
}

func (c *client) handleConnection() {
	defer c.conn.Close()

	firstMsg, err := protocol.Deserialize(c.conn)
	if err != nil {
		log.Error().Err(err).Msg("error reading from client")
		return
	}

	hi, ok := firstMsg.(*protocol.Hello)
	if !ok {
		log.Warn().Type("message_type", firstMsg).Msg("non-hello got as first message")
		resp, _ := (&protocol.Error{Message: "non-hello message as preamble"}).Serialize()
		c.conn.Write(resp)
		return
	}

	if hi.Protocol != "pestcontrol" || hi.Version != 1 {
		log.Warn().Str("protocol", hi.Protocol).Uint32("version", hi.Version).Msg("invalid protocol or version")
		resp, _ := (&protocol.Error{Message: "invalid protocol or version"}).Serialize()
		c.conn.Write(resp)
		return
	}

	resp, _ := (&protocol.Hello{Protocol: "pestcontrol", Version: 1}).Serialize()
	_, err = c.conn.Write(resp)
	if err != nil {
		log.Error().Err(err).Msg("could not write hello to client")
	}

	for {
		msg, err := protocol.Deserialize(c.conn)
		if err != nil {
			log.Error().Err(err).Msg("could not read data")
			resp, _ := (&protocol.Error{Message: "invalid message"}).Serialize()
			c.conn.Write(resp)
			return
		}

		switch m := msg.(type) {
		case *protocol.SiteVisit:
			log.Info().Msg("got site visit")
			s, err := c.hub.GetOrQuerySite(m.Site)
			if err != nil {
				log.Error().Err(err).Uint32("site", m.Site).Msg("could not query site")
			}
			s.HandleSiteVisit(m.Observations)
		}
	}
}
