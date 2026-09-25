package xsync

type Value[T any] struct {
	v  T
	mu RWMutex
}

func NewValue[T any](initValue T) *Value[T] {
	return &Value[T]{v: initValue}
}

func (v *Value[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.v
}

func (v *Value[T]) Change(change func(old T) T) {
	v.mu.WithLock(func() {
		v.v = change(v.v)
	})
}
