package pest

import (
	"fmt"
	"github.com/lthummus/protohackers/problems/pest/protocol"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type client struct {
	conn net.Conn
	hub  *Hub

	id int64
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

	log.Debug().Str("msg", msg.String()).Msg("message sent to client")
	return nil
}

func (c *client) handleConnection() {
	c.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, err := c.conn.Write(helloPayload)
	if err != nil {
		log.Error().Err(err).Int64("client_id", c.id).Msg("could not write hello to client")
		return
	}
	log.Debug().Msg("wrote hello")

	firstMsg, err := protocol.Deserialize(c.conn)
	if err != nil {
		log.Error().Err(err).Int64("client_id", c.id).Msg("error reading hello from client")
		err = c.sendMessage(&protocol.Error{Message: fmt.Sprintf("invalid message (%d)", c.id)})
		if err != nil {
			log.Fatal().Err(err).Msg("could not send back failed hello")
		}
		return
	}

	hi, ok := firstMsg.(*protocol.Hello)
	if !ok {
		log.Warn().Type("message_type", firstMsg).Int64("client_id", c.id).Msg("non-hello got as first message")
		resp := &protocol.Error{Message: fmt.Sprintf("non-hello message as preamble (%d)", c.id)}
		c.sendMessage(resp)
		return
	}

	if hi.Protocol != "pestcontrol" || hi.Version != 1 {
		log.Warn().Str("protocol", hi.Protocol).Int64("client_id", c.id).Uint32("version", hi.Version).Msg("invalid protocol or version")
		resp := &protocol.Error{Message: fmt.Sprintf("invalid protocol or version (%d)", c.id)}
		c.sendMessage(resp)
		return
	}

	var msg protocol.Message
	for {
		msg, err = protocol.Deserialize(c.conn)
		if err != nil {
			log.Error().Err(err).Int64("client_id", c.id).Msg("could not read message in loop")
			resp := &protocol.Error{Message: fmt.Sprintf("invalid message (%d)", c.id)}
			err = c.sendMessage(resp)
			return
		}

		switch m := msg.(type) {
		case *protocol.SiteVisit:
			obs, err := m.BuildMap()
			if err != nil {
				log.Warn().Err(err).Int64("client_id", c.id).Uint32("site", m.Site).Msg("invalid observation")
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
			log.Error().Int64("client_id", c.id).Type("message_type", m).Msg("invalid type gotten from client")
			return
		}
	}
}
