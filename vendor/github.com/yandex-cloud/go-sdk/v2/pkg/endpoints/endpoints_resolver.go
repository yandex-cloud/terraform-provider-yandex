package endpoints

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// EndpointsResolver defines an interface to resolve gRPC service endpoints dynamically based on method and options.
// The Endpoint method retrieves an endpoint instance, ensuring identical pointers for similar method requests.
// Endpoint instances are not cached by the caller and can be used for connection management and pooling.
type EndpointsResolver interface {
	Endpoint(ctx context.Context, method protoreflect.FullName, opts ...grpc.CallOption) (*Endpoint, error)
}
