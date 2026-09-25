package config

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"

	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"
	balancerConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/balancer/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/meta"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Config contains driver configuration.
type Config struct {
	config.Common

	trace              *trace.Driver
	dialTimeout        time.Duration
	connectionTTL      time.Duration
	balancerConfig     *balancerConfig.Config
	secure             bool
	endpoint           string
	database           string
	metaOptions        []meta.Option
	grpcOptions        []grpc.DialOption
	grpcMaxMessageSize int
	credentials        credentials.Credentials
	tlsConfig          *tls.Config
	meta               *meta.Meta

	excludeGRPCCodesForPessimization []grpcCodes.Code
	disableOptimisticUnban           bool
}

func (c *Config) Credentials() credentials.Credentials {
	return c.credentials
}

// ExcludeGRPCCodesForPessimization defines grpc codes for exclude its from pessimization trigger
func (c *Config) ExcludeGRPCCodesForPessimization() []grpcCodes.Code {
	return c.excludeGRPCCodesForPessimization
}

func (c *Config) DisableOptimisticUnban() bool {
	return c.disableOptimisticUnban
}

// GrpcDialOptions reports about used grpc dialing options
func (c *Config) GrpcDialOptions() []grpc.DialOption {
	return append(
		defaultGrpcOptions(c.secure, c.tlsConfig),
		c.grpcOptions...,
	)
}

// GrpcMaxMessageSize return client settings for max grpc message size
func (c *Config) GrpcMaxMessageSize() int {
	return c.grpcMaxMessageSize
}

// Meta reports meta information about database connection
func (c *Config) Meta() *meta.Meta {
	return c.meta
}

// ConnectionTTL defines interval for parking grpc connections.
//
// If ConnectionTTL is zero - connections are not park.
func (c *Config) ConnectionTTL() time.Duration {
	return c.connectionTTL
}

// Secure is a flag for secure connection
func (c *Config) Secure() bool {
	return c.secure
}

// Endpoint is a required starting endpoint for connect
func (c *Config) Endpoint() string {
	return c.endpoint
}

// TLSConfig reports about TLS configuration
func (c *Config) TLSConfig() *tls.Config {
	return c.tlsConfig
}

// DialTimeout is the maximum amount of time a dial will wait for a connect to
// complete.
//
// If DialTimeout is zero then no timeout is used.
func (c *Config) DialTimeout() time.Duration {
	return c.dialTimeout
}

// Database is a required database name.
func (c *Config) Database() string {
	return c.database
}

// Trace contains driver tracing options.
func (c *Config) Trace() *trace.Driver {
	return c.trace
}

// Balancer is an optional configuration related to selected balancer.
// That is, some balancing methods allow to be configured.
func (c *Config) Balancer() *balancerConfig.Config {
	return c.balancerConfig
}

type Option func(c *Config)

// WithInternalDNSResolver
//
// Deprecated: always used internal dns-resolver.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithInternalDNSResolver() Option {
	return func(c *Config) {}
}

func WithEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.endpoint = endpoint
	}
}

// WithSecure changes secure connection flag.
//
// Warning: if secure is false - TLS config options has no effect.
func WithSecure(secure bool) Option {
	return func(c *Config) {
		c.secure = secure
	}
}

func WithDatabase(database string) Option {
	return func(c *Config) {
		c.database = database
	}
}

// WithCertificate appends certificate to TLS config root certificates
func WithCertificate(certificate *x509.Certificate) Option {
	return func(c *Config) {
		c.tlsConfig.RootCAs.AddCert(certificate)
	}
}

// WithTLSConfig replaces older TLS config
//
// Warning: all early changes of TLS config will be lost
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *Config) {
		c.tlsConfig = tlsConfig
	}
}

func WithTrace(t trace.Driver, opts ...trace.DriverComposeOption) Option { //nolint:gocritic
	return func(c *Config) {
		c.trace = c.trace.Compose(&t, opts...)
	}
}

// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithRetryBudget(b budget.Budget) Option {
	return func(c *Config) {
		config.SetRetryBudget(&c.Common, b)
	}
}

func WithTraceRetry(t *trace.Retry, opts ...trace.RetryComposeOption) Option {
	return func(c *Config) {
		config.SetTraceRetry(&c.Common, t, opts...)
	}
}

// WithApplicationName add provided application name to all api requests
func WithApplicationName(applicationName string) Option {
	return func(c *Config) {
		c.metaOptions = append(c.metaOptions, meta.WithApplicationNameOption(applicationName))
	}
}

