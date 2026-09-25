package options

import (
	"math"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Coordination"
)

// WithDescription returns an SessionOption that specifies a user-defined description that may be used to describe
// the client.
func WithDescription(description string) SessionOption {
	return func(c *CreateSessionOptions) {
		c.Description = description
	}
}

// WithSessionTimeout returns an SessionOption that specifies the timeout during which client may restore a
// detached session. The client is forced to terminate the session if the last successful session request occurred
// earlier than this time.
//
// If this is not set, the client uses the default 5 seconds.
func WithSessionTimeout(timeout time.Duration) SessionOption {
	return func(c *CreateSessionOptions) {
		c.SessionTimeout = timeout
	}
}

// WithSessionStartTimeout returns an SessionOption that specifies the time that the client should wait for a
// response to the StartSession request from the server before it terminates the gRPC stream and tries to reconnect.
//
// If this is not set, the client uses the default time 1 second.
func WithSessionStartTimeout(timeout time.Duration) SessionOption {
	return func(c *CreateSessionOptions) {
		c.SessionStartTimeout = timeout
	}
}

// WithSessionStopTimeout returns an SessionOption that specifies the time that the client should wait for a
// response to the StopSession request from the server before it terminates the gRPC stream and tries to reconnect.
//
// If this is not set, the client uses the default time 1 second.
func WithSessionStopTimeout(timeout time.Duration) SessionOption {
	return func(c *CreateSessionOptions) {
		c.SessionStartTimeout = timeout
	}
}

// WithSessionKeepAliveTimeout returns an SessionOption that specifies the time that the client will wait before
// it terminates the gRPC stream and tries to reconnect if no successful responses have been received from the server.
//
// If this is not set, the client uses the default time 10 seconds.
func WithSessionKeepAliveTimeout(timeout time.Duration) SessionOption {
	return func(c *CreateSessionOptions) {
		c.SessionKeepAliveTimeout = timeout
	}
}

// WithSessionReconnectDelay returns an SessionOption that specifies the time that the client will wait before it
// tries to reconnect the underlying gRPC stream in case of error.
//
// If this is not set, the client uses the default time 500 milliseconds.
func WithSessionReconnectDelay(delay time.Duration) SessionOption {
	return func(c *CreateSessionOptions) {
		c.SessionReconnectDelay = delay
	}
}

// SessionOption configures how we create a new session.
type SessionOption func(c *CreateSessionOptions)

// CreateSessionOptions configure an Session call. CreateSessionOptions are set by the SessionOption values
// passed to the Session function.
type CreateSessionOptions struct {
	Description             string
	SessionTimeout          time.Duration
	SessionStartTimeout     time.Duration
	SessionStopTimeout      time.Duration
	SessionKeepAliveTimeout time.Duration
	SessionReconnectDelay   time.Duration
}

// WithEphemeral returns an AcquireSemaphoreOption that causes to create an ephemeral semaphore.
//
// Ephemeral semaphores are created with the first acquire operation and automatically deleted with the last release
// operation. Ephemeral semaphore are always created with the limit of coordination.MaxSemaphoreLimit.
func WithEphemeral(ephemeral bool) AcquireSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_AcquireSemaphore) {
		c.Ephemeral = ephemeral
	}
}

// WithAcquireTimeout returns an AcquireSemaphoreOption which sets the timeout after which the operation fails if it
// is still waiting in the queue. Use 0 to make the AcquireSemaphore method fail immediately if the semaphore is already
// acquired by another session.
//
// If this is not set, the client waits for the acquire operation result until the operation or session context is done.
// You can reset the default value of this timeout by calling the WithAcquireInfiniteTimeout method.
func WithAcquireTimeout(timeout time.Duration) AcquireSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_AcquireSemaphore) {
		c.TimeoutMillis = uint64(timeout.Milliseconds())
	}
}

// WithAcquireInfiniteTimeout returns an AcquireSemaphoreOption which disables the timeout after which the operation
// fails if it is still waiting in the queue.
//
// This is the default behavior. You can set the specific timeout by calling the WithAcquireTimeout method.
func WithAcquireInfiniteTimeout() AcquireSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_AcquireSemaphore) {
		c.TimeoutMillis = math.MaxUint64
	}
}

// WithAcquireData returns an AcquireSemaphoreOption which attaches user-defined data to the operation.
func WithAcquireData(data []byte) AcquireSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_AcquireSemaphore) {
		c.Data = data
	}
}

// AcquireSemaphoreOption configures how we acquire a semaphore.
type AcquireSemaphoreOption func(c *Ydb_Coordination.SessionRequest_AcquireSemaphore)

// WithForceDelete return a DeleteSemaphoreOption which allows to delete a semaphore even if it is currently acquired
// by other sessions.
func WithForceDelete(force bool) DeleteSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_DeleteSemaphore) {
		c.Force = force
	}
}

// DeleteSemaphoreOption configures how we delete a semaphore.
type DeleteSemaphoreOption func(c *Ydb_Coordination.SessionRequest_DeleteSemaphore)

// WithCreateData return a CreateSemaphoreOption which attaches user-defined data to the semaphore.
func WithCreateData(data []byte) CreateSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_CreateSemaphore) {
		c.Data = data
	}
}

// CreateSemaphoreOption configures how we create a semaphore.
type CreateSemaphoreOption func(c *Ydb_Coordination.SessionRequest_CreateSemaphore)

// WithUpdateData return a UpdateSemaphoreOption which changes user-defined data in the semaphore.
func WithUpdateData(data []byte) UpdateSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_UpdateSemaphore) {
		c.Data = data
	}
}

// UpdateSemaphoreOption configures how we update a semaphore.
type UpdateSemaphoreOption func(c *Ydb_Coordination.SessionRequest_UpdateSemaphore)

// WithDescribeOwners return a DescribeSemaphoreOption which causes server send the list of owners in the response
// to the DescribeSemaphore request.
func WithDescribeOwners(describeOwners bool) DescribeSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_DescribeSemaphore) {
		c.IncludeOwners = describeOwners
	}
}

// WithDescribeWaiters return a DescribeSemaphoreOption which causes server send the list of waiters in the response
// to the DescribeSemaphore request.
func WithDescribeWaiters(describeWaiters bool) DescribeSemaphoreOption {
	return func(c *Ydb_Coordination.SessionRequest_DescribeSemaphore) {
		c.IncludeWaiters = describeWaiters
	}
}

// DescribeSemaphoreOption configures how we update a semaphore.
type DescribeSemaphoreOption func(c *Ydb_Coordination.SessionRequest_DescribeSemaphore)
