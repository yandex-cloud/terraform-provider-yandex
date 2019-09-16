package yandex

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yandex-cloud/go-sdk/pkg/requestid"
	"github.com/yandex-cloud/go-sdk/pkg/retry"
)

const (
	defaultExponentialBackoffBase = 50 * time.Millisecond
	defaultExponentialBackoffCap  = 1 * time.Minute
)

type Config struct {
	Endpoint              string
	FolderID              string
	CloudID               string
	Zone                  string
	Token                 string
	ServiceAccountKeyFile string
	Plaintext             bool
	Insecure              bool
	MaxRetries            int

	// contextWithClientTraceID is a context that has client-trace-id in its metadata
	// It is initialized at the same time as ycsdk.SDK
	contextWithClientTraceID context.Context

	userAgent string
	sdk       *ycsdk.SDK
}

func (c *Config) ContextWithClientTraceID() context.Context {
	return c.contextWithClientTraceID
}

// Client configures and returns a fully initialized Yandex.Cloud sdk
func (c *Config) initAndValidate(terraformVersion string) error {
	if c.Token != "" && c.ServiceAccountKeyFile != "" {
		return fmt.Errorf("one of token or service account key file must be specified, not both (check your config AND environment variables)")
	}

	var credentials ycsdk.Credentials
	if c.Token != "" {
		credentials = ycsdk.OAuthToken(c.Token)
	} else if c.ServiceAccountKeyFile != "" {
		key, err := iamkey.ReadFromJSONFile(c.ServiceAccountKeyFile)
		if err != nil {
			return err
		}

		credentials, err = ycsdk.ServiceAccountKey(key)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("one of token or service account key file must be specified")
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

	c.userAgent = fmt.Sprintf("Terraform/%s (%s) %s", terraformVersion, terraformURL, providerNameAndVersion)

	headerMD := metadata.Pairs("user-agent", c.userAgent)

	requestIDInterceptor := requestid.Interceptor()

	retryInterceptor := retry.Interceptor(
		retry.WithMax(c.MaxRetries),
		retry.WithCodes(codes.Unavailable),
		retry.WithAttemptHeader(true),
		retry.WithBackoff(BackoffExponentialWithJitter(defaultExponentialBackoffBase, defaultExponentialBackoffCap)))

	// Make sure retry interceptor is above id interceptor.
	// Now we will have new request id for every retry attempt.
	interceptorChain := grpc_middleware.ChainUnaryClient(retryInterceptor, requestIDInterceptor)

	c.contextWithClientTraceID = contextWithClientTraceID(context.Background())

	var err error
	c.sdk, err = ycsdk.Build(c.contextWithClientTraceID, *yandexSDKConfig,
		grpc.WithUserAgent(c.userAgent),
		grpc.WithDefaultCallOptions(grpc.Header(&headerMD)),
		grpc.WithUnaryInterceptor(interceptorChain))

	return err
}

func BackoffExponentialWithJitter(base time.Duration, cap time.Duration) retry.BackoffFunc {
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

	if len(parts) != 2 {
		return "unknown/unknown"
	}

	return strings.Join(parts, "/")
}
