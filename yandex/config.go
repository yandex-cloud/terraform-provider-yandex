package yandex

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yandex-cloud/go-sdk/v2/pkg/options/retry"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/storage/s3"

	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/mitchellh/go-homedir"
	endpointpb "github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	ycsdkv2 "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/authentication"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	sdkerrors "github.com/yandex-cloud/go-sdk/v2/pkg/errors"
	iamkeyv2 "github.com/yandex-cloud/go-sdk/v2/pkg/iamkey"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"github.com/yandex-cloud/go-sdk/v2/pkg/transport/middleware/idempotency"
	"github.com/yandex-cloud/go-sdk/v2/pkg/transport/middleware/requestid"
	endpointsdk "github.com/yandex-cloud/go-sdk/v2/services/endpoints"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/logging"
)

type iamToken struct {
	Token     string
	expiresAt time.Time
}

func (t iamToken) IsValid() bool {
	return t.Token != "" && t.expiresAt.After(time.Now())
}

type Config struct {
	Endpoint                       string
	FolderID                       string
	CloudID                        string
	OrganizationID                 string
	Zone                           string
	Token                          string
	ServiceAccountKeyFileOrContent string
	Plaintext                      bool
	Insecure                       bool
	MaxRetries                     int
	StorageEndpoint                string
	YMQEndpoint                    string
	Region                         string

	// These storage access keys are optional and only used when
	// storage data/resource doesn't have own access keys explicitly specified.
	StorageAccessKey string
	StorageSecretKey string

	// These YMQ access keys are optional and only used when
	// Message Queue resource doesn't have own access keys explicitly specified.
	YMQAccessKey string
	YMQSecretKey string

	SharedCredentialsFile string
	Profile               string

	// contextWithClientTraceID is a context that has client-trace-id in its metadata
	// It is initialized from stopContext at the same time as SDK v2.
	contextWithClientTraceID context.Context

	userAgent         string
	SDK               *ycsdkv2.SDK
	sharedCredentials *SharedCredentials
	defaultS3Client   *s3.Client
	iamToken          *iamToken
}

// this function return context with added client trace id
func (c *Config) Context() context.Context {
	return c.contextWithClientTraceID
}

// this function returns context with client trace id AND timeout
func (c *Config) ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.contextWithClientTraceID, timeout)
}

