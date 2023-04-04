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

	for {
		msg, err := protocol.Deserialize(c.conn)
		if err != nil {
			log.Error().Err(err).Msg("could not read data")
			return
		}

		log.Info().Str("message", msg.String()).Msg("got message")
		_, ok := msg.(*protocol.Hello)
		if ok {
			resp, err := (&protocol.Hello{Protocol: "pestcontrol", Version: 1}).Serialize()
			if err != nil {
				log.Error().Err(err).Msg("could not serialize")
				return
			}
			c.conn.Write(resp)
		}
	}
}
