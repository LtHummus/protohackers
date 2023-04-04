package pest

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

func RunPest(port int) {
	h := NewHub()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Int("port", port).Msg("could not listen")
	}

	log.Info().Int("port", port).Msg("listening")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("could not accept connection")
		}

		log.Info().Stringer("remote_address", conn.RemoteAddr()).Msg("accepted connection")

		c := client{conn: conn, hub: h}

		go c.handleConnection()
	}
}
