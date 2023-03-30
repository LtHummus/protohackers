package insecure

import (
	"bufio"
	"fmt"
	"github.com/lthummus/protohackers/problems/insecure/isl"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	ToyEntry = regexp.MustCompile(`(\d+)x (.+)`)
)

func RunInsecure(port int) {
	log.Info().Int("port", port).Msg("starting server")

	l, err := isl.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("could not start server")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error().Err(err).Msg("could not accept connection")
			continue
		}

		log.Info().Stringer("remote_address", conn.RemoteAddr()).Msg("connected")

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scan := bufio.NewScanner(conn)
	for scan.Scan() {
		line := scan.Text()
		log.Trace().Str("line", line).Stringer("remote_address", conn.RemoteAddr()).Msg("read line")

		resp := getPopularToy(line)

		log.Debug().Str("line", line).Str("resp", resp).Stringer("remote_address", conn.RemoteAddr()).Msg("processed line")

		_, err := io.WriteString(conn, resp)
		if err != nil {
			log.Error().Err(err).Msg("coould not respond")
			break
		}
	}

	if err := scan.Err(); err != nil {
		log.Error().Err(err).Stringer("remote_address", conn.RemoteAddr()).Msg("could not deal with connection")
	}
}

func getPopularToy(line string) string {
	list := strings.Split(line, ",")

	var answer string
	max := -1

	for _, curr := range list {
		m := ToyEntry.FindStringSubmatch(curr)
		if m == nil {
			return "whoops"
		}

		count, err := strconv.Atoi(m[1])
		if err != nil {
			return "invalid toy count"
		}

		if count > max {
			max = count
			answer = curr
		}
	}

	return fmt.Sprintf("%s\n", answer)
}
