package config

import (
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/meta"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

const (
	DefaultInterval = time.Minute
)

type Config struct {
	config.Common

	endpoint       string
	database       string
	secure         bool
	meta           *meta.Meta
	addressMutator func(address string) string
	clock          clockwork.Clock

	interval time.Duration
	trace    *trace.Discovery
}

func New(opts ...Option) *Config {
	c := &Config{
		interval: DefaultInterval,
		trace:    &trace.Discovery{},
		addressMutator: func(address string) string {
			return address
		},
		clock: clockwork.NewRealClock(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	return c
}

func (c *Config) MutateAddress(fqdn string) string {
	return c.addressMutator(fqdn)
}

func (c *Config) Meta() *meta.Meta {
	return c.meta
}

func (c *Config) Clock() clockwork.Clock {
	return c.clock
}

func (c *Config) Interval() time.Duration {
	return c.interval
}

func (c *Config) Endpoint() string {
	return c.endpoint
}

func (c *Config) Database() string {
	return c.database
}

func (c *Config) Secure() bool {
	return c.secure
}

func (c *Config) Trace() *trace.Discovery {
	return c.trace
}

type Option func(c *Config)

// With applies common configuration params
func With(config config.Common) Option {
	return func(c *Config) {
		c.Common = config
	}
}

// WithEndpoint set a required starting endpoint for connect
func WithEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.endpoint = endpoint
	}
}

// WithDatabase set a required database name.
func WithDatabase(database string) Option {
	return func(c *Config) {
		c.database = database
	}
}

func WithClock(clock clockwork.Clock) Option {
	return func(c *Config) {
		c.clock = clock
	}
}

func WithAddressMutator(addressMutator func(address string) string) Option {
	return func(c *Config) {
		c.addressMutator = addressMutator
	}
}

// WithSecure set flag for secure connection
func WithSecure(ssl bool) Option {
	return func(c *Config) {
		c.secure = ssl
	}
}

// WithMeta is not for user.
//
// This option add meta information about database connection
func WithMeta(meta *meta.Meta) Option {
	return func(c *Config) {
		c.meta = meta
	}
}

// WithTrace configures discovery client calls tracing
func WithTrace(trace trace.Discovery, opts ...trace.DiscoveryComposeOption) Option {
	return func(c *Config) {
		c.trace = c.trace.Compose(&trace, opts...)
	}
}

// WithInterval set the frequency of background tasks of ydb endpoints discovery.
//
// If Interval is zero then the DefaultInterval is used.
//
// If Interval is negative, then no background discovery prepared.
func WithInterval(interval time.Duration) Option {
	return func(c *Config) {
		switch {
		case interval < 0:
			c.interval = 0
		case interval == 0:
			c.interval = DefaultInterval
		default:
			c.interval = interval
		}
	}
}
