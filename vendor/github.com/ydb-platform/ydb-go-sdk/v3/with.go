package ydb

import (
	"context"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var nextID atomic.Uint64 //nolint:gochecknoglobals

func (d *Driver) with(ctx context.Context, opts ...Option) (*Driver, uint64, error) {
	id := nextID.Add(1)

	child, err := driverFromOptions(
		ctx,
		append(
			append(
				d.opts,
				WithBalancer(
					d.config.Balancer(),
				),
				withOnClose(func(child *Driver) {
					d.childrenMtx.Lock()
					defer d.childrenMtx.Unlock()

					delete(d.children, id)
				}),
				withConnPool(d.pool),
			),
			opts...,
		)...,
	)
	if err != nil {
		return nil, 0, xerrors.WithStackTrace(err)
	}

	return child, id, nil
}

// With makes child Driver with the same options and another options
func (d *Driver) With(ctx context.Context, opts ...Option) (*Driver, error) {
	child, id, err := d.with(ctx, opts...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	onDone := trace.DriverOnWith(
		d.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/ydb.(*Driver).With"),
		d.config.Endpoint(), d.config.Database(), d.config.Secure(),
	)
	defer func() {
		onDone(err)
	}()

	if err = child.connect(ctx); err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	d.childrenMtx.Lock()
	defer d.childrenMtx.Unlock()

	d.children[id] = child

	return child, nil
}
