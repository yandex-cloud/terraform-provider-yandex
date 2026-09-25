package xsync

import (
	"context"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/closer"
)

func OnceFunc(f func(ctx context.Context) error) func(ctx context.Context) error {
	var once sync.Once

	return func(ctx context.Context) (err error) {
		once.Do(func() {
			err = f(ctx)
		})

		return err
	}
}

type Once[T closer.Closer] struct {
	f     func() (T, error)
	once  sync.Once
	mutex sync.RWMutex
	t     T
	err   error
}

func OnceValue[T closer.Closer](f func() (T, error)) *Once[T] {
	return &Once[T]{f: f}
}

func (v *Once[T]) Close(ctx context.Context) (err error) {
	has := true
	v.once.Do(func() {
		has = false
	})

	if has {
		v.mutex.RLock()
		defer v.mutex.RUnlock()

		return v.t.Close(ctx)
	}

	return nil
}

func (v *Once[T]) Get() (T, error) {
	v.once.Do(func() {
		v.mutex.Lock()
		defer v.mutex.Unlock()

		v.t, v.err = v.f()
	})

	v.mutex.RLock()
	defer v.mutex.RUnlock()

	return v.t, v.err
}

func (v *Once[T]) Must() T {
	t, err := v.Get()
	if err != nil {
		panic(err)
	}

	return t
}
