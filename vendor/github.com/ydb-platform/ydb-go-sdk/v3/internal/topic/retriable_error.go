package topic

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
)

const (
	DefaultStartTimeout          = value.InfiniteDuration
	connectionEstablishedTimeout = time.Minute
)

var errNil = xerrors.Wrap(errors.New("nil error is not retrieable"))

type RetrySettings struct {
	StartTimeout time.Duration // Full retry timeout
	CheckError   PublicCheckErrorRetryFunction
}

type PublicCheckErrorRetryFunction func(errInfo PublicCheckErrorRetryArgs) PublicCheckRetryResult

type PublicCheckErrorRetryArgs struct {
	Error error
}

func NewCheckRetryArgs(err error) PublicCheckErrorRetryArgs {
	return PublicCheckErrorRetryArgs{
		Error: err,
	}
}

type PublicCheckRetryResult struct {
	val int
}

var (
	PublicRetryDecisionDefault = PublicCheckRetryResult{val: 0}
	PublicRetryDecisionRetry   = PublicCheckRetryResult{val: 1}
	PublicRetryDecisionStop    = PublicCheckRetryResult{val: 2} //nolint:mnd
)

func CheckResetReconnectionCounters(lastTry, now time.Time, connectionTimeout time.Duration) bool {
	const resetAttemptEmpiricalCoefficient = 10
	if connectionTimeout == value.InfiniteDuration {
		return now.Sub(lastTry) > connectionEstablishedTimeout
	}

	return now.Sub(lastTry) > connectionTimeout*resetAttemptEmpiricalCoefficient
}

// RetryDecision check if err is retriable.
// if return nil stopRetryReason - err can be retried
// if return non nil stopRetryReason - err is not retriable and stopRetryReason contains reason,
// which should be used instead of err
func RetryDecision(checkErr error, settings RetrySettings, retriesDuration time.Duration) (
	_ backoff.Backoff,
	stopRetryReason error,
) {
	// nil is not error and doesn't need retry it.
	if checkErr == nil {
		return nil, xerrors.WithStackTrace(errNil)
	}

	// eof is retriable for topic
	if errors.Is(checkErr, io.EOF) && xerrors.RetryableError(checkErr) == nil {
		checkErr = xerrors.Retryable(checkErr, xerrors.WithName("TopicEOF"))
	}

	if retriesDuration > settings.StartTimeout {
		return nil, fmt.Errorf("ydb: topic reader reconnection timeout, last error: %w", xerrors.Unretryable(checkErr))
	}

	mode := retry.Check(checkErr)

	decision := PublicRetryDecisionDefault
	if settings.CheckError != nil {
		decision = settings.CheckError(NewCheckRetryArgs(checkErr))
	}

	switch decision {
	case PublicRetryDecisionDefault:
		isRetriable := mode.MustRetry(true)
		if !isRetriable {
			return nil, fmt.Errorf("ydb: topic reader unretriable error: %w", xerrors.Unretryable(checkErr))
		}
	case PublicRetryDecisionRetry:
		// pass
	case PublicRetryDecisionStop:
		return nil, fmt.Errorf(
			"ydb: topic reader unretriable error by check error callback: %w",
			xerrors.Unretryable(checkErr),
		)
	default:
		panic(fmt.Errorf("unexpected retry decision: %v", decision))
	}

	// checkErr is retryable error

	switch mode.BackoffType() {
	case backoff.TypeFast:
		return backoff.Fast, nil
	default:
		return backoff.Slow, nil
	}
}
