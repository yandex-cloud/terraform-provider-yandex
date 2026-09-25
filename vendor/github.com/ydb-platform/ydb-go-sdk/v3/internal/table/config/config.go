package config

import (
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

const (
	DefaultSessionPoolDeleteTimeout        = 500 * time.Millisecond
	DefaultSessionPoolCreateSessionTimeout = 5 * time.Second
	DefaultSessionPoolSizeLimit            = 50
	DefaultSessionPoolIdleThreshold        = 5 * time.Minute

	// Deprecated: table client do not supports background session keep-aliving now.
	// Will be removed after Oct 2024.
	// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
	DefaultKeepAliveMinSize = 10

	// Deprecated: table client do not supports background session keep-aliving now.
	// Will be removed after Oct 2024.
	// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
	DefaultIdleKeepAliveThreshold = 2

	// Deprecated: table client do not supports background session keep-aliving now.
	// Will be removed after Oct 2024.
	// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
	DefaultSessionPoolKeepAliveTimeout = 500 * time.Millisecond
)

func New(opts ...Option) *Config {
	c := defaults()
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	return c
}

type Option func(*Config)

// With applies common configuration params
func With(config config.Common) Option {
	return func(c *Config) {
		c.Common = config
	}
}

// WithSizeLimit defines upper bound of pooled sessions.
// If poolLimit is less than or equal to zero then the
// DefaultSessionPoolSizeLimit variable is used as a limit.
func WithSizeLimit(sizeLimit int) Option {
	return func(c *Config) {
		if sizeLimit > 0 {
			c.poolLimit = sizeLimit
		}
	}
}

// WithSessionPoolSessionUsageLimit set pool session max usage:
// - if argument type is uint64 - WithSessionPoolSessionUsageLimit limits max usage count of pool session
// - if argument type is time.Duration - WithSessionPoolSessionUsageLimit limits max time to live of pool session
func WithSessionPoolSessionUsageLimit[T interface{ uint64 | time.Duration }](limit T) Option {
	return func(c *Config) {
		switch v := any(limit).(type) {
		case uint64:
			c.poolSessionUsageLimit = v
		case time.Duration:
			c.poolSessionUsageTTL = v
		}
	}
}

// WithKeepAliveMinSize defines lower bound for sessions in the pool. If there are more sessions open, then
// the excess idle ones will be closed and removed after IdleKeepAliveThreshold is reached for each of them.
// If keepAliveMinSize is less than zero, then no sessions will be preserved
// If keepAliveMinSize is zero, the DefaultKeepAliveMinSize is used
//
// Deprecated: table client do not supports background session keep-aliving now.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithKeepAliveMinSize(keepAliveMinSize int) Option {
	return func(c *Config) {}
}

// WithIdleKeepAliveThreshold defines number of keepAlive messages to call before the
// session is removed if it is an excess session (see KeepAliveMinSize)
// This means that session will be deleted after the expiration of lifetime = IdleThreshold * IdleKeepAliveThreshold
// If IdleKeepAliveThreshold is less than zero then it will be treated as infinite and no sessions will
// be removed ever.
// If IdleKeepAliveThreshold is equal to zero, it will be set to DefaultIdleKeepAliveThreshold
//
// Deprecated: table client do not supports background session keep-aliving now.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithIdleKeepAliveThreshold(idleKeepAliveThreshold int) Option {
	return func(c *Config) {}
}

// WithIdleThreshold sets maximum duration between any activity within session.
// If this threshold reached, session will be closed.
//
// If idleThreshold is less than zero then there is no idle limit.
// If idleThreshold is zero, then the DefaultSessionPoolIdleThreshold value is used.
func WithIdleThreshold(idleThreshold time.Duration) Option {
	return func(c *Config) {
		if idleThreshold < 0 {
			idleThreshold = 0
		}
		c.idleThreshold = idleThreshold
	}
}

// WithKeepAliveTimeout limits maximum time spent on KeepAlive request
// If keepAliveTimeout is less than or equal to zero then the DefaultSessionPoolKeepAliveTimeout is used.
//
// Deprecated: table client do not supports background session keep-aliving now.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithKeepAliveTimeout(keepAliveTimeout time.Duration) Option {
	return func(c *Config) {}
}

// WithCreateSessionTimeout limits maximum time spent on Create session request
// If createSessionTimeout is less than or equal to zero then no used timeout on create session request
func WithCreateSessionTimeout(createSessionTimeout time.Duration) Option {
	return func(c *Config) {
		if createSessionTimeout > 0 {
			c.createSessionTimeout = createSessionTimeout
		} else {
			c.createSessionTimeout = 0
		}
	}
}

// WithDeleteTimeout limits maximum time spent on Delete request
// If deleteTimeout is less than or equal to zero then the DefaultSessionPoolDeleteTimeout is used.
func WithDeleteTimeout(deleteTimeout time.Duration) Option {
	return func(c *Config) {
		if deleteTimeout > 0 {
			c.deleteTimeout = deleteTimeout
		}
	}
}

// WithTrace appends table trace to early defined traces
func WithTrace(trace *trace.Table, opts ...trace.TableComposeOption) Option {
	return func(c *Config) {
		c.trace = c.trace.Compose(trace, opts...)
	}
}

// WithIgnoreTruncated disables errors on truncated flag
func WithIgnoreTruncated() Option {
	return func(c *Config) {
		c.ignoreTruncated = true
	}
}

