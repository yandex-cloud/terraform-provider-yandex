package topiclistenerinternal

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/background"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// extractSelectorNames extracts topic names from selectors for tracing
func extractSelectorNames(selectors []*topicreadercommon.PublicReadSelector) []string {
	result := make([]string, len(selectors))
	for i, selector := range selectors {
		result[i] = selector.Path
	}

	return result
}

type streamListener struct {
	cfg *StreamListenerConfig

	stream      topicreadercommon.RawTopicReaderStream
	streamClose context.CancelCauseFunc
	handler     EventHandler
	sessionID   string
	listenerID  string

	background       background.Worker
	sessions         *topicreadercommon.PartitionSessionStorage
	sessionIDCounter *atomic.Int64

	hasNewMessagesToSend empty.Chan
	syncCommitter        *topicreadercommon.Committer

	closing atomic.Bool
	tracer  *trace.Topic

	m              xsync.Mutex
	workers        map[rawtopicreader.PartitionSessionID]*PartitionWorker
	messagesToSend []rawtopicreader.ClientMessage
}

func newStreamListener(
	connectionCtx context.Context,
	client TopicClient,
	eventListener EventHandler,
	config *StreamListenerConfig,
	sessionIDCounter *atomic.Int64,
) (*streamListener, error) {
	// Generate unique listener ID
	listenerIDRand, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		listenerIDRand = big.NewInt(-1)
	}
	listenerID := "listener-" + listenerIDRand.String()

	res := &streamListener{
		cfg:              config,
		handler:          eventListener,
		background:       *background.NewWorker(xcontext.ValueOnly(connectionCtx), "topic reader stream listener"),
		sessionIDCounter: sessionIDCounter,
		listenerID:       listenerID,

		tracer: config.Tracer,
	}

	res.initVars(sessionIDCounter)

	logCtx := connectionCtx
	initDone := trace.TopicOnListenerInit(
		res.tracer, &logCtx, res.listenerID, res.cfg.Consumer, extractSelectorNames(res.cfg.Selectors),
	)

	if err := res.initStream(connectionCtx, client); err != nil {
		initDone("", err)
		res.goClose(connectionCtx, err)

		return nil, err
	}

	initDone(res.sessionID, nil)

	res.syncCommitter = topicreadercommon.NewCommitterStopped(
		res.tracer,
		res.background.Context(),
		topicreadercommon.CommitModeSync,
		res.stream.Send,
	)

	res.startBackground()
	res.sendDataRequest(config.BufferSize)

	return res, nil
}

func (l *streamListener) Close(ctx context.Context, reason error) error {
	if !l.closing.CompareAndSwap(false, true) {
		return errTopicListenerClosed
	}

	logCtx := ctx
	closeDone := trace.TopicOnListenerClose(l.tracer, &logCtx, l.listenerID, l.sessionID, reason)

	var resErrors []error

	// Stop all partition workers first
	// Copy workers to avoid holding mutex while closing them (to prevent deadlock)
	var workers []*PartitionWorker
	l.m.WithLock(func() {
		workers = make([]*PartitionWorker, 0, len(l.workers))
		for _, worker := range l.workers {
			workers = append(workers, worker)
		}
	})

	// Close workers without holding the mutex
	for _, worker := range workers {
		if err := worker.Close(ctx, reason); err != nil {
			resErrors = append(resErrors, err)
		}
	}

	// should be first because background wait stop of steams
	if l.stream != nil {
		l.streamClose(reason)
	}

	if err := l.background.Close(ctx, reason); err != nil {
		resErrors = append(resErrors, err)
	}

	if err := l.syncCommitter.Close(ctx, reason); err != nil {
		resErrors = append(resErrors, err)
	}

	for _, session := range l.sessions.GetAll() {
		session.Close()
		// For shutdown, we don't need to process stop partition requests through workers
		// since all workers are already being closed above
	}

	finalErr := errors.Join(resErrors...)

	closeDone(len(workers), finalErr)

	return finalErr
}

func (l *streamListener) goClose(ctx context.Context, reason error) {
	ctx, cancel := context.WithTimeout(xcontext.ValueOnly(ctx), time.Second)
	l.streamClose(reason)
	go func() {
		_ = l.background.Close(ctx, reason)
		cancel()
	}()
}

func (l *streamListener) startBackground() {
	l.background.Start("stream listener send loop", l.sendMessagesLoop)
	l.background.Start("stream listener receiver", l.receiveMessagesLoop)
	l.syncCommitter.Start()
}

