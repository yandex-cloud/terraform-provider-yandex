package trace

import (
	"context"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"
)

// tool gtrace used from ./internal/cmd/gtrace

//go:generate gtrace

type (
	// Topic specified trace of topic reader client activity.
	// gtrace:gen
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	Topic struct {
		// TopicReaderCustomerEvents - upper level, on bridge with customer code

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderStart func(info TopicReaderStartInfo)

		// TopicReaderStreamLifeCycleEvents

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderReconnect func(TopicReaderReconnectStartInfo) func(TopicReaderReconnectDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderReconnectRequest func(TopicReaderReconnectRequestInfo)

		// TopicReaderPartitionEvents

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderPartitionReadStartResponse func(
			TopicReaderPartitionReadStartResponseStartInfo,
		) func(
			TopicReaderPartitionReadStartResponseDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderPartitionReadStopResponse func(
			TopicReaderPartitionReadStopResponseStartInfo,
		) func(
			TopicReaderPartitionReadStopResponseDoneInfo,
		)

		OnReaderEndPartitionSession func(TopicReaderEndPartitionSessionInfo)

		// TopicReaderStreamEvents

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderCommit func(TopicReaderCommitStartInfo) func(TopicReaderCommitDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderSendCommitMessage func(TopicReaderSendCommitMessageStartInfo) func(TopicReaderSendCommitMessageDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderCommittedNotify func(TopicReaderCommittedNotifyInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderClose func(TopicReaderCloseStartInfo) func(TopicReaderCloseDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderInit func(TopicReaderInitStartInfo) func(TopicReaderInitDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderError func(TopicReaderErrorInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderUpdateToken func(
			OnReadUpdateTokenStartInfo,
		) func(
			OnReadUpdateTokenMiddleTokenReceivedInfo,
		) func(
			OnReadStreamUpdateTokenDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderPopBatchTx func(TopicReaderPopBatchTxStartInfo) func(TopicReaderPopBatchTxDoneInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderStreamPopBatchTx func(
			TopicReaderStreamPopBatchTxStartInfo,
		) func(
			TopicReaderStreamPopBatchTxDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderUpdateOffsetsInTransaction func(
			TopicReaderOnUpdateOffsetsInTransactionStartInfo,
		) func(
			TopicReaderOnUpdateOffsetsInTransactionDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderTransactionCompleted func(
			TopicReaderTransactionCompletedStartInfo,
		) func(
			TopicReaderTransactionCompletedDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderTransactionRollback func(
			TopicReaderTransactionRollbackStartInfo,
		) func(
			TopicReaderTransactionRollbackDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderSentGRPCMessage func(TopicReaderSentGRPCMessageInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderReceiveGRPCMessage func(TopicReaderReceiveGRPCMessageInfo)

		// TopicReaderMessageEvents

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderSentDataRequest func(TopicReaderSentDataRequestInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderReceiveDataResponse func(TopicReaderReceiveDataResponseStartInfo) func(TopicReaderReceiveDataResponseDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderReadMessages func(TopicReaderReadMessagesStartInfo) func(TopicReaderReadMessagesDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnReaderUnknownGrpcMessage func(OnReadUnknownGrpcMessageInfo)

		// TopicWriterStreamLifeCycleEvents

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterReconnect func(TopicWriterReconnectStartInfo) func(TopicWriterReconnectConnectedInfo) func(TopicWriterReconnectDoneInfo) //nolint:lll
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterInitStream func(TopicWriterInitStreamStartInfo) func(TopicWriterInitStreamDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterClose func(TopicWriterCloseStartInfo) func(TopicWriterCloseDoneInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterBeforeCommitTransaction func(
			TopicOnWriterBeforeCommitTransactionStartInfo,
		) func(TopicOnWriterBeforeCommitTransactionDoneInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterAfterFinishTransaction func(
			TopicOnWriterAfterFinishTransactionStartInfo,
		) func(TopicOnWriterAfterFinishTransactionDoneInfo)

		// TopicWriterStreamEvents

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterCompressMessages func(TopicWriterCompressMessagesStartInfo) func(TopicWriterCompressMessagesDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterSendMessages func(TopicWriterSendMessagesStartInfo) func(TopicWriterSendMessagesDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterReceiveResult func(TopicWriterResultMessagesInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterSentGRPCMessage func(TopicWriterSentGRPCMessageInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterReceiveGRPCMessage func(TopicWriterReceiveGRPCMessageInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWriterReadUnknownGrpcMessage func(TopicOnWriterReadUnknownGrpcMessageInfo)

		// TopicListenerEvents

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerStart func(TopicListenerStartInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerInit func(TopicListenerInitStartInfo) func(TopicListenerInitDoneInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerReceiveMessage func(TopicListenerReceiveMessageInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerRouteMessage func(TopicListenerRouteMessageInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerSplitMessage func(TopicListenerSplitMessageInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerError func(TopicListenerErrorInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerClose func(TopicListenerCloseStartInfo) func(TopicListenerCloseDoneInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnPartitionWorkerStart func(TopicPartitionWorkerStartInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnPartitionWorkerProcessMessage func(TopicPartitionWorkerProcessMessageStartInfo) func(
			TopicPartitionWorkerProcessMessageDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnPartitionWorkerHandlerCall func(TopicPartitionWorkerHandlerCallStartInfo) func(
			TopicPartitionWorkerHandlerCallDoneInfo,
		)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnPartitionWorkerStop func(TopicPartitionWorkerStopStartInfo) func(TopicPartitionWorkerStopDoneInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerSendDataRequest func(TopicListenerSendDataRequestInfo)

		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnListenerUnknownMessage func(TopicListenerUnknownMessageInfo)
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderPartitionReadStartResponseStartInfo struct {
		ReaderConnectionID string
		PartitionContext   *context.Context
		Topic              string
		PartitionID        int64
		PartitionSessionID int64
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderStartInfo struct {
		Context  *context.Context
		ReaderID int64
		Consumer string
		Error    error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderPartitionReadStartResponseDoneInfo struct {
		ReadOffset   *int64
		CommitOffset *int64
		Error        error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderPartitionReadStopResponseStartInfo struct {
		ReaderConnectionID string
		PartitionContext   context.Context //nolint:containedctx
		Topic              string
		PartitionID        int64
		PartitionSessionID int64
		CommittedOffset    int64
		Graceful           bool
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderPartitionReadStopResponseDoneInfo struct {
		Error error
	}

	TopicReaderEndPartitionSessionInfo struct {
		ReaderConnectionID   string
		PartitionContext     context.Context //nolint:containedctx
		Topic                string
		PartitionID          int64
		PartitionSessionID   int64
		AdjacentPartitionIDs []int64
		ChildPartitionIDs    []int64
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderSendCommitMessageStartInfo struct {
		Context     *context.Context
		CommitsInfo TopicReaderStreamSendCommitMessageStartMessageInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderStreamCommitInfo struct {
		Context            *context.Context
		Topic              string
		PartitionID        int64
		PartitionSessionID int64
		StartOffset        int64
		EndOffset          int64
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderStreamSendCommitMessageStartMessageInfo interface {
		GetCommitsInfo() []TopicReaderStreamCommitInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderSendCommitMessageDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderCommittedNotifyInfo struct {
		Context            *context.Context
		ReaderConnectionID string
		Topic              string
		PartitionID        int64
		PartitionSessionID int64
		CommittedOffset    int64
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderErrorInfo struct {
		Context            *context.Context
		ReaderConnectionID string
		Error              error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderSentDataRequestInfo struct {
		Context                  *context.Context
		ReaderConnectionID       string
		RequestBytes             int
		LocalBufferSizeAfterSent int
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReceiveDataResponseStartInfo struct {
		Context                     *context.Context
		ReaderConnectionID          string
		LocalBufferSizeAfterReceive int
		DataResponse                TopicReaderDataResponseInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderDataResponseInfo interface {
		GetBytesSize() int
		GetPartitionBatchMessagesCounts() (partitionCount, batchCount, messagesCount int)
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReceiveDataResponseDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReadMessagesStartInfo struct {
		Context            *context.Context
		MinCount           int
		MaxCount           int
		FreeBufferCapacity int
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReadMessagesDoneInfo struct {
		MessagesCount      int
		Topic              string
		PartitionID        int64
		PartitionSessionID int64
		OffsetStart        int64
		OffsetEnd          int64
		FreeBufferCapacity int
		Error              error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	OnReadUnknownGrpcMessageInfo struct {
		Context            *context.Context
		ReaderConnectionID string
		Error              error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReconnectStartInfo struct {
		Context *context.Context
		Reason  error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReconnectDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReconnectRequestInfo struct {
		Context *context.Context
		Reason  error
		WasSent bool
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderCommitStartInfo struct {
		Context            *context.Context
		Topic              string
		PartitionID        int64
		PartitionSessionID int64
		StartOffset        int64
		EndOffset          int64
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderCommitDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderCloseStartInfo struct {
		Context            *context.Context
		ReaderConnectionID string
		CloseReason        error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderCloseDoneInfo struct {
		CloseError error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderInitStartInfo struct {
		Context                   *context.Context
		PreInitReaderConnectionID string
		InitRequestInfo           TopicReadStreamInitRequestInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReadStreamInitRequestInfo interface {
		GetConsumer() string
		GetTopics() []string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderInitDoneInfo struct {
		ReaderConnectionID string
		Error              error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	OnReadUpdateTokenStartInfo struct {
		Context            *context.Context
		ReaderConnectionID string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	OnReadUpdateTokenMiddleTokenReceivedInfo struct {
		Context  *context.Context
		TokenLen int
		Error    error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	OnReadStreamUpdateTokenDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderPopBatchTxStartInfo struct {
		Context              *context.Context
		ReaderID             int64
		TransactionSessionID string
		Tx                   txInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderPopBatchTxDoneInfo struct {
		StartOffset   int64
		EndOffset     int64
		MessagesCount int
		Error         error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderStreamPopBatchTxStartInfo struct {
		Context              *context.Context
		ReaderID             int64
		ReaderConnectionID   string
		TransactionSessionID string
		Tx                   txInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderStreamPopBatchTxDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderOnUpdateOffsetsInTransactionStartInfo struct {
		Context              *context.Context
		ReaderID             int64
		ReaderConnectionID   string
		TransactionSessionID string
		Tx                   txInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderOnUpdateOffsetsInTransactionDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderTransactionCompletedStartInfo struct {
		Context              *context.Context
		ReaderID             int64
		ReaderConnectionID   string
		TransactionSessionID string
		Tx                   txInfo
		TransactionResult    error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderTransactionCompletedDoneInfo struct{}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderTransactionRollbackStartInfo struct {
		Context              *context.Context
		ReaderID             int64
		ReaderConnectionID   string
		TransactionSessionID string
		Tx                   txInfo
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderTransactionRollbackDoneInfo struct {
		RollbackError error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderSentGRPCMessageInfo struct {
		ReaderID      int64
		SessionID     string
		MessageNumber int
		Message       *Ydb_Topic.StreamReadMessage_FromClient // may be nil if err != nil
		Error         error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicReaderReceiveGRPCMessageInfo struct {
		ReaderID      int64
		SessionID     string
		MessageNumber int
		Message       *Ydb_Topic.StreamReadMessage_FromServer // may be nil if err != nil
		Error         error
	}

	////////////
	//////////// TopicWriter
	////////////

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterReconnectStartInfo struct {
		Context          *context.Context
		WriterInstanceID string
		Topic            string
		ProducerID       string
		Attempt          int
	}

	TopicWriterReconnectConnectedInfo struct {
		ConnectionResult error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterReconnectDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterInitStreamStartInfo struct {
		Context          *context.Context
		WriterInstanceID string
		Topic            string
		ProducerID       string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterInitStreamDoneInfo struct {
		SessionID string
		Error     error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterCloseStartInfo struct {
		Context          *context.Context
		WriterInstanceID string
		Reason           error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterCloseDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterCompressMessagesStartInfo struct {
		Context          *context.Context
		WriterInstanceID string
		SessionID        string
		Codec            int32
		FirstSeqNo       int64
		MessagesCount    int
		Reason           TopicWriterCompressMessagesReason
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterCompressMessagesDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterSendMessagesStartInfo struct {
		Context          *context.Context
		WriterInstanceID string
		SessionID        string
		Codec            int32
		FirstSeqNo       int64
		MessagesCount    int
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterSendMessagesDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterResultMessagesInfo struct {
		Context          *context.Context
		WriterInstanceID string
		SessionID        string
		PartitionID      int64
		Acks             TopicWriterResultMessagesInfoAcks
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterResultMessagesInfoAcks interface {
		GetAcks() struct {
			AcksCount        int
			SeqNoMin         int64
			SeqNoMax         int64
			WrittenOffsetMin int64
			WrittenOffsetMax int64
			WrittenCount     int
			WrittenInTxCount int
			SkipCount        int
		}
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicOnWriterBeforeCommitTransactionStartInfo struct {
		Context        *context.Context
		KqpSessionID   string
		TopicSessionID string
		TransactionID  string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicOnWriterBeforeCommitTransactionDoneInfo struct {
		Error          error
		TopicSessionID string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicOnWriterAfterFinishTransactionStartInfo struct {
		Context       *context.Context
		Error         error
		SessionID     string
		TransactionID string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicOnWriterAfterFinishTransactionDoneInfo struct {
		CloseError error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterSentGRPCMessageInfo struct {
		Context               *context.Context
		TopicStreamInternalID string
		SessionID             string
		MessageNumber         int
		Message               *Ydb_Topic.StreamWriteMessage_FromClient
		Error                 error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterReceiveGRPCMessageInfo struct {
		Context               *context.Context
		TopicStreamInternalID string
		SessionID             string
		MessageNumber         int
		Message               *Ydb_Topic.StreamWriteMessage_FromServer
		Error                 error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicOnWriterReadUnknownGrpcMessageInfo struct {
		Context          *context.Context
		WriterInstanceID string
		SessionID        string
		Error            error
	}
)

// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
type TopicWriterCompressMessagesReason string

const (
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterCompressMessagesReasonCompressData = TopicWriterCompressMessagesReason("compress-on-send")
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterCompressMessagesReasonCompressDataOnWriteReadData = TopicWriterCompressMessagesReason("compress-on-call-write") //nolint:lll
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicWriterCompressMessagesReasonCodecsMeasure = TopicWriterCompressMessagesReason("compress-on-codecs-measure") //nolint:lll
)

// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
func (r TopicWriterCompressMessagesReason) String() string {
	return string(r)
}

type (
	// TopicListener Events

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerStartInfo struct {
		Context    *context.Context
		ListenerID string
		Consumer   string
		Error      error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerInitStartInfo struct {
		Context        *context.Context
		ListenerID     string
		Consumer       string
		TopicSelectors []string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerInitDoneInfo struct {
		SessionID string
		Error     error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerReceiveMessageInfo struct {
		Context     *context.Context
		ListenerID  string
		SessionID   string
		MessageType string
		BytesSize   int
		Error       error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerRouteMessageInfo struct {
		Context            *context.Context
		ListenerID         string
		SessionID          string
		MessageType        string
		PartitionSessionID *int64
		WorkerFound        bool
		Error              error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerSplitMessageInfo struct {
		Context         *context.Context
		ListenerID      string
		SessionID       string
		MessageType     string
		TotalBatches    int
		TotalPartitions int
		SplitBatches    int
		RoutedBatches   int
		Error           error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerErrorInfo struct {
		Context    *context.Context
		ListenerID string
		SessionID  string
		Error      error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerCloseStartInfo struct {
		Context    *context.Context
		ListenerID string
		SessionID  string
		Reason     error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerCloseDoneInfo struct {
		WorkersClosed int
		Error         error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicPartitionWorkerStartInfo struct {
		Context            *context.Context
		ListenerID         string
		SessionID          string
		PartitionSessionID int64
		PartitionID        int64
		Topic              string
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicPartitionWorkerProcessMessageStartInfo struct {
		Context            *context.Context
		ListenerID         string
		SessionID          string
		PartitionSessionID int64
		PartitionID        int64
		Topic              string
		MessageType        string
		MessagesCount      int
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicPartitionWorkerProcessMessageDoneInfo struct {
		ProcessedMessages int
		Error             error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicPartitionWorkerHandlerCallStartInfo struct {
		Context            *context.Context
		ListenerID         string
		SessionID          string
		PartitionSessionID int64
		PartitionID        int64
		Topic              string
		HandlerType        string // "OnReadMessages", "OnStartPartition", "OnStopPartition"
		MessagesCount      int
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicPartitionWorkerHandlerCallDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicPartitionWorkerStopStartInfo struct {
		Context            *context.Context
		ListenerID         string
		SessionID          string
		PartitionSessionID int64
		PartitionID        int64
		Topic              string
		Reason             error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicPartitionWorkerStopDoneInfo struct {
		Error error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerSendDataRequestInfo struct {
		Context     *context.Context
		ListenerID  string
		SessionID   string
		MessageType string
		Error       error
	}

	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	TopicListenerUnknownMessageInfo struct {
		Context     *context.Context
		ListenerID  string
		SessionID   string
		MessageType string
		Error       error
	}
)
