package vcs

import (
	"bufio"
	"fmt"
	"github.com/lthummus/protohackers/problems/vcs/filesystem"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"strings"
)

var (
	ReadyText = []byte("READY\n")
	HelpText  = []byte("OK usage: HELP|GET|PUT|LIST\n")
)

var fs = filesystem.NewFilesystem()

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

		log.Debug().Str("line", line).Msg("read input")

		parts := strings.Split(strings.TrimSpace(line), " ")

		log.Info().Strs("parts", parts).Msg("read line from client")

		switch strings.ToUpper(parts[0]) {
		case "HELP":
			conn.Write(HelpText)
		case "LIST":
			handleList(conn, parts)
		case "PUT":
			handlePut(conn, parts)
		default:
			conn.Write([]byte(fmt.Sprintf("ERR illegal method: %s\n", parts[0])))
			break
		}

	}
}

func handleList(conn net.Conn, parts []string) {
	if len(parts) != 2 {
		log.Warn().Msg("invalid list syntax")
		conn.Write([]byte("ERR usage: LIST dir\n"))
		return
	}

	res, err := fs.List(parts[1])
	if err != nil {
		log.Warn().Err(err).Msg("could not list")
		conn.Write([]byte(fmt.Sprintf("ERR %s\n", err.Error())))
		return
	}

	log.Info().Msg("list successful")
	conn.Write([]byte(fmt.Sprintf("%s", res.String())))
}

func handlePut(conn net.Conn, parts []string) {
	if len(parts) != 3 {
		conn.Write([]byte("ERR usage: PUT file length newline data\n"))
		return
	}

	filename := parts[1]
	length, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Warn().Err(err).Str("len", parts[2]).Msg("illegal length")
		length = 0
	}

	log.Info().Str("name", filename).Int("len", length).Msg("putting file")
	r, err := fs.Put(filename, []byte("foo"))
	if err != nil {
		log.Error().Err(err).Msg("could not put")
		conn.Write([]byte(fmt.Sprintf("ERR %s\n", err.Error())))
		return
	}

	conn.Write([]byte(fmt.Sprintf("OK %s\n", r)))
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
