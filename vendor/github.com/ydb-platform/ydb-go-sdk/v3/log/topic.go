package log

import (
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/kv"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Topic returns trace.Topic with logging events from details
func Topic(l Logger, d trace.Detailer, opts ...Option) (t trace.Topic) {
	return internalTopic(wrapLogger(l, opts...), d)
}

//nolint:gocyclo,funlen
func internalTopic(l Logger, d trace.Detailer) (t trace.Topic) {
	t.OnReaderReconnect = func(
		info trace.TopicReaderReconnectStartInfo,
	) func(doneInfo trace.TopicReaderReconnectDoneInfo) {
		if d.Details()&trace.TopicReaderStreamLifeCycleEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "reconnect")
		start := time.Now()
		l.Log(ctx, "topic reader reconnect starting...")

		return func(doneInfo trace.TopicReaderReconnectDoneInfo) {
			l.Log(WithLevel(ctx, INFO), "topic reader reconnect done",
				kv.NamedError("reason", info.Reason),
				kv.Latency(start),
			)
		}
	}
	t.OnReaderReconnectRequest = func(info trace.TopicReaderReconnectRequestInfo) {
		if d.Details()&trace.TopicReaderStreamLifeCycleEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "reconnect", "request")
		l.Log(ctx, "topic reader reconnect request",
			kv.NamedError("reason", info.Reason),
			kv.Bool("was_sent", info.WasSent),
		)
	}
	t.OnReaderPartitionReadStartResponse = func(
		info trace.TopicReaderPartitionReadStartResponseStartInfo,
	) func(stopInfo trace.TopicReaderPartitionReadStartResponseDoneInfo) {
		if d.Details()&trace.TopicReaderPartitionEvents == 0 {
			return nil
		}
		ctx := with(*info.PartitionContext, TRACE, "ydb", "topic", "reader", "partition", "read", "start", "response")
		start := time.Now()
		l.Log(ctx, "topic reader start partition read response starting...",
			kv.String("topic", info.Topic),
			kv.String("reader_connection_id", info.ReaderConnectionID),
			kv.Int64("partition_id", info.PartitionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
		)

		return func(doneInfo trace.TopicReaderPartitionReadStartResponseDoneInfo) {
			fields := []Field{
				kv.String("topic", info.Topic),
				kv.String("reader_connection_id", info.ReaderConnectionID),
				kv.Int64("partition_id", info.PartitionID),
				kv.Int64("partition_session_id", info.PartitionSessionID),
				kv.Latency(start),
			}
			if doneInfo.CommitOffset != nil {
				fields = append(fields,
					kv.Int64("commit_offset", *doneInfo.CommitOffset),
				)
			}
			if doneInfo.ReadOffset != nil {
				fields = append(fields,
					kv.Int64("read_offset", *doneInfo.ReadOffset),
				)
			}
			if doneInfo.Error == nil {
				l.Log(WithLevel(ctx, INFO), "topic reader start partition read response done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic reader start partition read response failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}
	t.OnReaderPartitionReadStopResponse = func(
		info trace.TopicReaderPartitionReadStopResponseStartInfo,
	) func(trace.TopicReaderPartitionReadStopResponseDoneInfo) {
		if d.Details()&trace.TopicReaderPartitionEvents == 0 {
			return nil
		}
		ctx := with(info.PartitionContext, TRACE, "ydb", "topic", "reader", "partition", "read", "stop", "response")
		start := time.Now()
		l.Log(ctx, "topic reader stop partition read response starting...",
			kv.String("reader_connection_id", info.ReaderConnectionID),
			kv.String("topic", info.Topic),
			kv.Int64("partition_id", info.PartitionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
			kv.Int64("committed_offset", info.CommittedOffset),
			kv.Bool("graceful", info.Graceful))

		return func(doneInfo trace.TopicReaderPartitionReadStopResponseDoneInfo) {
			fields := []Field{
				kv.String("reader_connection_id", info.ReaderConnectionID),
				kv.String("topic", info.Topic),
				kv.Int64("partition_id", info.PartitionID),
				kv.Int64("partition_session_id", info.PartitionSessionID),
				kv.Int64("committed_offset", info.CommittedOffset),
				kv.Bool("graceful", info.Graceful),
				kv.Latency(start),
			}
			if doneInfo.Error == nil {
				l.Log(WithLevel(ctx, INFO), "topic reader stop partition read response done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic reader stop partition read response failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}
	t.OnReaderCommit = func(info trace.TopicReaderCommitStartInfo) func(doneInfo trace.TopicReaderCommitDoneInfo) {
		if d.Details()&trace.TopicReaderStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "commit")
		start := time.Now()
		l.Log(ctx, "topic reader commit starting...",
			kv.String("topic", info.Topic),
			kv.Int64("partition_id", info.PartitionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
			kv.Int64("commit_start_offset", info.StartOffset),
			kv.Int64("commit_end_offset", info.EndOffset),
		)

		return func(doneInfo trace.TopicReaderCommitDoneInfo) {
			fields := []Field{
				kv.String("topic", info.Topic),
				kv.Int64("partition_id", info.PartitionID),
				kv.Int64("partition_session_id", info.PartitionSessionID),
				kv.Int64("commit_start_offset", info.StartOffset),
				kv.Int64("commit_end_offset", info.EndOffset),
				kv.Latency(start),
			}
			if doneInfo.Error == nil {
				l.Log(ctx, "topic reader commit done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic reader commit failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}
	t.OnReaderSendCommitMessage = func(
		info trace.TopicReaderSendCommitMessageStartInfo,
	) func(trace.TopicReaderSendCommitMessageDoneInfo) {
		if d.Details()&trace.TopicReaderStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "send", "commit", "message")
		start := time.Now()

		commitInfo := info.CommitsInfo.GetCommitsInfo()
		for i := range commitInfo {
			l.Log(ctx, "topic reader send commit message starting...",
				kv.String("topic", commitInfo[i].Topic),
				kv.Int64("partitions_id", commitInfo[i].PartitionID),
				kv.Int64("partitions_session_id", commitInfo[i].PartitionSessionID),
				kv.Int64("commit_start_offset", commitInfo[i].StartOffset),
				kv.Int64("commit_end_offset", commitInfo[i].EndOffset),
			)
		}

		return func(doneInfo trace.TopicReaderSendCommitMessageDoneInfo) {
			for i := range commitInfo {
				fields := []Field{
					kv.String("topic", commitInfo[i].Topic),
					kv.Int64("partitions_id", commitInfo[i].PartitionID),
					kv.Int64("partitions_session_id", commitInfo[i].PartitionSessionID),
					kv.Int64("commit_start_offset", commitInfo[i].StartOffset),
					kv.Int64("commit_end_offset", commitInfo[i].EndOffset),
					kv.Latency(start),
				}
				if doneInfo.Error == nil {
					l.Log(ctx, "topic reader send commit message done", fields...)
				} else {
					l.Log(WithLevel(ctx, WARN), "topic reader send commit message failed",
						append(fields,
							kv.Error(doneInfo.Error),
							kv.Version(),
						)...,
					)
				}
			}
		}
	}
	t.OnReaderCommittedNotify = func(info trace.TopicReaderCommittedNotifyInfo) {
		if d.Details()&trace.TopicReaderStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "committed", "notify")
		l.Log(ctx, "topic reader received commit ack",
			kv.String("reader_connection_id", info.ReaderConnectionID),
			kv.String("topic", info.Topic),
			kv.Int64("partition_id", info.PartitionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
			kv.Int64("committed_offset", info.CommittedOffset),
		)
	}
	t.OnReaderClose = func(info trace.TopicReaderCloseStartInfo) func(doneInfo trace.TopicReaderCloseDoneInfo) {
		if d.Details()&trace.TopicReaderStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "close")
		start := time.Now()
		l.Log(ctx, "topic reader close starting...",
			kv.String("reader_connection_id", info.ReaderConnectionID),
			kv.NamedError("close_reason", info.CloseReason),
		)

		return func(doneInfo trace.TopicReaderCloseDoneInfo) {
			fields := []Field{
				kv.String("reader_connection_id", info.ReaderConnectionID),
				kv.Latency(start),
			}
			if doneInfo.CloseError == nil {
				l.Log(ctx, "topic reader close done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic reader close failed",
					append(fields,
						kv.Error(doneInfo.CloseError),
						kv.Version(),
					)...,
				)
			}
		}
	}

	t.OnReaderInit = func(info trace.TopicReaderInitStartInfo) func(doneInfo trace.TopicReaderInitDoneInfo) {
		if d.Details()&trace.TopicReaderStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "init")
		start := time.Now()
		l.Log(ctx, "topic reader init starting...",
			kv.String("pre_init_reader_connection_id", info.PreInitReaderConnectionID),
			kv.String("consumer", info.InitRequestInfo.GetConsumer()),
			kv.Strings("topics", info.InitRequestInfo.GetTopics()),
		)

		return func(doneInfo trace.TopicReaderInitDoneInfo) {
			fields := []Field{
				kv.String("pre_init_reader_connection_id", info.PreInitReaderConnectionID),
				kv.String("consumer", info.InitRequestInfo.GetConsumer()),
				kv.Strings("topics", info.InitRequestInfo.GetTopics()),
				kv.Latency(start),
			}
			if doneInfo.Error == nil {
				l.Log(ctx, "topic reader init done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic reader init failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}
	t.OnReaderError = func(info trace.TopicReaderErrorInfo) {
		if d.Details()&trace.TopicReaderStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "error")
		l.Log(WithLevel(ctx, INFO), "topic reader has grpc stream error",
			kv.Error(info.Error),
			kv.String("reader_connection_id", info.ReaderConnectionID),
			kv.Version(),
		)
	}
	t.OnReaderUpdateToken = func(
		info trace.OnReadUpdateTokenStartInfo,
	) func(
		updateTokenInfo trace.OnReadUpdateTokenMiddleTokenReceivedInfo,
	) func(doneInfo trace.OnReadStreamUpdateTokenDoneInfo) {
		if d.Details()&trace.TopicReaderStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "update", "token")
		start := time.Now()
		l.Log(ctx, "topic reader token update starting...",
			kv.String("reader_connection_id", info.ReaderConnectionID),
		)

		return func(
			updateTokenInfo trace.OnReadUpdateTokenMiddleTokenReceivedInfo,
		) func(doneInfo trace.OnReadStreamUpdateTokenDoneInfo) {
			if updateTokenInfo.Error == nil {
				l.Log(ctx, "topic reader token update: got token done",
					kv.String("reader_connection_id", info.ReaderConnectionID),
					kv.Int("token_len", updateTokenInfo.TokenLen),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic reader token update: got token failed",
					kv.Error(updateTokenInfo.Error),
					kv.String("reader_connection_id", info.ReaderConnectionID),
					kv.Int("token_len", updateTokenInfo.TokenLen),
					kv.Latency(start),
					kv.Version(),
				)
			}

			return func(doneInfo trace.OnReadStreamUpdateTokenDoneInfo) {
				if doneInfo.Error == nil {
					l.Log(ctx, "topic reader token update done",
						kv.String("reader_connection_id", info.ReaderConnectionID),
						kv.Int("token_len", updateTokenInfo.TokenLen),
						kv.Latency(start),
					)
				} else {
					l.Log(WithLevel(ctx, WARN), "topic reader token update failed",
						kv.Error(doneInfo.Error),
						kv.String("reader_connection_id", info.ReaderConnectionID),
						kv.Int("token_len", updateTokenInfo.TokenLen),
						kv.Latency(start),
						kv.Version(),
					)
				}
			}
		}
	}
	t.OnReaderSentDataRequest = func(info trace.TopicReaderSentDataRequestInfo) {
		if d.Details()&trace.TopicReaderMessageEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "sent", "data", "request")
		l.Log(ctx, "topic reader sent data request",
			kv.String("reader_connection_id", info.ReaderConnectionID),
			kv.Int("request_bytes", info.RequestBytes),
			kv.Int("local_capacity", info.LocalBufferSizeAfterSent),
		)
	}
	t.OnReaderReceiveDataResponse = func(
		info trace.TopicReaderReceiveDataResponseStartInfo,
	) func(trace.TopicReaderReceiveDataResponseDoneInfo) {
		if d.Details()&trace.TopicReaderMessageEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "receive", "data", "response")
		start := time.Now()
		partitionsCount, batchesCount, messagesCount := info.DataResponse.GetPartitionBatchMessagesCounts()
		l.Log(ctx, "topic reader data response received, process starting...",
			kv.String("reader_connection_id", info.ReaderConnectionID),
			kv.Int("received_bytes", info.DataResponse.GetBytesSize()),
			kv.Int("local_capacity", info.LocalBufferSizeAfterReceive),
			kv.Int("partitions_count", partitionsCount),
			kv.Int("batches_count", batchesCount),
			kv.Int("messages_count", messagesCount),
		)

		return func(doneInfo trace.TopicReaderReceiveDataResponseDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(ctx, "topic reader data response received and process done",
					kv.String("reader_connection_id", info.ReaderConnectionID),
					kv.Int("received_bytes", info.DataResponse.GetBytesSize()),
					kv.Int("local_capacity", info.LocalBufferSizeAfterReceive),
					kv.Int("partitions_count", partitionsCount),
					kv.Int("batches_count", batchesCount),
					kv.Int("messages_count", messagesCount),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic reader data response received and process failed",
					kv.Error(doneInfo.Error),
					kv.String("reader_connection_id", info.ReaderConnectionID),
					kv.Int("received_bytes", info.DataResponse.GetBytesSize()),
					kv.Int("local_capacity", info.LocalBufferSizeAfterReceive),
					kv.Int("partitions_count", partitionsCount),
					kv.Int("batches_count", batchesCount),
					kv.Int("messages_count", messagesCount),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnReaderReadMessages = func(
		info trace.TopicReaderReadMessagesStartInfo,
	) func(doneInfo trace.TopicReaderReadMessagesDoneInfo) {
		if d.Details()&trace.TopicReaderMessageEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "read", "messages")
		start := time.Now()
		l.Log(ctx, "topic read messages, waiting...",
			kv.Int("min_count", info.MinCount),
			kv.Int("max_count", info.MaxCount),
			kv.Int("local_capacity_before", info.FreeBufferCapacity),
		)

		return func(doneInfo trace.TopicReaderReadMessagesDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(ctx, "topic read messages done",
					kv.Int("min_count", info.MinCount),
					kv.Int("max_count", info.MaxCount),
					kv.Int("local_capacity_before", info.FreeBufferCapacity),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic read messages failed",
					kv.Error(doneInfo.Error),
					kv.Int("min_count", info.MinCount),
					kv.Int("max_count", info.MaxCount),
					kv.Int("local_capacity_before", info.FreeBufferCapacity),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnReaderUnknownGrpcMessage = func(info trace.OnReadUnknownGrpcMessageInfo) {
		if d.Details()&trace.TopicReaderMessageEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "reader", "unknown", "grpc", "message")
		l.Log(WithLevel(ctx, INFO), "topic reader received unknown grpc message",
			kv.Error(info.Error),
			kv.String("reader_connection_id", info.ReaderConnectionID),
		)
	}

	t.OnReaderPopBatchTx = func(
		startInfo trace.TopicReaderPopBatchTxStartInfo,
	) func(trace.TopicReaderPopBatchTxDoneInfo) {
		if d.Details()&trace.TopicReaderCustomerEvents == 0 {
			return nil
		}

		start := time.Now()
		ctx := with(*startInfo.Context, TRACE, "ydb", "topic", "reader", "customer", "popbatchtx")
		l.Log(WithLevel(ctx, TRACE), "topic reader pop batch tx starting...",
			kv.Int64("reader_id", startInfo.ReaderID),
			kv.String("transaction_session_id", startInfo.TransactionSessionID),
			kv.String("transaction_id", startInfo.Tx.ID()),
		)

		return func(doneInfo trace.TopicReaderPopBatchTxDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(
					WithLevel(ctx, DEBUG), "topic reader pop batch tx done",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Int("messaged_count", doneInfo.MessagesCount),
					kv.Int64("start_offset", doneInfo.StartOffset),
					kv.Int64("end_offset", doneInfo.EndOffset),
					kv.Latency(start),
					kv.Version(),
				)
			} else {
				l.Log(
					WithLevel(ctx, WARN), "topic reader pop batch tx failed",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Error(doneInfo.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}

	t.OnReaderStreamPopBatchTx = func(
		startInfo trace.TopicReaderStreamPopBatchTxStartInfo,
	) func(
		trace.TopicReaderStreamPopBatchTxDoneInfo,
	) {
		if d.Details()&trace.TopicReaderTransactionEvents == 0 {
			return nil
		}

		start := time.Now()
		ctx := with(*startInfo.Context, TRACE, "ydb", "topic", "reader", "transaction", "popbatchtx_on_stream")
		l.Log(WithLevel(ctx, TRACE), "topic reader pop batch tx on stream level starting...",
			kv.Int64("reader_id", startInfo.ReaderID),
			kv.String("reader_connection_id", startInfo.ReaderConnectionID),
			kv.String("transaction_session_id", startInfo.TransactionSessionID),
			kv.String("transaction_id", startInfo.Tx.ID()),
			kv.Version(),
		)

		return func(doneInfo trace.TopicReaderStreamPopBatchTxDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(
					WithLevel(ctx, DEBUG), "topic reader pop batch tx on stream level done",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Latency(start),
					kv.Version(),
				)
			} else {
				l.Log(
					WithLevel(ctx, WARN), "topic reader pop batch tx on stream level failed",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Error(doneInfo.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}

	t.OnReaderUpdateOffsetsInTransaction = func(
		startInfo trace.TopicReaderOnUpdateOffsetsInTransactionStartInfo,
	) func(
		trace.TopicReaderOnUpdateOffsetsInTransactionDoneInfo,
	) {
		if d.Details()&trace.TopicReaderTransactionEvents == 0 {
			return nil
		}

		start := time.Now()
		ctx := with(*startInfo.Context, TRACE, "ydb", "topic", "reader", "transaction", "update_offsets")
		l.Log(WithLevel(ctx, TRACE), "update offsets in transaction starting...",
			kv.Int64("reader_id", startInfo.ReaderID),
			kv.String("reader_connection_id", startInfo.ReaderConnectionID),
			kv.String("transaction_session_id", startInfo.TransactionSessionID),
			kv.String("transaction_id", startInfo.Tx.ID()),
			kv.Version(),
		)

		return func(doneInfo trace.TopicReaderOnUpdateOffsetsInTransactionDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(
					WithLevel(ctx, DEBUG), "update offsets in transaction starting done",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Latency(start),
					kv.Version(),
				)
			} else {
				l.Log(
					WithLevel(ctx, WARN), "update offsets in transaction starting failed",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Error(doneInfo.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}

	t.OnReaderTransactionRollback = func(
		startInfo trace.TopicReaderTransactionRollbackStartInfo,
	) func(
		trace.TopicReaderTransactionRollbackDoneInfo,
	) {
		if d.Details()&trace.TopicReaderTransactionEvents == 0 {
			return nil
		}

		start := time.Now()
		ctx := with(*startInfo.Context, TRACE, "ydb", "topic", "reader", "transaction", "update_offsets")
		l.Log(WithLevel(ctx, TRACE), "topic reader rollback transaction starting...",
			kv.Int64("reader_id", startInfo.ReaderID),
			kv.String("reader_connection_id", startInfo.ReaderConnectionID),
			kv.String("transaction_session_id", startInfo.TransactionSessionID),
			kv.String("transaction_id", startInfo.Tx.ID()),
			kv.Version(),
		)

		return func(doneInfo trace.TopicReaderTransactionRollbackDoneInfo) {
			if doneInfo.RollbackError == nil {
				l.Log(
					WithLevel(ctx, DEBUG), "topic reader rollback transaction done",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Latency(start),
					kv.Version(),
				)
			} else {
				l.Log(
					WithLevel(ctx, WARN), "topic reader rollback transaction failed",
					kv.Int64("reader_id", startInfo.ReaderID),
					kv.String("transaction_session_id", startInfo.TransactionSessionID),
					kv.String("transaction_id", startInfo.Tx.ID()),
					kv.Error(doneInfo.RollbackError),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}

	t.OnReaderTransactionCompleted = func(
		startInfo trace.TopicReaderTransactionCompletedStartInfo,
	) func(
		trace.TopicReaderTransactionCompletedDoneInfo,
	) {
		if d.Details()&trace.TopicReaderTransactionEvents == 0 {
			return nil
		}

		// expected as very short in memory operation without errors, no need log start separately
		start := time.Now()

		return func(doneInfo trace.TopicReaderTransactionCompletedDoneInfo) {
			ctx := with(*startInfo.Context, TRACE, "ydb", "topic", "reader", "transaction", "update_offsets")
			l.Log(WithLevel(ctx, TRACE), "topic reader transaction completed",
				kv.Int64("reader_id", startInfo.ReaderID),
				kv.String("reader_connection_id", startInfo.ReaderConnectionID),
				kv.String("transaction_session_id", startInfo.TransactionSessionID),
				kv.String("transaction_id", startInfo.Tx.ID()),
				kv.Latency(start),
				kv.Version(),
			)
		}
	}

	///
	/// Topic writer
	///
	t.OnWriterReconnect = func(
		info trace.TopicWriterReconnectStartInfo,
	) func(doneInfo trace.TopicWriterReconnectConnectedInfo) func(reconnectDoneInfo trace.TopicWriterReconnectDoneInfo) {
		if d.Details()&trace.TopicWriterStreamLifeCycleEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "reconnect")
		start := time.Now()
		l.Log(ctx, "connect to topic writer stream starting...",
			kv.String("topic", info.Topic),
			kv.String("producer_id", info.ProducerID),
			kv.String("writer_instance_id", info.WriterInstanceID),
			kv.Int("attempt", info.Attempt),
		)

		return func(doneInfo trace.TopicWriterReconnectConnectedInfo) func(reconnectDoneInfo trace.TopicWriterReconnectDoneInfo) { //nolint:lll
			connectedTime := time.Now()
			if doneInfo.ConnectionResult == nil {
				l.Log(WithLevel(ctx, DEBUG), "connect to topic writer stream completed",
					kv.String("topic", info.Topic),
					kv.String("producer_id", info.ProducerID),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.Int("attempt", info.Attempt),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "connect to topic writer stream completed",
					kv.Error(doneInfo.ConnectionResult),
					kv.String("topic", info.Topic),
					kv.String("producer_id", info.ProducerID),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.Int("attempt", info.Attempt),
					kv.Latency(start),
				)
			}

			return func(reconnectDoneInfo trace.TopicWriterReconnectDoneInfo) {
				l.Log(WithLevel(ctx, INFO), "stop topic writer stream reason",
					kv.String("topic", info.Topic),
					kv.String("producer_id", info.ProducerID),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.Duration("write with topic writer stream duration", time.Since(connectedTime)),
					kv.NamedError("reason", reconnectDoneInfo.Error))
			}
		}
	}
	t.OnWriterInitStream = func(
		info trace.TopicWriterInitStreamStartInfo,
	) func(doneInfo trace.TopicWriterInitStreamDoneInfo) {
		if d.Details()&trace.TopicWriterStreamLifeCycleEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "stream", "init")
		start := time.Now()
		l.Log(ctx, "topic writer init stream starting...",
			kv.String("topic", info.Topic),
			kv.String("producer_id", info.ProducerID),
			kv.String("writer_instance_id", info.WriterInstanceID),
		)

		return func(doneInfo trace.TopicWriterInitStreamDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(WithLevel(ctx, DEBUG), "topic writer init stream done",
					kv.Error(doneInfo.Error),
					kv.String("topic", info.Topic),
					kv.String("producer_id", info.ProducerID),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.Latency(start),
					kv.String("session_id", doneInfo.SessionID),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic writer init stream failed",
					kv.Error(doneInfo.Error),
					kv.String("topic", info.Topic),
					kv.String("producer_id", info.ProducerID),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.Latency(start),
					kv.String("session_id", doneInfo.SessionID),
				)
			}
		}
	}
	t.OnWriterBeforeCommitTransaction = func(
		info trace.TopicOnWriterBeforeCommitTransactionStartInfo,
	) func(trace.TopicOnWriterBeforeCommitTransactionDoneInfo) {
		if d.Details()&trace.TopicWriterStreamLifeCycleEvents == 0 {
			return nil
		}

		start := time.Now()

		return func(doneInfo trace.TopicOnWriterBeforeCommitTransactionDoneInfo) {
			ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "beforecommit")
			l.Log(ctx, "topic writer wait of flush messages before commit transaction",
				kv.String("kqp_session_id", info.KqpSessionID),
				kv.String("topic_session_id_start", info.TopicSessionID),
				kv.String("topic_session_id_finish", doneInfo.TopicSessionID),
				kv.String("tx_id", info.TransactionID),
				kv.Latency(start),
			)
		}
	}
	t.OnWriterAfterFinishTransaction = func(
		info trace.TopicOnWriterAfterFinishTransactionStartInfo,
	) func(trace.TopicOnWriterAfterFinishTransactionDoneInfo) {
		start := time.Now()

		return func(doneInfo trace.TopicOnWriterAfterFinishTransactionDoneInfo) {
			ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "beforecommit")
			l.Log(ctx, "topic writer close writer after transaction finished",
				kv.String("kqp_session_id", info.SessionID),
				kv.String("tx_id", info.TransactionID),
				kv.Latency(start),
			)
		}
	}
	t.OnWriterClose = func(info trace.TopicWriterCloseStartInfo) func(doneInfo trace.TopicWriterCloseDoneInfo) {
		if d.Details()&trace.TopicWriterStreamLifeCycleEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "close")
		start := time.Now()
		l.Log(ctx, "topic writer close starting...",
			kv.String("writer_instance_id", info.WriterInstanceID),
			kv.NamedError("reason", info.Reason),
		)

		return func(doneInfo trace.TopicWriterCloseDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(WithLevel(ctx, DEBUG), "topic writer close done",
					kv.Error(doneInfo.Error),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.NamedError("reason", info.Reason),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic writer close failed",
					kv.Error(doneInfo.Error),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.NamedError("reason", info.Reason),
					kv.Latency(start),
				)
			}
		}
	}
	t.OnWriterCompressMessages = func(
		info trace.TopicWriterCompressMessagesStartInfo,
	) func(doneInfo trace.TopicWriterCompressMessagesDoneInfo) {
		if d.Details()&trace.TopicWriterStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "compress", "messages")
		start := time.Now()
		l.Log(ctx, "topic writer compress messages starting...",
			kv.String("writer_instance_id", info.WriterInstanceID),
			kv.String("session_id", info.SessionID),
			kv.Any("reason", info.Reason),
			kv.Any("codec", info.Codec),
			kv.Int("messages_count", info.MessagesCount),
			kv.Int64("first_seqno", info.FirstSeqNo),
		)

		return func(doneInfo trace.TopicWriterCompressMessagesDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(ctx, "topic writer compress messages done",
					kv.Error(doneInfo.Error),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.String("session_id", info.SessionID),
					kv.Any("reason", info.Reason),
					kv.Any("codec", info.Codec),
					kv.Int("messages_count", info.MessagesCount),
					kv.Int64("first_seqno", info.FirstSeqNo),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "topic writer compress messages failed",
					kv.Error(doneInfo.Error),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.String("session_id", info.SessionID),
					kv.Any("reason", info.Reason),
					kv.Any("codec", info.Codec),
					kv.Int("messages_count", info.MessagesCount),
					kv.Int64("first_seqno", info.FirstSeqNo),
					kv.Latency(start),
				)
			}
		}
	}
	t.OnWriterSendMessages = func(
		info trace.TopicWriterSendMessagesStartInfo,
	) func(doneInfo trace.TopicWriterSendMessagesDoneInfo) {
		if d.Details()&trace.TopicWriterStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "send", "messages")
		start := time.Now()
		l.Log(ctx, "topic writer send messages starting...",
			kv.String("writer_instance_id", info.WriterInstanceID),
			kv.String("session_id", info.SessionID),
			kv.Any("codec", info.Codec),
			kv.Int("messages_count", info.MessagesCount),
			kv.Int64("first_seqno", info.FirstSeqNo),
		)

		return func(doneInfo trace.TopicWriterSendMessagesDoneInfo) {
			if doneInfo.Error == nil {
				l.Log(ctx, "topic writer send messages done",
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.String("session_id", info.SessionID),
					kv.Any("codec", info.Codec),
					kv.Int("messages_count", info.MessagesCount),
					kv.Int64("first_seqno", info.FirstSeqNo),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic writer send messages failed",
					kv.Error(doneInfo.Error),
					kv.String("writer_instance_id", info.WriterInstanceID),
					kv.String("session_id", info.SessionID),
					kv.Any("codec", info.Codec),
					kv.Int("messages_count", info.MessagesCount),
					kv.Int64("first_seqno", info.FirstSeqNo),
					kv.Latency(start),
				)
			}
		}
	}
	t.OnWriterReceiveResult = func(info trace.TopicWriterResultMessagesInfo) {
		if d.Details()&trace.TopicWriterStreamEvents == 0 {
			return
		}
		acks := info.Acks.GetAcks()
		ctx := with(*info.Context, DEBUG, "ydb", "topic", "writer", "receive", "result")
		l.Log(ctx, "topic writer received result from server",
			kv.String("writer_instance_id", info.WriterInstanceID),
			kv.String("session_id", info.SessionID),
			kv.Int("acks_count", acks.AcksCount),
			kv.Int64("seq_no_min", acks.SeqNoMin),
			kv.Int64("seq_no_max", acks.SeqNoMax),
			kv.Int64("written_offset_min", acks.WrittenOffsetMin),
			kv.Int64("written_offset_max", acks.WrittenOffsetMax),
			kv.Int("written_offset_count", acks.WrittenCount),
			kv.Int("written_in_tx_count", acks.WrittenInTxCount),
			kv.Int("skip_count", acks.SkipCount),
			kv.Version(),
		)
	}

	t.OnWriterSentGRPCMessage = func(info trace.TopicWriterSentGRPCMessageInfo) {
		if d.Details()&trace.TopicWriterStreamGrpcMessageEvents == 0 {
			return
		}

		ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "grpc")
		l.Log(
			ctx, "topic writer sent grpc message (message body and metadata are removed)",
			kv.String("topic_stream_internal_id", info.TopicStreamInternalID),
			kv.String("session_id", info.SessionID),
			kv.Int("message_number", info.MessageNumber),
			kv.Stringer("message", lazyProtoStringifer{info.Message}),
			kv.Error(info.Error),
			kv.Version(),
		)
	}
	t.OnWriterReceiveGRPCMessage = func(info trace.TopicWriterReceiveGRPCMessageInfo) {
		if d.Details()&trace.TopicWriterStreamGrpcMessageEvents == 0 {
			return
		}

		ctx := with(*info.Context, TRACE, "ydb", "topic", "writer", "grpc")
		l.Log(
			ctx, "topic writer received grpc message (message body and metadata are removed)",
			kv.String("topic_stream_internal_id", info.TopicStreamInternalID),
			kv.String("session_id", info.SessionID),
			kv.Int("message_number", info.MessageNumber),
			kv.Stringer("message", lazyProtoStringifer{info.Message}),
			kv.Error(info.Error),
			kv.Version(),
		)
	}
	t.OnWriterReadUnknownGrpcMessage = func(info trace.TopicOnWriterReadUnknownGrpcMessageInfo) {
		if d.Details()&trace.TopicWriterStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, DEBUG, "ydb", "topic", "writer", "read", "unknown", "grpc", "message")
		l.Log(ctx, "topic writer receive unknown grpc message from server",
			kv.Error(info.Error),
			kv.String("writer_instance_id", info.WriterInstanceID),
			kv.String("session_id", info.SessionID),
		)
	}

	///
	/// Topic Listener
	///
	t.OnListenerStart = func(info trace.TopicListenerStartInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, INFO, "ydb", "topic", "listener", "start")
		l.Log(ctx, "topic listener starting",
			kv.String("listener_id", info.ListenerID),
			kv.String("consumer", info.Consumer),
			kv.Error(info.Error),
		)
	}

	t.OnListenerInit = func(info trace.TopicListenerInitStartInfo) func(doneInfo trace.TopicListenerInitDoneInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "init")
		start := time.Now()
		l.Log(ctx, "topic listener init starting...",
			kv.String("listener_id", info.ListenerID),
			kv.String("consumer", info.Consumer),
			kv.Strings("topic_selectors", info.TopicSelectors),
		)

		return func(doneInfo trace.TopicListenerInitDoneInfo) {
			fields := []Field{
				kv.String("listener_id", info.ListenerID),
				kv.String("consumer", info.Consumer),
				kv.Strings("topic_selectors", info.TopicSelectors),
				kv.Latency(start),
			}
			if doneInfo.SessionID != "" {
				fields = append(fields, kv.String("session_id", doneInfo.SessionID))
			}
			if doneInfo.Error == nil {
				l.Log(WithLevel(ctx, INFO), "topic listener init done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic listener init failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}

	t.OnListenerReceiveMessage = func(info trace.TopicListenerReceiveMessageInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "receive", "message")
		fields := []Field{
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.String("message_type", info.MessageType),
			kv.Int("bytes_size", info.BytesSize),
		}
		if info.Error == nil {
			l.Log(ctx, "topic listener received message", fields...)
		} else {
			l.Log(WithLevel(ctx, WARN), "topic listener receive message failed",
				append(fields,
					kv.Error(info.Error),
					kv.Version(),
				)...,
			)
		}
	}

	t.OnListenerRouteMessage = func(info trace.TopicListenerRouteMessageInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "route", "message")
		fields := []Field{
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.String("message_type", info.MessageType),
			kv.Bool("worker_found", info.WorkerFound),
		}
		if info.PartitionSessionID != nil {
			fields = append(fields, kv.Int64("partition_session_id", *info.PartitionSessionID))
		}
		if info.Error == nil {
			l.Log(ctx, "topic listener routed message", fields...)
		} else {
			l.Log(WithLevel(ctx, ERROR), "topic listener route message failed",
				append(fields,
					kv.Error(info.Error),
					kv.Version(),
				)...,
			)
		}
	}

	t.OnListenerSplitMessage = func(info trace.TopicListenerSplitMessageInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "split", "message")
		fields := []Field{
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.String("message_type", info.MessageType),
			kv.Int("total_batches", info.TotalBatches),
			kv.Int("total_partitions", info.TotalPartitions),
			kv.Int("split_batches", info.SplitBatches),
			kv.Int("routed_batches", info.RoutedBatches),
		}
		if info.Error == nil {
			l.Log(ctx, "topic listener split message", fields...)
		} else {
			l.Log(WithLevel(ctx, ERROR), "topic listener split message failed",
				append(fields,
					kv.Error(info.Error),
					kv.Version(),
				)...,
			)
		}
	}

	t.OnListenerError = func(info trace.TopicListenerErrorInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, ERROR, "ydb", "topic", "listener", "error")
		l.Log(ctx, "topic listener error",
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.Error(info.Error),
			kv.Version(),
		)
	}

	t.OnListenerClose = func(info trace.TopicListenerCloseStartInfo) func(doneInfo trace.TopicListenerCloseDoneInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "close")
		start := time.Now()
		l.Log(ctx, "topic listener close starting...",
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.NamedError("reason", info.Reason),
		)

		return func(doneInfo trace.TopicListenerCloseDoneInfo) {
			fields := []Field{
				kv.String("listener_id", info.ListenerID),
				kv.String("session_id", info.SessionID),
				kv.NamedError("reason", info.Reason),
				kv.Int("workers_closed", doneInfo.WorkersClosed),
				kv.Latency(start),
			}
			if doneInfo.Error == nil {
				l.Log(WithLevel(ctx, INFO), "topic listener close done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic listener close failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}

	t.OnListenerSendDataRequest = func(info trace.TopicListenerSendDataRequestInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return
		}
		fields := []Field{
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.String("message_type", info.MessageType),
		}
		if info.Error == nil {
			ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "send", "data", "request")
			l.Log(ctx, "topic listener send data request", fields...)
		} else {
			ctx := with(*info.Context, WARN, "ydb", "topic", "listener", "send", "data", "request")
			l.Log(ctx, "topic listener send data request failed",
				append(fields,
					kv.Error(info.Error),
					kv.Version(),
				)...,
			)
		}
	}

	t.OnListenerUnknownMessage = func(info trace.TopicListenerUnknownMessageInfo) {
		if d.Details()&trace.TopicListenerStreamEvents == 0 {
			return
		}
		ctx := with(*info.Context, DEBUG, "ydb", "topic", "listener", "unknown", "message")
		l.Log(ctx, "topic listener received unknown message",
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.String("message_type", info.MessageType),
			kv.Error(info.Error),
			kv.Version(),
		)
	}

	///
	/// Topic Partition Worker
	///
	t.OnPartitionWorkerStart = func(info trace.TopicPartitionWorkerStartInfo) {
		if d.Details()&trace.TopicListenerWorkerEvents == 0 {
			return
		}
		ctx := with(*info.Context, INFO, "ydb", "topic", "listener", "partition", "worker", "start")
		l.Log(ctx, "topic partition worker starting",
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
			kv.Int64("partition_id", info.PartitionID),
			kv.String("topic", info.Topic),
		)
	}

	t.OnPartitionWorkerProcessMessage = func(
		info trace.TopicPartitionWorkerProcessMessageStartInfo,
	) func(doneInfo trace.TopicPartitionWorkerProcessMessageDoneInfo) {
		if d.Details()&trace.TopicListenerWorkerEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "partition", "worker", "process", "message")
		start := time.Now()
		l.Log(ctx, "topic partition worker process message starting...",
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
			kv.Int64("partition_id", info.PartitionID),
			kv.String("topic", info.Topic),
			kv.String("message_type", info.MessageType),
			kv.Int("messages_count", info.MessagesCount),
		)

		return func(doneInfo trace.TopicPartitionWorkerProcessMessageDoneInfo) {
			fields := []Field{
				kv.String("listener_id", info.ListenerID),
				kv.String("session_id", info.SessionID),
				kv.Int64("partition_session_id", info.PartitionSessionID),
				kv.Int64("partition_id", info.PartitionID),
				kv.String("topic", info.Topic),
				kv.String("message_type", info.MessageType),
				kv.Int("messages_count", info.MessagesCount),
				kv.Int("processed_messages", doneInfo.ProcessedMessages),
				kv.Latency(start),
			}
			if doneInfo.Error == nil {
				l.Log(ctx, "topic partition worker process message done", fields...)
			} else {
				l.Log(WithLevel(ctx, ERROR), "topic partition worker process message failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}

	t.OnPartitionWorkerHandlerCall = func(
		info trace.TopicPartitionWorkerHandlerCallStartInfo,
	) func(doneInfo trace.TopicPartitionWorkerHandlerCallDoneInfo) {
		if d.Details()&trace.TopicListenerWorkerEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "partition", "worker", "handler", "call")
		start := time.Now()
		l.Log(ctx, "topic partition worker handler call starting...",
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
			kv.Int64("partition_id", info.PartitionID),
			kv.String("topic", info.Topic),
			kv.String("handler_type", info.HandlerType),
			kv.Int("messages_count", info.MessagesCount),
		)

		return func(doneInfo trace.TopicPartitionWorkerHandlerCallDoneInfo) {
			fields := []Field{
				kv.String("listener_id", info.ListenerID),
				kv.String("session_id", info.SessionID),
				kv.Int64("partition_session_id", info.PartitionSessionID),
				kv.Int64("partition_id", info.PartitionID),
				kv.String("topic", info.Topic),
				kv.String("handler_type", info.HandlerType),
				kv.Int("messages_count", info.MessagesCount),
				kv.Latency(start),
			}
			if doneInfo.Error == nil {
				l.Log(ctx, "topic partition worker handler call done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic partition worker handler call failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}

	t.OnPartitionWorkerStop = func(
		info trace.TopicPartitionWorkerStopStartInfo,
	) func(doneInfo trace.TopicPartitionWorkerStopDoneInfo) {
		if d.Details()&trace.TopicListenerWorkerEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "topic", "listener", "partition", "worker", "stop")
		start := time.Now()
		l.Log(ctx, "topic partition worker stop starting...",
			kv.String("listener_id", info.ListenerID),
			kv.String("session_id", info.SessionID),
			kv.Int64("partition_session_id", info.PartitionSessionID),
			kv.Int64("partition_id", info.PartitionID),
			kv.String("topic", info.Topic),
			kv.NamedError("reason", info.Reason),
		)

		return func(doneInfo trace.TopicPartitionWorkerStopDoneInfo) {
			fields := []Field{
				kv.String("listener_id", info.ListenerID),
				kv.String("session_id", info.SessionID),
				kv.Int64("partition_session_id", info.PartitionSessionID),
				kv.Int64("partition_id", info.PartitionID),
				kv.String("topic", info.Topic),
				kv.NamedError("reason", info.Reason),
				kv.Latency(start),
			}
			if doneInfo.Error == nil {
				l.Log(WithLevel(ctx, INFO), "topic partition worker stop done", fields...)
			} else {
				l.Log(WithLevel(ctx, WARN), "topic partition worker stop failed",
					append(fields,
						kv.Error(doneInfo.Error),
						kv.Version(),
					)...,
				)
			}
		}
	}

	return t
}

type lazyProtoStringifer struct {
	message proto.Message
}

func (s lazyProtoStringifer) String() string {
	// cut message data
	if writeRequest, ok := s.message.(*Ydb_Topic.StreamWriteMessage_FromClient); ok {
		if data := writeRequest.GetWriteRequest(); data != nil {
			type messDataType struct {
				Data     []byte
				Metadata []*Ydb_Topic.MetadataItem
			}
			storage := make([]messDataType, len(data.GetMessages()))
			for i := range data.GetMessages() {
				storage[i].Data = data.GetMessages()[i].GetData()
				data.Messages[i] = nil

				storage[i].Metadata = data.GetMessages()[i].GetMetadataItems()
				data.Messages[i].MetadataItems = nil
			}

			defer func() {
				for i := range data.GetMessages() {
					data.Messages[i].Data = storage[i].Data
					data.Messages[i].MetadataItems = storage[i].Metadata
				}
			}()
		}
	}

	res := protojson.MarshalOptions{AllowPartial: true}.Format(s.message)

	return res
}
