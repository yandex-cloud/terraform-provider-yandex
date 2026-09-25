package topiclistenerinternal

import (
	"context"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/background"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

//go:generate mockgen -source partition_worker.go -destination partition_worker_mock_test.go --typed -package topiclistenerinternal -write_package_comment=false

var errPartitionQueueClosed = xerrors.Wrap(fmt.Errorf("ydb: partition messages queue closed"))

// MessageSender sends messages back to server
type MessageSender interface {
	SendRaw(msg rawtopicreader.ClientMessage)
}

// unifiedMessage wraps messages that PartitionWorker can handle
type unifiedMessage struct {
	// Only one of these should be set
	RawServerMessage *rawtopicreader.ServerMessage
	BatchMessage     *batchMessage
}

// batchMessage represents a ready PublicBatch message with metadata
type batchMessage struct {
	ServerMessageMetadata rawtopiccommon.ServerMessageMetadata
	Batch                 *topicreadercommon.PublicBatch
}

// WorkerStoppedCallback notifies when worker is stopped
type WorkerStoppedCallback func(sessionID rawtopicreader.PartitionSessionID, reason error)

// PartitionWorker processes messages for a single partition
type PartitionWorker struct {
	partitionSessionID rawtopicreader.PartitionSessionID
	partitionSession   *topicreadercommon.PartitionSession
	messageSender      MessageSender
	userHandler        EventHandler
	onStopped          WorkerStoppedCallback

	// Tracing and logging fields
	tracer     *trace.Topic
	listenerID string

	messageQueue *xsync.UnboundedChan[unifiedMessage]
	bgWorker     *background.Worker
}

// NewPartitionWorker creates a new PartitionWorker instance
func NewPartitionWorker(
	sessionID rawtopicreader.PartitionSessionID,
	session *topicreadercommon.PartitionSession,
	messageSender MessageSender,
	userHandler EventHandler,
	onStopped WorkerStoppedCallback,
	tracer *trace.Topic,
	listenerID string,
) *PartitionWorker {
	// Validate required parameters
	if userHandler == nil {
		panic("userHandler cannot be nil")
	}

	return &PartitionWorker{
		partitionSessionID: sessionID,
		partitionSession:   session,
		messageSender:      messageSender,
		userHandler:        userHandler,
		onStopped:          onStopped,
		tracer:             tracer,
		listenerID:         listenerID,
		messageQueue:       xsync.NewUnboundedChan[unifiedMessage](),
	}
}

// Start begins processing messages for this partition
func (w *PartitionWorker) Start(ctx context.Context) {
	// Add trace for worker creation
	trace.TopicOnPartitionWorkerStart(
		w.tracer,
		&ctx,
		w.listenerID,
		"", // sessionID - not available in worker context
		int64(w.partitionSessionID),
		w.partitionSession.PartitionID,
		w.partitionSession.Topic,
	)

	w.bgWorker = background.NewWorker(ctx, "partition worker")
	w.bgWorker.Start("partition worker message loop", func(bgCtx context.Context) {
		w.receiveMessagesLoop(bgCtx)
	})
}

// AddUnifiedMessage adds a unified message to the processing queue
func (w *PartitionWorker) AddUnifiedMessage(msg unifiedMessage) {
	w.messageQueue.SendWithMerge(msg, w.tryMergeMessages)
}

// AddRawServerMessage sends a raw server message
func (w *PartitionWorker) AddRawServerMessage(msg rawtopicreader.ServerMessage) {
	w.AddUnifiedMessage(unifiedMessage{RawServerMessage: &msg})
}

// AddMessagesBatch sends a ready batch message
func (w *PartitionWorker) AddMessagesBatch(
	metadata rawtopiccommon.ServerMessageMetadata,
	batch *topicreadercommon.PublicBatch,
) {
	w.AddUnifiedMessage(unifiedMessage{
		BatchMessage: &batchMessage{
			ServerMessageMetadata: metadata,
			Batch:                 batch,
		},
	})
}

// Close stops the worker gracefully
func (w *PartitionWorker) Close(ctx context.Context, reason error) error {
	w.messageQueue.Close()
	if w.bgWorker != nil {
		return w.bgWorker.Close(ctx, reason)
	}

	return nil
}

// receiveMessagesLoop is the main message processing loop
func (w *PartitionWorker) receiveMessagesLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			reason := xerrors.WithStackTrace(fmt.Errorf("ydb: partition worker panic: %v", r))
			w.onStopped(w.partitionSessionID, reason)
		}
	}()

	for {
		// Use context-aware Receive method
		msg, ok, err := w.messageQueue.Receive(ctx)
		if err != nil {
			// Context was cancelled or timed out
			if ctx.Err() == context.Canceled {
				reason := xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
					"graceful shutdown PartitionWorker: topic=%s, partition=%d, partitionSession=%d",
					w.partitionSession.Topic,
					w.partitionSession.PartitionID,
					w.partitionSessionID,
				)))
				w.onStopped(w.partitionSessionID, reason)
			} else {
				reason := xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
					"partition worker message queue context error: topic=%s, partition=%d, partitionSession=%d: %w",
					w.partitionSession.Topic,
					w.partitionSession.PartitionID,
					w.partitionSessionID,
					err,
				)))
				w.onStopped(w.partitionSessionID, reason)
			}

			return
		}
		if !ok {
			// Queue was closed - this is a normal stop reason during shutdown
			reason := xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
				"partition worker message queue closed: topic=%s, partition=%d, partitionSession=%d",
				w.partitionSession.Topic,
				w.partitionSession.PartitionID,
				w.partitionSessionID,
			)))
			w.onStopped(w.partitionSessionID, reason)

			return
		}

		if err := w.processUnifiedMessage(ctx, msg); err != nil {
			w.onStopped(w.partitionSessionID, err)

			return
		}
	}
}

