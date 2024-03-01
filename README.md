# DelayWheel 

A hierarchical timing wheel implemented in Go, based on DelayQueue and supporting graceful shutdown.

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
	dw.ScheduleFunc(3*time.Second, func(taskCtx *delaywheel.TaskCtx) {
		fmt.Println(taskCtx)
	})

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

