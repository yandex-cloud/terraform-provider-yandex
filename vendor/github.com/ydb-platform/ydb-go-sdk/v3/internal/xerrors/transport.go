package xerrors

import (
	"errors"
	"fmt"
	"strconv"

	grpcCodes "google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

type transportError struct {
	status  *grpcStatus.Status
	err     error
	address string
	nodeID  uint32
	traceID string
}

func (e *transportError) GRPCStatus() *grpcStatus.Status {
	return e.status
}

func (e *transportError) isYdbError() {}

func (e *transportError) Code() int32 {
	return int32(e.status.Code())
}

func (e *transportError) Name() string {
	return "transport/" + e.status.Code().String()
}

type teOpt interface {
	applyToTransportError(te *transportError)
}

type addressOption string

func (address addressOption) applyToTransportError(te *transportError) {
	te.address = string(address)
}

func WithAddress(address string) addressOption {
	return addressOption(address)
}

type nodeIDOption uint32

func (nodeID nodeIDOption) applyToTransportError(te *transportError) {
	te.nodeID = uint32(nodeID)
}

func WithNodeID(nodeID uint32) nodeIDOption {
	return nodeIDOption(nodeID)
}

func (e *transportError) Error() string {
	b := xstring.Buffer()
	defer b.Free()
	b.WriteString(e.Name())
	fmt.Fprintf(b, " (code = %d, source error = %q", e.status.Code(), e.err.Error())
	if len(e.address) > 0 {
		fmt.Fprintf(b, ", address: %q", e.address)
	}
	if e.nodeID > 0 {
		b.WriteString(", nodeID = ")
		b.WriteString(strconv.FormatUint(uint64(e.nodeID), 10))
	}
	if len(e.traceID) > 0 {
		fmt.Fprintf(b, ", traceID: %q", e.traceID)
	}
	b.WriteString(")")

	return b.String()
}

func (e *transportError) Unwrap() error {
	return e.err
}

func (e *transportError) Type() Type {
	switch e.status.Code() {
	case
		grpcCodes.Aborted,
		grpcCodes.ResourceExhausted:
		return TypeRetryable
	case
		grpcCodes.Internal,
		grpcCodes.Canceled,
		grpcCodes.DeadlineExceeded,
		grpcCodes.Unavailable:
		return TypeConditionallyRetryable
	default:
		return TypeUndefined
	}
}

func (e *transportError) BackoffType() backoff.Type {
	switch e.status.Code() {
	case
		grpcCodes.Internal,
		grpcCodes.Canceled,
		grpcCodes.DeadlineExceeded,
		grpcCodes.Unavailable:
		return backoff.TypeFast
	case grpcCodes.ResourceExhausted:
		return backoff.TypeSlow
	default:
		return backoff.TypeNoBackoff
	}
}

// IsTransportError reports whether err is transportError with given grpc codes
func IsTransportError(err error, codes ...grpcCodes.Code) bool {
	if err == nil {
		return false
	}
	var status *grpcStatus.Status
	if t := (*transportError)(nil); errors.As(err, &t) {
		status = t.status
	} else if t, has := grpcStatus.FromError(err); has {
		status = t
	}
	if status != nil {
		if len(codes) == 0 {
			return true
		}
		for _, code := range codes {
			if status.Code() == code {
				return true
			}
		}
	}

	return false
}

// Transport returns a new transport error with given options
func Transport(err error, opts ...teOpt) error {
	if err == nil {
		return nil
	}
	var te *transportError
	if errors.As(err, &te) {
		return te
	}
	if s, ok := grpcStatus.FromError(err); ok {
		te = &transportError{
			status: s,
			err:    err,
		}
	} else {
		te = &transportError{
			status: grpcStatus.New(grpcCodes.Unknown, stack.Record(1)),
			err:    err,
		}
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applyToTransportError(te)
		}
	}

	return te
}

func TransportError(err error) Error {
	if err == nil {
		return nil
	}
	var t *transportError
	if errors.As(err, &t) {
		return t
	}
	if s, ok := grpcStatus.FromError(err); ok {
		return &transportError{
			status: s,
			err:    err,
		}
	}

	return nil
}
