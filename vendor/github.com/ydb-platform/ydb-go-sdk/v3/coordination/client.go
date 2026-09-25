package coordination

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/coordination/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/scheme"
)

type Client interface {
	CreateNode(ctx context.Context, path string, config NodeConfig) (err error)
	AlterNode(ctx context.Context, path string, config NodeConfig) (err error)
	DropNode(ctx context.Context, path string) (err error)
	DescribeNode(ctx context.Context, path string) (_ *scheme.Entry, _ *NodeConfig, err error)

	// Session starts a new session. This method blocks until the server session is created. The context provided
	// may be used to cancel the invocation. If the method completes successfully, the session remains alive even if
	// the context is canceled.
	//
	// To ensure resources are not leaked, one of the following actions must be performed:
	//
	// - call Close on the Session,
	// - close the Client which the session was created with,
	// - call any method of the Session until the ErrSessionClosed is returned.
	Session(ctx context.Context, path string, opts ...options.SessionOption) (Session, error)
}

const (
	// MaxSemaphoreLimit defines the maximum value of the limit parameter in the Session.CreateSemaphore method.
	MaxSemaphoreLimit = math.MaxUint64

	// Exclusive is just a shortcut for the maximum semaphore limit value. You can use this to acquire a semaphore in
	// the exclusive mode if it was created with the limit value of MaxSemaphoreLimit, which is always true for
	// ephemeral semaphores.
	Exclusive = math.MaxUint64

	// Shared is just a shortcut for the minimum semaphore limit value (1). You can use this to acquire a semaphore in
	// the shared mode if it was created with the limit value of MaxSemaphoreLimit, which is always true for ephemeral
	// semaphores.
	Shared = 1
)

// Session defines a coordination service backed session.
//
// In general, Session API is concurrency-friendly, you can safely access all of its methods concurrently.
//
// The client guarantees that sequential calls of the methods are sent to the server in the same order. However, the
// session client may reorder and suspend some of the requests without violating correctness of the execution. This also
// applies to the situations when the underlying gRPC stream has been recreated due to network or server issues.
//
// The client automatically keeps the underlying gRPC stream alive by sending keep-alive (ping-pong) requests. If the
// client can no longer consider the session alive, it immediately cancels the session context which also leads to
// cancellation of contexts of all semaphore leases created by this session.
type Session interface {
	// Close closes the coordination service session. It cancels all active requests on the server and notifies every
	// pending or waiting for response request on the client side. It also cancels the session context and tries to
	// stop the session gracefully on the server. If the ctx is canceled, this will not wait for the server session to
	// become stopped and returns immediately with an error. Once this function returns with no error, all subsequent
	// calls will be noop.
	Close(ctx context.Context) error

	// Context returns the context of the session. It is canceled when the underlying server session is over or if the
	// client could not get any successful response from the server before the session timeout (see
	// options.WithSessionTimeout).
	Context() context.Context

	// CreateSemaphore creates a new semaphore. This method waits until the server successfully creates a new semaphore
	// or returns an error.
	//
	// This method is not idempotent. If the request has been sent to the server but no reply has been received, it
	// returns the ErrOperationStatusUnknown error.
	CreateSemaphore(ctx context.Context, name string, limit uint64, opts ...options.CreateSemaphoreOption) error

	// UpdateSemaphore changes semaphore data. This method waits until the server successfully updates the semaphore or
	// returns an error.
	//
	// This method is not idempotent. The client will automatically retry in the case of network or server failure
	// unless it leaves the client state inconsistent.
	UpdateSemaphore(ctx context.Context, name string, opts ...options.UpdateSemaphoreOption) error

	// DeleteSemaphore deletes an existing semaphore. This method waits until the server successfully deletes the
	// semaphore or returns an error.
	//
	// This method is not idempotent. If the request has been sent to the server but no reply has been received, it
	// returns the ErrOperationStatusUnknown error.
	DeleteSemaphore(ctx context.Context, name string, opts ...options.DeleteSemaphoreOption) error

	// DescribeSemaphore returns the state of the semaphore.
	//
	// This method is idempotent. The client will automatically retry in the case of network or server failure.
	DescribeSemaphore(
		ctx context.Context,
		name string,
		opts ...options.DescribeSemaphoreOption,
	) (*SemaphoreDescription, error)

	// AcquireSemaphore acquires the semaphore. If you acquire an ephemeral semaphore (see options.WithEphemeral), its
	// limit will be set to MaxSemaphoreLimit. Later requests override previous operations with the same semaphore, e.g.
	// to reduce acquired count, change timeout or attached data.
	//
	// This method blocks until the semaphore is acquired, an error is returned from the server or the session is
	// closed. If the operation context was canceled but the server replied that the semaphore was actually acquired,
	// the client will automatically release the semaphore.
	//
	// Semaphore waiting is fair: the semaphore guarantees that other sessions invoking the AcquireSemaphore method
	// acquire permits in the order which they were called (FIFO). If a session invokes the AcquireSemaphore method
	// multiple times while the first invocation is still in process, the position in the queue remains unchanged.
	//
	// This method is idempotent. The client will automatically retry in the case of network or server failure.
	AcquireSemaphore(
		ctx context.Context,
		name string,
		count uint64,
		opts ...options.AcquireSemaphoreOption,
	) (Lease, error)

	// SessionID returns a server-generated identifier of the session. This value is permanent and unique within the
	// coordination service node.
	SessionID() uint64

	// Reconnect forcibly shuts down the underlying gRPC stream and initiates a new one. This method is highly unlikely
	// to be of use in a typical application but is extremely useful for testing an API implementation.
	Reconnect()
}

