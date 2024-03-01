package delaywheel_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/saweima12/delaywheel"
)

func TestBasicFunctionality(t *testing.T) {
	dw, err := delaywheel.New(1*time.Millisecond, 20,
		delaywheel.WithAutoRun(),
	)
	if err != nil {
		t.Fatalf("Failed to create DelayWheel: %v", err)
	}

	dw.Start()

	var wg sync.WaitGroup
	wg.Add(1)

	_, err = dw.AfterFunc(10*time.Millisecond, func(ctx *delaywheel.TaskCtx) {
		wg.Done()
	})
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	wg.Wait()

	dw.Stop(nil)
}

func TestConcurrentSafety(t *testing.T) {
	dw, err := delaywheel.New(1*time.Millisecond, 20,
		delaywheel.WithAutoRun(),
	)
	if err != nil {
		t.Fatalf("Failed to create DelayWheel: %v", err)
	}
	dw.Start()

	var wg sync.WaitGroup
	taskCount := 100
	wg.Add(taskCount)
	for i := 0; i < taskCount; i++ {
		_, err := dw.AfterFunc(time.Duration(i+1)*time.Millisecond, func(ctx *delaywheel.TaskCtx) {
			wg.Done()
		})
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	wg.Wait()

	dw.Stop(nil)
}

// TestGracefulShutdownWithoutTimeout tests stopping the DelayWheel gracefully and waits for all tasks to complete.
func TestGracefulShutdownWithoutTimeout(t *testing.T) {
	dw, err := delaywheel.New(1*time.Millisecond, 20, delaywheel.WithAutoRun())
	if err != nil {
		t.Fatalf("Failed to create DelayWheel: %v", err)
	}

	dw.Start()

	// Add some tasks that will print a message
	for i := 0; i < 10; i++ {
		n := i
		_, err := dw.AfterFunc(time.Duration(i)*time.Second, func(ctx *delaywheel.TaskCtx) {
			// Task logic here
			fmt.Println("WithOutTimeout", n)
		})
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
		fmt.Println("Register WithOutTimeout", n)
	}

	fmt.Println("Start to stop WithOutTimeout")
	// Stop the DelayWheel and wait for all tasks to complete
	dw.Stop(func(ctx *delaywheel.StopCtx) error {
		ctx.WaitForDone(context.Background()) // Pass a fresh context to avoid mixing with the outer one
		return nil
	})
	fmt.Println("Finsihed")
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
		taskId, err := dw.AfterFunc(time.Duration(n)*100*time.Millisecond, func(ctx *delaywheel.TaskCtx) {
			// Longer task to simulate timeout
			fmt.Println("WithTimeout:", n)
		})
		fmt.Println("Register WithTimout", n, taskId)

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
		fmt.Println("Start to stop WithTimeout")
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
