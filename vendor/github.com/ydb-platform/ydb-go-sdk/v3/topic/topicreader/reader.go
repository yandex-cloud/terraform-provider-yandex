package topicreader

import (
	"context"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreaderinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Reader allow to read message from YDB topics.
// ReadMessage or ReadMessageBatch can call concurrency with Commit, other concurrency call is denied.
//
// In other words you can have one goroutine for read messages and one goroutine for commit messages.
//
// Concurrency table
// | Method           | ReadMessage | ReadMessageBatch | Commit | Close |
// | ReadMessage      |      -      |         -        |   +    | -     |
// | ReadMessageBatch |      -      |         -        |   +    | -     |
// | Commit           |      +      |         +        |   -    | -     |
// | Close            |      -      |         -        |   -    | -     |
type Reader struct {
	reader         topicreaderinternal.Reader
	readInFlyght   atomic.Bool
	commitInFlyght atomic.Bool
}

// NewReader
// create new reader, used internally only.
func NewReader(internalReader topicreaderinternal.Reader) *Reader {
	return &Reader{reader: internalReader}
}

// WaitInit waits until the reader is initialized
// or an error occurs
func (r *Reader) WaitInit(ctx context.Context) error {
	return r.reader.WaitInit(ctx)
}

// ReadMessage read exactly one message
// exactly one of message, error is nil
func (r *Reader) ReadMessage(ctx context.Context) (*Message, error) {
	if err := r.inCall(&r.readInFlyght); err != nil {
		return nil, err
	}
	defer r.outCall(&r.readInFlyght)

	return r.reader.ReadMessage(ctx)
}

// Message contains data and metadata, readed from the server
type Message = topicreadercommon.PublicMessage

// MessageContentUnmarshaler is interface for unmarshal message content to own struct
type MessageContentUnmarshaler = topicreadercommon.PublicMessageContentUnmarshaler

// Commit receive Message, Batch of single offset
// It can be fast (by default) or sync and waite response from server
// see topicoptions.CommitMode for details.
// Fast mode of commit (default) - store commit info to internal buffer only and send it to the server later.
// Close the reader for wait to send all commits to the server.
// For topicoptions.CommitModeSync mode sync the method can return ErrCommitToExpiredSession
// it means about the message/batch was not committed because connection broken or partition routed to
// other reader by server.
// Client code should continue work normally
func (r *Reader) Commit(ctx context.Context, obj CommitRangeGetter) error {
	if err := r.inCall(&r.commitInFlyght); err != nil {
		return err
	}
	defer r.outCall(&r.commitInFlyght)

	return r.reader.Commit(ctx, obj)
}

// PopMessagesBatchTx read messages batch and commit them within tx.
// If tx failed - the batch will be received again.
//
// Now it means reconnect to the server and re-read messages from the server to the readers buffer.
// It is expensive operation and will be good to minimize transaction failures.
//
// The reconnect is implementation detail and may be changed in the future.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (r *Reader) PopMessagesBatchTx(
	ctx context.Context,
	transaction tx.Identifier,
	opts ...ReadBatchOption,
) (
	resBatch *Batch,
	resErr error,
) {
	if err := r.inCall(&r.readInFlyght); err != nil {
		return nil, err
	}
	defer r.outCall(&r.readInFlyght)

	internalTx, err := tx.AsTransaction(transaction)
	if err != nil {
		return nil, err
	}

	tracer := r.reader.Tracer()

	traceCtx := ctx
	onDone := trace.TopicOnReaderPopBatchTx(tracer, &traceCtx, r.reader.ID(), internalTx.SessionID(), internalTx)
	ctx = traceCtx

	defer func() {
		var startOffset, endOffset int64
		var messagesCount int

		if resBatch != nil {
			messagesCount = len(resBatch.Messages)
			commitRange := topicreadercommon.GetCommitRange(resBatch)
			startOffset = commitRange.CommitOffsetStart.ToInt64()
			endOffset = commitRange.CommitOffsetEnd.ToInt64()
		}
		onDone(startOffset, endOffset, messagesCount, resErr)
	}()

	return r.reader.PopBatchTx(ctx, internalTx, opts...)
}

// CommitRangeGetter interface for get commit offsets
type CommitRangeGetter = topicreadercommon.PublicCommitRangeGetter

// ReadMessageBatch
//
// Deprecated: was experimental and not actual now.
// Use ReadMessagesBatch instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func (r *Reader) ReadMessageBatch(ctx context.Context, opts ...ReadBatchOption) (*Batch, error) {
	if err := r.inCall(&r.readInFlyght); err != nil {
		return nil, err
	}
	defer r.outCall(&r.readInFlyght)

	return r.reader.ReadMessageBatch(ctx, opts...)
}

// ReadMessagesBatch read batch of messages
// Batch is ordered message group from one partition
// exactly one of Batch, err is nil
// if Batch is not nil - reader guarantee about all Batch.Messages are not nil
func (r *Reader) ReadMessagesBatch(ctx context.Context, opts ...ReadBatchOption) (*Batch, error) {
	if err := r.inCall(&r.readInFlyght); err != nil {
		return nil, err
	}
	defer r.outCall(&r.readInFlyght)

	return r.reader.ReadMessageBatch(ctx, opts...)
}

// Batch is ordered group of messages from one partition
type Batch = topicreadercommon.PublicBatch

// ReadBatchOption is type for options of read batch
type ReadBatchOption = topicreaderinternal.PublicReadBatchOption

// Close stop work with reader.
// return when reader complete internal works, flush commit buffer. You should close the Reader after use and before
// exit from a program for prevent lost last commits.
func (r *Reader) Close(ctx context.Context) error {
	// close must be non-concurrent with read and commit

	if err := r.inCall(&r.readInFlyght); err != nil {
		return err
	}
	defer r.outCall(&r.readInFlyght)

	if err := r.inCall(&r.commitInFlyght); err != nil {
		return err
	}
	defer r.outCall(&r.commitInFlyght)

	return r.reader.Close(ctx)
}

func (r *Reader) inCall(inFlight *atomic.Bool) error {
	if inFlight.CompareAndSwap(false, true) {
		return nil
	}

	return xerrors.WithStackTrace(ErrConcurrencyCall)
}

func (r *Reader) outCall(inFlight *atomic.Bool) {
	if inFlight.CompareAndSwap(true, false) {
		return
	}

	panic("ydb: topic reader out call without in call, must be never")
}
