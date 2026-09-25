package topicreaderinternal

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/background"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var (
	errReconnectRequestOutdated = xerrors.Wrap(errors.New("ydb: reconnect request outdated"))
	errReconnect                = xerrors.Wrap(errors.New("ydb: reconnect to topic grpc stream"))
	errConnectionTimeout        = xerrors.Wrap(errors.New("ydb: topic reader connection timeout for stream"))
)

type readerConnectFunc func(ctx context.Context) (batchedStreamReader, error)

type readerReconnector struct {
	background                 background.Worker
	clock                      clockwork.Clock
	logContext                 context.Context //nolint:containedctx
	retrySettings              topic.RetrySettings
	streamVal                  batchedStreamReader
	streamContextCancel        context.CancelCauseFunc
	streamErr                  error
	initErr                    error
	tracer                     *trace.Topic
	readerConnect              readerConnectFunc
	reconnectFromBadStream     chan reconnectRequest
	connectTimeout             time.Duration
	readerID                   int64
	streamConnectionInProgress empty.Chan // opened if connection in progress, closed if connection established
	initDoneCh                 empty.Chan
	m                          xsync.RWMutex
	closeOnce                  sync.Once
	initDone                   bool
}

func newReaderReconnector(
	logContext context.Context,
	readerID int64,
	connector readerConnectFunc,
	connectTimeout time.Duration,
	retrySettings topic.RetrySettings,
	tracer *trace.Topic,
) *readerReconnector {
	res := &readerReconnector{
		readerID:       readerID,
		clock:          clockwork.NewRealClock(),
		readerConnect:  connector,
		streamErr:      errUnconnected,
		connectTimeout: connectTimeout,
		logContext:     logContext,
		tracer:         tracer,
		retrySettings:  retrySettings,
	}

	if res.connectTimeout == 0 {
		res.connectTimeout = value.InfiniteDuration
	}

	res.initChannelsAndClock()
	res.start()

	return res
}

func (r *readerReconnector) TopicOnReaderStart(consumer string, err error) {
	logCtx := r.logContext
	trace.TopicOnReaderStart(r.tracer, &logCtx, r.readerID, consumer, err)
}

func (r *readerReconnector) SetLogContext(ctx context.Context) {
	r.logContext = ctx
}

func (r *readerReconnector) PopMessagesBatchTx(
	ctx context.Context,
	tx tx.Transaction,
	opts ReadMessageBatchOptions,
) (
	*topicreadercommon.PublicBatch,
	error,
) {
	return r.readWithReconnections(
		ctx,
		func(
			ctx context.Context,
			stream batchedStreamReader,
		) (
			*topicreadercommon.PublicBatch,
			error,
		) {
			return stream.PopMessagesBatchTx(ctx, tx, opts)
		},
	)
}

func (r *readerReconnector) ReadMessageBatch(
	ctx context.Context,
	opts ReadMessageBatchOptions,
) (
	*topicreadercommon.PublicBatch,
	error,
) {
	return r.readWithReconnections(
		ctx,
		func(
			ctx context.Context,
			stream batchedStreamReader,
		) (
			*topicreadercommon.PublicBatch,
			error,
		) {
			return stream.ReadMessageBatch(ctx, opts)
		},
	)
}

func (r *readerReconnector) readWithReconnections(
	ctx context.Context,
	read func(
		ctx context.Context,
		stream batchedStreamReader,
	) (*topicreadercommon.PublicBatch, error),
) (
	*topicreadercommon.PublicBatch,
	error,
) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	attempt := 0

	for {
		if attempt > 0 {
			if err := func() error {
				t := r.clock.NewTimer(backoff.Fast.Delay(attempt))
				defer t.Stop()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-t.Chan():
					return nil
				}
			}(); err != nil {
				return nil, err
			}
		}

		attempt++
		stream, err := r.stream(ctx)
		switch {
		case r.isRetriableError(err):
			r.fireReconnectOnRetryableError(stream, err)
			runtime.Gosched()

			continue
		case err != nil:
			return nil, err
		default:
			// pass
		}

		res, err := read(ctx, stream)
		if r.isRetriableError(err) {
			r.fireReconnectOnRetryableError(stream, err)
			runtime.Gosched()

			continue
		}

		return res, err
	}
}

