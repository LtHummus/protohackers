package reversal

import (
	"github.com/lthummus/protohackers/problems/reversal/lrcp"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"time"
)

const (
	RetransmissionInterval = 3 * time.Second
	SessionTimeout         = 1 * time.Minute

	CleanupCheckInterval = 5 * time.Second
)

type Server struct {
	lock       *sync.Mutex
	sessionMap map[int]*Session

	conn *net.UDPConn
}

func NewServer(conn *net.UDPConn) *Server {
	s := &Server{
		lock:       &sync.Mutex{},
		sessionMap: map[int]*Session{},
		conn:       conn,
	}

	go func() {
		log.Info().Msg("starting ticker goroutine")

		t := time.NewTicker(CleanupCheckInterval)
		for {
			<-t.C
			s.lock.Lock()
			for id, sess := range s.sessionMap {
				if sess.Closed || sess.shouldBeCleaned {
					log.Info().Int("session", id).Msg("cleaning session")
					delete(s.sessionMap, id)
				}
			}
			s.lock.Unlock()
		}
	}()

	return s
}

func (s *Server) WritePacket(p lrcp.Packet, dest *net.UDPAddr, session int) {
	_, err := s.conn.WriteToUDP(p.Serialize(), dest)
	if err != nil {
		log.Error().Err(err).Str("kind", p.Kind()).Stringer("remote_address", dest).Int("session", session).Msg("could not write packet")
	}
}

func (s *Server) GetOrCreateSession(id int, remote *net.UDPAddr) (*Session, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	sess := s.sessionMap[id]
	if sess != nil {
		return sess, false
	}

	log.Info().Int("id", id).Msg("creating session")
	sess = NewSession(id, remote, s)
	s.sessionMap[id] = sess
	return sess, true
}
