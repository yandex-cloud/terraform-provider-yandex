package transport

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	transportgrpc "github.com/yandex-cloud/go-sdk/v2/pkg/transport/grpc"
)

// Connector is an interface for retrieving gRPC client connections to specified methods with context and options.
type Connector interface {
	GetConnection(ctx context.Context, method protoreflect.FullName, opts ...grpc.CallOption) (grpc.ClientConnInterface, error)
}

var _ Connector = &ConnectorImpl{}

// ConnectorImpl is an implementation of the Connector interface, managing gRPC connections using endpoints resolution and pooling.
type ConnectorImpl struct {
	endpointsResolver endpoints.EndpointsResolver
	connectionPool    *transportgrpc.ConnPool
}

// NewConnector initializes and returns a ConnectorImpl instance with the provided endpoints resolver and connection pool.
func NewConnector(endpoints endpoints.EndpointsResolver, connectionPool *transportgrpc.ConnPool) *ConnectorImpl {
	return &ConnectorImpl{
		endpointsResolver: endpoints,
		connectionPool:    connectionPool,
	}
}

// GetConnection retrieves a gRPC client connection for the given method, resolving the endpoint via the endpointsResolver.
func (c *ConnectorImpl) GetConnection(ctx context.Context, method protoreflect.FullName, opts ...grpc.CallOption) (grpc.ClientConnInterface, error) {
	ep, err := c.endpointsResolver.Endpoint(ctx, method, opts...)
	if err != nil {
		return nil, err
	}

	return c.connectionPool.GetConn(ctx, ep)
}

// SingleConnector provides a gRPC connection handler using a single predefined endpoint resolver.
type SingleConnector struct {
	endpointsResolver endpoints.EndpointsResolver
}

// NewSingleConnector creates a SingleConnector that utilizes a single gRPC endpoint with the specified address and options.
func NewSingleConnector(add string, opts ...grpc.DialOption) *SingleConnector {
	return &SingleConnector{
		endpointsResolver: endpoints.NewSingleEndpointResolver(add, opts...),
	}
}

// GetConnection retrieves a gRPC client connection for the specified method and options, resolving the appropriate endpoint.
func (s *SingleConnector) GetConnection(ctx context.Context, method protoreflect.FullName, opts ...grpc.CallOption) (grpc.ClientConnInterface, error) {
	ep, err := s.endpointsResolver.Endpoint(ctx, method, opts...)
	if err != nil {
		return nil, err
	}

	return grpc.NewClient(ep.Addr, ep.DialOptions...)
}
