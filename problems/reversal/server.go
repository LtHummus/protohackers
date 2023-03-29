package reversal

import (
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"time"
)

const (
	RetransmissionInterval = 3 * time.Second
	SessionTimeout         = 1 * time.Minute
)

type Server struct {
	lock       *sync.Mutex
	sessionMap map[int]any

	conn *net.UDPConn
}

func NewServer() *Server {
	return &Server{
		lock:       &sync.Mutex{},
		sessionMap: map[int]any{},
	}
}

func (s *Server) WritePacket(p Packet, dest *net.UDPAddr, session int) {
	_, err := s.conn.WriteToUDP(p.Serialize(), dest)
	if err != nil {
		log.Error().Err(err).Str("kind", p.Kind()).Stringer("remote_address", dest).Int("session", session).Msg("could not write packet")
	}
}