// Lease is the object which defines the rights of the session to the acquired semaphore. Lease is alive until its
// context is not canceled. This may happen implicitly, when the associated session becomes lost or closed, or
// explicitly, if someone calls the Release method of the lease.
type Lease interface {
	// Context returns the context of the lease. It is canceled when the session it was created by was lost or closed,
	// or if the lease was released by calling the Release method.
	Context() context.Context

	// Release releases the acquired lease to the semaphore. It also cancels the context of the lease. This method does
	// not take a ctx argument, but you can cancel the execution of it by closing the session or canceling its context.
	Release() error

	// Session returns the session which this lease was created by.
	Session() Session
}

// SemaphoreDescription describes the state of a semaphore.
type SemaphoreDescription struct {
	// Name is the name of the semaphore.
	Name string

	// Limit is the maximum number of tokens that may be acquired.
	Limit uint64

	// Count is the number of tokens currently acquired by its owners.
	Count uint64

	// Ephemeral semaphores are deleted when there are no owners and waiters left.
	Ephemeral bool

	// Data is user-defined data attached to the semaphore.
	Data []byte

	// Owner is the list of current owners of the semaphore.
	Owners []*SemaphoreSession

	// Waiter is the list of current waiters of the semaphore.
	Waiters []*SemaphoreSession
}

// SemaphoreSession describes an owner or a waiter of this semaphore.
type SemaphoreSession struct {
	// SessionID is the id of the session which tried to acquire the semaphore.
	SessionID uint64

	// Count is the number of tokens for the acquire operation.
	Count uint64

	// OrderId is a monotonically increasing id which determines locking order.
	OrderID uint64

	// Data is user-defined data attached to the acquire operation.
	Data []byte

	// Timeout is the timeout for the operation in the waiter queue. If this is time.Duration(math.MaxInt64) the session
	// will wait for the semaphore until the operation is canceled.
	Timeout time.Duration
}

func (d *SemaphoreDescription) String() string {
	return fmt.Sprintf(
		"{Name: %q Limit: %d Count: %d Ephemeral: %t Data: %q Owners: %s Waiters: %s}",
		d.Name, d.Limit, d.Count, d.Ephemeral, d.Data, d.Owners, d.Waiters)
}

func (s *SemaphoreSession) String() string {
	return fmt.Sprintf("{SessionID: %d Count: %d OrderID: %d Data: %q TimeoutMillis: %v}",
		s.SessionID, s.Count, s.OrderID, s.Data, s.Timeout)
}
