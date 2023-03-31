package jobs

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

func RunJobs(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Int("port", port).Err(err).Msg("could not start server")
	}

	s := NewServer()

	log.Info().Int("port", port).Msg("starting server")

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("could not accept connection")
			continue
		}

		log.Info().Stringer("remote_addr", conn.RemoteAddr()).Msg("got connection")

		c := &Client{
			Id:           generateJobId(),
			Conn:         conn,
			Server:       s,
			AssignedJobs: map[uint64]*JobEntry{},
		}

		s.RegisterClient(c)

		go c.handleConnection()
	}
}
