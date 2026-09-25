package balancers

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/endpoint"
)

// WithNodeID returns the copy of context with NodeID which the client balancer will
// prefer on step of choose YDB endpoint step
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithNodeID(ctx context.Context, nodeID uint32) context.Context {
	return endpoint.WithNodeID(ctx, nodeID)
}
