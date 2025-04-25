package connection

import (
	"context"

	"google.golang.org/grpc"
)

type ConnectionDialer interface {
	GetConnection(ctx context.Context, addr string) (*grpc.ClientConn, error)
}

type SimpleConnectionDialer struct { // TODO: cache connections
	dialOpts []grpc.DialOption
}

func NewSimpleConnectionDialer(opts ...grpc.DialOption) *SimpleConnectionDialer {
	return &SimpleConnectionDialer{
		dialOpts: opts,
	}
}

func (c *SimpleConnectionDialer) GetConnection(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, addr, c.dialOpts...)
}