func (l *streamListener) initVars(sessionIDCounter *atomic.Int64) {
	l.hasNewMessagesToSend = make(empty.Chan, 1)
	l.sessions = &topicreadercommon.PartitionSessionStorage{}
	l.sessionIDCounter = sessionIDCounter
	l.workers = make(map[rawtopicreader.PartitionSessionID]*PartitionWorker)
	if l.cfg == nil {
		l.cfg = &StreamListenerConfig{}
	}
}

//nolint:funlen
func (l *streamListener) initStream(ctx context.Context, client TopicClient) error {
	streamCtx, streamClose := context.WithCancelCause(xcontext.ValueOnly(ctx))
	l.streamClose = streamClose
	initDone := make(empty.Chan)
	defer close(initDone)

	go func() {
		select {
		case <-ctx.Done():
			err := xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
				"ydb: topic listener stream init timeout: %w", ctx.Err(),
			)))
			l.goClose(ctx, err)
			l.streamClose(err)
		case <-initDone:
			// pass
		}
	}()

	stream, err := client.StreamRead(streamCtx, -1, l.tracer)
	if err != nil {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: topic listener failed connect to a stream: %w",
			err,
		)))
	}
	l.stream = topicreadercommon.NewSyncedStream(stream)

	initMessage := topicreadercommon.CreateInitMessage(l.cfg.Consumer, false, l.cfg.Selectors)
	err = stream.Send(initMessage)
	if err != nil {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: failed to send init request for read stream in the listener: %w", err)))
	}

	resp, err := l.stream.Recv()
	if err != nil {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: failed to receive init response for read stream in the listener: %w",
			err,
		)))
	}

	if status := resp.StatusData(); !status.Status.IsSuccess() {
		// wrap initialization error as operation status error - for handle with retrier
		// https://github.com/ydb-platform/ydb-go-sdk/issues/1361
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: received bad status on init the topic stream listener: %v (%v)",
			status.Status,
			status.Issues,
		)))
	}

	initResp, ok := resp.(*rawtopicreader.InitResponse)
	if !ok {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"bad message type on session init: %v (%v)",
			resp,
			reflect.TypeOf(resp),
		)))
	}

	l.sessionID = initResp.SessionID

	return nil
}

func (l *streamListener) sendMessagesLoop(ctx context.Context) {
	chDone := ctx.Done()
	for {
		select {
		case <-chDone:
			return
		case <-l.hasNewMessagesToSend:
			var messages []rawtopicreader.ClientMessage
			l.m.WithLock(func() {
				messages = l.messagesToSend
				if len(messages) > 0 {
					l.messagesToSend = make([]rawtopicreader.ClientMessage, 0, cap(messages))
				}
			})

			if len(messages) == 0 {
				continue
			}

			logCtx := l.background.Context()

			for i, m := range messages {
				messageType := l.getMessageTypeName(m)

				if err := l.stream.Send(m); err != nil {
					// Trace send error
					l.traceMessageSend(&logCtx, messageType, err)
					trace.TopicOnListenerError(l.tracer, &logCtx, l.listenerID, l.sessionID, err)

					reason := xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
						"ydb: failed send message by grpc to topic reader stream from listener: "+
							"message_type=%s, message_index=%d, total_messages=%d: %w",
						messageType, i, len(messages), err,
					)))
					l.goClose(ctx, reason)

					return
				}

				// Trace successful send
				l.traceMessageSend(&logCtx, messageType, nil)
			}
		}
	}
}

// traceMessageSend provides consistent tracing for message sends
func (l *streamListener) traceMessageSend(ctx *context.Context, messageType string, err error) {
	// Use TopicOnListenerSendDataRequest for all message send tracing
	trace.TopicOnListenerSendDataRequest(l.tracer, ctx, l.listenerID, l.sessionID, messageType, err)
}

// getMessageTypeName returns a human-readable name for the message type
func (l *streamListener) getMessageTypeName(m rawtopicreader.ClientMessage) string {
	switch m.(type) {
	case *rawtopicreader.ReadRequest:
		return "ReadRequest"
	case *rawtopicreader.StartPartitionSessionResponse:
		return "StartPartitionSessionResponse"
	case *rawtopicreader.StopPartitionSessionResponse:
		return "StopPartitionSessionResponse"
	case *rawtopicreader.CommitOffsetRequest:
		return "CommitOffsetRequest"
	case *rawtopicreader.PartitionSessionStatusRequest:
		return "PartitionSessionStatusRequest"
	case *rawtopicreader.UpdateTokenRequest:
		return "UpdateTokenRequest"
	default:
		return fmt.Sprintf("Unknown(%T)", m)
	}
}

