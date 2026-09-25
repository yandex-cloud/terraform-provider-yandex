package meta

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/meta"
)

// WithTraceID returns a copy of parent context with traceID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return meta.WithTraceID(ctx, traceID)
}

// WithUserAgent returns a copy of parent context with custom user-agent info
//
// Deprecated: use WithApplicationName instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithUserAgent(ctx context.Context, _ string) context.Context {
	return ctx
}

// WithApplicationName returns a copy of parent context with application name
func WithApplicationName(ctx context.Context, applicationName string) context.Context {
	return meta.WithApplicationName(ctx, applicationName)
}

// WithRequestType returns a copy of parent context with custom request type
func WithRequestType(ctx context.Context, requestType string) context.Context {
	return meta.WithRequestType(ctx, requestType)
}

// WithAllowFeatures returns a copy of parent context with allowed client feature
func WithAllowFeatures(ctx context.Context, features ...string) context.Context {
	return meta.WithAllowFeatures(ctx, features...)
}

// WithTrailerCallback attaches callback to context for listening incoming metadata
func WithTrailerCallback(
	ctx context.Context,
	callback func(md metadata.MD),
) context.Context {
	return meta.WithTrailerCallback(ctx, callback)
}

// WithTraceParent returns a copy of parent context with traceparent header
func WithTraceParent(ctx context.Context, traceparent string) context.Context {
	return meta.WithTraceParent(ctx, traceparent)
}
