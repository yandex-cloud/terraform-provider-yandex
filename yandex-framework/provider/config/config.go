package config

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
	"time"

	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/go-homedir"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yandex-cloud/go-sdk/pkg/idempotency"
	"github.com/yandex-cloud/go-sdk/pkg/requestid"
	"github.com/yandex-cloud/go-sdk/pkg/retry/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/logging"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/yqsdk"
)

const (
	DefaultTimeout = 1 * time.Minute
)

type State struct {
	Endpoint                       types.String `tfsdk:"endpoint"`
	YQEndpoint                     types.String `tfsdk:"yq_endpoint"`
	FolderID                       types.String `tfsdk:"folder_id"`
	CloudID                        types.String `tfsdk:"cloud_id"`
	OrganizationID                 types.String `tfsdk:"organization_id"`
	Zone                           types.String `tfsdk:"zone"`
	Token                          types.String `tfsdk:"token"`
	ServiceAccountKeyFileOrContent types.String `tfsdk:"service_account_key_file"`
	Plaintext                      types.Bool   `tfsdk:"plaintext"`
	Insecure                       types.Bool   `tfsdk:"insecure"`
	MaxRetries                     types.Int64  `tfsdk:"max_retries"`
	StorageEndpoint                types.String `tfsdk:"storage_endpoint"`
	YMQEndpoint                    types.String `tfsdk:"ymq_endpoint"`
	Region                         types.String `tfsdk:"region_id"`

	// These storage access keys are optional and only used when
	// storage data/resource doesn't have own access keys explicitly specified.
	StorageAccessKey types.String `tfsdk:"storage_access_key"`
	StorageSecretKey types.String `tfsdk:"storage_secret_key"`

	// These YMQ access keys are optional and only used when
	// Message Queue resource doesn't have own access keys explicitly specified.
	YMQAccessKey types.String `tfsdk:"ymq_access_key"`
	YMQSecretKey types.String `tfsdk:"ymq_secret_key"`

	SharedCredentialsFile types.String `tfsdk:"shared_credentials_file"`
	Profile               types.String `tfsdk:"profile"`
	//
	//sharedCredentials *SharedCredentials
	//defaultS3Client   *s3.S3
}

// TODO: remove yandex.Config when it is not used
type iamToken struct {
	Token     string
	expiresAt time.Time
}

func (t iamToken) IsValid() bool {
	return t.Token != "" && t.expiresAt.After(time.Now())
}

type Config struct {
	ProviderState State

	UserAgent types.String
	SDK       *ycsdk.SDK
	YqSdk     *yqsdk.SDK
	iamToken  *iamToken
}

// Client configures and returns a fully initialized Yandex Cloud SDK
func (c *Config) InitAndValidate(ctx context.Context, terraformVersion string, sweeper bool) error {
	ctx = requestid.ContextWithClientTraceID(ctx, uuid.New().String())

	credentials, err := c.Credentials(ctx)
	if err != nil {
		return err
	}

	yandexSDKConfig := &ycsdk.Config{
		Credentials: credentials,
		Endpoint:    c.ProviderState.Endpoint.ValueString(),
		Plaintext:   c.ProviderState.Plaintext.ValueBool(),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: c.ProviderState.Insecure.ValueBool(),
		},
	}

	c.UserAgent = types.StringValue(config.BuildUserAgent(terraformVersion, sweeper))

	headerMD := metadata.Pairs("user-agent", c.UserAgent.ValueString())

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

	retryOptions, err := retry.RetryDialOption(
		retry.WithRetries(retry.DefaultNameConfig(), int(c.ProviderState.MaxRetries.ValueInt64())),
		retry.WithThrottlingMode(retry.ThrottlingModeTemporary),
	)
	if err != nil {
		return err
	}

	c.SDK, err = ycsdk.Build(ctx, *yandexSDKConfig,
		grpc.WithUserAgent(c.UserAgent.ValueString()),
		grpc.WithDefaultCallOptions(grpc.Header(&headerMD)),
		grpc.WithUnaryInterceptor(interceptorChain),
		retryOptions)

	if err != nil {
		return err
	}

	yqSDKConfig := &yqsdk.Config{
		AuthTokenProvider: func(ctx context.Context) (string, error) { return c.getIAMToken(ctx) },
		FolderID:          c.ProviderState.FolderID.ValueString(),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: c.ProviderState.Insecure.ValueBool(),
		},
		Endpoint:  c.ProviderState.YQEndpoint.ValueString(),
		Plaintext: c.ProviderState.Plaintext.ValueBool(),
	}

	c.YqSdk, err = yqsdk.NewYQSDK(ctx, *yqSDKConfig)
	if err != nil {
		return err
	}

	return err
}

func (c *Config) Credentials(ctx context.Context) (ycsdk.Credentials, error) {
	if c.ProviderState.ServiceAccountKeyFileOrContent.ValueString() != "" {
		contents, _, err := pathOrContents(c.ProviderState.ServiceAccountKeyFileOrContent.ValueString())
		if err != nil {
			return nil, fmt.Errorf("Error loading Credentials: %s", err)
		}

		key, err := iamKeyFromJSONContent(contents)
		if err != nil {
			return nil, err
		}
		return ycsdk.ServiceAccountKey(key)
	}

	if c.ProviderState.Token.ValueString() != "" {
		if strings.HasPrefix(
			c.ProviderState.Token.ValueString(), "t1.",
		) && strings.Count(
			c.ProviderState.Token.ValueString(), ".",
		) == 2 {
			return ycsdk.NewIAMTokenCredentials(c.ProviderState.Token.ValueString()), nil
		}
		return ycsdk.OAuthToken(c.ProviderState.Token.ValueString()), nil
	}

	if sa := ycsdk.InstanceServiceAccount(); checkServiceAccountAvailable(ctx, sa) {
		return sa, nil
	}

	return nil, fmt.Errorf("one of 'token' or 'service_account_key_file' should be specified;" +
		" if you are inside compute instance, you can attach service account to it in order to " +
		"authenticate via instance service account")
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
		Token: resp.IamToken,
	}
	if resp.ExpiresAt != nil && resp.ExpiresAt.IsValid() {
		c.iamToken.expiresAt = resp.ExpiresAt.AsTime()
	}

	return c.iamToken.Token, nil
}

func iamKeyFromJSONContent(content string) (*iamkey.Key, error) {
	key := &iamkey.Key{}
	err := json.Unmarshal([]byte(content), key)
	if err != nil {
		return nil, fmt.Errorf("key unmarshal fail: %s", err)
	}
	return key, nil
}

func checkServiceAccountAvailable(ctx context.Context, sa ycsdk.NonExchangeableCredentials) bool {
	dialer := net.Dialer{Timeout: 50 * time.Millisecond}
	conn, err := dialer.Dial("tcp", net.JoinHostPort(ycsdk.InstanceMetadataAddr, "80"))
	if err != nil {
		return false
	}
	_ = conn.Close()
	_, err = sa.IAMToken(ctx)
	return err == nil
}

// copy of github.com/hashicorp/terraform-plugin-SDK/helper/pathorcontents.Read()
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
