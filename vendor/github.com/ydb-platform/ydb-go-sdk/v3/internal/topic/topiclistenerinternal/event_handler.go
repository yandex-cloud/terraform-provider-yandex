package topiclistenerinternal

import (
	"context"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
)

//go:generate mockgen -source event_handler.go -destination event_handler_mock_test.go --typed -package topiclistenerinternal -write_package_comment=false

// CommitHandler interface for PublicReadMessages commit operations
type CommitHandler interface {
	sendCommit(b *topicreadercommon.PublicBatch) error
	getSyncCommitter() SyncCommitter
}

// SyncCommitter interface for ConfirmWithAck support
type SyncCommitter interface {
	Commit(ctx context.Context, commitRange topicreadercommon.CommitRange) error
}

type EventHandler interface {
	// OnStartPartitionSessionRequest called when server send start partition session request method.
	// You can use it to store read progress on your own side.
	// You must call event.Confirm(...) for start to receive messages from the partition.
	// You can set topiclistener.StartPartitionSessionConfirm for change default settings.
	//
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	OnStartPartitionSessionRequest(ctx context.Context, event *PublicEventStartPartitionSession) error

	// OnReadMessages called with batch of messages. Max count of messages limited by internal buffer size
	//
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	OnReadMessages(ctx context.Context, event *PublicReadMessages) error

	// OnStopPartitionSessionRequest called when the server send stop partition message.
	// It means that no more OnReadMessages calls for the partition session.
	// You must call event.Confirm() for allow the server to stop the partition session (if event.Graceful=true).
	// Confirm is optional for event.Graceful=false
	// The method can be called twice: with event.Graceful=true, then event.Graceful=false.
	// It is guaranteed about the method will be called least once.
	//
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	OnStopPartitionSessionRequest(ctx context.Context, event *PublicEventStopPartitionSession) error
}

// PublicReadMessages
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type PublicReadMessages struct {
	PartitionSession topicreadercommon.PublicPartitionSession
	Batch            *topicreadercommon.PublicBatch
	commitHandler    CommitHandler
	committed        atomic.Bool
}

func NewPublicReadMessages(
	session topicreadercommon.PublicPartitionSession,
	batch *topicreadercommon.PublicBatch,
	commitHandler CommitHandler,
) *PublicReadMessages {
	return &PublicReadMessages{
		PartitionSession: session,
		Batch:            batch,
		commitHandler:    commitHandler,
	}
}

// Confirm of the process messages from the batch.
// Send commit message the server in background. The method returns fast, without wait commits ack.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (e *PublicReadMessages) Confirm() {
	if e.committed.Swap(true) {
		return
	}

	_ = e.commitHandler.sendCommit(e.Batch)
}

// ConfirmWithAck commit the batch and wait ack from the server. The method will be blocked until
// receive ack, error or expire ctx.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (e *PublicReadMessages) ConfirmWithAck(ctx context.Context) error {
	if e.committed.Swap(true) {
		return nil
	}

	return e.commitHandler.getSyncCommitter().Commit(ctx, topicreadercommon.GetCommitRange(e.Batch))
}

// PublicEventStartPartitionSession
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type PublicEventStartPartitionSession struct {
	PartitionSession topicreadercommon.PublicPartitionSession
	CommittedOffset  int64
	PartitionOffsets PublicOffsetsRange
	confirm          confirmStorage[PublicStartPartitionSessionConfirm]
}

func NewPublicStartPartitionSessionEvent(
	session topicreadercommon.PublicPartitionSession,
	committedOffset int64,
	partitionOffsets PublicOffsetsRange,
) *PublicEventStartPartitionSession {
	return &PublicEventStartPartitionSession{
		PartitionSession: session,
		CommittedOffset:  committedOffset,
		PartitionOffsets: partitionOffsets,
	}
}

// Confirm
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (e *PublicEventStartPartitionSession) Confirm() {
	e.ConfirmWithParams(PublicStartPartitionSessionConfirm{})
}

func (e *PublicEventStartPartitionSession) ConfirmWithParams(p PublicStartPartitionSessionConfirm) {
	e.confirm.Set(p)
}

// PublicStartPartitionSessionConfirm
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type PublicStartPartitionSessionConfirm struct {
	readOffset   *int64
	CommitOffset *int64 ``
}

// WithReadOffet
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c PublicStartPartitionSessionConfirm) WithReadOffet(val int64) PublicStartPartitionSessionConfirm {
	c.readOffset = &val

	return c
}

// WithCommitOffset
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c PublicStartPartitionSessionConfirm) WithCommitOffset(val int64) PublicStartPartitionSessionConfirm {
	c.CommitOffset = &val

	return c
}

// PublicOffsetsRange
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type PublicOffsetsRange struct {
	Start int64
	End   int64
}

// PublicEventStopPartitionSession
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type PublicEventStopPartitionSession struct {
	PartitionSession topicreadercommon.PublicPartitionSession

	// Graceful mean about server is waiting for client finish work with the partition and confirm stop the work
	// if the field is false it mean about server stop lease the partition to the client and can assignee the partition
	// to other read session (on this or other connection).
	Graceful        bool
	CommittedOffset int64

	confirm confirmStorage[empty.Struct]
}

func NewPublicStopPartitionSessionEvent(
	partitionSession topicreadercommon.PublicPartitionSession,
	graceful bool,
	committedOffset int64,
) *PublicEventStopPartitionSession {
	return &PublicEventStopPartitionSession{
		PartitionSession: partitionSession,
		Graceful:         graceful,
		CommittedOffset:  committedOffset,
	}
}

// Confirm
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (e *PublicEventStopPartitionSession) Confirm() {
	e.confirm.Set(empty.Struct{})
}
