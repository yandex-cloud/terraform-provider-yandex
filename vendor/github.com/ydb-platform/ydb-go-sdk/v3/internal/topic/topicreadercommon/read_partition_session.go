package topicreadercommon

import (
	"context"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
)

type PartitionSession struct {
	Topic       string
	PartitionID int64

	ReaderID     int64
	connectionID string

	ctx                      context.Context //nolint:containedctx
	ctxCancel                context.CancelFunc
	StreamPartitionSessionID rawtopicreader.PartitionSessionID
	ClientPartitionSessionID int64

	lastReceivedOffsetEndVal atomic.Int64
	committedOffsetVal       atomic.Int64
	noMoreMessages           atomic.Bool
}

func NewPartitionSession(
	partitionContext context.Context,
	topic string,
	partitionID int64,
	readerID int64,
	connectionID string,
	partitionSessionID rawtopicreader.PartitionSessionID,
	clientPartitionSessionID int64,
	committedOffset rawtopiccommon.Offset,
) *PartitionSession {
	partitionContext, cancel := xcontext.WithCancel(partitionContext)

	res := &PartitionSession{
		Topic:                    topic,
		PartitionID:              partitionID,
		ReaderID:                 readerID,
		connectionID:             connectionID,
		ctx:                      partitionContext,
		ctxCancel:                cancel,
		StreamPartitionSessionID: partitionSessionID,
		ClientPartitionSessionID: clientPartitionSessionID,
	}
	res.SetInitialCommitOffset(committedOffset)

	return res
}

func (s *PartitionSession) Context() context.Context {
	return s.ctx
}

func (s *PartitionSession) SetContext(ctx context.Context) {
	s.ctx, s.ctxCancel = xcontext.WithCancel(ctx)
}

func (s *PartitionSession) Close() {
	s.ctxCancel()
}

func (s *PartitionSession) CommittedOffset() rawtopiccommon.Offset {
	v := s.committedOffsetVal.Load()

	var res rawtopiccommon.Offset
	res.FromInt64(v)

	return res
}

// SetCommittedOffsetForward set new offset if new offset greater, then old
func (s *PartitionSession) SetCommittedOffsetForward(v rawtopiccommon.Offset) {
	newVal := int64(v)
	for {
		old := s.committedOffsetVal.Load()
		if newVal <= old {
			return
		}

		if s.committedOffsetVal.CompareAndSwap(old, newVal) {
			return
		}
	}
}

func (s *PartitionSession) LastReceivedMessageOffset() rawtopiccommon.Offset {
	v := s.lastReceivedOffsetEndVal.Load()

	var res rawtopiccommon.Offset
	res.FromInt64(v)

	return res
}

func (s *PartitionSession) SetLastReceivedMessageOffset(v rawtopiccommon.Offset) {
	s.lastReceivedOffsetEndVal.Store(v.ToInt64())
}

func (s *PartitionSession) SetInitialCommitOffset(committedOffset rawtopiccommon.Offset) {
	s.committedOffsetVal.Store(committedOffset.ToInt64())
	s.lastReceivedOffsetEndVal.Store(committedOffset.ToInt64() - 1)
}

func (s *PartitionSession) NoMoreMessages() bool {
	return s.noMoreMessages.Load()
}

func (s *PartitionSession) SetNoMoreMessages() {
	s.noMoreMessages.Store(true)
}

func (s *PartitionSession) ToPublic() PublicPartitionSession {
	return PublicPartitionSession{
		PartitionSessionID: s.ClientPartitionSessionID,
		TopicPath:          s.Topic,
		PartitionID:        s.PartitionID,
	}
}

// PublicPartitionSession contains information about partition session for the event
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type PublicPartitionSession struct {
	// PartitionSessionID is unique session ID per listener object
	PartitionSessionID int64

	// TopicPath contains path for the topic
	TopicPath string

	// PartitionID contains partition id. It can be repeated for one reader if the partition will stop/start few times.
	PartitionID int64
}
