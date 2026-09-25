package config

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Config struct {
	config.Common

	trace *trace.Scripting
}

// Trace defines trace over scripting client calls
func (c Config) Trace() *trace.Scripting {
	return c.trace
}

type Option func(c *Config)

// WithTrace appends scripting trace to early added traces
func WithTrace(trace trace.Scripting, opts ...trace.ScriptingComposeOption) Option {
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
		trace: &trace.Scripting{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}

	return c
}
