package pqueue_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/saweima12/delaywheel/internal/pqueue"
)

type TestDelayerItem struct {
	expiration int64
	Name       string
}

func (te *TestDelayerItem) SetExpiration(d int64) bool {
	return atomic.SwapInt64(&te.expiration, d) != d

}

func (te *TestDelayerItem) Expiration() int64 {
	return te.expiration
}

func TestDelayQueue(t *testing.T) {
	fmt.Println("==Start to test DelayQueue==")
	dq := pqueue.NewDelayQueue[*TestDelayerItem](func() int64 {
		return timeToMs(time.Now().UTC())
	}, 64)
	dq.Start()

	nb := TestDelayerItem{
		expiration: timeToMs(time.Now().Add(1 * time.Second)),
		Name:       "John",
	}
	nb2 := TestDelayerItem{
		expiration: timeToMs(time.Now().Add(2 * time.Second)),
		Name:       "Lee",
	}
	dq.Offer(&nb)
	dq.Offer(&nb2)

	rtn := []string{}

	for dq.Len() > 0 {
		item := Poll(dq.ExpiredCh())
		rtn = append(rtn, item.Name)
	}

	answer := []string{"John", "Lee"}
	ok := compareArr(rtn, answer)
	if !ok {
		t.Errorf("The result must be %v, got %v", answer, rtn)
		t.Fail()

	}

	dq.Stop()
}

func Poll(ch <-chan *TestDelayerItem) *TestDelayerItem {
	select {
	case item := <-ch:
		return item
	}
}

func compareArr[T comparable](arr1, arr2 []T) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	for i := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}

	return true
}

// timeToMs returns an integer number, which represents t in milliseconds.
func timeToMs(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// msToTime returns the UTC time corresponding to the given Unix time,
// t milliseconds since January 1, 1970 UTC.
func msToTime(t int64) time.Time {
	return time.Unix(0, t*int64(time.Millisecond)).UTC()
}
