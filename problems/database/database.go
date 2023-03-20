package database

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"strings"
)

func RunDatabase(port int) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not resolve udp address")
	}
	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start server")
	}

	log.Info().Int("port", port).Msg("listening")

	defer l.Close()
	buf := make([]byte, 1024)
	for {
		n, raddr, err := l.ReadFromUDP(buf)
		if err != nil {
			log.Error().Err(err).Msg("could not read from UDP socket")
			continue
		}

		log.Info().Int("len", n).Stringer("remote_address", raddr).Msg("message recieved")

		message := string(buf[:n])
		parts := strings.SplitN(message, "=", 2)
		if len(parts) == 1 {
			// retrieve
			resp := fmt.Sprintf("%s=%s", parts[0], Query(parts[0]))
			_, err := l.WriteToUDP([]byte(resp), raddr)
			if err != nil {
				log.Warn().Err(err).Msg("could not write response")
			}
		} else {
			// set
			Set(parts[0], parts[1])
		}
	}
}
