<a href="https://pkg.go.dev/github.com/saweima12/delaywheel"><img src="https://pkg.go.dev/badge/github.com/saweima12/delaywheel.svg" alt="Go Reference"></a>

# DelayWheel 

A Go-implemented, Kafka-style timing wheel that utilizes a DelayQueue to minimize frequent tick operations, making it suitable for managing large volumes of frequent delayed and scheduled tasks. 

It features built-in support for Graceful Shutdown, allowing it to wait for all tasks to finish upon shutdown, thus ensuring tasks are executed correctly.

The aim of this project is to create a universal, efficient, and user-friendly timing wheel solution. 

## Features

- A Kafka-like timing wheel implementation, based on a min heap DelayQueue.
- Supports graceful shutdown, waiting for all tasks to complete before terminating.
- Supports both one-time execution and scheduled execution.
- Supports custom task execution by implementing the Executor interface.
- Features built-in logging functionality, with options to customize the logger used.

## Getting Started

Delaywheel requires Go 1.18 or later. Start by installing in your project:

```sh
go get github.com/saweima12/delaywheel
```

### Exmaple
```go
func main() {
	// Create and startup the delaywheel
	dw, _ := delaywheel.New(1*time.Second, 20,
		delaywheel.WithLogLevel(delaywheel.LOG_DEBUG),
	)
	dw.Start()

	// Insert a new task
	dw.AfterFunc(2*time.Second, func(taskCtx *delaywheel.TaskCtx) {
		fmt.Println(taskCtx)
	})

	// Insert a scheduled task to execute every 3 seconds.
	taskId, _ := dw.ScheduleFunc(3*time.Second, func(taskCtx *delaywheel.TaskCtx) {
		fmt.Println(taskCtx)
	})
    // Cancel the target task.
	wheel.CancelTask(taskId)
    
	// Launch a separate goroutine for executing tasks, 
    // which can be replaced with a goroutine pool.
	go func() {
		for task := range dw.PendingChan() {
			task()
		}
	}()

	// Shutdown the delaywheel
	dw.Stop(func(ctx *delaywheel.StopCtx) error {
		ctx.WaitForDone(context.Background())
		return nil
	})
}
```

### Submit a task

Use AfterFunc to submit a one-time delayed task or ScheduleFunc to submit a regularly scheduled task.

```go
// Insert a new task
dw.AfterFunc(2*time.Second, func(taskCtx *delaywheel.TaskCtx) {
    fmt.Println(taskCtx)
})

// Insert a scheduled task to execute every 3 seconds.
dw.ScheduleFunc(3*time.Second, func(taskCtx *delaywheel.TaskCtx) {
    fmt.Println(taskCtx)
})
```

### Customize the Executor.

After creating a struct and implementing the Executor interface, use AfterExecute and ScheduleExecute to submit tasks. This allows tasks to carry more parameters, enabling the implementation of more complex tasks.

```go
type TestExecutor struct {
	Name string
}

func (te *TestExecutor) Execute(taskCtx *delaywheel.TaskCtx) {
	fmt.Println(taskCtx.TaskID(), te.Name)
}

func main() {
	wheel, _ := delaywheel.New(time.Second, 20, delaywheel.WithAutoRun())
	wheel.Start()
    // Execute once
	wheel.AfterExecute(time.Second, &TestExecutor{Name: "John"})
    // Execute per second
	wheel.ScheduleExecute(time.Second, &TestExecutor{Name: "John"})
}
```

### About TaskCtx

TaskCtx provides auxiliary functions, such as:

- `TaskID()` -> Obtaining the task ID
- `ExpireTime()` -> Getting the expiration time 
- `Cancel()` -> Canceling the task
- `ReSchedule(time.Duration)` -> Rescheduling the task
(Note: ReSchedule will not have an effect when using ScheduleExec and ScheduleFunc)

```go
func (te *TestExecutor) Execute(taskCtx *delaywheel.TaskCtx) {
	fmt.Println(taskCtx.TaskID(), taskCtx.ExpireTime())
	if te.Name == "Herry" {
		taskCtx.Cancel()
	} else {
		taskCtx.ReSchedule(5 * time.Second)
	}
}
```

### About StopCtx

StopCtx assists with graceful shutdown-related functionalities, such as:

- `GetAllTask()` -> Retrieves a list of all scheduled tasks
- `WaitForDone(context.Context)` -> Block until all tasks are completed.
```go
// Wait until all tasks are completed or the context times out, allowing for flexible flow control in combination with the context.
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
wheel.Stop(func(stopCtx *delaywheel.StopCtx) error {
    stopCtx.WaitForDone(ctx)
    return nil
})
```

### About Options

DelayWheel provides a few simple configuration options:
- `WithAutoRun()` -> Automatically launches a goroutine to execute tasks when they expire. This is suitable for scenarios where performance demands are low and there is no desire to introduce a separate goroutine pool.
- `WithCurTaskID()` -> Sets the initial task ID number, with the default being 0. This is useful for cases where synchronization with database IDs is desired.
- `WithPendingBufferSize()` -> Sets the size of the pendingTask channel, with the default being 1. Evaluate the number of tasks and set accordingly, recommended to match the size of the corresponding goroutine pool.
- `WithLogLevel()` -> Sets the logging level, with WARN as the default. Adjusting to INFO or DEBUG can provide more detailed log information.
- `WithLogger()` -> Integrates a custom logger by implementing the Logger interface, making it compatible with other popular logging packages such as logrus, zerolog, or zap.

```go
type TestLogger struct {
	logger *zap.SugaredLogger
}

func newTestLogger() *TestLogger {
	logger, _ := zap.NewProduction()
	return &TestLogger{
		logger: logger.Sugar(),
	}
}

// Implement from delaywheel.Logger
func (te *TestLogger) Debug(format string, args ...any) {
	te.logger.Debugf(format, args...)
}

// Implement from delaywheel.Logger
func (te *TestLogger) Info(format string, args ...any) {
	te.logger.Infof(format, args...)
}

// Implement from delaywheel.Logger
func (te *TestLogger) Error(format string, args ...any) {
	te.logger.Errorf(format, args...)
}

// Implement from delaywheel.Logger
func (te *TestLogger) Warn(format string, args ...any) {
	te.logger.Warnf(format, args...)
}

func main() {
	wheel, _ := delaywheel.New(time.Second, 20,
		delaywheel.WithAutoRun(),
		delaywheel.WithCurTaskID(30),
		delaywheel.WithLogLevel(delaywheel.LOG_ERROR),
		delaywheel.WithPendingBufferSize(30),
		delaywheel.WithLogger(newTestLogger()),
	)
	wheel.Start()
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

