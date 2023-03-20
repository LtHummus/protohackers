package chat

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

func RunChat(port int) {
	h := NewHub()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen")
	}

	log.Info().Int("port", port).Msg("listening")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("could not accept connection")
		}

		log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("connected")

		go ConnectClient(conn, h)
	}
}
