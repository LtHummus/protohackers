package jobs

import (
	"container/heap"
	"github.com/rs/zerolog/log"
	"math/rand"
	"sync"
	"time"
)

type JobList []*JobEntry

func (h JobList) Len() int           { return len(h) }
func (h JobList) Less(i, j int) bool { return h[j].Priority < h[i].Priority }
func (h JobList) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h JobList) Get(i int) any      { return h[i] }

func (h *JobList) Push(x any) {
	*h = append(*h, x.(*JobEntry))
}

func (h *JobList) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type JobEntry struct {
	Id       uint64
	Queue    string
	Priority int
	Job      any
}

type Queue struct {
	Name string
	Jobs *JobList

	lock sync.Mutex
}

func NewQueue(name string) *Queue {
	jl := &JobList{}
	heap.Init(jl)

	return &Queue{
		Name: name,
		Jobs: jl,
	}
}

func (q *Queue) NextJob() *JobEntry {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.Jobs.Len() == 0 {
		return nil
	}

	next := heap.Pop(q.Jobs)
	return next.(*JobEntry)
}

func (q *Queue) QueueJob(je *JobEntry) *JobEntry {
	q.lock.Lock()
	defer q.lock.Unlock()

	heap.Push(q.Jobs, je)
	log.Info().Str("queue", q.Name).Int("priority", je.Priority).Uint64("job_id", je.Id).Msg("queued job")

	return je
}

func (q *Queue) DeleteJob(id uint64) bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	for i := 0; i < q.Jobs.Len(); i++ {
		j := q.Jobs.Get(i)
		if j.(*JobEntry).Id == id {
			log.Info().Uint64("id", id).Str("queue", q.Name).Msg("deleting job")
			heap.Remove(q.Jobs, i)

			return true
		}
	}

	return false
}

func (q *Queue) PeekJob() *JobEntry {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.Jobs.Len() == 0 {
		return nil
	}
	return q.Jobs.Get(0).(*JobEntry)
}

func generateJobId() uint64 {
	return (uint64(time.Now().UnixMicro()) ^ rand.Uint64()) & uint64(0x7FFFFFFFFFFFFFFF)
}
