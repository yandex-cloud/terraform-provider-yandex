package config

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Config is a configuration of scheme client
type Config struct {
	config.Common

	databaseName string
	trace        *trace.Scheme
}

// Trace returns trace over scheme client calls
func (c *Config) Trace() *trace.Scheme {
	return c.trace
}

// Database returns database name
func (c *Config) Database() string {
	return c.databaseName
}

type Option func(c *Config)

// WithTrace appends scheme trace to early defined traces
func WithTrace(trace trace.Scheme, opts ...trace.SchemeComposeOption) Option {
	return func(c *Config) {
		c.trace = c.trace.Compose(&trace, opts...)
	}
}

// WithDatabaseName applies database name
func WithDatabaseName(dbName string) Option {
	return func(c *Config) {
		c.databaseName = dbName
	}
}

// With applies common configuration params
func With(config config.Common) Option {
	return func(c *Config) {
		c.Common = config
	}
}

func New(opts ...Option) *Config {
	c := &Config{
		trace: &trace.Scheme{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	return c
}
