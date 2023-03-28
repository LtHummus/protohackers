package speeding

import "github.com/rs/zerolog/log"

type Dispatcher struct {
	Roads         []uint16
	TicketChannel chan []byte
}

func (c *Client) dispatchChannelCreator() {
	for _, r := range c.dispatcher.Roads {
		go func(road uint16) {
			for curr := range c.system.roadChannels[road] {
				log.Info().Hex("ticket", curr).Msg("read message, forwarding")
				c.dispatcher.TicketChannel <- curr
			}
		}(r)
	}
}

func (c *Client) dispatchSendChannel() {
	for curr := range c.dispatcher.TicketChannel {
		log.Info().Hex("ticket", curr).Msg("sending ticket to client")
		_, err := c.conn.Write(curr)
		if err != nil {
			log.Error().Err(err).Msg("could not write to client")
		}

	}
}
