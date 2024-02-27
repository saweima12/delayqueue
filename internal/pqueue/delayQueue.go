package pqueue

import (
	"time"
)

type Delayer interface {
	SetExpiration(d int64) bool
	Expiration() int64
}

type DelayQueue[T Delayer] interface {
	Offer(item T)
	ExpiredCh() <-chan T
	Len() int
	Start()
	Stop()
}

type delayQueue[T Delayer] struct {
	pq *PriorityQueue[T]
	t  *time.Timer

	nowFunc func() int64

	expireCh chan T
	offerCh  chan T
	stopCh   chan struct{}
}

func delayerLess[T Delayer](i, j T) bool {
	return i.Expiration() < j.Expiration()
}

func NewDelayQueue[T Delayer](nowFunc func() int64, size int) DelayQueue[T] {
	initTimer := time.NewTimer(time.Hour * 24)
	initTimer.Stop()

	return &delayQueue[T]{
		pq:      NewPriorityQueue[T](delayerLess, size),
		t:       initTimer,
		nowFunc: nowFunc,

		expireCh: make(chan T, 1),
		offerCh:  make(chan T, 1),
		stopCh:   make(chan struct{}, 1),
	}
}

func (dq *delayQueue[T]) Offer(item T) {
	dq.offerCh <- item
}

func (dq *delayQueue[T]) ExpiredCh() <-chan T {
	return dq.expireCh
}

func (dq *delayQueue[T]) Len() int {
	return dq.pq.Len()
}

func (dq *delayQueue[T]) Start() {
	go dq.run()
}

func (dq *delayQueue[T]) Stop() {
	dq.stopCh <- struct{}{}
}

func (dq *delayQueue[T]) run() {
	for {
		select {
		case <-dq.stopCh:
			return
		case nOffer := <-dq.offerCh:
			dq.handleOffer(nOffer)
		case <-dq.t.C:
			// When current event timeout, Pop expired item and push into expireCh
			dq.handleTimeout()
		}
	}
}

func (dq *delayQueue[T]) handleOffer(item T) {
	topExp := int64(0)

	oldLen := dq.Len()
	if oldLen > 0 {
		topExp = int64(dq.pq.Peek().Expiration())
	}
	dq.pq.Push(item)

	if item.Expiration() < topExp || oldLen == 0 {
		// when offer a "earlier" event, refresh timer.
		dq.refreshTimer(item.Expiration())
	}
}

func (dq *delayQueue[T]) handleTimeout() {
	now := dq.nowFunc()

	for dq.pq.Len() > 0 {
		// take the earliest item
		topExp := dq.pq.Peek().Expiration()
		if now < topExp {
			// if current time earlier than task time, reset timer & return.
			dq.refreshTimer(topExp)
			return
		}
		// refershTimer.
		expired := dq.pq.Pop()
		dq.expireCh <- expired
	}
}

func (dq *delayQueue[T]) refreshTimer(expireTime int64) {
	// Ensure the timer is fully reset.
	if !dq.t.Stop() {
		select {
		case <-dq.t.C:
		default:
		}
	}

	delta := expireTime - dq.nowFunc()
	dq.t.Reset(time.Duration(delta))
}

func (dq *delayQueue[T]) isTimeout(expire int64) bool {
	return dq.nowFunc() > expire
}
