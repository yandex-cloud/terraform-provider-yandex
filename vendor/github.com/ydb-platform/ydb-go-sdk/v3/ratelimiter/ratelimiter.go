package ratelimiter

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/ratelimiter/options"
)

type Client interface {
	CreateResource(
		ctx context.Context,
		coordinationNodePath string,
		resource Resource,
	) (err error)
	AlterResource(
		ctx context.Context,
		coordinationNodePath string,
		resource Resource,
	) (err error)
	DropResource(
		ctx context.Context,
		coordinationNodePath string,
		resourcePath string,
	) (err error)
	ListResource(
		ctx context.Context,
		coordinationNodePath string,
		resourcePath string,
		recursive bool,
	) (_ []string, err error)
	DescribeResource(
		ctx context.Context,
		coordinationNodePath string,
		resourcePath string,
	) (_ *Resource, err error)
	AcquireResource(
		ctx context.Context,
		coordinationNodePath string,
		resourcePath string,
		amount uint64,
		opts ...options.AcquireOption,
	) (err error)
}

func WithAcquire() options.AcquireOption {
	return options.WithAcquire()
}

func WithReport() options.AcquireOption {
	return options.WithReport()
}

func WithOperationTimeout(operationTimeout time.Duration) options.AcquireOption {
	return options.WithOperationTimeout(operationTimeout)
}

func WithOperationCancelAfter(operationCancelAfter time.Duration) options.AcquireOption {
	return options.WithOperationCancelAfter(operationCancelAfter)
}
