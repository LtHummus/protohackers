package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"net"
)

func readRoutine(conn net.Conn, name string) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Error().Err(err).Str("client", name).Msg("error reading")
			return
		}

		log.Info().Str("client", name).Hex("data", buf[:n]).Msg("read data")
	}
}

var speedTestCmd = &cobra.Command{
	Use: "speedtest",
	Run: func(cmd *cobra.Command, args []string) {
		a, err := net.Dial("tcp", "127.0.0.1:28172")
		if err != nil {
			panic(err)
		}

		b, err := net.Dial("tcp", "127.0.0.1:28172")
		if err != nil {
			panic(err)
		}

		c, err := net.Dial("tcp", "127.0.0.1:28172")
		if err != nil {
			panic(err)
		}

		go readRoutine(a, "a")
		go readRoutine(b, "b")
		go readRoutine(c, "c")

		c.Write([]byte{0x40, 0x00, 0x00, 0x00, 0x0a})

		/*
			<-- 80 00 7b 00 08 00 3c
			<-- 20 04 55 4e 31 58 00 00 00 00
		*/
		a.Write([]byte{0x80, 0x00, 0x7b, 0x00, 0x08, 0x00, 0x3c, 0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x00, 0x00})

		/*
			<-- 80 00 7b 00 09 00 3c
			<-- 20 04 55 4e 31 58 00 00 00 2d
		*/
		b.Write([]byte{0x80, 0x00, 0x7b, 0x00, 0x09, 0x00, 0x3c, 0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x00, 0x2d})

		for {
		}
	},
}
