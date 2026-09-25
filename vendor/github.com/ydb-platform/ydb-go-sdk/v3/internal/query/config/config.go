package config

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/pool"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

const (
	DefaultSessionDeleteTimeout = 500 * time.Millisecond
	DefaultSessionCreateTimeout = 500 * time.Millisecond
	DefaultPoolMaxSize          = pool.DefaultLimit
)

type Config struct {
	config.Common

	poolLimit             int
	poolSessionUsageLimit uint64
	poolSessionUsageTTL   time.Duration

	sessionCreateTimeout   time.Duration
	sessionDeleteTimeout   time.Duration
	sessionIddleTimeToLive time.Duration

	allowImplicitSessions bool

	lazyTx bool

	trace *trace.Query
}

func New(opts ...Option) *Config {
	c := defaults()
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	return c
}

func defaults() *Config {
	return &Config{
		poolLimit:            DefaultPoolMaxSize,
		sessionCreateTimeout: DefaultSessionCreateTimeout,
		sessionDeleteTimeout: DefaultSessionDeleteTimeout,
		trace:                &trace.Query{},
	}
}

// Trace defines trace over table client calls
func (c *Config) Trace() *trace.Query {
	return c.trace
}

// PoolLimit is an upper bound of pooled sessions.
// If PoolLimit is less than or equal to zero then the
// DefaultPoolMaxSize variable is used as a pool limit.
func (c *Config) PoolLimit() int {
	return c.poolLimit
}

func (c *Config) AllowImplicitSessions() bool {
	return c.allowImplicitSessions
}

func (c *Config) PoolSessionUsageLimit() uint64 {
	return c.poolSessionUsageLimit
}

func (c *Config) PoolSessionUsageTTL() time.Duration {
	return c.poolSessionUsageTTL
}

// SessionCreateTimeout limits maximum time spent on Create session request
func (c *Config) SessionCreateTimeout() time.Duration {
	return c.sessionCreateTimeout
}

// SessionDeleteTimeout limits maximum time spent on Delete request
//
// If SessionDeleteTimeout is less than or equal to zero then the DefaultSessionDeleteTimeout is used.
func (c *Config) SessionDeleteTimeout() time.Duration {
	return c.sessionDeleteTimeout
}

// SessionIdleTimeToLive limits maximum time to live of idle session
// If idleTimeToLive is less than or equal to zero then sessions will not be closed by idle
func (c *Config) SessionIdleTimeToLive() time.Duration {
	return c.sessionIddleTimeToLive
}

func (c *Config) LazyTx() bool {
	return c.lazyTx
}
