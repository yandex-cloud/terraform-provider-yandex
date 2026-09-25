package xsync

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Map[K comparable, V any] struct {
	m    sync.Map
	size atomic.Int32
}

func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, false
	}
	value, ok = v.(V)

	return value, ok
}

func (m *Map[K, V]) Must(key K) (value V) {
	v, ok := m.m.Load(key)
	if !ok {
		panic(fmt.Sprintf("unexpected key = %v", key))
	}
	value, ok = v.(V)
	if !ok {
		panic(fmt.Sprintf("unexpected type of value = %T", v))
	}

	return value
}

func (m *Map[K, V]) Has(key K) bool {
	_, ok := m.m.Load(key)

	return ok
}

func (m *Map[K, V]) Set(key K, value V) {
	_, exists := m.m.LoadOrStore(key, value)

	if !exists {
		m.size.Add(1)
	}
}

func (m *Map[K, V]) Delete(key K) bool {
	_, exists := m.Extract(key)

	if !exists {
		m.size.Add(-1)
	}

	return exists
}

func (m *Map[K, V]) Len() int {
	return int(m.size.Load())
}

func (m *Map[K, V]) Extract(key K) (value V, ok bool) {
	v, exists := m.m.LoadAndDelete(key)
	if !exists {
		return value, false
	}

	m.size.Add(-1)

	value, ok = v.(V)

	return value, ok
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V)) //nolint:forcetypeassert
	})
}

func (m *Map[K, V]) Clear() (removed int) {
	m.m.Range(func(k, v any) bool {
		removed++

		m.m.Delete(k)

		return true
	})

	m.size.Add(int32(-removed))

	return removed
}
