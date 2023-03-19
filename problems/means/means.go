package means

import (
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

type Client struct {
	conn    net.Conn
	entries map[int32]int32
}

func (c *Client) handleConnection() {
	defer c.conn.Close()

	var packet inputMessage
	for {
		err := binary.Read(c.conn, binary.BigEndian, &packet)
		if err != nil {
			log.Warn().Err(err).Msg("error reading from client")
			return
		}

		if !packet.IsValid() {
			log.Warn().Interface("packet", packet).Msg("invalid packet")
			return
		}

		log.Info().Interface("packet", packet).Msg("valid packet")

		if packet.IsInsert() {
			c.entries[packet.Timestamp()] = packet.Price()
		} else {
			// query
			total := 0
			count := 0

			min := packet.MinTime()
			max := packet.MaxTime()
			for k, v := range c.entries {
				if min <= k && k <= max {
					total += int(v)
					count++
				}
			}

			var avg int32 = 0
			if count != 0 {
				avg = int32(total / count)
			}

			log.Info().Int32("average", avg).Msg("sending response")

			err = binary.Write(c.conn, binary.BigEndian, avg)
			if err != nil {
				log.Warn().Err(err).Msg("could not write result")
				return
			}
		}

	}
}

func RunMeans(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen")
	}

	defer l.Close()

	log.Info().Int("port", port).Msg("listening")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Warn().Err(err).Msg("could not accept connection")
			continue
		}

		log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("connected")

		c := &Client{
			conn:    conn,
			entries: map[int32]int32{},
		}

		go c.handleConnection()
	}
}
