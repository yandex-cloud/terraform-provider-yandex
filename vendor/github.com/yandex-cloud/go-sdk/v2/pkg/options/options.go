package options

import (
	"crypto/tls"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/authentication"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options/retry"
)

// defaultEndpoint specifies the default gRPC endpoint for connecting to the Yandex Cloud API.
const defaultEndpoint = "api.cloud.yandex.net:443"

// defaultKeepalive prevents intermediate load balancers from closing idle
// HTTP/2 streams during long-running server-streaming RPCs (e.g. operation
// progress streams). Tuned conservatively to stay well above typical server
// MinTime policies.
var defaultKeepalive = keepalive.ClientParameters{
	Time:                30 * time.Second,
	Timeout:             10 * time.Second,
	PermitWithoutStream: true,
}

// Option defines a function that modifies the configuration of an Options instance.
type Option func(*Options)

// Options defines a configuration structure for customizing SDK behavior and connections.
type Options struct {
	// Credentials is used to sign and authenticate API requests.
	Credentials credentials.Credentials
	// EndpointsResolver provides or overrides service endpoints.
	// By default, the SDK uses built-in endpoints, but you can
	// supply a custom resolver to target pre-release or private APIs.
	EndpointsResolver endpoints.EndpointsResolver
	// DiscoveryEndpoint specifies a custom URL to retrieve default
	// service endpoints. If unset, the SDK uses its default discovery service.
	DiscoveryEndpoint string
	// Authenticator signs requests and injects auth headers.
	Authenticator authentication.Authenticator
	// TLSConfig allows customizing TLS settings for gRPC connections.
	// If nil, the SDK uses the system default configuration.
	TlsConfig *tls.Config
	// Plaintext, when true, disables TLS and connects over Plaintext.
	// This is useful for local testing or when an external proxy handles TLS.
	Plaintext bool
	// Logger provides structured logging functionality using zap Logger.
	// If not set, no logging will be performed.
	Logger *zap.Logger

	CustomDialOpts      []grpc.DialOption
	RetryOptions        []retry.RetryOption
	DefaultRetryOptions bool

	// Keepalive configures client-side HTTP/2 keepalive pings on every
	// gRPC connection in the pool. Nil disables keepalive entirely.
	Keepalive *keepalive.ClientParameters
}

// DefaultOptions initializes and returns an Options struct with default configuration for endpoint and connection timeout.
func DefaultOptions() *Options {
	ka := defaultKeepalive
	return &Options{
		DiscoveryEndpoint: defaultEndpoint,
		Keepalive:         &ka,
	}
}

// WithCredentials sets the provided Credentials to the Options, used for signing and authenticating API requests.
func WithCredentials(creds credentials.Credentials) Option {
	return func(o *Options) {
		o.Credentials = creds
	}
}

// WithEndpointsResolver configures a custom EndpointsResolver to dynamically resolve gRPC service endpoints for the SDK.
func WithEndpointsResolver(resolver endpoints.EndpointsResolver) Option {
	return func(o *Options) {
		o.EndpointsResolver = resolver
	}
}

// WithDiscoveryEndpoint sets a custom discovery endpoint URL for resolving service endpoints in the SDK configuration.
func WithDiscoveryEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.DiscoveryEndpoint = endpoint
	}
}

// WithAuthenticator sets the provided Authenticator for authentication in the Options configuration.
func WithAuthenticator(auth authentication.Authenticator) Option {
	return func(o *Options) {
		o.Authenticator = auth
	}
}

// WithTLSConfig sets a custom TLS configuration for gRPC connections by assigning it to the Options struct.
func WithTLSConfig(config *tls.Config) Option {
	return func(o *Options) {
		o.TlsConfig = config
	}
}

// WithPlaintext is an Option that configures the SDK to use Plaintext communication, disabling TLS for gRPC connections.
func WithPlaintext() Option {
	return func(o *Options) {
		o.Plaintext = true
	}
}

// WithDefaultRetryOptions enables default retry handling by setting the `DefaultRetryOptions` field to true in Options.
// it could be used in combination with WithRetryOptions to override default retry options.
func WithDefaultRetryOptions() Option {
	return func(o *Options) {
		o.DefaultRetryOptions = true
	}
}

// WithRetryOptions applies retry options to the SDK's configuration.
func WithRetryOptions(opts ...retry.RetryOption) Option {
	return func(o *Options) {
		o.RetryOptions = opts
	}
}

// WithCustomDialOptions injects custom gRPC dial options into the SDK's configuration.
// It ovverides any other dial options that may have been set by other Options.
func WithCustomDialOptions(opts ...grpc.DialOption) Option {
	return func(o *Options) {
		o.CustomDialOpts = opts
	}
}

// WithLogger sets the structured logger used by the SDK and its middleware
// (auth, request-id, grpcdebug). When unset, the SDK uses a no-op logger.
func WithLogger(logger *zap.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithKeepalive overrides the default client-side HTTP/2 keepalive parameters
// applied to every gRPC connection. Use this to tune ping cadence for very
// long-running streams or to comply with stricter server policies.
func WithKeepalive(params keepalive.ClientParameters) Option {
	return func(o *Options) {
		o.Keepalive = &params
	}
}

// WithoutKeepalive disables client-side HTTP/2 keepalive pings.
func WithoutKeepalive() Option {
	return func(o *Options) {
		o.Keepalive = nil
	}
}