// this function adds client trace id to provided context
func (c *Config) ContextWithClientTraceID(ctx context.Context) context.Context {
	if md, ok := metadata.FromOutgoingContext(c.contextWithClientTraceID); ok && md != nil {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// Client configures and returns a fully initialized Yandex Cloud sdk
func (c *Config) initAndValidate(stopContext context.Context, terraformVersion string, sweeper bool) error {
	c.contextWithClientTraceID = requestid.ContextWithClientTraceID(stopContext, uuid.New().String())

	c.userAgent = config.BuildUserAgent(terraformVersion, sweeper)

	headerMD := metadata.Pairs("user-agent", c.userAgent)

	requestIDInterceptor := requestid.Interceptor()
	idempotencyIntepceptor := idempotency.Interceptor()

	var interceptors = []grpc.UnaryClientInterceptor{
		idempotencyIntepceptor,
		requestIDInterceptor,
	}

	// Support deep API logging in case user has requested it.
	if os.Getenv("TF_ENABLE_API_LOGGING") != "" {
		log.Print("[INFO] API logging has been requested, turning on")
		interceptors = append(interceptors, logging.NewAPILoggingUnaryInterceptor())
	}

	// Make sure retry interceptor is above id interceptor.
	// Now we will have new request id for every retry attempt.
	interceptorChain := grpc_middleware.ChainUnaryClient(interceptors...)

	grpcOptions := []grpc.DialOption{
		grpc.WithUserAgent(c.userAgent),
		grpc.WithDefaultCallOptions(grpc.Header(&headerMD)),
		grpc.WithUnaryInterceptor(interceptorChain),
	}
	if c.MaxRetries > 1 {
		retryOptions, err := retry.RetryDialOption(
			retry.WithRetries(retry.DefaultNameConfig(), c.MaxRetries),
			retry.WithThrottlingMode(retry.ThrottlingModeTemporary),
		)
		if err != nil {
			return err
		}
		grpcOptions = append(grpcOptions, retryOptions)
	}

	var endpointCredentials grpccredentials.TransportCredentials = grpccredentials.NewTLS(&tls.Config{})
	if c.Plaintext {
		endpointCredentials = insecure.NewCredentials()
	} else if c.Insecure {
		endpointCredentials = grpccredentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}
	endpointOptions := make([]grpc.DialOption, 0, len(grpcOptions)+1)
	endpointOptions = append(endpointOptions, grpc.WithTransportCredentials(endpointCredentials))
	endpointOptions = append(endpointOptions, grpcOptions...)

	discoveryEndpoint := c.Endpoint
	if discoveryEndpoint == "" {
		discoveryEndpoint = common.DefaultEndpoint
	}
	credentialsV2, err := c.credentialsV2()
	if err != nil {
		return err
	}
	endpointResolver := &sdkV2EndpointsResolver{
		discoveryEndpoint: discoveryEndpoint,
		dialOptions:       endpointOptions,
	}
	optionsV2 := []options.Option{
		options.WithCredentials(credentialsV2),
		options.WithEndpointsResolver(endpointResolver),
		options.WithAuthenticator(&sdkV2Authenticator{
			credentials: credentialsV2,
			resolver:    endpointResolver,
		}),
		options.WithoutKeepalive(),
	}
	if c.Plaintext {
		optionsV2 = append(optionsV2, options.WithPlaintext())
	}
	if c.Insecure {
		optionsV2 = append(optionsV2, options.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}
	c.SDK, err = ycsdkv2.Build(c.contextWithClientTraceID, optionsV2...)
	if err != nil {
		return err
	}

	if err := c.initSharedCredentials(); err != nil {
		return err
	}

	return c.initializeDefaultS3Client(stopContext)
}

type sdkV2EndpointsResolver struct {
	mutex             sync.Mutex
	discoveryEndpoint string
	dialOptions       []grpc.DialOption
	endpoints         map[protoreflect.FullName]*endpoints.Endpoint
}

const sdkV2DialTimeout = 20 * time.Second

func (r *sdkV2EndpointsResolver) Endpoint(
	ctx context.Context,
	method protoreflect.FullName,
	_ ...grpc.CallOption,
) (*endpoints.Endpoint, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.endpoints == nil {
		resolved, err := discoverSDKV2Endpoints(ctx, r.discoveryEndpoint, r.dialOptions)
		if err != nil {
			return nil, err
		}
		r.endpoints = resolved
	}

	for name := method; name != ""; name = name.Parent() {
		if endpoint, ok := r.endpoints[name]; ok {
			return endpoint, nil
		}
	}
	return nil, &sdkerrors.EndpointNotFoundError{Method: method}
}

type sdkV2Authenticator struct {
	mutex         sync.Mutex
	credentials   credentials.Credentials
	resolver      endpoints.EndpointsResolver
	authenticator authentication.Authenticator
}

func (a *sdkV2Authenticator) CreateIAMToken(ctx context.Context) (authentication.IamToken, error) {
	authenticator, err := a.get(ctx)
	if err != nil {
		return nil, err
	}
	return authenticator.CreateIAMToken(ctx)
}

func (a *sdkV2Authenticator) CreateIAMTokenForServiceAccount(
	ctx context.Context,
	serviceAccountID string,
) (authentication.IamToken, error) {
	authenticator, err := a.get(ctx)
	if err != nil {
		return nil, err
	}
	return authenticator.CreateIAMTokenForServiceAccount(ctx, serviceAccountID)
}

func (a *sdkV2Authenticator) get(ctx context.Context) (authentication.Authenticator, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.authenticator != nil {
		return a.authenticator, nil
	}
	if _, ok := a.credentials.(credentials.NonExchangeableCredentials); ok {
		a.authenticator = authentication.NewAuthenticator(zap.NewNop(), a.credentials, nil)
		return a.authenticator, nil
	}
	endpoint, err := a.resolver.Endpoint(ctx, ycsdkv2.IamTokenCreateEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth endpoint: %w", err)
	}
	a.authenticator, err = authentication.NewAuthenticatorFromEndpoint(zap.NewNop(), a.credentials, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}
	return a.authenticator, nil
}

func discoverSDKV2Endpoints(
	ctx context.Context,
	discoveryEndpoint string,
	dialOptions []grpc.DialOption,
) (map[protoreflect.FullName]*endpoints.Endpoint, error) {
	ctx, cancel := context.WithTimeout(ctx, sdkV2DialTimeout)
	defer cancel()

	conn, err := grpc.NewClient(discoveryEndpoint, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API endpoint %q: %w", discoveryEndpoint, err)
	}
	defer conn.Close()

	response, err := endpointpb.NewApiEndpointServiceClient(conn).List(ctx, &endpointpb.ListApiEndpointsRequest{
		PageSize: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list API endpoints: %w", err)
	}

	addresses := make(map[string]string, len(response.Endpoints))
	for _, endpoint := range response.Endpoints {
		addresses[endpoint.Id] = endpoint.Address
	}

	resolved := make(map[protoreflect.FullName]*endpoints.Endpoint, len(endpointsdk.DynamicEndpoints)+1)
	for prefix, serviceID := range endpointsdk.DynamicEndpoints {
		if address, ok := addresses[serviceID]; ok {
			resolved[prefix] = &endpoints.Endpoint{Addr: address, DialOptions: dialOptions}
		}
	}
	resolved["yandex.cloud.endpoint"] = &endpoints.Endpoint{
		Addr:        discoveryEndpoint,
		DialOptions: dialOptions,
	}

	return resolved, nil
}

func (c *Config) initSharedCredentials() error {
	if c.SharedCredentialsFile == "" {
		return nil
	}

	sharedCredentialsProvider := SharedCredentialsProvider{c.SharedCredentialsFile, c.Profile}
	sharedCredentials, err := sharedCredentialsProvider.Retrieve()
	if err != nil {
		return err
	}
	c.sharedCredentials = sharedCredentials
	return nil
}

func (c *Config) resolveStorageAccessKeys() (string, string) {
	if c.sharedCredentials == nil || (c.StorageAccessKey != "" && c.StorageSecretKey != "") {
		return c.StorageAccessKey, c.StorageSecretKey // from 'provider "yandex" {...}' or ENV vars
	}
	return c.sharedCredentials.StorageAccessKey, c.sharedCredentials.StorageSecretKey
}

func (c *Config) initializeDefaultS3Client(ctx context.Context) (err error) {
	if c.StorageEndpoint == "" {
		return nil
	}

	accessKey, secretKey := c.resolveStorageAccessKeys()
	iamToken, err := c.getIAMToken(ctx)
	if err != nil {
		log.Println("[WARN] Failed to get IAM token for default storage client:", err)
		iamToken = ""
	}

	if (accessKey == "" || secretKey == "") && iamToken == "" {
		return nil
	}

	c.defaultS3Client, err = s3.NewClient(ctx, accessKey, secretKey, iamToken, c.StorageEndpoint)
	return err
}

func (c *Config) credentialsV2() (credentials.Credentials, error) {
	if c.ServiceAccountKeyFileOrContent != "" {
		contents, _, err := pathOrContents(c.ServiceAccountKeyFileOrContent)
		if err != nil {
			return nil, fmt.Errorf("Error loading credentials: %s", err)
		}

		key, err := iamKeyV2FromJSONContent(contents)
		if err != nil {
			return nil, err
		}
		return credentials.ServiceAccountKey(key)
	}

	if c.Token != "" {
		if strings.HasPrefix(c.Token, "t1.") && strings.Count(c.Token, ".") == 2 {
			return credentials.IAMToken(c.Token), nil
		}
		return credentials.OAuthToken(c.Token), nil
	}

	if sa := credentials.InstanceServiceAccount(); checkServiceAccountV2Available(c.Context(), sa) {
		return sa, nil
	}

	return nil, fmt.Errorf(
		"one of 'token' or 'service_account_key_file' should be specified; if you are inside compute instance, you can attach service account to it in order to authenticate via instance service account",
	)
}

func (c *Config) getIAMToken(ctx context.Context) (string, error) {
	if c.iamToken != nil && c.iamToken.IsValid() {
		return c.iamToken.Token, nil
	}

	resp, err := c.SDK.CreateIAMToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get IAM token: %w", err)
	}

	c.iamToken = &iamToken{
		Token:     resp.GetIamToken(),
		expiresAt: resp.GetExpiresAt(),
	}

	return c.iamToken.Token, nil
}

func iamKeyV2FromJSONContent(content string) (*iamkeyv2.Key, error) {
	key := &iamkeyv2.Key{}
	err := json.Unmarshal([]byte(content), key)
	if err != nil {
		return nil, fmt.Errorf("key unmarshal fail: %s", err)
	}
	return key, nil
}

func checkServiceAccountV2Available(ctx context.Context, sa credentials.NonExchangeableCredentials) bool {
	dialer := net.Dialer{Timeout: 50 * time.Millisecond}
	conn, err := dialer.Dial("tcp", net.JoinHostPort(credentials.InstanceMetadataAddr, "80"))
	if err != nil {
		return false
	}
	_ = conn.Close()
	_, err = sa.IAMToken(ctx)
	return err == nil
}

// copy of github.com/hashicorp/terraform-plugin-sdk/helper/pathorcontents.Read()
func pathOrContents(poc string) (string, bool, error) {
	if len(poc) == 0 {
		return poc, false, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, true, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		return string(contents), true, err
	}

	return poc, false, nil
}
