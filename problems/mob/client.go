package mob

import (
	"bufio"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"sync"
)

type Mitm struct {
	chatConn   net.Conn
	clientConn net.Conn

	closeLock *sync.Once
}

func NewMitm(client net.Conn, upstream string) *Mitm {
	chatConn, err := net.Dial("tcp", upstream)
	if err != nil {
		log.Error().Err(err).Msg("could not connect to upstream")
		client.Close()
		return nil
	}

	return &Mitm{
		chatConn:   chatConn,
		clientConn: client,
		closeLock:  &sync.Once{},
	}
}

func (m *Mitm) clientToServerHijack() {
	r := bufio.NewReader(m.clientConn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				m.Close()
				return
			}
			log.Warn().Err(err).Msg("could not read from client")
			return
		}

		patched := swapAddress(line)

		log.Debug().Str("line", line).Str("patched", patched).Msg("read line from client")

		_, err = m.chatConn.Write([]byte(patched))
		if err != nil {
			if err == io.EOF {
				m.Close()
				return
			}
			log.Warn().Err(err).Msg("could not pass to upstream")
			return
		}
	}
}

func (m *Mitm) serverToClientHijack() {
	r := bufio.NewReader(m.chatConn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				m.Close()
				return
			}
			log.Warn().Err(err).Msg("could not read from server")
			return
		}

		patched := swapAddress(line)
		log.Debug().Str("line", line).Str("patched", patched).Msg("read line from server")

		_, err = m.clientConn.Write([]byte(patched))
		if err != nil {
			if err == io.EOF {
				m.Close()
				return
			}
			log.Warn().Err(err).Msg("could not pass to client")
			return
		}
	}
}

func (m *Mitm) Close() {
	m.closeLock.Do(func() {
		m.chatConn.Close()
		m.clientConn.Close()
		log.Info().Msg("connection closed")
	})
}
