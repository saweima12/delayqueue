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
	t.executor = nil
	t.interval = 0
	t.isCancelled.Store(false)
	t.elm = nil
	t.bucket = nil
	tp.pool.Put(t)
}

type Task struct {
	taskID      uint64
	expiration  int64
	interval    time.Duration
	isCancelled atomic.Bool
	once        sync.Once

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
	if dt.isCancelled.Load() {
		return
	}

	isSchedule := dt.run()

	if dt.wg != nil {
		dt.wg.Done()
	}

	if !isSchedule {
		dt.de.recycleTaskCh <- dt
	}
}

// Cancel the task
func (dt *Task) Cancel() {
	dt.once.Do(func() {
		if dt.wg != nil {
			dt.wg.Done()
		}
		dt.isCancelled.Store(true)
	})
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
