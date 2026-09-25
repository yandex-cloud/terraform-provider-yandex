package xsync

import (
	"sync"
)

func (guard *lastUsage) Start() (stop func()) {
	guard.locks.Add(1)

	return sync.OnceFunc(func() {
		if guard.locks.Add(-1) == 0 {
			now := guard.clock.Now()
			guard.t.Store(&now)
		}
	})
}
