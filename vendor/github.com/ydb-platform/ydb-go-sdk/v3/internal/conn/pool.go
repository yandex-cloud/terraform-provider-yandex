package conn

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/closer"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/endpoint"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xresolver"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Pool struct {
	usages      int64
	config      Config
	dialOptions []grpc.DialOption
	conns       xsync.Map[string, *conn]
	done        chan struct{}
}

func (p *Pool) DialTimeout() time.Duration {
	return p.config.DialTimeout()
}

func (p *Pool) Trace() *trace.Driver {
	return p.config.Trace()
}

func (p *Pool) GrpcDialOptions() []grpc.DialOption {
	return p.dialOptions
}

func (p *Pool) Get(endpoint endpoint.Endpoint) Conn {
	var (
		address = endpoint.Address()
		cc      *conn
		has     bool
	)

	cc, has = p.conns.Get(address)
	if has && cc.Endpoint().NodeID() == endpoint.NodeID() && cc.Endpoint().OverrideHost() == endpoint.OverrideHost() {
		return cc
	}

	cc = newConn(endpoint, p,
		withOnClose(p.remove),
		withOnTransportError(p.Ban),
	)

	p.conns.Set(address, cc)

	return cc
}

func (p *Pool) remove(c *conn) {
	p.conns.Delete(c.Address())
}

func (p *Pool) isClosed() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

func (p *Pool) Ban(ctx context.Context, cc Conn, cause error) {
	if p.isClosed() {
		return
	}

	if !xerrors.IsTransportError(cause,
		grpcCodes.ResourceExhausted,
		grpcCodes.Unavailable,
		// grpcCodes.OK,
		// grpcCodes.Canceled,
		// grpcCodes.Unknown,
		// grpcCodes.InvalidArgument,
		// grpcCodes.DeadlineExceeded,
		// grpcCodes.NotFound,
		// grpcCodes.AlreadyExists,
		// grpcCodes.PermissionDenied,
		// grpcCodes.FailedPrecondition,
		// grpcCodes.Aborted,
		// grpcCodes.OutOfRange,
		// grpcCodes.Unimplemented,
		// grpcCodes.Internal,
		// grpcCodes.DataLoss,
		// grpcCodes.Unauthenticated,
	) {
		return
	}

	e := cc.Endpoint().Copy()

	cc, ok := p.conns.Get(e.Address())
	if !ok {
		return
	}

	trace.DriverOnConnBan(
		p.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/conn.(*Pool).Ban"),
		e, cc.GetState(), cause,
	)(cc.SetState(ctx, Banned))
}

func (p *Pool) Allow(ctx context.Context, cc Conn) {
	if p.isClosed() {
		return
	}

	e := cc.Endpoint().Copy()

	cc, ok := p.conns.Get(e.Address())
	if !ok {
		return
	}

	trace.DriverOnConnAllow(
		p.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/conn.(*Pool).Allow"),
		e, cc.GetState(),
	)(cc.Unban(ctx))
}

func (p *Pool) Take(context.Context) error {
	atomic.AddInt64(&p.usages, 1)

	return nil
}

func (p *Pool) Release(ctx context.Context) (finalErr error) {
	onDone := trace.DriverOnPoolRelease(p.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/conn.(*Pool).Release"),
	)
	defer func() {
		onDone(finalErr)
	}()

	if atomic.AddInt64(&p.usages, -1) > 0 {
		return nil
	}

	close(p.done)

	var (
		errCh = make(chan error, p.conns.Len())
		wg    sync.WaitGroup
	)

	wg.Add(cap(errCh))
	p.conns.Range(func(_ string, c *conn) bool {
		go func(c closer.Closer) {
			defer wg.Done()
			if err := c.Close(ctx); err != nil {
				errCh <- err
			}
		}(c)

		return true
	})
	wg.Wait()
	close(errCh)

	issues := make([]error, 0, cap(errCh))
	for err := range errCh {
		issues = append(issues, err)
	}

	if len(issues) > 0 {
		return xerrors.WithStackTrace(xerrors.NewWithIssues("connection pool close failed", issues...))
	}

	return nil
}

func (p *Pool) connParker(ctx context.Context, ttl, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.conns.Range(func(_ string, c *conn) bool {
				if time.Since(c.LastUsage()) > ttl {
					switch c.GetState() {
					case Online, Banned:
						_ = c.park(ctx)
					default:
						// nop
					}
				}

				return true
			})
		}
	}
}

func NewPool(ctx context.Context, config Config) *Pool {
	onDone := trace.DriverOnPoolNew(config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/conn.NewPool"),
	)
	defer onDone()

	p := &Pool{
		usages:      1,
		config:      config,
		dialOptions: config.GrpcDialOptions(),
		done:        make(chan struct{}),
	}

	p.dialOptions = append(p.dialOptions,
		grpc.WithResolvers(
			xresolver.New("", config.Trace().Compose(&trace.Driver{
				OnResolve: func(info trace.DriverResolveStartInfo) func(trace.DriverResolveDoneInfo) {
					target := info.Target
					resolved := info.Resolved

					return func(info trace.DriverResolveDoneInfo) {
						if info.Error != nil || len(resolved) == 0 {
							p.conns.Range(func(address string, cc *conn) bool {
								if u, err := url.Parse(address); err == nil && u.Host == target && cc.grpcConn != nil {
									_ = cc.grpcConn.Close()
									_ = p.conns.Delete(address)
								}

								return true
							})
						}
					}
				},
			})),
		),
	)

	if ttl := config.ConnectionTTL(); ttl > 0 {
		go p.connParker(xcontext.ValueOnly(ctx), ttl, ttl/2) //nolint:mnd
	}

	return p
}
