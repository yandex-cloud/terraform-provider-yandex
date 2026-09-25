package xsync

import (
	"context"

	"golang.org/x/sync/semaphore"
)

// SoftWeightedSemaphore extends semaphore.Weighted with ability to acquire
// one request over capacity if semaphore is completely free
type SoftWeightedSemaphore struct {
	sem      *semaphore.Weighted
	capacity int64
	mu       Mutex
	overflow int64 // Number of tokens in overflow state
}

// NewSoftWeightedSemaphore creates new SoftWeightedSemaphore with given capacity
func NewSoftWeightedSemaphore(n int64) *SoftWeightedSemaphore {
	return &SoftWeightedSemaphore{
		sem:      semaphore.NewWeighted(n),
		capacity: n,
	}
}

// Acquire acquires the semaphore with a weight of n.
// If the semaphore is completely free, the acquisition will succeed regardless of weight.
func (s *SoftWeightedSemaphore) Acquire(ctx context.Context, n int64) error {
	// If request doesn't exceed capacity, use normal path
	if n <= s.capacity {
		return s.sem.Acquire(ctx, n)
	}

	// For large requests, try to acquire entire semaphore
	if err := s.sem.Acquire(ctx, s.capacity); err != nil {
		return err
	}

	s.setOverflow(n)

	return nil
}

// Release releases n tokens back to the semaphore.
func (s *SoftWeightedSemaphore) Release(n int64) {
	s.mu.WithLock(func() {
		if n >= s.overflow {
			n -= s.overflow
			s.overflow = 0
			if n > 0 {
				s.sem.Release(n)
			}
		} else {
			s.overflow -= n
		}
	})
}

// TryAcquire tries to acquire the semaphore with a weight of n without blocking.
// If the semaphore is completely free, the acquisition will succeed regardless of weight.
func (s *SoftWeightedSemaphore) TryAcquire(n int64) bool {
	// If request doesn't exceed capacity, use normal path
	if n <= s.capacity {
		return s.sem.TryAcquire(n)
	}

	// For large requests, try to acquire entire semaphore
	if !s.sem.TryAcquire(s.capacity) {
		return false
	}

	s.setOverflow(n)

	return true
}

// setOverflow sets the overflow value when entire semaphore is successfully acquired
func (s *SoftWeightedSemaphore) setOverflow(n int64) {
	s.mu.WithLock(func() {
		s.overflow = n - s.capacity
	})
}