func (l *streamListener) receiveMessagesLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		mess, err := l.stream.Recv()

		logCtx := ctx
		if err != nil {
			trace.TopicOnListenerReceiveMessage(l.tracer, &logCtx, l.listenerID, l.sessionID, "", 0, err)
			trace.TopicOnListenerError(l.tracer, &logCtx, l.listenerID, l.sessionID, err)
			l.goClose(ctx, xerrors.WithStackTrace(xerrors.Wrap(
				fmt.Errorf("ydb: failed read message from the stream in the topic reader listener: %w", err),
			)))

			return
		}

		messageType := reflect.TypeOf(mess).String()
		bytesSize := 0
		if mess, ok := mess.(*rawtopicreader.ReadResponse); ok {
			bytesSize = mess.BytesSize
		}

		trace.TopicOnListenerReceiveMessage(l.tracer, &logCtx, l.listenerID, l.sessionID, messageType, bytesSize, nil)

		if err := l.routeMessage(ctx, mess); err != nil {
			trace.TopicOnListenerError(l.tracer, &logCtx, l.listenerID, l.sessionID, err)
			l.goClose(ctx, err)
		}
	}
}

// routeMessage routes messages to appropriate handlers/workers
func (l *streamListener) routeMessage(ctx context.Context, mess rawtopicreader.ServerMessage) error {
	switch m := mess.(type) {
	case *rawtopicreader.StartPartitionSessionRequest:
		return l.handleStartPartition(ctx, m)
	case *rawtopicreader.StopPartitionSessionRequest:
		return l.routeToWorker(m.PartitionSessionID, func(worker *PartitionWorker) {
			worker.AddRawServerMessage(m)
		})
	case *rawtopicreader.ReadResponse:
		return l.splitAndRouteReadResponse(m)
	case *rawtopicreader.CommitOffsetResponse:
		return l.onCommitResponse(m)
	default:
		// Ignore unknown message types
		return nil
	}
}

// handleStartPartition creates a new worker and routes StartPartition message to it
func (l *streamListener) handleStartPartition(
	ctx context.Context,
	m *rawtopicreader.StartPartitionSessionRequest,
) error {
	session := topicreadercommon.NewPartitionSession(
		ctx,
		m.PartitionSession.Path,
		m.PartitionSession.PartitionID,
		l.cfg.readerID,
		l.sessionID,
		m.PartitionSession.PartitionSessionID,
		l.sessionIDCounter.Add(1),
		m.CommittedOffset,
	)
	if err := l.sessions.Add(session); err != nil {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf("ydb: failed to add partition session: %w", err)))
	}

	// Create worker for this partition
	worker := l.createWorkerForPartition(session)

	// Send StartPartition message to the worker
	worker.AddRawServerMessage(m)

	return nil
}

// splitAndRouteReadResponse splits ReadResponse into batches and routes to workers
func (l *streamListener) splitAndRouteReadResponse(m *rawtopicreader.ReadResponse) error {
	batches, err := topicreadercommon.ReadRawBatchesToPublicBatches(m, l.sessions, l.cfg.Decoders)
	if err != nil {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: failed to convert raw batches to public batches: %w", err)))
	}

	// Route each batch to its partition worker
	for _, batch := range batches {
		partitionSession := topicreadercommon.BatchGetPartitionSession(batch)
		err := l.routeToWorker(partitionSession.StreamPartitionSessionID, func(worker *PartitionWorker) {
			worker.AddMessagesBatch(m.ServerMessageMetadata, batch)
		})
		if err != nil {
			return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
				"ydb: failed to route batch to worker: %w", err)))
		}
	}

	return nil
}

// onCommitResponse processes CommitOffsetResponse directly in streamListener
// This prevents blocking commits in PartitionWorker threads
func (l *streamListener) onCommitResponse(msg *rawtopicreader.CommitOffsetResponse) error {
	for i := range msg.PartitionsCommittedOffsets {
		commit := &msg.PartitionsCommittedOffsets[i]

		var worker *PartitionWorker
		l.m.WithLock(func() {
			worker = l.workers[commit.PartitionSessionID]
		})

		if worker == nil {
			// Session not found - this can happen during shutdown, log but don't fail
			continue
		}

		session := worker.partitionSession

		// Update committed offset in the session
		session.SetCommittedOffsetForward(commit.CommittedOffset)

		// Notify the syncCommitter about the commit
		l.syncCommitter.OnCommitNotify(session, commit.CommittedOffset)

		// Emit trace event - use partition context instead of background
		logCtx := session.Context()
		trace.TopicOnReaderCommittedNotify(
			l.tracer,
			&logCtx,
			l.listenerID,
			session.Topic,
			session.PartitionID,
			session.StreamPartitionSessionID.ToInt64(),
			commit.CommittedOffset.ToInt64(),
		)
	}

	return nil
}

