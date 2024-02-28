package delaywheel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/saweima12/delaywheel/internal/pqueue"
	"github.com/saweima12/delaywheel/internal/shardmap"
)

type StopFunc func(ctx *StopCtx) error

func defaultNow() int64 {
	return timeToMs(time.Now().UTC())
}

func New(tick time.Duration, wheelSize int, options ...Option) (*DelayWheel, error) {
	startMs := timeToMs(time.Now().UTC())
	tickMs := int64(tick) / int64(time.Millisecond)

	if tickMs < 1 {
		return nil, fmt.Errorf("The tick must be greater than or equal to 1 Millisecond")
	}

	result := create(
		tickMs,
		wheelSize,
		startMs,
	)

	for _, opt := range options {
		opt(result)
	}

	return result, nil
}

func create(tickMs int64, wheelSize int, startMs int64) *DelayWheel {
	initInterval := tickMs * int64(wheelSize)

	dw := &DelayWheel{
		tickMs:      tickMs,
		wheelSize:   wheelSize,
		startMs:     startMs,
		curInterval: initInterval,
		autoRun:     false,

		taskMap: shardmap.NewNum[uint64, *Task](),
		queue:   pqueue.NewDelayQueue[*bucket](defaultNow, wheelSize),
		wheel:   newWheel(tickMs, wheelSize, startMs),

		pendingTaskCh: make(chan func(), 1),
		addTaskCh:     make(chan *Task, 1),
		recycleTaskCh: make(chan *Task, 1),
		stopCh:        make(chan struct{}, 1),
	}
	dw.curTime.Store(startMs)
	dw.ctxPool = newCtxPool(dw)
	dw.taskPool = newTaskPool()
	dw.isClosing.Store(false)
	return dw
}

type DelayWheel struct {
	wheelSize    int
	tickMs       int64
	curInterval  int64
	startMs      int64
	curTime      atomic.Int64
	curTaskID    atomic.Uint64
	autoRun      bool
	shuttingDown uint32
	isClosing    atomic.Bool
	wg           sync.WaitGroup

	wheel    *tWheel
	taskPool *taskPool
	ctxPool  *ctxPool
	taskMap  *shardmap.Map[uint64, *Task]
	queue    pqueue.DelayQueue[*bucket]

	pendingTaskCh chan func()
	cancelTaskCh  chan uint64
	addTaskCh     chan *Task
	recycleTaskCh chan *Task
	stopCh        chan struct{}
}

// Start the delaywheel.
func (de *DelayWheel) Start() {
	go de.run()
}

// Send a stop signal to delaywheel
func (de *DelayWheel) Stop(stopFunc StopFunc) error {
	de.isClosing.Store(true)

	stopCtx := &StopCtx{
		de: de,
	}
	err := stopFunc(stopCtx)
	if err != nil {
		return err
	}

	de.stopCh <- struct{}{}
	return nil
}

func (de *DelayWheel) PendingChan() <-chan func() {
	return de.pendingTaskCh
}

// Submit a delayed execution of a function.
func (de *DelayWheel) AfterFunc(d time.Duration, f func(task *TaskCtx)) (taskId uint64, err error) {
	t := de.createTask(d)
	t.executor = pureExec(f)
	if err := de.pushTask(t); err != nil {
		return 0, err
	}
	return t.taskID, nil
}

// Submit a delayed execution of a executor.
func (de *DelayWheel) AfterExecute(d time.Duration, executor Executor) (taskId uint64, err error) {
	t := de.createTask(d)
	t.executor = executor
	if err := de.pushTask(t); err != nil {
		return 0, err
	}
	return t.taskID, nil
}

// Schedule a delayed execution of a function with a time interval.
func (de *DelayWheel) ScheduleFunc(d time.Duration, f func(ctx *TaskCtx)) (taskId uint64, err error) {
	// Create the task and wrpper auto reSchedul
	t := de.createTask(d)
	sch := pureScheduler{d: d}
	t.executor = pureExec(func(ctx *TaskCtx) {
		// Calculate new expiration.
		expiration := sch.Next(msToTime(ctx.Expiration()))
		if !expiration.IsZero() {
			ctx.t.expiration = timeToMs(expiration)
			de.addTaskCh <- ctx.t
			ctx.isSechuled = true
		}
		f(ctx)
	})

	if err := de.pushTask(t); err != nil {
		return 0, err
	}
	return t.taskID, nil
}

// Schedule a delayed execution of a executor with a time interval.
func (de *DelayWheel) ScheduleExecute(d time.Duration, executor Executor) (taskId uint64, err error) {
	t := de.createTask(d)

	sch := pureScheduler{d: d}
	t.executor = pureExec(func(ctx *TaskCtx) {
		// Calculate new expiration.
		expiration := sch.Next(msToTime(ctx.t.expiration))
		if !expiration.IsZero() {
			ctx.t.expiration = timeToMs(expiration)
			de.addTaskCh <- ctx.t
			ctx.isSechuled = true
		}
		executor.Execute(ctx)
	})

	if err := de.pushTask(t); err != nil {
		return 0, err
	}
	return t.taskID, nil
}

