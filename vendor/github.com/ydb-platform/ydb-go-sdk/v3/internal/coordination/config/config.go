package config

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Config is an configuration of coordination client
type Config struct {
	config.Common

	trace *trace.Coordination
}

// Trace returns trace over coordination client calls
func (c Config) Trace() *trace.Coordination {
	return c.trace
}

type Option func(c *Config)

// WithTrace appends coordination trace to early defined traces
func WithTrace(trace *trace.Coordination, opts ...trace.CoordinationComposeOption) Option {
	return func(c *Config) {
		c.trace = c.trace.Compose(trace, opts...)
	}
}

// With applies common configuration params
func With(config config.Common) Option {
	return func(c *Config) {
		c.Common = config
	}
}

func New(opts ...Option) Config {
	c := Config{
		trace: &trace.Coordination{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}

	return c
}
