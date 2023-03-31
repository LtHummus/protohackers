package jobs

import (
	"container/heap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJobList_Len(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		h := &JobList{}
		heap.Init(h)
		heap.Push(h, &JobEntry{Job: "one", Priority: 1})
		heap.Push(h, &JobEntry{Job: "two", Priority: 2})

		assert.Equal(t, 2, h.Len())
	})

	t.Run("empty", func(t *testing.T) {
		h := &JobList{}
		heap.Init(h)

		assert.Equal(t, 0, h.Len())
	})
}

func TestJobList_Less(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		h := &JobList{}
		heap.Init(h)
		heap.Push(h, &JobEntry{Job: "one", Priority: 1})
		heap.Push(h, &JobEntry{Job: "thousand", Priority: 1000})
		heap.Push(h, &JobEntry{Job: "two", Priority: 2})

		top := heap.Pop(h).(*JobEntry)

		assert.Equal(t, "thousand", top.Job)
		assert.Equal(t, 1000, top.Priority)

		next := heap.Pop(h).(*JobEntry)

		assert.Equal(t, "two", next.Job)
		assert.Equal(t, 2, next.Priority)
	})

	t.Run("complex", func(t *testing.T) {
		h := &JobList{}
		heap.Init(h)

		heap.Push(h, &JobEntry{Job: "a", Priority: 9763})
	})
}
