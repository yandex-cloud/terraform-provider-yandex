package ycsdk

import (
	"context"
	"crypto/tls"
	"fmt"
	"runtime/debug"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpccreds "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/reflect/protoreflect"

	endpointpb "github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/authentication"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/log"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options/retry"
	"github.com/yandex-cloud/go-sdk/v2/pkg/transport"
	transportgrpc "github.com/yandex-cloud/go-sdk/v2/pkg/transport/grpc"
	transportauth "github.com/yandex-cloud/go-sdk/v2/pkg/transport/middleware/authentication"
	endpointsdk "github.com/yandex-cloud/go-sdk/v2/services/endpoint"
	endpointssdk "github.com/yandex-cloud/go-sdk/v2/services/endpoints"
)

var IamTokenCreateEndpoint = protoreflect.FullName("yandex.cloud.iam.v1.IamTokenService.Create")

var _ transport.Connector = (*SDK)(nil)

// SDK provides a client connection wrapper managing connection pooling and endpoint resolution for gRPC services.
type SDK struct {
	ctx              context.Context
	connPool         *transportgrpc.ConnPool
	conn             transport.Connector
	authenticator    authentication.Authenticator
	endpointResolver endpoints.EndpointsResolver
}

// Build initializes and configures an SDK instance with the provided options and context.
// It applies default configurations and validates necessary parameters like credentials and endpoints.
// Returns an SDK instance or an error if initialization fails.
func Build(ctx context.Context, opts ...options.Option) (*SDK, error) {
	buildOptions := options.DefaultOptions()
	for _, opt := range opts {
		opt(buildOptions)
	}
	if buildOptions.Credentials == nil {
		return nil, fmt.Errorf("credentials must be provided")
	}

	logger := zap.NewNop()
	if buildOptions.Logger != nil {
		logger = buildOptions.Logger
	}

	if injector, ok := buildOptions.Credentials.(log.LogInjector); ok {
		injector.InjectLogger(logger)
	}

	var err error
	if buildOptions.EndpointsResolver == nil {
		buildOptions.EndpointsResolver, err = buildEndpointsResolver(ctx, buildOptions.DiscoveryEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to get endpointsResolver: %w", err)
		}
	}

	if buildOptions.Authenticator == nil {
		buildOptions.Authenticator, err = defaultAuthenticator(ctx, logger, buildOptions.Credentials, buildOptions.EndpointsResolver)
		if err != nil {
			return nil, fmt.Errorf("failed to create authenticator: %w", err)
		}
	}

	if injector, ok := buildOptions.Authenticator.(log.LogInjector); ok {
		injector.InjectLogger(
			logger,
		)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithUserAgent(userAgent()),
	}

	if buildOptions.Keepalive != nil {
		dialOpts = append(dialOpts, grpc.WithKeepaliveParams(*buildOptions.Keepalive))
	}

	if _, ok := buildOptions.Credentials.(*credentials.NoCredentials); !ok {
		tokenMiddleware := transportauth.NewIAMTokenMiddleware(buildOptions.Authenticator, transportauth.WithLogger(logger))
		dialOpts = append(dialOpts,
			grpc.WithChainUnaryInterceptor(tokenMiddleware.InterceptUnary),
			grpc.WithChainStreamInterceptor(tokenMiddleware.InterceptStream),
		)
	}

	if buildOptions.Plaintext {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := buildOptions.TlsConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(grpccreds.NewTLS(tlsConfig)))
	}

	if buildOptions.DefaultRetryOptions {
		retryOpt, err := retry.DefaultRetryDialOption()
		if err != nil {
			return nil, fmt.Errorf("failed to apply default retry options: %w", err)
		}
		dialOpts = append(dialOpts, retryOpt)
	}

	if len(buildOptions.RetryOptions) > 0 {
		retryOpt, err := retry.RetryDialOption(buildOptions.RetryOptions...)
		if err != nil {
			return nil, fmt.Errorf("failed to apply retry options: %w", err)
		}
		dialOpts = append(dialOpts, retryOpt)
	}

	dialOpts = append(dialOpts, buildOptions.CustomDialOpts...)

	connectionPool := transportgrpc.NewConnPool(dialOpts)
	return &SDK{
		ctx:              ctx,
		conn:             transport.NewConnector(buildOptions.EndpointsResolver, connectionPool),
		endpointResolver: buildOptions.EndpointsResolver,
		connPool:         connectionPool,
		authenticator:    buildOptions.Authenticator,
	}, nil
}

// GetConnection retrieves a gRPC client connection for the specified method with optional call options.
func (sdk *SDK) GetConnection(ctx context.Context, method protoreflect.FullName, opts ...grpc.CallOption) (grpc.ClientConnInterface, error) {
	return sdk.conn.GetConnection(ctx, method, opts...)
}

