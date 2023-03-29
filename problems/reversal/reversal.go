package reversal

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

func RunReversal(port int) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not resolve udp address")
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start server")
	}

	log.Info().Int("port", port).Msg("listening")

	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Error().Err(err).Msg("could not read from UDP socket")
			continue
		}

		log.Trace().Int("len", n).Stringer("remote_address", raddr).Msg("packet recieved")

		packet, err := decodePacket(buf[:n])
		if err != nil {
			log.Warn().Err(err).Msg("invalid packet")
			continue
		}

		log.Info().Stringer("packet", packet).Msg("packet got")
	}
}
