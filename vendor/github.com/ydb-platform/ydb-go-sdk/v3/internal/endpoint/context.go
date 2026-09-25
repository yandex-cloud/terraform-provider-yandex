package endpoint

import "context"

type (
	ctxEndpointKey struct{}
)

func WithNodeID(ctx context.Context, nodeID uint32) context.Context {
	return context.WithValue(ctx, ctxEndpointKey{}, nodeID)
}

func ContextNodeID(ctx context.Context) (nodeID uint32, ok bool) {
	if nodeID, ok = ctx.Value(ctxEndpointKey{}).(uint32); ok {
		return nodeID, true
	}

	return 0, false
}
