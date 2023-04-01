package jobs

import (
	"errors"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

var (
	ErrQueueExists = errors.New("queue already exists")
)

type awaitingClients struct {
	client *Client
	queues []string
}

type Server struct {
	Queues  map[string]*Queue
	Clients map[uint64]*Client

	pendingClients    map[uint64]awaitingClients
	pendingClientLock sync.Mutex

	lock sync.Mutex
}

func NewServer() *Server {
	s := &Server{
		Queues:  map[string]*Queue{},
		Clients: map[uint64]*Client{},

		pendingClients: map[uint64]awaitingClients{},
	}

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			s.pendingClientLock.Lock()

			for _, curr := range s.pendingClients {
				j := s.FindJob(curr.queues)
				if j != nil {
					log.Info().Uint64("client_id", curr.client.Id).Uint64("job_id", j.Id).Msg("assigning awaiting client job")
					curr.client.writeJobExternal(j)
					delete(s.pendingClients, curr.client.Id)
				}
			}

			s.pendingClientLock.Unlock()
		}
	}()

	return s
}

func (s *Server) GetOrCreateQueue(name string) *Queue {
	q := s.Queues[name]
	if q != nil {
		return q
	}

	q = NewQueue(name)

	log.Debug().Str("name", name).Msg("creating queue")

	s.Queues[name] = q
	return q
}

func (s *Server) QueueJob(queue string, je *JobEntry) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.GetOrCreateQueue(queue).QueueJob(je)
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

func (s *Server) AwaitForJob(c *Client, queues []string) {
	s.pendingClientLock.Lock()
	defer s.pendingClientLock.Unlock()

	log.Info().Uint64("client_id", c.Id).Msg("adding ourselves as a waiting client")

	s.pendingClients[c.Id] = awaitingClients{
		client: c,
		queues: queues,
	}
}

func (s *Server) UnawaitForJob(c *Client) {
	s.pendingClientLock.Lock()
	defer s.pendingClientLock.Unlock()

	delete(s.pendingClients, c.Id)
}
