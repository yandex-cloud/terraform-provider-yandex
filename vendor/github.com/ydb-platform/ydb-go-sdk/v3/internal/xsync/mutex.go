package xsync

import (
	"sync"
)

type Mutex struct { //nolint:gocritic
	sync.Mutex
}

func (l *Mutex) WithLock(f func()) {
	l.Lock()
	defer l.Unlock()

	f()
}

type RWMutex struct { //nolint:gocritic
	sync.RWMutex
}

func (l *RWMutex) WithLock(f func()) {
	l.Lock()
	defer l.Unlock()

	f()
}

func (l *RWMutex) WithRLock(f func()) {
	l.RLock()
	defer l.RUnlock()

	f()
}

type (
	mutex interface {
		Lock()
		Unlock()
	}
	rwMutex interface {
		RLock()
		RUnlock()
	}
)

func WithLock[T any](l mutex, f func() T) T {
	l.Lock()
	defer l.Unlock()

	return f()
}

func WithRLock[T any](l rwMutex, f func() T) T {
	l.RLock()
	defer l.RUnlock()

	return f()
}
