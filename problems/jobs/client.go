package jobs

import (
	"bufio"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"time"
)

type Client struct {
	Id     uint64
	Server *Server
	Conn   net.Conn

	AssignedJobs map[uint64]*JobEntry

	closed bool

	lock sync.Mutex
}

func (c *Client) handleConnection() {
	scan := bufio.NewScanner(c.Conn)
	for scan.Scan() {
		l := scan.Bytes()

		log.Debug().Bytes("payload", l).Msg("read line from client")

		msg, err := decodeMessage(l)
		if err != nil {
			log.Error().Err(err).Msg("could not decode message")
			c.Conn.Write(InvalidMessageTypeBytes)
			continue
		}

		log.Debug().Type("message_type", msg).Msg("message decoded")

		switch p := msg.(type) {
		case *PutRequest:
			c.handlePutMessage(p)
		case *GetRequest:
			c.handleGetMessage(p)
		case *DeleteRequest:
			c.handleDeleteMessage(p)
		case *AbortRequest:
			c.handleAbortMessage(p)
		default:
			log.Warn().Msg("unknown message type")
			c.Conn.Write(UnknownRequestTypeBytes)
		}
	}

	if err := scan.Err(); err != nil {
		log.Error().Err(err).Msg("could not read")
	}

	log.Info().Stringer("remote_addr", c.Conn.RemoteAddr()).Msg("client disconnected")

	c.closed = true
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, curr := range c.AssignedJobs {
		c.Server.GetOrCreateQueue(curr.Queue).QueueJob(curr)
		log.Info().Uint64("id", curr.Id).Str("queue_name", curr.Queue).Msg("requeued job")
	}
}

func (c *Client) handlePutMessage(p *PutRequest) {
	log.Trace().Str("queue_name", p.Queue).Int("priority", p.Priority).Msg("got put request")

	if p.Priority < 0 {
		response, err := serializeMessage(&ErrorResponse{
			Status: "error",
			Error:  "Priority cannot be negative",
		})
		if err != nil {
			log.Error().Err(err).Msg("could not create error response")
			return
		}
		c.Conn.Write(response)
		return
	}

	q := c.Server.GetOrCreateQueue(p.Queue)
	job := q.QueueJob(&JobEntry{
		Id:       generateJobId(),
		Queue:    p.Queue,
		Priority: p.Priority,
		Job:      p.Job,
	})
	response, err := serializeMessage(&PutResponse{Status: "ok", Id: job.Id})
	if err != nil {
		log.Error().Err(err).Msg("could not serialize response")
	}
	c.Conn.Write(response)
}

func (c *Client) handleGetMessage(g *GetRequest) {
	log.Trace().Strs("queue_names", g.Queues).Bool("wait", g.Wait).Msg("got get request")

	if len(g.Queues) == 0 {
		c.Conn.Write(NoJobResponseBytes)
		return
	}

	if g.Wait {
		go func() {
			log.Info().Uint64("client_id", c.Id).Strs("queues", g.Queues).Msg("starting wait thread")
			for !c.closed {
				time.Sleep(5 * time.Second)
				j := c.Server.FindJob(g.Queues)
				if j != nil {
					c.assignJob(j)
					response, err := serializeMessage(&GetResponse{
						Status:   "ok",
						Id:       j.Id,
						Job:      j.Job,
						Priority: j.Priority,
						Queue:    j.Queue,
					})
					if err != nil {
						log.Error().Err(err).Msg("could not serialize job response")
						return
					}
					log.Info().Uint64("client_id", c.Id).Uint64("job_id", j.Id).Msg("assigning from thread")
					c.Conn.Write(response)
					return
				}
			}
			log.Warn().Uint64("client_id", c.Id).Msg("stopping job wait thread on disconnect")
		}()
	} else {
		j := c.Server.FindJob(g.Queues)
		if j != nil {
			c.assignJob(j)
			response, err := serializeMessage(&GetResponse{
				Status:   "ok",
				Id:       j.Id,
				Job:      j.Job,
				Priority: j.Priority,
				Queue:    j.Queue,
			})
			if err != nil {
				log.Error().Err(err).Msg("could not serialize job response")
				return
			}
			c.Conn.Write(response)
			return
		} else {
			c.Conn.Write(NoJobResponseBytes)
		}
	}
}

func (c *Client) handleDeleteMessage(d *DeleteRequest) {
	log.Trace().Uint64("id", d.Id).Msg("got delete request")
	deleted := c.Server.DeleteJob(d.Id)
	if deleted {
		log.Info().Uint64("client_id", c.Id).Uint64("job_id", d.Id).Msg("sending delete successful response")
		c.Conn.Write(EmptyOkResponseBytes)
	} else {
		log.Info().Uint64("client_id", c.Id).Uint64("job_id", d.Id).Msg("sending delete fail response")
		c.Conn.Write(NoJobResponseBytes)
	}
}

func (c *Client) handleAbortMessage(a *AbortRequest) {
	log.Trace().Uint64("id", a.Id).Msg("got abort request")

	c.lock.Lock()
	defer c.lock.Unlock()

	j := c.AssignedJobs[a.Id]
	if j == nil {
		log.Warn().Uint64("client_id", a.Id).Uint64("client_id", c.Id).Msg("attempted to abort unowned job")
		c.Conn.Write(NoJobResponseBytes)
		return
	}

	delete(c.AssignedJobs, a.Id)
	log.Info().Uint64("job_id", j.Id).Str("queue", j.Queue).Msg("requeued job")
	c.Server.GetOrCreateQueue(j.Queue).QueueJob(j)
	c.Conn.Write(EmptyOkResponseBytes)
}

func (c *Client) DeleteJob(id uint64) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.AssignedJobs[id] != nil {
		delete(c.AssignedJobs, id)
		return true
	} else {
		return false
	}
}

func (c *Client) assignJob(je *JobEntry) {
	log.Info().Uint64("client_i", c.Id).Uint64("job_id", je.Id).Int("priority", je.Priority).Msg("assigning job")
	c.lock.Lock()
	defer c.lock.Unlock()
	c.AssignedJobs[je.Id] = je
}
