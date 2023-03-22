package mob

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

func RunMob(port int, upstream string) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen")
	}

	defer l.Close()
	log.Info().Int("port", port).Str("upstream", upstream).Msg("listening")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("could not accept connection")
		}

		log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("connected")

		go handleConnection(conn, upstream)
	}
}

func handleConnection(conn net.Conn, upstream string) {
	m := NewMitm(conn, upstream)
	if m == nil {
		return
	}

	go m.clientToServerHijack()
	go m.serverToClientHijack()

}
