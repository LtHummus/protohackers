package reversal

import (
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type Session struct {
	ID     int
	Closed bool

	Remote *net.UDPAddr

	Server *Server

	amtSoFar int

	lastAction time.Time
}

func NewSession(id int, server *Server) *Session {
	return &Session{
		ID:     id,
		Server: server,
	}
}

func (s *Session) HandleConnectMessage(packet *ConnectPacket) {
	log.Debug().Int("session", s.ID).Msg("handing connect packet")
	s.Server.WritePacket(&AckPacket{Session: s.ID}, s.Remote, s.ID)
}

func (s *Session) HandleClose(packet *ClosePacket) {
	log.Debug().Int("session", s.ID).Msg("handling close packet")
	s.Closed = true
	s.Server.WritePacket(packet, s.Remote, s.ID)
}
