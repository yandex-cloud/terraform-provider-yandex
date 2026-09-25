package retry

import "context"

type (
	ctxIsOperationIdempotentKey struct{}
)

// WithIdempotentOperation returns a copy of parent context with idempotent operation feature
//
// Deprecated: use retry.WithIdempotent option instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithIdempotentOperation(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxIsOperationIdempotentKey{}, true)
}

// WithNonIdempotentOperation returns a copy of parent context with non-idempotent operation feature
//
// Deprecated: idempotent flag is false by default.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithNonIdempotentOperation(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxIsOperationIdempotentKey{}, false)
}

// IsOperationIdempotent returns the flag for retry with no idempotent errors
//
// Deprecated: context cannot store idempotent value now.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func IsOperationIdempotent(ctx context.Context) bool {
	v, ok := ctx.Value(ctxIsOperationIdempotentKey{}).(bool)

	return ok && v
}