func (de *DelayWheel) CancelTask(taskID uint64) {
	de.cancelTaskCh <- taskID
}

func (de *DelayWheel) run() {
	de.queue.Start()

	for {
		select {
		case task := <-de.addTaskCh:
			de.addOrRun(task)
		case bu := <-de.queue.ExpiredCh():
			de.handleExipredBucket(bu)
		case task := <-de.recycleTaskCh:
			de.recycleTask(task)
		case taskID := <-de.cancelTaskCh:
			de.cancelTask(taskID)
		case <-de.stopCh:
			de.queue.Stop()
			return
		}
	}
}

// Advance the timing wheel to a specified time point.
func (de *DelayWheel) advanceClock(expiration int64) {
	if expiration < de.curTime.Load() {
		// if the expiration is less than current, ignore it.
		return
	}

	// update currentTime
	de.curTime.Store(expiration)
	de.wheel.advanceClock(expiration)
}

func (de *DelayWheel) pushTask(t *Task) error {
	if de.isClosing.Load() {
		return fmt.Errorf("The delayWheel is shutting down and will not accept new tasks.")
	}
	de.addTaskCh <- t
	return nil
}

func (de *DelayWheel) handleExipredBucket(b *bucket) {
	// Advance wheel's currentTime.
	de.advanceClock(b.Expiration())

	for {
		// Extract all task
		task, ok := b.tasks.PopFront()
		if !ok {
			break
		}
		if task.isCancelled.Load() {
			continue
		}
		de.addOrRun(task)
	}

	b.SetExpiration(-1)
}

func (de *DelayWheel) add(task *Task) bool {
	curTime := de.curTime.Load()
	// When the expiration time less than one tick. execute directly.
	if task.expiration < curTime+de.tickMs {
		return false
	}

	// if the expiration is exceeds the curInterval, expand the wheel.
	if task.expiration > curTime+de.curInterval {
		// When the wheel need be extended, advanceClock to now.
		de.advanceClock(defaultNow())
		for task.expiration > curTime+de.curInterval {
			de.expandWheel()
		}
	}

	// Choose a suitable wheel
	wheel := de.wheel
	for wheel.next != nil {
		if task.expiration <= wheel.curTime.Load()+wheel.interval {
			break
		}
		wheel = wheel.next
	}

	// Insert the task into bucket and setting expiration.
	bucket := wheel.addTask(task)
	if bucket.SetExpiration(task.expiration) {
		de.queue.Offer(bucket)
	}

	// Insert into taskmap
	de.taskMap.Set(task.taskID, task)
	return true
}

func (de *DelayWheel) addOrRun(task *Task) {
	if task.isCancelled.Load() {
		de.recycleTask(task)
		return
	}

	if !de.add(task) {
		// Add into waitGroup.
		task.wg = &de.wg
		de.wg.Add(1)
		// If autoRun is enabled, It will directly start a goroutine for automatic execution.
		if de.autoRun {
			go task.Execute()
			return
		}
		de.pendingTaskCh <- task.Execute
	}
}

func (de *DelayWheel) expandWheel() {
	// Find last wheel
	target := de.wheel
	for target.next != nil {
		target = target.next
	}
	// Loading currnet time & create next layer wheel.
	curTime := de.curTime.Load()
	next := newWheel(target.interval, de.wheelSize, curTime)
	target.next = next

	atomic.StoreInt64(&de.curInterval, next.interval)
}

func (de *DelayWheel) createTask(d time.Duration) *Task {
	t := de.taskPool.Get()
	t.taskID = de.curTaskID.Add(1)
	t.expiration = timeToMs(time.Now().UTC().Add(d))
	t.de = de
	return t
}

func (de *DelayWheel) createContext(t *Task) *TaskCtx {
	ctx := de.ctxPool.Get(t)
	ctx.t = t
	ctx.taskCh = de.addTaskCh
	return ctx
}

func (de *DelayWheel) getAllTask() []*Task {
	rtn := make([]*Task, 0)
	for tuple := range de.taskMap.Iter() {
		rtn = append(rtn, tuple.Value)
	}
	return rtn
}

func (de *DelayWheel) cancelTask(taskID uint64) {
	t, ok := de.taskMap.Get(taskID)
	if !ok {
		return
	}
	t.Cancel()
}

func (de *DelayWheel) recycleTask(t *Task) {
	// Remove task from taskMap
	de.taskMap.Remove(t.taskID)
	de.taskPool.Put(t)
}

func (de *DelayWheel) recycleContext(ctx *TaskCtx) {
	de.ctxPool.Put(ctx)
}