// WithUserAgent add provided user agent to all api requests
//
// Deprecated: use WithApplicationName instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithUserAgent(userAgent string) Option {
	return func(c *Config) {
		c.metaOptions = append(c.metaOptions, meta.WithApplicationNameOption(userAgent))
	}
}

func WithConnectionTTL(ttl time.Duration) Option {
	return func(c *Config) {
		c.connectionTTL = ttl
	}
}

func WithCredentials(credentials credentials.Credentials) Option {
	return func(c *Config) {
		c.credentials = credentials
	}
}

// WithOperationTimeout defines the maximum amount of time a YDB server will process
// an operation. After timeout exceeds YDB will try to cancel operation and
// regardless of the cancellation appropriate error will be returned to
// the client.
//
// If OperationTimeout is zero then no timeout is used.
func WithOperationTimeout(operationTimeout time.Duration) Option {
	return func(c *Config) {
		config.SetOperationTimeout(&c.Common, operationTimeout)
	}
}

// WithOperationCancelAfter sets the maximum amount of time a YDB server will process an
// operation. After timeout exceeds YDB will try to cancel operation and if
// it succeeds appropriate error will be returned to the client; otherwise
// processing will be continued.
//
// If OperationCancelAfter is zero then no timeout is used.
func WithOperationCancelAfter(operationCancelAfter time.Duration) Option {
	return func(c *Config) {
		config.SetOperationCancelAfter(&c.Common, operationCancelAfter)
	}
}

// WithNoAutoRetry disable auto-retry calls from YDB sub-clients
func WithNoAutoRetry() Option {
	return func(c *Config) {
		config.SetAutoRetry(&c.Common, false)
	}
}

// WithPanicCallback applies panic callback to config
func WithPanicCallback(panicCallback func(e interface{})) Option {
	return func(c *Config) {
		config.SetPanicCallback(&c.Common, panicCallback)
	}
}

func WithDialTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.dialTimeout = timeout
	}
}

func WithGrpcMaxMessageSize(sizeBytes int) Option {
	return func(c *Config) {
		c.grpcMaxMessageSize = sizeBytes
		c.grpcOptions = append(c.grpcOptions,
			// limit size of outgoing and incoming packages
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(c.grpcMaxMessageSize),
				grpc.MaxCallSendMsgSize(c.grpcMaxMessageSize),
			),
		)
	}
}

func WithBalancer(balancer *balancerConfig.Config) Option {
	return func(c *Config) {
		c.balancerConfig = balancer
	}
}

func WithRequestsType(requestsType string) Option {
	return func(c *Config) {
		c.metaOptions = append(c.metaOptions, meta.WithRequestTypeOption(requestsType))
	}
}

// WithMinTLSVersion applies minimum TLS version that is acceptable.
func WithMinTLSVersion(minVersion uint16) Option {
	return func(c *Config) {
		c.tlsConfig.MinVersion = minVersion
	}
}

// WithTLSSInsecureSkipVerify applies InsecureSkipVerify flag to TLS config
func WithTLSSInsecureSkipVerify() Option {
	return func(c *Config) {
		c.tlsConfig.InsecureSkipVerify = true
	}
}

// WithGrpcOptions appends custom grpc dial options to defaults
func WithGrpcOptions(option ...grpc.DialOption) Option {
	return func(c *Config) {
		c.grpcOptions = append(c.grpcOptions, option...)
	}
}

func ExcludeGRPCCodesForPessimization(codes ...grpcCodes.Code) Option {
	return func(c *Config) {
		c.excludeGRPCCodesForPessimization = append(
			c.excludeGRPCCodesForPessimization,
			codes...,
		)
	}
}

// WithDisableOptimisticUnban disables optimistic unban of nodes after a successful call
// immediately following pessimization.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithDisableOptimisticUnban() Option {
	return func(c *Config) {
		c.disableOptimisticUnban = true
	}
}

func New(opts ...Option) *Config {
	c := defaultConfig()

	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	c.meta = meta.New(c.database, c.credentials, c.trace, c.metaOptions...)

	return c
}

// With makes copy of current Config with specified options
func (c *Config) With(opts ...Option) *Config {
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	c.meta = meta.New(
		c.database,
		c.credentials,
		c.trace,
		c.metaOptions...,
	)

	return c
}
