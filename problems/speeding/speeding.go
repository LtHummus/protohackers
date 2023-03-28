package speeding

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

func RunSpeeding(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen")
	}

	system := NewSystem()

	defer l.Close()
	log.Info().Int("port", port).Msg("listening")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("could not accept connection")
		}

		log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("connection established")

		c := NewClient(system, conn)

		go c.Run()
	}
}
