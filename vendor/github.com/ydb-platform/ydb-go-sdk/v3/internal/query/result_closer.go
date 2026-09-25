package query

import (
	"context"
	"io"
	"sync"
)

var errResultCloserNilReason = io.EOF

// ResultCloser provides a mechanism to close query results and handle cleanup operations.
// It tracks the reason for closing, provides a done channel for synchronization,
// and allows registering cleanup functions to be called on close.
type ResultCloser struct {
	reason  error
	done    chan struct{}
	mu      sync.Mutex
	onClose []func()
}

// NewResultCloser creates and returns a new ResultCloser instance.
func NewResultCloser() *ResultCloser {
	return &ResultCloser{
		done: make(chan struct{}),
	}
}

// Close closes the ResultCloser with the specified reason error.
// If the ResultCloser is already closed, this method does nothing.
// If reason is nil, it will be set to io.EOF.
// All registered onClose functions will be called in LIFO order.
// After calling Close, the done channel will be closed to signal completion.
func (r *ResultCloser) Close(reason error) {
	if r.doneWithReason(reason) { // only first [r.Close] invoke runs callbacks
		r.runOnCloseCallbacks()
	}
}

// doneWithReason sets the closure reason and signals completion by closing the done channel.
// The method is idempotent - subsequent calls after the first successful call are no-ops.
// If reason is nil, it defaults to errResultCloserNilReason.
// The method uses mutex synchronization to ensure safe concurrent access.
// Returns true if the close operation was performed (first call), false otherwise.
func (r *ResultCloser) doneWithReason(reason error) (realClose bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.reason != nil {
		return false
	}

	if reason == nil {
		reason = errResultCloserNilReason
	}

	r.reason = reason

	close(r.done)

	return true
}

// runOnCloseCallbacks executes registered cleanup callbacks in reverse order (LIFO).
// This method is NOT safe for concurrent access.
func (r *ResultCloser) runOnCloseCallbacks() {
	for i := range r.onClose { // descending calls for LIFO
		r.onClose[len(r.onClose)-i-1]()
	}
}

// Err returns the reason error that was passed to Close.
// If Close has not been called yet, it returns nil.
func (r *ResultCloser) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.reason
}

// Done returns a channel that will be closed when the ResultCloser is closed.
// This channel can be used to wait for the ResultCloser to be closed.
func (r *ResultCloser) Done() <-chan struct{} {
	return r.done
}

// CloseOnContextCancel registers a callback function that closes the ResultCloser
// when the provided context is cancelled.
// It returns a function that can be used to stop the callback.
func (r *ResultCloser) CloseOnContextCancel(ctx context.Context) func() bool {
	return context.AfterFunc(ctx, func() {
		r.Close(ctx.Err())
	})
}

// OnClose registers a function to be called when the ResultCloser is closed.
// Multiple functions can be registered and they will be called in LIFO order
// when Close is called.
func (r *ResultCloser) OnClose(f func()) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onClose = append(r.onClose, f)
}
