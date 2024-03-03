package delaywheel_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/saweima12/delaywheel"
)

func TestCreateUnavaliableWheel(t *testing.T) {
	_, err := delaywheel.New(time.Microsecond, 20)
	t.Run("Submission should fail, because tickMs is less than 1 ms", func(t *testing.T) {
		if err == nil {
			t.Fail()
		}
	})
}

func TestDelayWheelWithWorker(t *testing.T) {
	// Create and startup the delaywheel
	dw, _ := delaywheel.New(1*time.Second, 20,
		delaywheel.WithLogLevel(delaywheel.LOG_DEBUG),
	)

	t.Run("Submission should fail because it has not started yet", func(t *testing.T) {
		// Insert a new task
		_, err := dw.AfterFunc(time.Second, func(taskCtx *delaywheel.TaskCtx) {})
		if err == nil {
			t.Fail()
		}
	})
	dw.Start()

	t.Run("WaitForDone should ensure that the tasks run properly.", func(t *testing.T) {
		rtn := int32(0)
		// Insert a scheduled task to execute every 3 seconds.
		dw.ScheduleFunc(3*time.Second, func(taskCtx *delaywheel.TaskCtx) {
			atomic.AddInt32(&rtn, 1)
			fmt.Println("EventTrigger")
		})

		// Launch a separate goroutine for executing tasks, which can be replaced with a goroutine pool.
		done := make(chan struct{}, 1)
		go func() {
			for task := range dw.PendingChan() {
				task()
			}
			done <- struct{}{}
		}()

		// Shutdown the delaywheel
		dw.Stop(func(ctx *delaywheel.StopCtx) error {
			ctx.WaitForDone(context.Background())
			return nil
		})

		<-done

		if atomic.LoadInt32(&rtn) != 1 {
			t.Errorf("Because it stops shortly after starting, ScheduleFunc should only trigger once. got: %v", rtn)
			t.Fail()
		}
	})
}

func TestBasicFunctionality(t *testing.T) {
	dw, err := delaywheel.New(time.Millisecond, 20,
		delaywheel.WithAutoRun(),
	)
	if err != nil {
		t.Fatalf("Failed to create DelayWheel: %v", err)
	}

	dw.Start()

	var wg sync.WaitGroup
	wg.Add(1)

	_, err = dw.AfterFunc(2*time.Millisecond, func(ctx *delaywheel.TaskCtx) {
		wg.Done()
	})

	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	wg.Wait()

	dw.Stop(nil)
}

