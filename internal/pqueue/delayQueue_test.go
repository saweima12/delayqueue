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

	nb := TestDelayerItem{expiration: timeToMs(time.Now().Add(5 * time.Second))}
	nb2 := TestDelayerItem{expiration: timeToMs(time.Now().Add(6 * time.Second))}
	dq.Offer(&nb)
	dq.Offer(&nb2)

	for dq.Len() > 0 {
		fmt.Println(Poll(dq.ExpiredCh()))
	}

}

func Poll(ch <-chan *TestDelayerItem) *TestDelayerItem {
	select {
	case item := <-ch:
		return item
	}
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
