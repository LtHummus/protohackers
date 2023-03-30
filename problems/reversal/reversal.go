package reversal

import (
	"bufio"
	"fmt"
	"github.com/lthummus/protohackers/problems/reversal/lrcp"
	"github.com/rs/zerolog/log"
)

func RunReversal(port int) {
	log.Info().Int("port", port).Msg("starting server")
	l, err := lrcp.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen:")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("could not accept")
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn *lrcp.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		l := scanner.Bytes()
		log.Info().Str("line", string(l)).Msg("read line")
		rev := reverse(l)
		rev = append(rev, '\n')
		_, err := conn.Write(rev)
		if err != nil {
			log.Error().Err(err).Msg("could not write response")
			return
		}
	}
}