func TestLargeTasks(t *testing.T) {
	dw, err := delaywheel.New(1*time.Millisecond, 20,
		delaywheel.WithCurTaskID(10),
		delaywheel.WithPendingBufferSize(10),
	)
	if err != nil {
		t.Fatalf("Failed to create DelayWheel: %v", err)
	}
	dw.Start()

	go func() {
		for f := range dw.PendingChan() {
			f()
		}
	}()

	rtn := int32(0)
	// Simulate concurrent tasks.
	var wg sync.WaitGroup
	taskCount := 100000
	wg.Add(taskCount)
	for i := 0; i < taskCount; i++ {
		_, err := dw.AfterFunc(time.Duration(10)*time.Millisecond, func(ctx *delaywheel.TaskCtx) {
			atomic.AddInt32(&rtn, 1)
			wg.Done()
		})
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	dw.Stop(func(stopCtx *delaywheel.StopCtx) error {
		stopCtx.WaitForDone(context.Background())
		return nil
	})

	fmt.Println("Count", rtn)
	if rtn != int32(taskCount) {
		t.Errorf("The result must be %d, got %d", taskCount, rtn)
		t.Fail()
	}

	wg.Wait()
}

// TestGracefulShutdownWithoutTimeout tests stopping the DelayWheel gracefully and waits for all tasks to complete.
func TestGracefulShutdownWithoutTimeout(t *testing.T) {
	dw, err := delaywheel.New(1*time.Millisecond, 20,
		delaywheel.WithAutoRun(),
		delaywheel.WithLogLevel(delaywheel.LOG_DEBUG),
	)
	if err != nil {
		t.Fatalf("Failed to create DelayWheel: %v", err)
	}

	dw.Start()

	answer := 10
	result := uint32(0)
	// Add some tasks that will print a message
	for i := 0; i < 10; i++ {
		_, err := dw.AfterFunc(time.Duration(i)*time.Millisecond, func(ctx *delaywheel.TaskCtx) {
			// Task logic here
			atomic.AddUint32(&result, 1)
		})

		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	// Testing cancelTask
	var fail atomic.Bool
	taskId, err := dw.AfterFunc(2*time.Second, func(ctx *delaywheel.TaskCtx) {
		fail.Swap(false)
	})
	dw.CancelTask(taskId)

	// Stop the DelayWheel and wait for all tasks to complete
	dw.Stop(func(ctx *delaywheel.StopCtx) error {
		ctx.WaitForDone(context.Background()) // Pass a fresh context to avoid mixing with the outer one
		return nil
	})

	if fail.Load() {
		t.Errorf("The fail trigger task, must be canceled.")
		t.Fail()
	}

	if result != uint32(answer) {
		t.Errorf("The rtn must be %d, got %d", answer, result)
		t.Fail()
	}
}

// TestGracefulShutdownWithTimeout tests the behavior of stopping the DelayWheel gracefully under a timeout condition.
func TestGracefulShutdownWithTimeout(t *testing.T) {
	dw, err := delaywheel.New(1*time.Millisecond, 20, delaywheel.WithAutoRun())
	if err != nil {
		t.Fatalf("Failed to create DelayWheel: %v", err)
	}
	dw.Start()

	// Add some tasks
	for i := 0; i < 10; i++ {
		n := i
		_, err := dw.AfterFunc(time.Duration(n)*100*time.Millisecond, func(ctx *delaywheel.TaskCtx) {
			// Longer task to simulate timeout
		})

		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	// Create a context with timeout
	cctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Stop the DelayWheel with the context that has a timeout
	err = dw.Stop(func(ctx *delaywheel.StopCtx) error {
		// Wait for tasks to complete or for the context to be canceled due to timeout
		ctx.WaitForDone(cctx)
		if cctx.Err() != nil {
			return cctx.Err()
		}
		return nil
	})

	if err == nil {
		t.Fatal("Expected Stop to fail due to timeout, but it did not")
	}
}

type TestLogger struct{}

func (te *TestLogger) Debug(format string, args ...any) {
	ptn := format + "\n"
	fmt.Printf(ptn, args...)
}

func (te *TestLogger) Info(format string, args ...any) {
	ptn := format + "\n"
	fmt.Printf(ptn, args...)
}

func (te *TestLogger) Error(format string, args ...any) {
	ptn := format + "\n"
	fmt.Printf(ptn, args...)
}

func (te *TestLogger) Warn(format string, args ...any) {
	ptn := format + "\n"
	fmt.Printf(ptn, args...)
}

type TestExecutor struct {
	Count int32
}

func (te *TestExecutor) Execute(taskCtx *delaywheel.TaskCtx) {
	fmt.Println(taskCtx.TaskID(), taskCtx.Expiration(), taskCtx.ExpireTime())
	atomic.AddInt32(&te.Count, 1)

	if te.Count > 10 {
		taskCtx.Cancel()
	}
}

func TestGrcefulShutdownWithCancelTask(t *testing.T) {
	dw, err := delaywheel.New(time.Millisecond, 20,
		delaywheel.WithAutoRun(),
		delaywheel.WithLogger(&TestLogger{}),
	)
	if err != nil {
		fmt.Println(err)
	}

	t.Run("The tasks will be failed, because the delaywheel not running", func(t *testing.T) {
		_, err := dw.AfterFunc(time.Second, func(task *delaywheel.TaskCtx) {
			// This task will be failed.
		})

		if err == nil {
			t.Fail()
		}

		_, err = dw.ScheduleFunc(time.Second, func(task *delaywheel.TaskCtx) {
			// This task will be failed.
		})

		if err == nil {
			t.Fail()
		}
	})

	dw.Start()

	t.Run("The executor count should be 1", func(t *testing.T) {
		executor := TestExecutor{}
		_, err := dw.AfterExecute(time.Second, &executor)
		if err != nil {
			t.Fail()
		}
		<-time.After(4 * time.Second)
		if executor.Count != 1 {
			t.Fail()
		}
	})

	t.Run("The executor count should be 3", func(t *testing.T) {
		executor := TestExecutor{}
		taskId, err := dw.ScheduleExecute(time.Second, &executor)
		if err != nil {
			t.Fail()
		}
		<-time.After(3 * time.Second)
		dw.CancelTask(taskId)
		<-time.After(time.Second)
		if executor.Count != 3 {
			t.Fail()
		}
	})

}
