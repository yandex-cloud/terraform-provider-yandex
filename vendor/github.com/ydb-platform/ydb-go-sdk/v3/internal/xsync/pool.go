package xsync

import "sync"

// pool interface uses for testing with mock or standard sync.Pool for runtime
type pool interface {
	Get() (v any)
	Put(v any)
}

type Pool[T any] struct {
	p   pool
	New func() *T

	once sync.Once
}

func (p *Pool[T]) init() {
	p.once.Do(func() {
		if p.p == nil {
			p.p = &sync.Pool{}
		}
	})
}

func (p *Pool[T]) GetOrNew() *T {
	p.init()

	if v := p.p.Get(); v != nil {
		return v.(*T) //nolint:forcetypeassert
	}

	if p.New != nil {
		return p.New()
	}

	return new(T)
}

func (p *Pool[T]) GetOrNil() *T {
	p.init()

	if v := p.p.Get(); v != nil {
		return v.(*T) //nolint:forcetypeassert
	}

	return nil
}

func (p *Pool[T]) Put(t *T) {
	p.init()

	p.p.Put(t)
}
