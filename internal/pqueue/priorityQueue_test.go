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

	q.Init()

	popNum := q.Pop().Num
	if popNum != -50 {
		t.Errorf("The pop number must be -50, val: %d", popNum)
		t.Fail()
	}

	if q.Peek().Num != -20 {
		t.Errorf("The peek number must be -20, val: %d", q.Peek().Num)
		t.Fail()
	}

	answer := []int{-20, -10, 10, 20, 30, 40}
	rtn := []int{}
	l := q.Len()
	for i := 0; i < l; i++ {
		item := q.Pop()
		rtn = append(rtn, item.Num)
	}

	ok := compareArr(rtn, answer)
	if !ok {
		t.Errorf("The result should be %v, got %v", answer, rtn)
		t.Fail()
	}

}
