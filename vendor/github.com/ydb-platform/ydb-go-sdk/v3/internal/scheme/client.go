package scheme

import (
	"context"
	"errors"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Scheme_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Scheme"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/scheme"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

//nolint:gofumpt
//nolint:nolintlint
var errNilClient = xerrors.Wrap(errors.New("scheme client is not initialized"))

type Client struct {
	config  *config.Config
	service Ydb_Scheme_V1.SchemeServiceClient
}

func (c *Client) Database() string {
	return c.config.Database()
}

func (c *Client) Close(_ context.Context) error {
	if c == nil {
		return xerrors.WithStackTrace(errNilClient)
	}

	return nil
}

func New(ctx context.Context, cc grpc.ClientConnInterface, config *config.Config) *Client {
	return &Client{
		config:  config,
		service: Ydb_Scheme_V1.NewSchemeServiceClient(cc),
	}
}

func (c *Client) MakeDirectory(ctx context.Context, path string) (finalErr error) {
	onDone := trace.SchemeOnMakeDirectory(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme.(*Client).MakeDirectory"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()
	call := func(ctx context.Context) error {
		return xerrors.WithStackTrace(c.makeDirectory(ctx, path))
	}
	if !c.config.AutoRetry() {
		return call(ctx)
	}

	return retry.Retry(ctx, call,
		retry.WithStackTrace(),
		retry.WithIdempotent(true),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)
}

func (c *Client) makeDirectory(ctx context.Context, path string) (err error) {
	_, err = c.service.MakeDirectory(
		ctx,
		&Ydb_Scheme.MakeDirectoryRequest{
			Path: path,
			OperationParams: operation.Params(
				ctx,
				c.config.OperationTimeout(),
				c.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)

	return xerrors.WithStackTrace(err)
}

func (c *Client) RemoveDirectory(ctx context.Context, path string) (finalErr error) {
	onDone := trace.SchemeOnRemoveDirectory(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme.(*Client).RemoveDirectory"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()
	call := func(ctx context.Context) error {
		return xerrors.WithStackTrace(c.removeDirectory(ctx, path))
	}
	if !c.config.AutoRetry() {
		return call(ctx)
	}

	return retry.Retry(ctx, call,
		retry.WithStackTrace(),
		retry.WithIdempotent(true),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)
}

func (c *Client) removeDirectory(ctx context.Context, path string) (err error) {
	_, err = c.service.RemoveDirectory(
		ctx,
		&Ydb_Scheme.RemoveDirectoryRequest{
			Path: path,
			OperationParams: operation.Params(
				ctx,
				c.config.OperationTimeout(),
				c.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)

	return xerrors.WithStackTrace(err)
}

func (c *Client) ListDirectory(ctx context.Context, path string) (d scheme.Directory, finalErr error) {
	onDone := trace.SchemeOnListDirectory(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme.(*Client).ListDirectory"),
	)
	defer func() {
		onDone(finalErr)
	}()
	call := func(ctx context.Context) (err error) {
		d, err = c.listDirectory(ctx, path)

		return xerrors.WithStackTrace(err)
	}
	if !c.config.AutoRetry() {
		err := call(ctx)

		return d, xerrors.WithStackTrace(err)
	}
	err := retry.Retry(ctx, call,
		retry.WithIdempotent(true),
		retry.WithStackTrace(),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)

	return d, xerrors.WithStackTrace(err)
}

func (c *Client) listDirectory(ctx context.Context, path string) (scheme.Directory, error) {
	var (
		d        scheme.Directory
		err      error
		response *Ydb_Scheme.ListDirectoryResponse
		result   Ydb_Scheme.ListDirectoryResult
	)
	response, err = c.service.ListDirectory(
		ctx,
		&Ydb_Scheme.ListDirectoryRequest{
			Path: path,
			OperationParams: operation.Params(
				ctx,
				c.config.OperationTimeout(),
				c.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return d, xerrors.WithStackTrace(err)
	}
	err = response.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return d, xerrors.WithStackTrace(err)
	}
	d.From(result.GetSelf())
	d.Children = make([]scheme.Entry, len(result.GetChildren()))
	putEntry(d.Children, result.GetChildren())

	return d, nil
}

func (c *Client) DescribePath(ctx context.Context, path string) (e scheme.Entry, finalErr error) {
	onDone := trace.SchemeOnDescribePath(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme.(*Client).DescribePath"),
		path,
	)
	defer func() {
		onDone(e.Type.String(), finalErr)
	}()
	call := func(ctx context.Context) (err error) {
		e, err = c.describePath(ctx, path)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}
	if !c.config.AutoRetry() {
		err := call(ctx)

		return e, err
	}
	err := retry.Retry(ctx, call,
		retry.WithIdempotent(true),
		retry.WithStackTrace(),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)

	return e, xerrors.WithStackTrace(err)
}

func (c *Client) describePath(ctx context.Context, path string) (e scheme.Entry, err error) {
	var (
		response *Ydb_Scheme.DescribePathResponse
		result   Ydb_Scheme.DescribePathResult
	)
	response, err = c.service.DescribePath(
		ctx,
		&Ydb_Scheme.DescribePathRequest{
			Path: path,
			OperationParams: operation.Params(
				ctx,
				c.config.OperationTimeout(),
				c.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return e, xerrors.WithStackTrace(err)
	}
	err = response.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return e, xerrors.WithStackTrace(err)
	}
	e.From(result.GetSelf())

	return e, nil
}

func (c *Client) ModifyPermissions(
	ctx context.Context, path string, opts ...scheme.PermissionsOption,
) (finalErr error) {
	onDone := trace.SchemeOnModifyPermissions(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme.(*Client).ModifyPermissions"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()
	var desc permissionsDesc
	for _, opt := range opts {
		if opt != nil {
			opt(&desc)
		}
	}
	call := func(ctx context.Context) error {
		return xerrors.WithStackTrace(c.modifyPermissions(ctx, path, desc))
	}
	if !c.config.AutoRetry() {
		return call(ctx)
	}

	return retry.Retry(ctx, call,
		retry.WithStackTrace(),
		retry.WithIdempotent(true),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)
}

func (c *Client) modifyPermissions(ctx context.Context, path string, desc permissionsDesc) (err error) {
	_, err = c.service.ModifyPermissions(
		ctx,
		&Ydb_Scheme.ModifyPermissionsRequest{
			Path:             path,
			Actions:          desc.actions,
			ClearPermissions: desc.clear,
			OperationParams: operation.Params(
				ctx,
				c.config.OperationTimeout(),
				c.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func putEntry(dst []scheme.Entry, src []*Ydb_Scheme.Entry) {
	for i, e := range src {
		(dst[i]).From(e)
	}
}
