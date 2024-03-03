package delaywheel

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// Executor contains an Execute() method,
// > where TaskCtx is passed in to obtain the relevant parameters of the current task.
type Executor interface {
	Execute(taskCtx *TaskCtx)
}

type taskPool struct {
	pool sync.Pool
}

func newTaskPool() *taskPool {
	return &taskPool{
		pool: sync.Pool{
			New: func() any {
				return new(Task)
			},
		},
	}
}

func (tp *taskPool) Get() *Task {
	return tp.pool.Get().(*Task)
}

func (tp *taskPool) Put(t *Task) {
	t.taskID = 0
	t.expiration = 0
	t.interval = 0
	t.state.Store(0)

	t.executor = nil
	t.elm = nil
	t.wg = nil
	t.bucket = nil
	t.de = nil
	tp.pool.Put(t)
}

type taskState uint32

const (
	taskReady taskState = iota
	taskExecuted
	taskCanceled
)

type Task struct {
	taskID     uint64
	expiration int64
	interval   time.Duration
	state      atomic.Uint32
	mu         sync.Mutex

	executor Executor
	elm      *list.Element
	wg       *sync.WaitGroup
	bucket   *bucket
	de       *DelayWheel
}

// Get the taskID
func (dt *Task) TaskID() uint64 {
	return dt.taskID
}

// Get the task expiration.
func (dt *Task) Expiration() int64 {
	return dt.expiration
}

// Execute the task;
// Notice: The task will self-recycle and clear relevant data after execution.
func (dt *Task) Execute() {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if dt.state.Load() != uint32(taskReady) {
		dt.waitDone()
		return
	}

	dt.run()

	dt.state.Store(uint32(taskExecuted))
	dt.waitDone()

	dt.de.recycleTask(dt)
}

// Cancel the task
func (dt *Task) Cancel() {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if dt.IsCanceled() {
		return
	}
	dt.state.Store(uint32(taskCanceled))
}

func (dt *Task) IsCanceled() bool {
	return dt.state.Load() == uint32(taskCanceled)
}

// Get the executor.
func (dt *Task) Executor() Executor {
	return dt.executor
}

func (dt *Task) run() (isSchedule bool) {
	ctx := dt.de.createContext(dt)
	if dt.interval > 0 {
		ctx.ReSchedule(dt.interval)
	}

	dt.executor.Execute(ctx)
	result := ctx.isScheduled
	dt.de.recycleContext(ctx)
	return result
}

func (dt *Task) setWaitGroup(wg *sync.WaitGroup) bool {
	if dt.IsCanceled() {
		return false
	}
	dt.mu.Lock()
	dt.wg = wg
	dt.mu.Unlock()
	return true
}

func (dt *Task) waitDone() {
	if dt.wg == nil {
		return
	}

	dt.wg.Done()
	dt.wg = nil
}

// Create a simple executor function wrapper.
func pureExec(f func(task *TaskCtx)) *pureExecutor {
	return &pureExecutor{
		f: f,
	}
}

type pureExecutor struct {
	f func(task *TaskCtx)
}

func (we *pureExecutor) Execute(task *TaskCtx) {
	we.f(task)
}
