package prime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"math/big"
	"net"
)

var malformedResponse = []byte("no")
var (
	yesPrimeResponse []byte
	noPrimeResponse  []byte
)

func init() {
	yesPrimeResponse, _ = json.Marshal(Output{
		Method:  "isPrime",
		IsPrime: true,
	})
	noPrimeResponse, _ = json.Marshal(Output{
		Method:  "isPrime",
		IsPrime: false,
	})

	yesPrimeResponse = append(yesPrimeResponse, '\n')
	noPrimeResponse = append(noPrimeResponse, '\n')
}

func handleConnection(conn net.Conn) {
	for {
		defer conn.Close()
		read := bufio.NewReader(conn)
		line, err := read.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("could not read")
			conn.Write(malformedResponse)
			return
		}

		log.Info().Str("line", line).Msg("read line")

		var in Input
		err = json.Unmarshal([]byte(line), &in)
		if err != nil {
			log.Error().Err(err).Msg("bad json decode")
			conn.Write(malformedResponse)
			return
		}

		if in.Method != "isPrime" || in.Number == nil {
			log.Error().Str("method", in.Method).Interface("number", in.Number).Msg("bad json")
			conn.Write(malformedResponse)
			return
		}

		n := int(*in.Number)
		if float64(n) != *in.Number {
			log.Warn().Float64("number", *in.Number).Msg("non-int input")
			conn.Write(noPrimeResponse)
			return
		}

		isPrime := big.NewInt(int64(n)).ProbablyPrime(0)
		log.Info().Int("number", n).Bool("is_prime", isPrime).Msg("prime calculated")

		if isPrime {
			conn.Write(yesPrimeResponse)
		} else {
			conn.Write(noPrimeResponse)
		}
	}

}

func RunPrime(port int) {
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
