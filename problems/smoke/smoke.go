package smoke

import (
	"fmt"
	"io"
	"net"

	"github.com/rs/zerolog/log"
)

const BufferSizeBytes = 32 * 1024 * 1024

func handleConnection(conn net.Conn) {
	_, err := io.Copy(conn, conn)
	if err != nil {
		log.Fatal().Err(err).Msg("could not copy")
	}

	conn.Close()
}

func RunSmoke(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen")
	}

	defer l.Close()
	log.Info().Int("port", port).Msg("listening...")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("could not accept connection")
		}

		log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("connection accepted")

		go handleConnection(conn)
	}
}