// processUnifiedMessage handles a single unified message by routing to appropriate processor
func (w *PartitionWorker) processUnifiedMessage(ctx context.Context, msg unifiedMessage) error {
	switch {
	case msg.RawServerMessage != nil:
		return w.processRawServerMessage(ctx, *msg.RawServerMessage)
	case msg.BatchMessage != nil:
		return w.processBatchMessage(ctx, msg.BatchMessage)
	default:
		// Ignore empty messages
		return nil
	}
}

// processRawServerMessage handles raw server messages (StartPartition, StopPartition)
func (w *PartitionWorker) processRawServerMessage(ctx context.Context, msg rawtopicreader.ServerMessage) error {
	switch m := msg.(type) {
	case *rawtopicreader.StartPartitionSessionRequest:
		return w.handleStartPartitionRequest(ctx, m)
	case *rawtopicreader.StopPartitionSessionRequest:
		return w.handleStopPartitionRequest(ctx, m)
	default:
		// Ignore unknown raw message types (e.g., ReadResponse which is now handled via BatchMessage)
		return nil
	}
}

// calculateBatchMetrics calculates metrics for a batch message
func (w *PartitionWorker) calculateBatchMetrics(msg *batchMessage) (messagesCount int, accountBytes int) {
	if msg.Batch != nil && len(msg.Batch.Messages) > 0 {
		messagesCount = len(msg.Batch.Messages)

		for i := range msg.Batch.Messages {
			accountBytes += topicreadercommon.MessageGetBufferBytesAccount(msg.Batch.Messages[i])
		}
	}

	return messagesCount, accountBytes
}

// validateBatchMetadata checks if the batch metadata status is successful
func (w *PartitionWorker) validateBatchMetadata(msg *batchMessage) error {
	if !msg.ServerMessageMetadata.Status.IsSuccess() {
		err := xerrors.WithStackTrace(fmt.Errorf(
			"ydb: batch message contains error status: %v",
			msg.ServerMessageMetadata.Status,
		))

		return err
	}

	return nil
}

// callUserHandler calls the user handler for a batch message with tracing
func (w *PartitionWorker) callUserHandler(
	ctx context.Context,
	msg *batchMessage,
	commitHandler CommitHandler,
	messagesCount int,
) error {
	event := NewPublicReadMessages(
		w.partitionSession.ToPublic(),
		msg.Batch,
		commitHandler,
	)

	// Add handler call tracing
	handlerTraceDone := trace.TopicOnPartitionWorkerHandlerCall(
		w.tracer,
		&ctx,
		w.listenerID,
		"", // sessionID - not available in worker context
		int64(w.partitionSessionID),
		w.partitionSession.PartitionID,
		w.partitionSession.Topic,
		"OnReadMessages",
		messagesCount,
	)

	if err := w.userHandler.OnReadMessages(ctx, event); err != nil {
		handlerErr := xerrors.WithStackTrace(err)
		handlerTraceDone(handlerErr)

		return handlerErr
	}

	handlerTraceDone(nil)

	return nil
}

