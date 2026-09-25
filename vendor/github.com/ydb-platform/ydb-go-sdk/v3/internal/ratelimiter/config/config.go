package config

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Config is a configuration of ratelimiter client
type Config struct {
	config.Common

	trace *trace.Ratelimiter
}

// Trace returns trace over ratelimiter calls
func (c Config) Trace() *trace.Ratelimiter {
	return c.trace
}

type Option func(c *Config)

// WithTrace appends ratelimiter trace to early defined traces
func WithTrace(trace trace.Ratelimiter, opts ...trace.RatelimiterComposeOption) Option {
	return func(c *Config) {
		c.trace = c.trace.Compose(&trace, opts...)
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
		trace: &trace.Ratelimiter{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}

	return c
}
