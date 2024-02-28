package delaywheel

import (
	"sync"
	"sync/atomic"
)

type bucket struct {
	expiration int64
	tasks      *genericList[*Task]
	mu         sync.Mutex
}

func newBucket() *bucket {
	return &bucket{
		expiration: -1,
		tasks:      newGeneric[*Task](),
	}
}

func (bu *bucket) SetExpiration(d int64) bool {
	return atomic.SwapInt64(&bu.expiration, d) != d
}

func (bu *bucket) Expiration() int64 {
	return atomic.LoadInt64(&bu.expiration)
}

func (bu *bucket) AddTask(task *Task) {
	bu.mu.Lock()
	elm := bu.tasks.PushBack(task)
	task.elm = elm
	task.bucket = bu
	bu.mu.Unlock()
}
