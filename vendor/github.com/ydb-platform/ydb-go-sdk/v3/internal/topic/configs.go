package topic

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Config struct {
	config.Common
	Trace              *trace.Topic
	MaxGrpcMessageSize int
}

type Option func(c *Config)

func PublicWithTrace(trace trace.Topic, opts ...trace.TopicComposeOption) Option { //nolint:gocritic
	return func(c *Config) {
		c.Trace = c.Trace.Compose(&trace, opts...)
	}
}

func PublicWithOperationTimeout(operationTimeout time.Duration) Option {
	return func(c *Config) {
		config.SetOperationTimeout(&c.Common, operationTimeout)
	}
}

func PublicWithOperationCancelAfter(operationCancelAfter time.Duration) Option {
	return func(c *Config) {
		config.SetOperationCancelAfter(&c.Common, operationCancelAfter)
	}
}

func WithGrpcMessageSize(sizeBytes int) Option {
	return func(c *Config) {
		c.MaxGrpcMessageSize = sizeBytes
	}
}
