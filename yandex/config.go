package yandex

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/mitchellh/go-homedir"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yandex-cloud/go-sdk/pkg/requestid"
	"github.com/yandex-cloud/go-sdk/pkg/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/logging"
)

const (
	defaultExponentialBackoffBase = 50 * time.Millisecond
	defaultExponentialBackoffCap  = 1 * time.Minute
)

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

	// contextWithClientTraceID is a context that has client-trace-id in its metadata
	// It is initialized from stopContext at the same time as ycsdk.SDK
	contextWithClientTraceID context.Context

	userAgent       string
	sdk             *ycsdk.SDK
	defaultS3Client *s3.S3
}

// this function return context with added client trace id
func (c *Config) Context() context.Context {
	return c.contextWithClientTraceID
}

// this function returns context with client trace id AND timeout
func (c *Config) ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.contextWithClientTraceID, timeout)
}

// Client configures and returns a fully initialized Yandex.Cloud sdk
func (c *Config) initAndValidate(stopContext context.Context, terraformVersion string, sweeper bool) error {
	c.contextWithClientTraceID = requestid.ContextWithClientTraceID(stopContext, uuid.New().String())

	credentials, err := c.credentials()
	if err != nil {
		return err
	}

	yandexSDKConfig := &ycsdk.Config{
		Credentials: credentials,
		Endpoint:    c.Endpoint,
		Plaintext:   c.Plaintext,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: c.Insecure,
		},
	}

	providerNameAndVersion := getProviderNameAndVersion()
	terraformURL := "https://www.terraform.io"

	if sweeper {
		c.userAgent = "Terraform Sweeper"
	} else {
		c.userAgent = fmt.Sprintf("Terraform/%s (%s) %s", terraformVersion, terraformURL, providerNameAndVersion)
	}

	headerMD := metadata.Pairs("user-agent", c.userAgent)

	requestIDInterceptor := requestid.Interceptor()

	retryInterceptor := retry.Interceptor(
		retry.WithMax(c.MaxRetries),
		retry.WithCodes(codes.Unavailable),
		retry.WithAttemptHeader(true),
		retry.WithBackoff(backoffExponentialWithJitter(defaultExponentialBackoffBase, defaultExponentialBackoffCap)))

	var interceptors = []grpc.UnaryClientInterceptor{
		retryInterceptor,
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

	c.sdk, err = ycsdk.Build(c.contextWithClientTraceID, *yandexSDKConfig,
		grpc.WithUserAgent(c.userAgent),
		grpc.WithDefaultCallOptions(grpc.Header(&headerMD)),
		grpc.WithUnaryInterceptor(interceptorChain))

	if err == nil {
		err = c.initializeDefaultS3Client()
	}

	return err
}

func (c *Config) initializeDefaultS3Client() (err error) {
	if c.StorageEndpoint == "" || (c.StorageAccessKey == "" && c.StorageSecretKey == "") {
		return nil
	}

	if c.StorageAccessKey == "" || c.StorageSecretKey == "" {
		return fmt.Errorf("both storage access key and storage secret key should be specified or not specified")
	}

	c.defaultS3Client, err = newS3Client(c.StorageEndpoint, c.StorageAccessKey, c.StorageSecretKey)

	return err
}

func (c *Config) credentials() (ycsdk.Credentials, error) {
	if c.ServiceAccountKeyFileOrContent != "" {
		contents, _, err := pathOrContents(c.ServiceAccountKeyFileOrContent)
		if err != nil {
			return nil, fmt.Errorf("Error loading credentials: %s", err)
		}

		key, err := iamKeyFromJSONContent(contents)
		if err != nil {
			return nil, err
		}
		return ycsdk.ServiceAccountKey(key)
	}

	if c.Token != "" {
		if strings.HasPrefix(c.Token, "t1.") && strings.Count(c.Token, ".") == 2 {
			return ycsdk.NewIAMTokenCredentials(c.Token), nil
		}
		return ycsdk.OAuthToken(c.Token), nil
	}

	if sa := ycsdk.InstanceServiceAccount(); checkServiceAccountAvailable(c.Context(), sa) {
		return sa, nil
	}

	return nil, fmt.Errorf("one of 'token' or 'service_account_key_file' should be specified; if you are inside compute instance, you can attach service account to it in order to authenticate via instance service account")
}

func iamKeyFromJSONContent(content string) (*iamkey.Key, error) {
	key := &iamkey.Key{}
	err := json.Unmarshal([]byte(content), key)
	if err != nil {
		return nil, fmt.Errorf("key unmarshal fail: %s", err)
	}
	return key, nil
}

func backoffExponentialWithJitter(base time.Duration, cap time.Duration) retry.BackoffFunc {
	return func(attempt int) time.Duration {
		// First call of BackoffFunc would be with attempt arq equal 0
		log.Printf("[DEBUG] API call retry attempt %d", attempt+1)

		to := getExponentialTimeout(attempt, base)
		// Using float types here, because exponential time can be really big, and converting it to time.Duration may
		// result in undefined behaviour. Its safe conversion, when we have compared it to our 'cap' value.
		if to > float64(cap) {
			to = float64(cap)
		}

		return time.Duration(to * rand.Float64())
	}
}

func getExponentialTimeout(attempt int, base time.Duration) float64 {
	mult := math.Pow(2, float64(attempt))
	return float64(base) * mult
}

func getProviderNameAndVersion() string {
	// version is part of binary name
	// https://www.terraform.io/docs/configuration/providers.html#plugin-names-and-versions
	fullBinaryPath := os.Args[0]
	binaryName := filepath.Base(fullBinaryPath)
	parts := strings.Split(binaryName, "_")

	if len(parts) < 2 {
		return "unknown/unknown"
	}

	parts[1] = strings.TrimPrefix(parts[1], "v")

	return strings.Join(parts[:2], "/")
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
