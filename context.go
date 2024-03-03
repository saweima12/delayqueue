package delaywheel

import (
	"context"
	"sync"
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
	ctx.pushFunc = nil
	ctx.isScheduled = false
	c.pool.Put(ctx)
}

type TaskCtx struct {
	t        *Task
	pushFunc func(*Task) error

	isScheduled bool
	mu          sync.Mutex
}

func (ctx *TaskCtx) IsCancelled() bool {
	return ctx.t.IsCanceled()
}

func (ctx *TaskCtx) Cancel() {
	ctx.t.Cancel()
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

	newTask := ctx.t.de.createTask(d)
	newTask.executor = ctx.t.executor
	newTask.interval = ctx.t.interval
	// Attempt to push the new task into delaywheel
	if err := ctx.pushFunc(newTask); err == nil {
		ctx.isScheduled = true
	}
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
	prevExp := 0

	for {
		exp, isEmpty := ctx.getMaxExpiration()
		if isEmpty {
			break
		}
		// Calculate the watting time.
		now := time.Now().UTC()
		nowMs := timeToMs(now)

		// concurrent scenarios: after the timing wheel retrieves a task but before it is added to the waitGroup,
		// if it exits directly, there's a high chance the last task will not be executed.
		if !refreshTimer(timer, nowMs, exp) {
			if exp == int64(prevExp) {
				break
			}
			// Current solution: Add a delay and wait until the next cycle check.
			nextExp := now.Add(200 * time.Millisecond)
			nextExpMs := timeToMs(nextExp)
			refreshTimer(timer, nowMs, nextExpMs)
			prevExp = int(nextExpMs)
		}

		ctx.de.logger.Info("WaitForDone: %d Ms, exp: %v", time.Duration(exp-nowMs), exp)
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
	rtn := int64(0)
	isFound := false
	count := 0

	for tuple := range ctx.de.getAllTask() {
		count++
		if tuple.Value.IsCanceled() {
			continue
		}
		if tuple.Value.expiration > rtn {
			rtn = tuple.Value.expiration
			isFound = true
		}
	}
	return rtn, !isFound
}
