package topicwriterinternal

import (
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicwriter"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type PublicWriterOption func(cfg *WriterReconnectorConfig)

func WithAddEncoder(codec rawtopiccommon.Codec, encoderFunc PublicCreateEncoderFunc) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		if cfg.AdditionalEncoders == nil {
			cfg.AdditionalEncoders = map[rawtopiccommon.Codec]PublicCreateEncoderFunc{}
		}
		cfg.AdditionalEncoders[codec] = encoderFunc
	}
}

func WithAutoSetSeqNo(val bool) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.AutoSetSeqNo = val
	}
}

func WithAutoCodec() PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.forceCodec = rawtopiccommon.CodecUNSPECIFIED
	}
}

func WithCompressorCount(num int) PublicWriterOption {
	if num <= 0 {
		panic("ydb: compressor count must be > 0")
	}

	return func(cfg *WriterReconnectorConfig) {
		cfg.compressorCount = num
	}
}

func WithMaxGrpcMessageBytes(num int) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.maxBytesPerMessage = num
	}
}

func WithTokenUpdateInterval(interval time.Duration) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.credUpdateInterval = interval
	}
}

// WithCommonConfig
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithCommonConfig(common config.Common) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.Common = common
	}
}

// WithCredentials for internal usage only
// no proxy to public interface
func WithCredentials(cred credentials.Credentials) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		if cred == nil {
			cred = credentials.NewAnonymousCredentials()
		}
		cfg.cred = cred
	}
}

// WithRawClient for internal usage only
// no proxy to public interface
func WithRawClient(rawClient *rawtopic.Client) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.rawTopicClient = rawClient
	}
}

func WithCodec(codec rawtopiccommon.Codec) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.forceCodec = codec
	}
}

func WithConnectFunc(connect ConnectFunc) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.Connect = connect
	}
}

func WithConnectTimeout(timeout time.Duration) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.connectTimeout = timeout
	}
}

func WithAutosetCreatedTime(enable bool) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.AutoSetCreatedTime = enable
	}
}

func WithMaxQueueLen(num int) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.MaxQueueLen = num
	}
}

func WithPartitioning(partitioning PublicFuturePartitioning) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.defaultPartitioning = partitioning.ToRaw()
	}
}

func WithProducerID(producerID string) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.producerID = producerID
		oldPartitioningType := cfg.defaultPartitioning.Type
		if oldPartitioningType == rawtopicwriter.PartitioningUndefined ||
			oldPartitioningType == rawtopicwriter.PartitioningMessageGroupID {
			WithPartitioning(NewPartitioningWithMessageGroupID(producerID))(cfg)
		}
	}
}

func WithSessionMeta(meta map[string]string) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		if len(meta) == 0 {
			cfg.writerMeta = nil
		} else {
			cfg.writerMeta = make(map[string]string, len(meta))
			for k, v := range meta {
				cfg.writerMeta[k] = v
			}
		}
	}
}

func WithStartTimeout(timeout time.Duration) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.RetrySettings.StartTimeout = timeout
	}
}

func WithWaitAckOnWrite(val bool) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.WaitServerAck = val
	}
}

func WithTrace(tracer *trace.Topic) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.Tracer = cfg.Tracer.Compose(tracer)
	}
}

func WithTopic(topic string) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.topic = topic
	}
}

// WithClock is private option for tests
func WithClock(clock clockwork.Clock) PublicWriterOption {
	return func(cfg *WriterReconnectorConfig) {
		cfg.clock = clock
	}
}