// Shutdown gracefully terminates the SDK by closing all active gRPC connections in the connection pool.
func (sdk *SDK) Shutdown(ctx context.Context) error {
	return sdk.connPool.Shutdown(ctx)
}

// userAgent returns the User-Agent string that includes the SDK name and its version, read from the build info.
func userAgent() string {
	cloudUserAgent := "yandex-cloud/go-sdk-v2"

	build, _ := debug.ReadBuildInfo()
	version := "unknown"

	if build.Main.Version != "" {
		version = build.Main.Version
	}

	return fmt.Sprintf("%s/%s", cloudUserAgent, version)
}

// defaultAuthenticator initializes an Authenticator using provided credentials and an endpoint resolver in the given context.
func defaultAuthenticator(ctx context.Context, logger *zap.Logger, creds credentials.Credentials, resolver endpoints.EndpointsResolver) (authentication.Authenticator, error) {
	authEndpoint, err := resolver.Endpoint(ctx, IamTokenCreateEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth endpoint: %w", err)
	}

	authernticator, err := authentication.NewAuthenticatorFromEndpoint(logger, creds, authEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}
	return authernticator, nil
}

// BuildEndpointsResolver creates an EndpointsResolver using a discovery endpoint to dynamically map service prefixes.
func buildEndpointsResolver(ctx context.Context, discoveryEndpoint string) (endpoints.EndpointsResolver, error) {
	conn := transport.NewSingleConnector(discoveryEndpoint, grpc.WithTransportCredentials(grpccreds.NewTLS(&tls.Config{})))
	client := endpointsdk.NewApiEndpointClient(conn)

	resp, err := client.List(ctx, &endpointpb.ListApiEndpointsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}

	// Map endpoint IDs to addresses
	endpointMap := make(map[string]string, len(resp.Endpoints))
	for _, ep := range resp.Endpoints {
		endpointMap[ep.Id] = ep.Address
	}

	// Build resolver from dynamic endpoints
	p2e := make(endpoints.PrefixToEndpoint, len(endpointssdk.DynamicEndpoints))
	for prefix, id := range endpointssdk.DynamicEndpoints {
		if addr, ok := endpointMap[id]; ok {
			p2e[prefix] = endpoints.NewEndpointParams(addr)
		}
	}
	p2e["yandex.cloud.endpoint"] = endpoints.NewEndpointParams(discoveryEndpoint)

	return endpoints.NewPrefixEndpointsResolver(p2e), nil
}

func (sdk *SDK) CreateIAMToken(ctx context.Context) (authentication.IamToken, error) {
	return sdk.authenticator.CreateIAMToken(ctx)
}

func (sdk *SDK) GetEndpoint(method protoreflect.FullName) (*endpoints.Endpoint, error) {
	return sdk.endpointResolver.Endpoint(sdk.ctx, method)
}

func (sdk *SDK) WithEndpoints(endpoints *endpoints.PrefixEndpointsResolver) *SDK {
	return &SDK{
		ctx:              sdk.ctx,
		conn:             transport.NewConnector(endpoints, sdk.connPool),
		endpointResolver: endpoints,
		connPool:         sdk.connPool,
		authenticator:    sdk.authenticator,
	}
}

func (sdk *SDK) WithEnpointsResolver(endpointsResolver endpoints.EndpointsResolver) *SDK {
	return &SDK{
		ctx:              sdk.ctx,
		conn:             transport.NewConnector(endpointsResolver, sdk.connPool),
		endpointResolver: endpointsResolver,
		connPool:         sdk.connPool,
		authenticator:    sdk.authenticator,
	}
}

// WithEndpoint returns a new SDK that overrides a single endpoint by proto prefix,
// keeping all other endpoints from the current resolver intact. Falls back to a
// resolver containing only the override if the current resolver is not prefix-based.
func (sdk *SDK) WithEndpoint(prefix protoreflect.FullName, params *endpoints.EndpointParams) *SDK {
	override := endpoints.PrefixToEndpoint{prefix: params}

	var merged endpoints.PrefixToEndpoint
	if base, ok := sdk.endpointResolver.(*endpoints.PrefixEndpointsResolver); ok {
		merged = base.PrefixToEndpoint().Merge(override)
	} else {
		merged = override
	}
	resolver := endpoints.NewPrefixEndpointsResolver(merged)

	return &SDK{
		ctx:              sdk.ctx,
		conn:             transport.NewConnector(resolver, sdk.connPool),
		endpointResolver: resolver,
		connPool:         sdk.connPool,
		authenticator:    sdk.authenticator,
	}
}