// processBatchMessage handles ready PublicBatch messages
func (w *PartitionWorker) processBatchMessage(ctx context.Context, msg *batchMessage) error {
	// Add tracing for batch processing
	messagesCount, accountBytes := w.calculateBatchMetrics(msg)

	traceDone := trace.TopicOnPartitionWorkerProcessMessage(
		w.tracer,
		&ctx,
		w.listenerID,
		"", // sessionID - not available in worker context
		int64(w.partitionSessionID),
		w.partitionSession.PartitionID,
		w.partitionSession.Topic,
		"BatchMessage",
		messagesCount,
	)

	// Check for errors in the metadata
	if err := w.validateBatchMetadata(msg); err != nil {
		traceDone(0, err)

		return err
	}

	// Cast messageSender to CommitHandler (it's the streamListener)
	commitHandler, ok := w.messageSender.(CommitHandler)
	if !ok {
		err := xerrors.WithStackTrace(fmt.Errorf("ydb: messageSender does not implement CommitHandler"))
		traceDone(0, err)

		return err
	}

	// Call user handler with tracing
	if err := w.callUserHandler(ctx, msg, commitHandler, messagesCount); err != nil {
		traceDone(0, err)

		return err
	}

	// Use estimated bytes size for flow control
	w.messageSender.SendRaw(&rawtopicreader.ReadRequest{BytesSize: accountBytes})

	traceDone(messagesCount, nil)

	return nil
}

// handleStartPartitionRequest processes StartPartitionSessionRequest
func (w *PartitionWorker) handleStartPartitionRequest(
	ctx context.Context,
	m *rawtopicreader.StartPartitionSessionRequest,
) error {
	event := NewPublicStartPartitionSessionEvent(
		w.partitionSession.ToPublic(),
		m.CommittedOffset.ToInt64(),
		PublicOffsetsRange{
			Start: m.PartitionOffsets.Start.ToInt64(),
			End:   m.PartitionOffsets.End.ToInt64(),
		},
	)

	err := w.userHandler.OnStartPartitionSessionRequest(ctx, event)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	// Wait for user confirmation
	var userResp PublicStartPartitionSessionConfirm
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-event.confirm.Done():
		userResp, _ = event.confirm.Get()
	}

	// Build response
	resp := &rawtopicreader.StartPartitionSessionResponse{
		PartitionSessionID: m.PartitionSession.PartitionSessionID,
	}

	if userResp.readOffset != nil {
		resp.ReadOffset.Offset.FromInt64(*userResp.readOffset)
		resp.ReadOffset.HasValue = true
	}
	if userResp.CommitOffset != nil {
		resp.CommitOffset.Offset.FromInt64(*userResp.CommitOffset)
		resp.CommitOffset.HasValue = true
	}

	w.messageSender.SendRaw(resp)

	return nil
}

// handleStopPartitionRequest processes StopPartitionSessionRequest
func (w *PartitionWorker) handleStopPartitionRequest(
	ctx context.Context,
	m *rawtopicreader.StopPartitionSessionRequest,
) error {
	event := NewPublicStopPartitionSessionEvent(
		w.partitionSession.ToPublic(),
		m.Graceful,
		m.CommittedOffset.ToInt64(),
	)

	if err := w.userHandler.OnStopPartitionSessionRequest(ctx, event); err != nil {
		return xerrors.WithStackTrace(err)
	}

	// Wait for user confirmation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-event.confirm.Done():
		// pass
	}

	// Only send response if graceful
	if m.Graceful {
		resp := &rawtopicreader.StopPartitionSessionResponse{
			PartitionSessionID: w.partitionSession.StreamPartitionSessionID,
		}
		w.messageSender.SendRaw(resp)
	}

	return nil
}

// tryMergeMessages attempts to merge messages when possible
func (w *PartitionWorker) tryMergeMessages(last, new unifiedMessage) (unifiedMessage, bool) {
	// Only merge batch messages for now
	if last.BatchMessage != nil && new.BatchMessage != nil {
		// Validate metadata compatibility before merging
		if !last.BatchMessage.ServerMessageMetadata.Equals(&new.BatchMessage.ServerMessageMetadata) {
			return new, false // Don't merge messages with different metadata
		}

		var err error
		result, err := topicreadercommon.BatchAppend(last.BatchMessage.Batch, new.BatchMessage.Batch)
		if err != nil {
			w.Close(context.Background(), err)

			return new, false
		}

		return unifiedMessage{BatchMessage: &batchMessage{
			ServerMessageMetadata: last.BatchMessage.ServerMessageMetadata,
			Batch:                 result,
		}}, true
	}

	// Don't merge other types of messages
	return new, false
}
