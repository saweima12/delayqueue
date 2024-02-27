package delaywheel

import "sync/atomic"

type tWheel struct {
	tickMs    int64
	wheelSize int
	interval  int64
	curTime   atomic.Int64
	buckets   []*bucket

	next *tWheel
}

func newWheel(tickMs int64, wheelSize int, curTime int64) *tWheel {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}

	curTime = truncate(curTime, tickMs)

	result := &tWheel{
		tickMs:    tickMs,
		wheelSize: wheelSize,
		interval:  tickMs * int64(wheelSize),
		buckets:   buckets,
	}
	result.curTime.Store(curTime)

	return result
}

func (tw *tWheel) addTask(task *Task) *bucket {
	blocks := task.expiration / tw.tickMs
	bucket := tw.buckets[blocks%int64(tw.wheelSize)]

	bucket.AddTask(task)
	return bucket
}

func (tw *tWheel) advanceClock(expiration int64) {
	curTime := tw.curTime.Load()
	if expiration < curTime+tw.tickMs {
		return
	}
	curTime = truncate(expiration, tw.tickMs)
	tw.curTime.Store(curTime)

	if tw.next != nil {
		tw.next.advanceClock(curTime)
	}
}
