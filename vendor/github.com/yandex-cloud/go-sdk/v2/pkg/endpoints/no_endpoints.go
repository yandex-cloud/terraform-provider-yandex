package endpoints

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/yandex-cloud/go-sdk/v2/pkg/errors"
)

// NoEndpointsResolver returns an EndpointsResolver that always fails to resolve endpoints with a not found error.
func NoEndpointsResolver() EndpointsResolver { return noEndpointsResolver{} }

// noEndpointsResolver is a type implementing the EndpointsResolver interface that always returns EndpointNotFoundError.
type noEndpointsResolver struct{}

func (noEndpointsResolver) Endpoint(ctx context.Context, method protoreflect.FullName, opts ...grpc.CallOption) (*Endpoint, error) {
	return nil, &errors.EndpointNotFoundError{Method: method}
}
