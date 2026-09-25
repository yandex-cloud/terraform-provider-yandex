package endpoints

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// EndpointParams represents the parameters required to configure an endpoint, including its address and options.
type EndpointParams struct {
	Addr    string
	Options []EndpointOption
}

// NewEndpointParams creates a new EndpointParams instance with the specified address and optional EndpointOptions.
func NewEndpointParams(addr string, opts ...EndpointOption) *EndpointParams {
	return &EndpointParams{Addr: addr, Options: opts}
}

// Append adds one or more EndpointOption instances to the Options slice and returns the updated EndpointParams.
func (ep *EndpointParams) Append(opts ...EndpointOption) *EndpointParams {
	ep.Options = append(ep.Options, opts...)
	return ep
}

// Build constructs an Endpoint using the provided address and dial options from the EndpointParams configuration.
func (ep *EndpointParams) Build() *Endpoint {
	var creds grpc.DialOption
	dialOptions := make([]grpc.DialOption, 0, len(ep.Options)+1)

	// Separate transport credentials from other options.
	for _, opt := range ep.Options {
		if tco, ok := opt.(TransportCredentialEndpointOption); ok {
			// Always use the last-specified transport credential option.
			creds = tco.DialOption()
		} else {
			dialOptions = append(dialOptions, opt.DialOption())
		}
	}

	// Use default TLS credentials if none provided.
	if creds == nil {
		// grpc.DialOption for TLS using the default system roots.
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}
	// Prepend the transport credentials.
	dialOptions = append([]grpc.DialOption{creds}, dialOptions...)

	return &Endpoint{
		Addr:        ep.Addr,
		DialOptions: dialOptions,
	}
}

// EndpointOption is an interface that represents a configurable option for creating a gRPC endpoint connection.
type EndpointOption interface {
	DialOption() grpc.DialOption
}

// TransportCredentialEndpointOption defines an interface for specifying transport-layer security options for an endpoint.
type TransportCredentialEndpointOption interface {
	EndpointOption
	isTransportCredential()
}

func Plaintext() TransportCredentialEndpointOption { return plaintextOption{} }

// SkipTLSVerify returns a TransportCredentialEndpointOption that disables TLS certificate verification for secure connections.
func SkipTLSVerify() TransportCredentialEndpointOption { return skipTLSVerifyOption{} }

// Keepalive configures client-side keepalive parameters for gRPC connections using the provided parameters.
func Keepalive(params keepalive.ClientParameters) EndpointOption {
	return keepaliveOption(params)
}

// plaintextOption is a struct type used to configure plaintext transport credentials for gRPC connections.
type plaintextOption struct{}

// skipTLSVerifyOption is a type that represents an option to skip TLS certificate verification during a transport.
type skipTLSVerifyOption struct{}

// keepaliveOption is a type alias for keepalive.ClientParameters, used to configure keepalive settings in gRPC clients.
type keepaliveOption keepalive.ClientParameters

// isTransportCredential marks plaintextOption as a type that satisfies the transport credentials interface.
func (plaintextOption) isTransportCredential() {}

// isTransportCredential marks skipTLSVerifyOption as a credential option for transport security use.
func (skipTLSVerifyOption) isTransportCredential() {}

// DialOption returns a grpc.DialOption configured with insecure transport credentials.
func (plaintextOption) DialOption() grpc.DialOption {
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}

// DialOption returns a gRPC DialOption that configures the transport credentials to skip TLS verification.
func (skipTLSVerifyOption) DialOption() grpc.DialOption {
	return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	}))
}

// DialOption returns a grpc.DialOption configured with the keepalive parameters from the keepaliveOption receiver.
func (kp keepaliveOption) DialOption() grpc.DialOption {
	return grpc.WithKeepaliveParams(keepalive.ClientParameters(kp))
}
