package topicoptions

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreaderinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// ReadSelector set rules for reader: set of topic, partitions, start time filted, etc.
type ReadSelector = topicreadercommon.PublicReadSelector

// ReadSelectors slice of rules for topic reader
type ReadSelectors []ReadSelector

// ReadTopic create simple selector for read topics, if no need special settings.
func ReadTopic(path string) ReadSelectors {
	return ReadSelectors{{Path: path}}
}

// ReaderOption options for topic reader
type ReaderOption = topicreaderinternal.PublicReaderOption

// WithReaderOperationTimeout
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithReaderOperationTimeout(timeout time.Duration) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		config.SetOperationTimeout(&cfg.Common, timeout)
	}
}

// WithReaderStartTimeout mean timeout for connect to reader stream and work some time without errors
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithReaderStartTimeout(timeout time.Duration) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.RetrySettings.StartTimeout = timeout
	}
}

// WithReaderCheckRetryErrorFunction can override default error retry policy
// use CheckErrorRetryDecisionDefault for use default behavior for the error
// callback func must be fast and deterministic: always result same result for same error - it can be called
// few times for every error
func WithReaderCheckRetryErrorFunction(callback CheckErrorRetryFunction) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.RetrySettings.CheckError = callback
	}
}

// WithReaderOperationCancelAfter
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithReaderOperationCancelAfter(cancelAfter time.Duration) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		config.SetOperationCancelAfter(&cfg.Common, cancelAfter)
	}
}

// WithCommonConfig
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithCommonConfig(common config.Common) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.Common = common
	}
}

// WithCommitTimeLagTrigger
//
// Deprecated: was experimental and not actual now.
// Use WithReaderCommitTimeLagTrigger instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithCommitTimeLagTrigger(lag time.Duration) ReaderOption {
	return WithReaderCommitTimeLagTrigger(lag)
}

// WithReaderCommitTimeLagTrigger set time lag from first commit message before send commit to server
// for accumulate many similar-time commits to one server request
// 0 mean no additional lag and send commit soon as possible
// Default value: 1 second
func WithReaderCommitTimeLagTrigger(lag time.Duration) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.CommitterBatchTimeLag = lag
	}
}

// WithCommitCountTrigger
//
// Deprecated: was experimental and not actual now.
// Use WithReaderCommitCountTrigger instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithCommitCountTrigger(count int) ReaderOption {
	return WithReaderCommitCountTrigger(count)
}

// WithReaderCommitCountTrigger set count trigger for send batch to server
// if count > 0 and sdk count of buffered commits >= count - send commit request to server
// 0 mean no count limit and use timer lag trigger only
func WithReaderCommitCountTrigger(count int) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.CommitterBatchCounterTrigger = count
	}
}

// WithBatchReadMinCount
// prefer min count messages in batch
// sometimes batch can contain fewer messages, for example if local buffer is full and SDK can't receive more messages
//
// Deprecated: was experimental and not actual now.
// The option will be removed for simplify code internals.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithBatchReadMinCount(count int) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.DefaultBatchConfig.MinCount = count
	}
}

// WithBatchReadMaxCount
//
// Deprecated: was experimental and not actual now.
// Use WithReaderBatchMaxCount instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithBatchReadMaxCount(count int) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.DefaultBatchConfig.MaxCount = count
	}
}

// WithReaderBatchMaxCount set max messages count, returned by topic.TopicReader.ReadBatch method
func WithReaderBatchMaxCount(count int) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.DefaultBatchConfig.MaxCount = count
	}
}

// WithMessagesBufferSize
//
// Deprecated: was experimental and not actual now.
// Use WithReaderBufferSizeBytes instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithMessagesBufferSize(size int) ReaderOption {
	return WithReaderBufferSizeBytes(size)
}

// WithReaderBufferSizeBytes set size of internal buffer for read ahead messages.
func WithReaderBufferSizeBytes(size int) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.BufferSizeProtoBytes = size
	}
}

// CreateDecoderFunc interface for fabric of message decoders
type CreateDecoderFunc = topicreadercommon.PublicCreateDecoderFunc

// WithAddDecoder add decoder for a codec.
// It allows to set decoders fabric for custom codec and replace internal decoders.
func WithAddDecoder(codec topictypes.Codec, decoderCreate CreateDecoderFunc) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.Decoders.AddDecoder(rawtopiccommon.Codec(codec), decoderCreate)
	}
}

