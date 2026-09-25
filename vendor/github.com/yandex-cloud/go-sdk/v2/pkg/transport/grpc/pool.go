package grpc

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	pkgerrors "github.com/yandex-cloud/go-sdk/v2/pkg/errors"
)

// ConnPool manages a pool of gRPC client connections with support for connection reuse and context-based lifecycle.
type ConnPool struct {
	dialOptions []grpc.DialOption

	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	conns  sync.Map
	closed bool
}

// NewConnPool creates a new connection pool with the provided gRPC dial options. It initializes context and internal state.
func NewConnPool(opts []grpc.DialOption) *ConnPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConnPool{
		dialOptions: opts,
		ctx:         ctx,
		cancel:      cancel,
		conns:       sync.Map{},
	}
}

// GetConn retrieves or establishes a gRPC connection for the specified endpoint, respecting the provided context constraints.
func (cc *ConnPool) GetConn(ctx context.Context, ep *endpoints.Endpoint) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		// fail fast: don't even try
		return nil, status.FromContextError(err).Err()
	}

	if conn, err := cc.pooledConn(ep); conn != nil || err != nil {
		return conn, err
	}

	type dialResult struct {
		conn *grpc.ClientConn
		err  error
	}

	resultCh := make(chan dialResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultCh <- dialResult{nil, r.(error)}
			}
		}()

		v, err := cc.makeDial(ep)
		resultCh <- dialResult{v, err}
	}()

	select {
	case res := <-resultCh:
		return res.conn, res.err
	case <-ctx.Done():
		return nil, status.FromContextError(ctx.Err()).Err()
	}
}

// makeDial establishes a new gRPC client connection to the specified endpoint and stores it in the connection pool.
func (cc *ConnPool) makeDial(ep *endpoints.Endpoint) (*grpc.ClientConn, error) {
	if conn, err := cc.pooledConn(ep); conn != nil || err != nil {
		return conn, err
	}

	opts := make([]grpc.DialOption, 0, len(cc.dialOptions)+len(ep.DialOptions))
	opts = append(opts, cc.dialOptions...)
	opts = append(opts, ep.DialOptions...)

	conn, err := grpc.NewClient(ep.Addr, opts...)
	if err != nil {
		if cc.ctx.Err() != nil {
			return nil, pkgerrors.ErrConnContextClosed
		}

		return nil, &pkgerrors.DialError{Err: err, Addr: ep.Addr}
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.closed {
		_ = conn.Close()
		return nil, nil
	}

	cc.conns.Store(ep, conn)

	return conn, nil
}

// pooledConn retrieves a gRPC client connection from the pool for the given endpoint if available, else returns nil.
// It ensures thread-safe access and returns an error if the connection context is closed.
func (cc *ConnPool) pooledConn(ep *endpoints.Endpoint) (*grpc.ClientConn, error) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if cc.closed {
		return nil, pkgerrors.ErrConnContextClosed
	}

	value, ok := cc.conns.Load(ep)
	if !ok {
		return nil, nil
	}

	conn := value.(*grpc.ClientConn)

	return conn, nil
}

// Shutdown closes all active connections in the pool, cancels the context, and marks the pool as closed. Returns errors if any occur during the connection closures.
func (cc *ConnPool) Shutdown(ctx context.Context) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.closed {
		return nil
	}

	defer func() {
		cc.closed = true
		cc.cancel()
	}()

	var errs error

	cc.conns.Range(
		func(_, value any) bool {
			conn := value.(*grpc.ClientConn)
			if err := conn.Close(); err != nil {
				errs = errors.Join(errs, err)
			}

			return true
		},
	)

	return errs
}
