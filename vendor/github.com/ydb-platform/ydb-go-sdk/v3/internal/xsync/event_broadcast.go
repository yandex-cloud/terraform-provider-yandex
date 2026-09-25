package xsync

import (
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
)

// EventBroadcast is implementation of broadcast notify about event
// Zero value is usable, must not copy after first call any method
type EventBroadcast struct {
	m sync.Mutex

	nextEventChannel empty.Chan
}

func (b *EventBroadcast) initNeedLock() {
	if b.nextEventChannel == nil {
		b.nextEventChannel = make(empty.Chan)
	}
}

// Waiter return channel, that will close when next event will be broadcast.
// For prevent race between subscribe and event client code must subscribe at first, then check condition
// if false - wait closing channed and check condition again
func (b *EventBroadcast) Waiter() OneTimeWaiter {
	b.m.Lock()
	defer b.m.Unlock()

	b.initNeedLock()

	return OneTimeWaiter{b.nextEventChannel}
}

func (b *EventBroadcast) Broadcast() {
	b.m.Lock()
	defer b.m.Unlock()

	b.initNeedLock()

	close(b.nextEventChannel)
	b.nextEventChannel = make(empty.Chan)
}

type OneTimeWaiter struct {
	ch empty.Chan
}

func (w *OneTimeWaiter) Done() <-chan struct{} {
	return w.ch
}