// CommitMode variants of commit mode of the reader
type CommitMode = topicreadercommon.PublicCommitMode

const (
	// CommitModeAsync - commit return true if commit success add to internal send buffer (but not sent to server)
	// now it is grpc buffer, in feature it may be internal sdk buffer
	CommitModeAsync = topicreadercommon.CommitModeAsync // default

	// CommitModeNone - reader will not be commit operation
	CommitModeNone = topicreadercommon.CommitModeNone

	// CommitModeSync - commit return true when sdk receive ack of commit from server
	// The mode needs strong ordering client code for prevent deadlock.
	// Example:
	// Good:
	// CommitOffset(1)
	// CommitOffset(2)
	//
	// Error:
	// CommitOffset(2) - server will wait commit offset 1 before send ack about offset 1 and 2 committed.
	// CommitOffset(1)
	// SDK will detect the problem and return error instead of deadlock.
	CommitModeSync = topicreadercommon.CommitModeSync
)

// WithCommitMode
//
// Deprecated: was experimental and not actual now.
// Use WithReaderCommitMode instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithCommitMode(mode CommitMode) ReaderOption {
	return WithReaderCommitMode(mode)
}

// WithReaderCommitMode set commit mode to the reader
func WithReaderCommitMode(mode CommitMode) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.CommitMode = mode
	}
}

type (
	// GetPartitionStartOffsetFunc callback function for optional handle start partition event and manage read progress
	// at own side. It can call multiply times in parallel.
	GetPartitionStartOffsetFunc = topicreaderinternal.PublicGetPartitionStartOffsetFunc

	// GetPartitionStartOffsetRequest info about the partition
	GetPartitionStartOffsetRequest = topicreaderinternal.PublicGetPartitionStartOffsetRequest

	// GetPartitionStartOffsetResponse optional set offset for start reade messages for the partition
	GetPartitionStartOffsetResponse = topicreaderinternal.PublicGetPartitionStartOffsetResponse
)

// WithGetPartitionStartOffset
//
// Deprecated: was experimental and not actual now.
// Use WithReaderGetPartitionStartOffset instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithGetPartitionStartOffset(f GetPartitionStartOffsetFunc) ReaderOption {
	return WithReaderGetPartitionStartOffset(f)
}

// WithReaderGetPartitionStartOffset set optional handler for own manage progress of read partitions
// instead of/additional to commit messages
func WithReaderGetPartitionStartOffset(f GetPartitionStartOffsetFunc) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.GetPartitionStartOffsetCallback = f
	}
}

// WithReaderTrace set tracer for the topic reader
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithReaderTrace(t trace.Topic) ReaderOption { //nolint:gocritic
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.Trace = cfg.Trace.Compose(&t)
	}
}

// WithReaderUpdateTokenInterval set custom interval for send update token message to the server
func WithReaderUpdateTokenInterval(interval time.Duration) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.CredUpdateInterval = interval
	}
}

// WithReaderWithoutConsumer allow read topic without consumer.
// Read without consumer is special read mode on a server. In the mode every reader without consumer receive all
// messages from a topic and can't commit them.
// The mode work good if every reader process need all messages (for example for cache invalidation) and no need
// scale process messages by readers count.
//
// saveStateOnReconnection
// - if true: simulate one unbroken stream without duplicate messages (unimplemented)
// - if false: need store progress on client side for prevent re-read messages on internal reconnections to the server.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
// https://github.com/ydb-platform/ydb-go-sdk/issues/905
func WithReaderWithoutConsumer(saveStateOnReconnection bool) ReaderOption {
	if saveStateOnReconnection {
		panic("ydb: saveStateOnReconnection mode doesn't implemented yet")
	}

	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.ReadWithoutConsumer = true
		cfg.CommitMode = CommitModeNone
	}
}

// WithReaderSupportSplitMergePartitions set support of split and merge partitions on client side.
// Default is true, set false for disable the support.
func WithReaderSupportSplitMergePartitions(enableSupport bool) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.EnableSplitMergeSupport = enableSupport
	}
}

// WithReaderLogContext allows providing a context.Context instance which will be used
// in log/topic events.
func WithReaderLogContext(ctx context.Context) ReaderOption {
	return func(cfg *topicreaderinternal.ReaderConfig) {
		cfg.BaseContext = ctx
	}
}
