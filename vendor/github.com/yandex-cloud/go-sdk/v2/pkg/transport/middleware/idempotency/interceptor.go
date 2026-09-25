package idempotency

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// idempotencyTokenMetadataKey is the metadata key used to track the idempotency token in outgoing context.
const idempotencyTokenMetadataKey = "idempotency-key"

// Interceptor returns a grpc.UnaryClientInterceptor that adds an idempotency token to the outgoing context if not present.
func Interceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = addIdempotencyToken(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// addIdempotencyToken ensures an idempotency token is present in the outgoing context metadata.
// If not present, it generates a new token and appends it to the context.
func addIdempotencyToken(ctx context.Context) context.Context {
	idempotencyTokenPresent := false
	md, ok := metadata.FromOutgoingContext(ctx)

	if ok {
		_, idempotencyTokenPresent = md[idempotencyTokenMetadataKey]
	}

	if !idempotencyTokenPresent {
		ctx = metadata.AppendToOutgoingContext(ctx, idempotencyTokenMetadataKey, uuid.New().String())
	}

	return ctx
}
