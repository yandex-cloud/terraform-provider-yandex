package xsync

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
)

// UnboundedChan is a generic unbounded channel implementation that supports
// message merging and concurrent access.
type UnboundedChan[T any] struct {
	signal empty.Chan // buffered channel with capacity 1

	mutex  Mutex
	buffer []T
	closed bool
}

// NewUnboundedChan creates a new UnboundedChan instance.
func NewUnboundedChan[T any]() *UnboundedChan[T] {
	return &UnboundedChan[T]{
		signal: make(empty.Chan, 1),
		buffer: make([]T, 0),
	}
}

// Send adds a message to the channel.
// The operation is non-blocking and thread-safe.
func (c *UnboundedChan[T]) Send(msg T) {
	c.mutex.WithLock(func() {
		if !c.closed {
			c.buffer = append(c.buffer, msg)
		}
	})

	// Signal that something happened
	select {
	case c.signal <- struct{}{}:
	default: // channel already has signal, skip
	}
}

// SendWithMerge adds a message to the channel with optional merging.
// If mergeFunc returns true, the new message will be merged with the last message.
// The merge operation is atomic and preserves message order.
func (c *UnboundedChan[T]) SendWithMerge(msg T, mergeFunc func(last, new T) (T, bool)) {
	c.mutex.WithLock(func() {
		if !c.closed {
			if len(c.buffer) > 0 {
				if merged, shouldMerge := mergeFunc(c.buffer[len(c.buffer)-1], msg); shouldMerge {
					c.buffer[len(c.buffer)-1] = merged

					return
				}
			}

			c.buffer = append(c.buffer, msg)
		}
	})

	// Signal that something happened
	select {
	case c.signal <- empty.Struct{}:
	default: // channel already has signal, skip
	}
}

// Receive retrieves a message from the channel with context support.
// Returns (message, true, nil) if a message is available.
// Returns (zero_value, false, nil) if the channel is closed and empty.
// Returns (zero_value, false, context.Canceled) if context is cancelled.
// Returns (zero_value, false, context.DeadlineExceeded) if context times out.
func (c *UnboundedChan[T]) Receive(ctx context.Context) (T, bool, error) {
	for {
		var msg T
		var hasMsg, isClosed bool

		c.mutex.WithLock(func() {
			if len(c.buffer) > 0 {
				msg = c.buffer[0]
				c.buffer = c.buffer[1:]
				hasMsg = true
			}
			isClosed = c.closed
		})

		if hasMsg {
			return msg, true, nil
		}
		if isClosed {
			return msg, false, nil
		}

		// Wait for signal that something happened or context cancellation
		select {
		case <-ctx.Done():
			return msg, false, ctx.Err()
		case <-c.signal:
			// Loop back to check state again
		}
	}
}

// Close closes the channel.
// After closing, Send and SendWithMerge operations will be ignored,
// and Receive will return (zero_value, false) once the buffer is empty.
func (c *UnboundedChan[T]) Close() {
	var isClosed bool
	c.mutex.WithLock(func() {
		if !c.closed {
			c.closed = true
			isClosed = true
		}
	})

	if !isClosed {
		return
	}

	close(c.signal)
}
