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

func (c *client) sendMessage(msg protocol.Message) error {
	payload, err := msg.Serialize()
	if err != nil {
		log.Error().Type("message_type", msg).Err(err).Msg("could not serialize message")
		return err
	}

	_, err = c.conn.Write(payload)
	if err != nil {
		log.Error().Err(err).Msg("could not write response")
		return err
	}

	return nil
}

func (c *client) handleConnection() {
	defer c.conn.Close()

	_, err := c.conn.Write(helloPayload)
	if err != nil {
		log.Error().Err(err).Msg("could not write hello to client")
		return
	}
	log.Debug().Msg("wrote hello")

	firstMsg, err := protocol.Deserialize(c.conn)
	if err != nil {
		log.Error().Err(err).Msg("error reading hello from client")
		err = c.sendMessage(&protocol.Error{Message: fmt.Sprintf("could not deserialize: %s", err.Error())})
		if err != nil {
			log.Fatal().Err(err).Msg("could not send back failed hello")
		}
		return
	}

	hi, ok := firstMsg.(*protocol.Hello)
	if !ok {
		log.Warn().Type("message_type", firstMsg).Msg("non-hello got as first message")
		resp := &protocol.Error{Message: "non-hello message as preamble"}
		c.sendMessage(resp)
		return
	}

	if hi.Protocol != "pestcontrol" || hi.Version != 1 {
		log.Warn().Str("protocol", hi.Protocol).Uint32("version", hi.Version).Msg("invalid protocol or version")
		resp := &protocol.Error{Message: "invalid protocol or version"}
		c.sendMessage(resp)
		return
	}

	var msg protocol.Message
	for {
		msg, err = protocol.Deserialize(c.conn)
		if err != nil {
			log.Error().Err(err).Msg("could not read message in loop")
			resp := &protocol.Error{Message: fmt.Sprintf("invalid message: %s", err.Error())}
			err = c.sendMessage(resp)
			return
		}

		switch m := msg.(type) {
		case *protocol.SiteVisit:
			obs, err := m.BuildMap()
			if err != nil {
				log.Warn().Err(err).Uint32("site", m.Site).Msg("invalid observation")
				resp := &protocol.Error{Message: err.Error()}
				c.sendMessage(resp)
				return
			}
			c.hub.visitChan <- &Visit{
				Site:         m.Site,
				Observations: obs,
			}
		default:
			c.sendMessage(&protocol.Error{Message: "invalid message type"})
			log.Error().Type("message_type", m).Msg("invalid type gotten from client")
			return
		}
	}
}
