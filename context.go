package delaywheel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type ctxPool struct {
	pool sync.Pool
}

func newCtxPool() *ctxPool {
	return &ctxPool{
		pool: sync.Pool{
			New: func() any {
				return new(TaskCtx)
			},
		},
	}
}

func (c *ctxPool) Get(t *Task) *TaskCtx {
	return c.pool.Get().(*TaskCtx)
}

func (c *ctxPool) Put(ctx *TaskCtx) {
	ctx.t = nil
	ctx.taskCh = nil
	ctx.isScheduled = false
	c.pool.Put(ctx)
}

type TaskCtx struct {
	t      *Task
	taskCh chan *Task

	mu          sync.Mutex
	isScheduled bool
}

func (ctx *TaskCtx) IsCancelled() bool {
	return ctx.t.isCancelled.Load()
}

func (ctx *TaskCtx) Cancel() {
	ctx.t.isCancelled.Store(true)
}

func (ctx *TaskCtx) TaskID() uint64 {
	return ctx.t.taskID
}

func (ctx *TaskCtx) Expiration() int64 {
	return ctx.t.expiration
}

func (ctx *TaskCtx) ExpireTime() time.Time {
	return msToTime(ctx.t.expiration)
}

func (ctx *TaskCtx) ReSchedule(d time.Duration) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.isScheduled {
		return
	}
	newExp := ctx.ExpireTime().Add(d)
	atomic.SwapInt64(&ctx.t.expiration, timeToMs(newExp))

	ctx.taskCh <- ctx.t
	ctx.isScheduled = true
}

type StopCtx struct {
	de *DelayWheel
}

func (ctx *StopCtx) GetAllTask() []*Task {
	rtn := make([]*Task, 0)
	for tuple := range ctx.de.getAllTask() {
		rtn = append(rtn, tuple.Value)
	}
	return rtn
}

func (ctx *StopCtx) WaitForDone(inputCtx context.Context) {
	timer := time.NewTimer(0)
	for {
		exp, isEmpty := ctx.getMaxExpiration()
		if isEmpty {
			break
		}
		// Calculate the watting time.
		now := timeToMs(time.Now().UTC())
		if !refreshTimer(timer, now, exp) {
			break
		}
		// Waitting for expiration.
		select {
		case <-timer.C:
			continue
		case <-inputCtx.Done():
			return
		}
	}

	done := make(chan struct{}, 1)
	go func() {
		ctx.de.wg.Wait()
		done <- struct{}{}
	}()

	timer.Stop()
	// Wait for all tasks to complete or for the context to be canceled.
	select {
	case <-done:
	case <-inputCtx.Done():
	}

}

func (ctx *StopCtx) getMaxExpiration() (ms int64, isEmpty bool) {
	rtn := int64(-1)
	isFound := false

	for tuple := range ctx.de.getAllTask() {
		if tuple.Value.isCancelled.Load() {
			continue
		}
		if tuple.Value.expiration > rtn {
			rtn = tuple.Value.expiration
			isFound = true
		}
	}

	return rtn, !isFound
}
