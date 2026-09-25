package topicoptions

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// TopicOption
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type TopicOption = topic.Option // func(c *topic.Config)

// WithTrace defines trace over persqueue client calls
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithTrace(trace trace.Topic, opts ...trace.TopicComposeOption) TopicOption { //nolint:gocritic
	return topic.PublicWithTrace(trace, opts...)
}

// WithOperationTimeout set the maximum amount of time a YDB server will process
// an operation. After timeout exceeds YDB will try to cancel operation and
// regardless of the cancellation appropriate error will be returned to
// the client.
// If OperationTimeout is zero then no timeout is used.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithOperationTimeout(operationTimeout time.Duration) TopicOption {
	return topic.PublicWithOperationTimeout(operationTimeout)
}

// WithOperationCancelAfter set the maximum amount of time a YDB server will process an
// operation. After timeout exceeds YDB will try to cancel operation and if
// it succeeds appropriate error will be returned to the client; otherwise
// processing will be continued.
// If OperationCancelAfter is zero then no timeout is used.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithOperationCancelAfter(operationCancelAfter time.Duration) TopicOption {
	return topic.PublicWithOperationCancelAfter(operationCancelAfter)
}
