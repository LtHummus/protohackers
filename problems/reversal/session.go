package reversal

import (
	"bufio"
	"bytes"
	"github.com/lthummus/protohackers/problems/reversal/lrcp"
	"github.com/rs/zerolog/log"
	"net"
)

type Session struct {
	ID     int
	Closed bool

	Remote *net.UDPAddr

	Server *Server

	shouldBeCleaned bool

	// application layer stuff

	recvCount int
	recvData  *bytes.Buffer
}

func NewSession(id int, remote *net.UDPAddr, server *Server) *Session {
	s := &Session{
		ID:     id,
		Server: server,
		Remote: remote,

		recvData: &bytes.Buffer{},
	}

	go func() {
		scanner := bufio.NewScanner(s.recvData)
		for scanner.Scan() {
			l := scanner.Bytes()
			l = reverse(l)
			l = append(l, '\n')

			s.respond(&lrcp.DataPacket{
				Session:  id,
				Position: 0,
				Data:     l,
			})
		}
	}()
	return s
}

func (s *Session) HandlePacket(packet lrcp.Packet) {
	switch p := packet.(type) {
	case *lrcp.ConnectPacket:
		s.handleConnect(p)
	case *lrcp.DataPacket:
		s.handleData(p)
	case *lrcp.ClosePacket:
		s.handleClose(p)
	case *lrcp.AckPacket:
		s.handleAck(p)
	default:
		log.Warn().Msg("unknown packet kind")
	}
}

func (s *Session) handleConnect(packet *lrcp.ConnectPacket) {
	log.Debug().Int("session", s.ID).Msg("handing connect packet")
	s.respond(&lrcp.AckPacket{Session: s.ID})
}

func (s *Session) handleClose(packet *lrcp.ClosePacket) {
	log.Debug().Int("session", s.ID).Msg("handling close packet")
	s.Closed = true
	s.respond(packet)
}

func (s *Session) handleData(packet *lrcp.DataPacket) {
	log.Debug().Int("session", s.ID).Msg("handing data packet")
	if s.Closed {
		s.respond(&lrcp.ClosePacket{Session: s.ID})
		return
	}

	if s.recvCount >= (packet.Position + len(packet.Data)) {
		// extra data transmission, ignore it
		log.Debug().Int("session", s.ID).Msg("got extra data retransmission")
		return
	}

	if s.recvCount >= packet.Position {
		log.Debug().Int("session", s.ID).Msg("got new data")

		idx := s.recvCount - packet.Position
		s.recvData.Write(packet.Data[idx:])
		s.recvCount = packet.Position + len(packet.Data)

		s.respond(&lrcp.AckPacket{Session: s.ID, Length: s.recvCount})
		return
	}

	// if we got here, we are missing some data, request retransmit
	s.respond(&lrcp.AckPacket{Session: s.ID, Length: s.recvCount})
}

func (s *Session) handleAck(packet *lrcp.AckPacket) {
	log.Debug().Int("session", s.ID).Msg("handling ack packet")
	// TODO: something
}

func (s *Session) respond(packet lrcp.Packet) {
	s.Server.WritePacket(packet, s.Remote, s.ID)
}

func reverse(x []byte) []byte {
	r := x[:]
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return r
}
