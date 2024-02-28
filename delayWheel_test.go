package delaywheel_test

import (
	"fmt"
	"testing"
	"time"
	"unsafe"

	"github.com/saweima12/delaywheel"
)

func TestAfterFuncAndAutoRun(t *testing.T) {
	// Create delaywheel and start it.
	dw, err := delaywheel.New(
		time.Millisecond, 20,
		delaywheel.WithAutoRun(),
	)
	if err != nil {
		t.Error(err)
		return
	}
	dw.Start()
	fmt.Println("delaywheel size:", unsafe.Sizeof(*dw))

	dw.AfterFunc(time.Second, func(ctx *delaywheel.TaskCtx) {
		fmt.Println("AutoRun: Hello")
	})

	dw.AfterFunc(time.Second*3, func(ctx *delaywheel.TaskCtx) {
		fmt.Println("AutoRun: Hello wait 3")
	})

	fmt.Println("Start to shuting down....")
	go dw.Stop(func(ctx *delaywheel.StopCtx) error {
		ctx.WaitForDone()
		return nil
	})
	<-time.After(time.Second)
	_, err = dw.AfterFunc(time.Second*3, func(ctx *delaywheel.TaskCtx) {
		fmt.Println("AutoRun: Hello wait 3")
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Terminated....")
	<-time.After(time.Second * 5)
}

type TestExecutor struct {
	count int
}

func (te *TestExecutor) Execute(task *delaywheel.TaskCtx) {
	te.count += 1
	fmt.Println("Test count", te.count)
}

func TestAfterExec(t *testing.T) {
	// Create delaywheel and start it.
	dw, err := delaywheel.New(time.Millisecond, 20)
	if err != nil {
		t.Error(err)
		return
	}
	dw.Start()

	dw.AfterExecute(time.Second, &TestExecutor{})

	task := <-dw.PendingChan()
	task()
}

func TestScheduleFunc(t *testing.T) {
	// Create delaywheel and start it.
	dw, err := delaywheel.New(time.Millisecond, 20)
	if err != nil {
		t.Error(err)
		return
	}

	dw.Start()
	dw.ScheduleFunc(time.Second*3, func(ctx *delaywheel.TaskCtx) {
		fmt.Println("Hello")
	})

	limit := time.NewTimer(time.Second * 10)
	for {
		select {
		case item := <-dw.PendingChan():
			item()
			fmt.Println("point")
		case <-limit.C:
			fmt.Println("After")
			return
		}
	}

}
