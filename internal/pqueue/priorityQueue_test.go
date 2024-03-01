package pqueue_test

import (
	"fmt"
	"testing"

	"github.com/saweima12/delaywheel/internal/pqueue"
)

type TestItem struct {
	Num int
}

func LessItem(i, j *TestItem) bool {
	return i.Num < j.Num
}

func TestProirityQueue(t *testing.T) {
	fmt.Println("==Start to test PriorityQueue==")
	q := pqueue.NewPriorityQueue(LessItem, 64)
	q.Push(&TestItem{Num: 10})
	q.Push(&TestItem{Num: 20})
	q.Push(&TestItem{Num: 30})
	q.Push(&TestItem{Num: 40})
	q.Push(&TestItem{Num: -10})
	q.Push(&TestItem{Num: -20})
	q.Push(&TestItem{Num: -50})

	popNum := q.Pop().Num
	if popNum != -50 {
		t.Errorf("The pop number must be -50, val: %d", popNum)
	}

	if q.Peek().Num != -20 {
		t.Errorf("The peek number must be -20, val: %d", q.Peek().Num)
	}

	l := q.Len()
	for i := 0; i < l; i++ {
		fmt.Println(q.Pop(), q.Len())
	}
}
