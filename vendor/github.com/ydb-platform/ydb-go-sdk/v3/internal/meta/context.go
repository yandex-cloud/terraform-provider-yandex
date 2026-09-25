package meta

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// WithTraceID returns a copy of parent context with traceID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if md, has := metadata.FromOutgoingContext(ctx); !has || len(md[HeaderTraceID]) == 0 {
		return metadata.AppendToOutgoingContext(ctx, HeaderTraceID, traceID)
	}

	return ctx
}

func traceID(ctx context.Context) (string, bool) {
	if md, has := metadata.FromOutgoingContext(ctx); has && len(md[HeaderTraceID]) > 0 {
		return md[HeaderTraceID][0], true
	}

	return "", false
}

// WithApplicationName returns a copy of parent context with custom user-agent info
func WithApplicationName(ctx context.Context, applicationName string) context.Context {
	md, has := metadata.FromOutgoingContext(ctx)
	if !has {
		md = metadata.MD{}
	}
	md.Set(HeaderApplicationName, applicationName)

	return metadata.NewOutgoingContext(ctx, md)
}

// WithRequestType returns a copy of parent context with custom request type
func WithRequestType(ctx context.Context, requestType string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, HeaderRequestType, requestType)
}

// WithAllowFeatures returns a copy of parent context with allowed client feature
func WithAllowFeatures(ctx context.Context, features ...string) context.Context {
	kv := make([]string, 0, len(features)*2) //nolint:mnd
	for _, feature := range features {
		kv = append(kv, HeaderClientCapabilities, feature)
	}

	return metadata.AppendToOutgoingContext(ctx, kv...)
}

// WithTraceParent returns a copy of parent context with traceparent header
func WithTraceParent(ctx context.Context, traceparent string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, HeaderTraceParent, traceparent)
}
