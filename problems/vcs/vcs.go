package vcs

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"strings"
)

var (
	ReadyText = []byte("READY\n")
	HelpText  = []byte("OK usage: HELP|GET|PUT|LIST\n")
)

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		_, err := conn.Write(ReadyText)
		if err != nil {
			log.Error().Err(err).Msg("could not write ready text")
			break
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("error reading")
			break
		}

		log.Info().Str("line", line).Msg("read input")

		parts := strings.Split(line, " ")

		switch strings.ToUpper(parts[0]) {
		case "HELP":
			conn.Write(HelpText)
		default:
			conn.Write([]byte(fmt.Sprintf("ERR illegal method: %s", parts[0])))
			break
		}

	}
}

func RunVCS(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Int("port", port).Msg("could not listen")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("error on accept")
			continue
		}

		log.Info().Stringer("remote_address", conn.RemoteAddr()).Msg("client connected")

		go handleConnection(conn)

	}
}