// WithMaxRequestMessageSize sets the maximum size of request message in bytes.
func WithMaxRequestMessageSize(maxMessageSize int) Option {
	return func(c *Config) {
		c.maxRequestMessageSize = maxMessageSize
	}
}

// ExecuteDataQueryOverQueryService overrides Execute handle with query service execute with materialized result
func ExecuteDataQueryOverQueryService(b bool) Option {
	return func(c *Config) {
		c.executeDataQueryOverQueryService = b
		if b {
			c.useQuerySession = true
		}
	}
}

// UseQuerySession creates session using query service client
func UseQuerySession(b bool) Option {
	return func(c *Config) {
		c.useQuerySession = b
	}
}

func WithDisableSessionBalancer() Option {
	return func(c *Config) {
		c.SetDisableSessionBalancer()
	}
}

// WithClock replaces default clock
func WithClock(clock clockwork.Clock) Option {
	return func(c *Config) {
		c.clock = clock
	}
}

// Config is a configuration of table client
type Config struct {
	config.Common

	poolLimit             int
	poolSessionUsageLimit uint64
	poolSessionUsageTTL   time.Duration

	createSessionTimeout time.Duration
	deleteTimeout        time.Duration
	idleThreshold        time.Duration

	ignoreTruncated                  bool
	useQuerySession                  bool
	executeDataQueryOverQueryService bool

	maxRequestMessageSize int

	trace *trace.Table

	clock clockwork.Clock
}

// Trace defines trace over table client calls
func (c *Config) Trace() *trace.Table {
	return c.trace
}

// Clock defines clock
func (c *Config) Clock() clockwork.Clock {
	return c.clock
}

// SizeLimit is an upper bound of pooled sessions.
// If SizeLimit is less than or equal to zero then the
// DefaultSessionPoolSizeLimit variable is used as a limit.
func (c *Config) SizeLimit() int {
	return c.poolLimit
}

func (c *Config) SessionUsageLimit() uint64 {
	return c.poolSessionUsageLimit
}

func (c *Config) SessionUsageTTL() time.Duration {
	return c.poolSessionUsageTTL
}

// KeepAliveMinSize is a lower bound for sessions in the pool. If there are more sessions open, then
// the excess idle ones will be closed and removed after IdleKeepAliveThreshold is reached for each of them.
// If KeepAliveMinSize is less than zero, then no sessions will be preserved
// If KeepAliveMinSize is zero, the DefaultKeepAliveMinSize is used
//
// Deprecated: table client do not supports background session keep-aliving now.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func (c *Config) KeepAliveMinSize() int {
	return DefaultKeepAliveMinSize
}

// IgnoreTruncated specifies behavior on truncated flag
func (c *Config) IgnoreTruncated() bool {
	return c.ignoreTruncated
}

// UseQuerySession specifies behavior on create/delete session
func (c *Config) UseQuerySession() bool {
	return c.useQuerySession
}

// ExecuteDataQueryOverQueryService specifies behavior on execute handle
func (c *Config) ExecuteDataQueryOverQueryService() bool {
	return c.executeDataQueryOverQueryService
}

// IdleKeepAliveThreshold is a number of keepAlive messages to call before the
// session is removed if it is an excess session (see KeepAliveMinSize)
// This means that session will be deleted after the expiration of lifetime = IdleThreshold * IdleKeepAliveThreshold
// If IdleKeepAliveThreshold is less than zero then it will be treated as infinite and no sessions will
// be removed ever.
// If IdleKeepAliveThreshold is equal to zero, it will be set to DefaultIdleKeepAliveThreshold
//
// Deprecated: table client do not supports background session keep-aliving now.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func (c *Config) IdleKeepAliveThreshold() int {
	return DefaultIdleKeepAliveThreshold
}

// IdleThreshold is a maximum duration between any activity within session.
// If this threshold reached, idle session will be closed
//
// If IdleThreshold is less than zero then there is no idle limit.
// If IdleThreshold is zero, then the DefaultSessionPoolIdleThreshold value is used.
func (c *Config) IdleThreshold() time.Duration {
	return c.idleThreshold
}

// KeepAliveTimeout limits maximum time spent on KeepAlive request
// If KeepAliveTimeout is less than or equal to zero then the DefaultSessionPoolKeepAliveTimeout is used.
//
// Deprecated: table client do not supports background session keep-aliving now.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func (c *Config) KeepAliveTimeout() time.Duration {
	return DefaultSessionPoolKeepAliveTimeout
}

// CreateSessionTimeout limits maximum time spent on Create session request
func (c *Config) CreateSessionTimeout() time.Duration {
	return c.createSessionTimeout
}

// DeleteTimeout limits maximum time spent on Delete request
//
// If DeleteTimeout is less than or equal to zero then the DefaultSessionPoolDeleteTimeout is used.
func (c *Config) DeleteTimeout() time.Duration {
	return c.deleteTimeout
}

// MaxRequestMessageSize returns the maximum size in bytes for a single request message.
//
// If the value is exceeded, the request will be split into several parts.
func (c *Config) MaxRequestMessageSize() int {
	return c.maxRequestMessageSize
}

func defaults() *Config {
	return &Config{
		poolLimit:            DefaultSessionPoolSizeLimit,
		createSessionTimeout: DefaultSessionPoolCreateSessionTimeout,
		deleteTimeout:        DefaultSessionPoolDeleteTimeout,
		idleThreshold:        DefaultSessionPoolIdleThreshold,
		clock:                clockwork.NewRealClock(),
		trace:                &trace.Table{},
	}
}
