package pest

import (
	"fmt"
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
	"net"
)

type client struct {
	conn net.Conn
	hub  *Hub
}

var (
	helloPayload []byte
)

func init() {
	helloPayload, _ = (&protocol.Hello{Protocol: "pestcontrol", Version: 1}).Serialize()
}

func (c *client) handleConnection() {
	defer c.conn.Close()

	firstMsg, err := protocol.Deserialize(c.conn)
	_, herr := c.conn.Write(helloPayload)
	if herr != nil {
		log.Fatal().Err(err).Msg("could not write hello to client")
	}
	log.Debug().Msg("wrote hello")
	if err != nil {
		log.Error().Err(err).Msg("error reading hello from client")
		resp := &protocol.Error{Message: fmt.Sprintf("error: %s", err.Error())}
		respBytes, _ := resp.Serialize()
		c.conn.Write(respBytes)
		return
	}

	hi, ok := firstMsg.(*protocol.Hello)
	if !ok {
		log.Warn().Type("message_type", firstMsg).Msg("non-hello got as first message")
		resp := &protocol.Error{Message: "non-hello message as preamble"}
		respBytes, _ := resp.Serialize()
		c.conn.Write(respBytes)
		return
	}

	if hi.Protocol != "pestcontrol" || hi.Version != 1 {
		log.Warn().Str("protocol", hi.Protocol).Uint32("version", hi.Version).Msg("invalid protocol or version")
		resp := &protocol.Error{Message: "non-hello message as preamble"}
		respBytes, _ := resp.Serialize()
		c.conn.Write(respBytes)
		return
	}

	var msg protocol.Message
	for {
		msg, err = protocol.Deserialize(c.conn)
		if err != nil {
			log.Error().Err(err).Msg("could not read message in loop")
			resp, _ := (&protocol.Error{Message: "invalid message"}).Serialize()
			c.conn.Write(resp)
			return
		}

		switch m := msg.(type) {
		case *protocol.SiteVisit:
			obs, err := m.BuildMap()
			if err != nil {
				log.Warn().Err(err).Uint32("site", m.Site).Msg("invalid observation")
				resp := &protocol.Error{Message: err.Error()}
				respBytes, _ := resp.Serialize()
				c.conn.Write(respBytes)
				continue
			}
			c.hub.visitChan <- &Visit{
				Site:         m.Site,
				Observations: obs,
			}
		default:
			log.Fatal().Type("message_type", m).Msg("invalid type gotten from client")
		}
	}
}
