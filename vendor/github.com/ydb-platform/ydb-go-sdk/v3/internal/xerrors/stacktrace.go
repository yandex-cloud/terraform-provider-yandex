package xerrors

import (
	grpcStatus "google.golang.org/grpc/status"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
)

const KeywordAt = "at"

type withStackTraceOptions struct {
	skipDepth int
}

type WithStackTraceOption func(o *withStackTraceOptions)

func WithSkipDepth(skipDepth int) WithStackTraceOption {
	return func(o *withStackTraceOptions) {
		o.skipDepth = skipDepth
	}
}

// WithStackTrace is a wrapper over original err with file:line identification
func WithStackTrace(err error, opts ...WithStackTraceOption) error {
	if err == nil {
		return nil
	}
	options := withStackTraceOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	if s, has := grpcStatus.FromError(err); has {
		return &stackTransportError{
			stackError: stackError{
				stackRecord: stack.Record(options.skipDepth + 1),
				err:         err,
			},
			status: s,
		}
	}

	return &stackError{
		stackRecord: stack.Record(options.skipDepth + 1),
		err:         err,
	}
}

type stackError struct {
	stackRecord string
	err         error
}

func (e *stackError) Error() string {
	return e.err.Error() + " " + KeywordAt + " `" + e.stackRecord + "`"
}

func (e *stackError) Unwrap() error {
	return e.err
}

type stackTransportError struct {
	stackError
	status *grpcStatus.Status
}

func (e *stackTransportError) GRPCStatus() *grpcStatus.Status {
	return e.status
}
