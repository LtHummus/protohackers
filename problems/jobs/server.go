package jobs

import (
	"errors"
	"github.com/rs/zerolog/log"
	"sync"
)

var (
	ErrQueueExists = errors.New("queue already exists")
)

type Server struct {
	Queues  map[string]*Queue
	Clients map[uint64]*Client

	lock sync.Mutex
}

func NewServer() *Server {
	return &Server{
		Queues:  map[string]*Queue{},
		Clients: map[uint64]*Client{},
	}
}

func (s *Server) GetOrCreateQueue(name string) *Queue {
	s.lock.Lock()
	defer s.lock.Unlock()

	q := s.Queues[name]
	if q != nil {
		return q
	}

	q = NewQueue(name)

	log.Debug().Str("name", name).Msg("creating queue")

	s.Queues[name] = q
	return q
}

func (s *Server) FindJob(queues []string) *JobEntry {
	s.lock.Lock()
	defer s.lock.Unlock()
	var job *JobEntry

	for _, curr := range queues {
		q := s.Queues[curr]
		if q == nil {
			continue
		}
		j := q.PeekJob()

		if j == nil {
			continue
		}

		if job == nil || job.Priority < j.Priority {
			job = j
		}
	}

	if job != nil {
		s.Queues[job.Queue].DeleteJob(job.Id)
	}

	return job
}

func (s *Server) DeleteJob(id uint64) bool {
	// just go through everything
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, curr := range s.Queues {
		if curr.DeleteJob(id) {
			log.Info().Uint64("id", id).Str("queue_name", curr.Name).Msg("deleted job from queue")
			return true
		}
	}

	for _, curr := range s.Clients {
		if curr.DeleteJob(id) {
			log.Info().Uint64("id", id).Uint64("client_id", curr.Id).Msg("deleted job from client")
			return true
		}
	}

	log.Warn().Uint64("id", id).Msg("could not find job to delete")
	return false
}

func (s *Server) RegisterClient(c *Client) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Clients[c.Id] = c
}

func (s *Server) DeregisterClient(c *Client) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.Clients, c.Id)
}
