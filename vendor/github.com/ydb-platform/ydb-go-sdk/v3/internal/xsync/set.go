package xsync

import (
	"sync"
	"sync/atomic"
)

type Set[T comparable] struct {
	m    sync.Map
	size atomic.Int32
}

func (s *Set[T]) Has(key T) bool {
	_, ok := s.m.Load(key)

	return ok
}

func (s *Set[T]) Add(key T) bool {
	_, exists := s.m.LoadOrStore(key, struct{}{})

	if !exists {
		s.size.Add(1)
	}

	return !exists
}

func (s *Set[T]) Size() int {
	return int(s.size.Load())
}

func (s *Set[T]) Range(f func(key T) bool) {
	s.m.Range(func(k, v any) bool {
		return f(k.(T)) //nolint:forcetypeassert
	})
}

func (s *Set[T]) Values() []T {
	values := make([]T, 0, s.size.Load())

	s.m.Range(func(k, v any) bool {
		values = append(values, k.(T)) //nolint:forcetypeassert

		return true
	})

	return values
}

func (s *Set[T]) Remove(key T) bool {
	_, exists := s.m.LoadAndDelete(key)

	if exists {
		s.size.Add(-1)
	}

	return exists
}

func (s *Set[T]) Clear() (removed int) {
	s.m.Range(func(k, v any) bool {
		removed++

		s.m.Delete(k)

		return true
	})
	s.size.Add(int32(-removed))

	return removed
}