func (l *streamListener) sendCommit(b *topicreadercommon.PublicBatch) error {
	commitRanges := topicreadercommon.CommitRanges{
		Ranges: []topicreadercommon.CommitRange{topicreadercommon.GetCommitRange(b)},
	}

	if err := l.stream.Send(commitRanges.ToRawMessage()); err != nil {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf("ydb: failed to send commit message: %w", err)))
	}

	return nil
}

// getSyncCommitter returns the syncCommitter for CommitHandler interface compatibility
func (l *streamListener) getSyncCommitter() SyncCommitter {
	return l.syncCommitter
}

func (l *streamListener) sendDataRequest(bytesCount int) {
	l.sendMessage(&rawtopicreader.ReadRequest{BytesSize: bytesCount})
}

func (l *streamListener) sendMessage(m rawtopicreader.ClientMessage) {
	l.m.WithLock(func() {
		l.messagesToSend = append(l.messagesToSend, m)
	})

	select {
	case l.hasNewMessagesToSend <- empty.Struct{}:
	default:
	}
}

type confirmStorage[T any] struct {
	doneChan      empty.Chan
	confirmed     atomic.Bool
	val           T
	confirmAction sync.Once
	initAction    sync.Once
}

func (c *confirmStorage[T]) init() {
	c.initAction.Do(func() {
		c.doneChan = make(empty.Chan)
	})
}

func (c *confirmStorage[T]) Set(val T) {
	c.init()
	c.confirmAction.Do(func() {
		c.val = val
		c.confirmed.Store(true)
		close(c.doneChan)
	})
}

func (c *confirmStorage[T]) Done() empty.ChanReadonly {
	c.init()

	return c.doneChan
}

func (c *confirmStorage[T]) Get() (val T, ok bool) {
	c.init()

	if c.confirmed.Load() {
		return c.val, true
	}

	return val, false
}

// SendRaw implements MessageSender interface for PartitionWorkers
func (l *streamListener) SendRaw(msg rawtopicreader.ClientMessage) {
	l.sendMessage(msg)
}

// onWorkerStopped handles worker stopped notifications
func (l *streamListener) onWorkerStopped(sessionID rawtopicreader.PartitionSessionID, reason error) {
	// Remove worker from workers map
	l.m.WithLock(func() {
		delete(l.workers, sessionID)
	})

	// Remove corresponding session
	for _, session := range l.sessions.GetAll() {
		if session.StreamPartitionSessionID == sessionID {
			_, _ = l.sessions.Remove(session.StreamPartitionSessionID)

			break
		}
	}

	// If reason from worker, propagate to streamListener shutdown
	// But avoid cascading shutdowns for normal lifecycle events like queue closure during shutdown
	if reason != nil && !l.closing.Load() {
		// Only propagate reason if we're not already closing
		// and if it's not a normal queue closure reason (which can happen during shutdown)
		if !xerrors.Is(reason, errPartitionQueueClosed) {
			l.goClose(l.background.Context(), reason)
		}
	}
}

// createWorkerForPartition creates a new PartitionWorker for the given session
func (l *streamListener) createWorkerForPartition(session *topicreadercommon.PartitionSession) *PartitionWorker {
	worker := NewPartitionWorker(
		session.StreamPartitionSessionID,
		session,
		l, // streamListener implements MessageSender and CommitHandler
		l.handler,
		l.onWorkerStopped,
		l.tracer,
		l.listenerID,
	)

	// Store worker in map
	l.m.WithLock(func() {
		l.workers[session.StreamPartitionSessionID] = worker
	})

	// Start worker
	worker.Start(l.background.Context())

	return worker
}

// routeToWorker routes a message to the appropriate worker
func (l *streamListener) routeToWorker(
	partitionSessionID rawtopicreader.PartitionSessionID,
	routeFunc func(*PartitionWorker),
) error {
	// Find worker by session
	var targetWorker *PartitionWorker
	l.m.WithLock(func() {
		targetWorker = l.workers[partitionSessionID]
	})

	if targetWorker != nil {
		routeFunc(targetWorker)
	}

	// Log error for missing worker but don't fail - this indicates a serious protocol/state issue
	// In production, this should be extremely rare and indicates server/client state mismatch
	return nil
}