func (r *readerReconnector) Commit(
	ctx context.Context,
	commitRange topicreadercommon.CommitRange,
) error {
	stream, err := r.stream(ctx)
	if err != nil {
		return err
	}

	err = stream.Commit(ctx, commitRange)
	r.fireReconnectOnRetryableError(stream, err)

	return err
}

func (r *readerReconnector) CloseWithError(ctx context.Context, reason error) error {
	var closeErr error
	r.closeOnce.Do(func() {
		closeErr = r.background.Close(ctx, reason)

		if r.streamVal != nil {
			streamCloseErr := r.streamVal.CloseWithError(ctx, xerrors.WithStackTrace(errReaderClosed))
			r.streamContextCancel(errReaderClosed)
			if closeErr == nil {
				closeErr = streamCloseErr
			}
		}

		r.m.WithLock(func() {
			if !r.initDone {
				r.initErr = reason
				close(r.initDoneCh)
			}
		})
	})

	return closeErr
}

func (r *readerReconnector) start() {
	r.background.Start("reconnector-loop", r.reconnectionLoop)

	// start first connection
	r.reconnectFromBadStream <- newReconnectRequest(nil, nil)
}

func (r *readerReconnector) initChannelsAndClock() {
	if r.clock == nil {
		r.clock = clockwork.NewRealClock()
	}
	r.reconnectFromBadStream = make(chan reconnectRequest, 1)
	r.streamConnectionInProgress = make(empty.Chan)
	r.initDoneCh = make(empty.Chan)
	close(r.streamConnectionInProgress) // no progress at start
}

func (r *readerReconnector) reconnectionLoop(ctx context.Context) {
	defer r.handlePanic()

	var retriesStarted time.Time
	lastTime := time.Time{}
	attempt := 0
	for {
		now := r.clock.Now()
		if topic.CheckResetReconnectionCounters(lastTime, now, r.connectTimeout) {
			attempt = 0
			retriesStarted = time.Now()
		} else {
			attempt++
		}
		lastTime = now

		var request reconnectRequest
		select {
		case <-ctx.Done():
			return

		case request = <-r.reconnectFromBadStream:
			if retriesStarted.IsZero() {
				retriesStarted = time.Now()
			}
		}

		if request.reason != nil {
			retryBackoff, stopRetryReason := r.checkErrRetryMode(
				request.reason,
				r.clock.Since(retriesStarted),
			)
			if stopRetryReason == nil {
				if err := func() error {
					t := r.clock.NewTimer(retryBackoff.Delay(attempt))
					defer t.Stop()

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-t.Chan():
						return nil
					}
				}(); err != nil {
					return
				}
			} else {
				_ = r.CloseWithError(ctx, stopRetryReason)
				logCtx := r.logContext
				trace.TopicOnReaderReconnect(r.tracer, &logCtx, request.reason)(stopRetryReason)

				return
			}
		}

		err := r.reconnect(ctx, request.reason, request.oldReader)
		logCtx := r.logContext
		trace.TopicOnReaderReconnect(r.tracer, &logCtx, request.reason)(err)
	}
}

//nolint:funlen
func (r *readerReconnector) reconnect(ctx context.Context, reason error, oldReader batchedStreamReader) (err error) {
	defer func() {
		logCtx := r.logContext
		trace.TopicOnReaderReconnect(r.tracer, &logCtx, reason)(err)
	}()

	if err = ctx.Err(); err != nil {
		return err
	}

	var closedErr error
	r.m.WithRLock(func() {
		closedErr = r.background.CloseReason()
	})
	if closedErr != nil {
		return err
	}

	if stream, _ := r.stream(ctx); oldReader != stream {
		return xerrors.WithStackTrace(errReconnectRequestOutdated)
	}

	connectionInProgress := make(empty.Chan)
	defer close(connectionInProgress)

	r.m.WithLock(func() {
		r.streamConnectionInProgress = connectionInProgress
	})

	if oldReader != nil {
		_ = oldReader.CloseWithError(ctx, xerrors.WithStackTrace(errReconnect))
	}

	newStream, newStreamClose, err := r.connectWithTimeout()

	switch {
	case err == nil:
		// pass
	case r.isRetriableError(err):
		sendReason := err
		r.background.Start("ydb topic reader send reconnect message", func(ctx context.Context) {
			select {
			case r.reconnectFromBadStream <- newReconnectRequest(oldReader, sendReason):
				{
					logCtx := r.logContext
					trace.TopicOnReaderReconnectRequest(r.tracer, &logCtx, err, true)
				}
			case <-ctx.Done():
				{
					logCtx := r.logContext
					trace.TopicOnReaderReconnectRequest(r.tracer, &logCtx, err, false)
				}
			}
		})
	default:
		// unretriable error
		_ = r.CloseWithError(ctx, err)
	}

	r.m.WithLock(func() {
		r.streamErr = err
		if err == nil {
			r.streamVal = newStream
			r.streamContextCancel = newStreamClose
			if !r.initDone {
				r.initDone = true
				close(r.initDoneCh)
			}
		}
	})

	return err
}

