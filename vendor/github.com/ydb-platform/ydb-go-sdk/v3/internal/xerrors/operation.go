package xerrors

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Issue"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

// operationError reports about operation fail.
type operationError struct {
	code    Ydb.StatusIds_StatusCode
	nodeID  uint32
	issues  issues
	address string
	traceID string
}

func (e *operationError) isYdbError() {}

func (e *operationError) Code() int32 {
	return int32(e.code)
}

func (e *operationError) Name() string {
	return "operation/" + e.code.String()
}

type issuesOption []*Ydb_Issue.IssueMessage

func (issues issuesOption) applyToOperationError(oe *operationError) {
	oe.issues = []*Ydb_Issue.IssueMessage(issues)
}

// WithIssues is an option for construct operation error with issues list
// WithIssues must use as `Operation(WithIssues(issues))`
func WithIssues(issues []*Ydb_Issue.IssueMessage) issuesOption {
	return issues
}

type statusCodeOption Ydb.StatusIds_StatusCode

func (code statusCodeOption) applyToOperationError(oe *operationError) {
	oe.code = Ydb.StatusIds_StatusCode(code)
}

// WithStatusCode is an option for construct operation error with reason code
// WithStatusCode must use as `Operation(WithStatusCode(reason))`
func WithStatusCode(code Ydb.StatusIds_StatusCode) statusCodeOption {
	return statusCodeOption(code)
}

func (address addressOption) applyToOperationError(oe *operationError) {
	oe.address = string(address)
}

func (nodeID nodeIDOption) applyToOperationError(oe *operationError) {
	oe.nodeID = uint32(nodeID)
}

type traceIDOption string

func (traceID traceIDOption) applyToTransportError(te *transportError) {
	te.traceID = string(traceID)
}

func (traceID traceIDOption) applyToOperationError(oe *operationError) {
	oe.traceID = string(traceID)
}

// WithTraceID is an option for construct operation error with traceID
func WithTraceID(traceID string) traceIDOption {
	return traceIDOption(traceID)
}

type operationOption struct {
	code   Ydb.StatusIds_StatusCode
	issues issues
}

func (e *operationOption) applyToOperationError(oe *operationError) {
	oe.code = e.code
	oe.issues = e.issues
}

// FromOperation is an option for construct operation error from operation.Status
// FromOperation must use as `Operation(FromOperation(operation.Status))`
func FromOperation(operation operation.Status) *operationOption {
	return &operationOption{
		code:   operation.GetStatus(),
		issues: operation.GetIssues(),
	}
}

type oeOpt interface {
	applyToOperationError(oe *operationError)
}

func Operation(opts ...oeOpt) error {
	oe := &operationError{
		code: Ydb.StatusIds_STATUS_CODE_UNSPECIFIED,
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applyToOperationError(oe)
		}
	}

	return oe
}

func (e *operationError) Issues() []*Ydb_Issue.IssueMessage {
	return e.issues
}

func (e *operationError) Error() string {
	b := xstring.Buffer()
	defer b.Free()
	b.WriteString(e.Name())
	fmt.Fprintf(b, " (code = %d", e.code)
	if len(e.address) > 0 {
		b.WriteString(", address = ")
		b.WriteString(e.address)
	}
	if e.nodeID > 0 {
		b.WriteString(", nodeID = ")
		b.WriteString(strconv.FormatUint(uint64(e.nodeID), 10))
	}
	if len(e.issues) > 0 {
		b.WriteString(", issues = ")
		b.WriteString(e.issues.String())
	}
	b.WriteString(")")

	return b.String()
}

// IsOperationError reports whether err is operationError with given errType codes.
func IsOperationError(err error, codes ...Ydb.StatusIds_StatusCode) bool {
	var op *operationError
	if !errors.As(err, &op) {
		return false
	}
	if len(codes) == 0 {
		return true
	}
	for _, code := range codes {
		if op.code == code {
			return true
		}
	}

	return false
}

func IsOperationErrorTransactionLocksInvalidated(err error) (isTLI bool) {
	return IsOperationError(err, Ydb.StatusIds_ABORTED) &&
		iterateByIssues(err, func(message string, code Ydb.StatusIds_StatusCode, severity uint32) (stop bool) {
			return code == IssueCodeTransactionLocksInvalidated
		})
}

func (e *operationError) Type() Type {
	switch e.code {
	case
		Ydb.StatusIds_UNAVAILABLE,
		Ydb.StatusIds_OVERLOADED,
		Ydb.StatusIds_BAD_SESSION,
		Ydb.StatusIds_SESSION_BUSY:
		return TypeRetryable
	case
		Ydb.StatusIds_UNDETERMINED,
		Ydb.StatusIds_SESSION_EXPIRED:
		return TypeConditionallyRetryable
	case Ydb.StatusIds_UNAUTHORIZED:
		return TypeNonRetryable
	case Ydb.StatusIds_ABORTED:
		if e.hasIssueCodes(IssueCodeDatashardProgramSizeLimitExceeded) {
			return TypeNonRetryable
		}

		return TypeRetryable
	default:
		return TypeUndefined
	}
}

func (e *operationError) hasIssueCodes(codes ...Ydb.StatusIds_StatusCode) bool {
	return iterateByIssues(e, func(message string, code Ydb.StatusIds_StatusCode, severity uint32) (stop bool) {
		return slices.Contains(codes, code)
	})
}

func (e *operationError) BackoffType() backoff.Type {
	switch e.code {
	case Ydb.StatusIds_OVERLOADED:
		return backoff.TypeSlow
	case
		Ydb.StatusIds_UNAVAILABLE,
		Ydb.StatusIds_CANCELLED,
		Ydb.StatusIds_SESSION_BUSY,
		Ydb.StatusIds_UNDETERMINED:
		return backoff.TypeFast
	case Ydb.StatusIds_ABORTED:
		if e.hasIssueCodes(IssueCodeDatashardProgramSizeLimitExceeded) {
			return backoff.TypeNoBackoff
		}

		return backoff.TypeFast
	default:
		return backoff.TypeNoBackoff
	}
}

func OperationError(err error) Error {
	var o *operationError
	if errors.As(err, &o) {
		return o
	}

	return nil
}
