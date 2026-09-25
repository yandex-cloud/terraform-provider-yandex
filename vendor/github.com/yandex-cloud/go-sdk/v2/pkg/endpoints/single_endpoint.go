package endpoints

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewSingleEndpointResolver creates an EndpointsResolver that resolves to a single gRPC endpoint with specified options.
func NewSingleEndpointResolver(addr string, opts ...grpc.DialOption) EndpointsResolver {
	return singleEndpointResolver{&Endpoint{
		Addr:        addr,
		DialOptions: opts,
	}}
}

// SingleEndpointResolver returns an EndpointsResolver that always resolves to the specified single Endpoint.
func SingleEndpointResolver(e *Endpoint) EndpointsResolver { return singleEndpointResolver{e} }

// singleEndpointResolver is a resolver that always returns a single predefined Endpoint instance.
type singleEndpointResolver struct{ e *Endpoint }

// Endpoint resolves and returns a gRPC endpoint along with its connection options, based on the resolver configuration.
func (o singleEndpointResolver) Endpoint(_ context.Context, _ protoreflect.FullName, _ ...grpc.CallOption) (*Endpoint, error) {
	return o.e, nil
}
