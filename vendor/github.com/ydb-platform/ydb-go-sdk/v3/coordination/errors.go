package coordination

import "errors"

var (
	// ErrOperationStatusUnknown indicates that the request has been sent to the server but no reply has been received.
	// The client usually automatically retries calls of that kind, but there are cases when it is not possible:
	// - the request is not idempotent, non-idempotent requests are never retried,
	// - the session was lost and its context is canceled.
	ErrOperationStatusUnknown = errors.New("operation status is unknown")

	// ErrSessionClosed indicates that the Session object is closed.
	ErrSessionClosed = errors.New("session is closed")

	// ErrAcquireTimeout indicates that the Session.AcquireSemaphore method could not acquire the semaphore before the
	// operation timeout (see options.WithAcquireTimeout).
	ErrAcquireTimeout = errors.New("acquire semaphore timeout")
)