func (r *readerReconnector) isRetriableError(err error) bool {
	_, stopReason := topic.RetryDecision(err, r.retrySettings, 0)

	return stopReason == nil
}

func (r *readerReconnector) checkErrRetryMode(err error, retriesDuration time.Duration) (
	backoffType backoff.Backoff,
	stopRetryReason error,
) {
	return topic.RetryDecision(err, r.retrySettings, retriesDuration)
}

func (r *readerReconnector) connectWithTimeout() (_ batchedStreamReader, _ context.CancelCauseFunc, err error) {
	bgContext := r.background.Context()

	if err = bgContext.Err(); err != nil {
		return nil, nil, err
	}

	connectionContext, cancel := context.WithCancelCause(xcontext.ValueOnly(bgContext))

	type connectResult struct {
		stream batchedStreamReader
		err    error
	}
	result := make(chan connectResult, 1)

	go func() {
		stream, err := r.readerConnect(connectionContext)
		result <- connectResult{stream: stream, err: err}
	}()

	connectionTimoutTimer := r.clock.NewTimer(r.connectTimeout)
	defer connectionTimoutTimer.Stop()

	var res connectResult
	select {
	case <-connectionTimoutTimer.Chan():
		// cancel connection context only if timeout exceed while connection
		// because if cancel context after connect - it will break
		cancel(xerrors.WithStackTrace(errConnectionTimeout))
		res = <-result
	case res = <-result:
		// pass
	}

	if res.err == nil {
		return res.stream, cancel, nil
	}

	cancel(res.err)

	return nil, nil, res.err
}

func (r *readerReconnector) WaitInit(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.initDoneCh:
		return r.initErr
	case <-r.background.Done():
		return r.background.CloseReason()
	}
}

func (r *readerReconnector) fireReconnectOnRetryableError(stream batchedStreamReader, err error) {
	if !r.isRetriableError(err) {
		return
	}

	select {
	case r.reconnectFromBadStream <- newReconnectRequest(stream, err):
		{
			// send signal
			logCtx := r.logContext
			trace.TopicOnReaderReconnectRequest(r.tracer, &logCtx, err, true)
		}
	default:
		{
			// previous reconnect signal in process, no need sent signal more
			logCtx := r.logContext
			trace.TopicOnReaderReconnectRequest(r.tracer, &logCtx, err, false)
		}
	}
}

func (r *readerReconnector) stream(ctx context.Context) (batchedStreamReader, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var err error
	var connectionChan empty.Chan
	r.m.WithRLock(func() {
		connectionChan = r.streamConnectionInProgress
		err = r.background.CloseReason()
	})
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.background.Done():
		return nil, r.background.CloseReason()
	case <-connectionChan:
		var reader batchedStreamReader
		r.m.WithRLock(func() {
			reader = r.streamVal
			err = r.streamErr
		})
		r.fireReconnectOnRetryableError(reader, err)

		return reader, err
	}
}

func (r *readerReconnector) handlePanic() {
	if p := recover(); p != nil {
		_ = r.CloseWithError(context.Background(), xerrors.WithStackTrace(fmt.Errorf("handled panic: %v", p)))
	}
}

type reconnectRequest struct {
	oldReader batchedStreamReader
	reason    error
}

func newReconnectRequest(oldReader batchedStreamReader, reason error) reconnectRequest {
	return reconnectRequest{
		oldReader: oldReader,
		reason:    reason,
	}
}
