package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

const (
	metadataIamTokenValidityThreshold = time.Second * 5
	metadataIamTokenRefreshTimeout    = time.Second * 10

	// metadataTokenCappedExpirationSeconds is a constant for a case you can find below. This implementation overrides
	// the expiration duration to 60 seconds max to handle cases when the token gets invalid: E.g. instance sa is
	// stripped of all permissions (or even worse, deleted), instance is updated to use another sa with valid
	// permissions; however, since the token is not expired it is still used for subsequent requests resulting in
	// Permission Denied. Shorter expiration time is a workaround
	metadataTokenCappedExpirationDuration = time.Second * 60
	metadataTokenRefreshInterval          = metadataTokenCappedExpirationDuration - metadataIamTokenRefreshTimeout
)

var _ NonExchangeableCredentials = (*metadataServiceCredentialProvider)(nil)
var _ MetadataServiceCredentialProvider = (*metadataServiceCredentialProvider)(nil)

type MetadataServiceCredentialProvider interface {
	NonExchangeableCredentials

	Addr() string
	Available(ctx context.Context) bool
}

type metadataServiceCredentialProvider struct {
	metadataServiceAddr string
	client              retryablehttp.Client

	currentTokenMutex          sync.RWMutex
	currentToken               string
	currentTokenRealExpiration time.Time

	refreshMutex    sync.Mutex
	lastRefreshTime time.Time

	refreshStop   chan struct{}
	refreshTicker *time.Ticker
	logger        *zap.Logger
}

var providerInstances = make(map[string]MetadataServiceCredentialProvider)
var providerInstancesLock sync.Mutex

// MetadataService returns credentials provider that queries local metadata service for IAM tokens
// This is currently available on Yandex Cloud Compute Instances instances with a Service Account attached
// https://yandex.cloud/ru/docs/compute/concepts/vm-metadata
func MetadataService() MetadataServiceCredentialProvider {
	return NewMetadataServiceCredentialProvider(GetMetadataServiceAddr())
}

func NewMetadataServiceCredentialProvider(metadataServiceAddr string) MetadataServiceCredentialProvider {
	providerInstancesLock.Lock()
	defer providerInstancesLock.Unlock()

	if provider, ok := providerInstances[metadataServiceAddr]; ok {
		return provider
	}

	provider := metadataServiceCredentialProvider{
		metadataServiceAddr: metadataServiceAddr,
		client: retryablehttp.Client{
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					DialContext: (&net.Dialer{
						Timeout:   time.Second, // One second should be enough for localhost connection.
						KeepAlive: -1,          // No keep alive. Near token per hour requested.
					}).DialContext,
					ResponseHeaderTimeout: 5 * time.Second, // Prevent hanging after successful request write.
				},
			},
			RetryMax:   5,
			CheckRetry: retryablehttp.DefaultRetryPolicy,
			Backoff:    retryablehttp.DefaultBackoff,
		},
		refreshStop:   make(chan struct{}),
		refreshTicker: time.NewTicker(metadataTokenRefreshInterval),
		logger:        zap.NewNop(),
	}
	go provider.tokenRefreshLoop()

	providerInstances[metadataServiceAddr] = &provider

	return &provider
}

func (c *metadataServiceCredentialProvider) InjectLogger(logger *zap.Logger) {
	c.logger = logger
}

func (c *metadataServiceCredentialProvider) tokenRefreshLoop() {
	for {
		select {
		case <-c.refreshStop:
			c.logger.Info("Token refresh loop stopped")
			c.refreshTicker.Stop()
			return
		case <-c.refreshTicker.C:
			c.logger.Debug("Starting token refresh")
			c.refreshTokenWithTimeout()
		}
	}
}

func (c *metadataServiceCredentialProvider) refreshTokenWithTimeout() {
	refreshContext, cancel := context.WithTimeout(context.Background(), metadataIamTokenRefreshTimeout)
	defer cancel()

	if err := c.refreshToken(refreshContext); err != nil {
		c.logger.Error("Failed to refresh token", zap.Error(err))
	}
}

func (c *metadataServiceCredentialProvider) refreshToken(ctx context.Context) error {
	c.refreshMutex.Lock()
	defer c.refreshMutex.Unlock()

	if c.recentlyRefreshedAndValid() {
		c.logger.Debug("Token is still valid, skipping refresh")
		return nil
	}

	c.logger.Debug("Getting new IAM token")
	token, actualExpiration, err := c.iamToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get compute instance service account token from instance metadata service: GET %s: %w", c.url(), err)
	}

	c.currentTokenMutex.Lock()
	defer c.currentTokenMutex.Unlock()

	c.currentTokenRealExpiration = actualExpiration
	c.currentToken = token
	c.lastRefreshTime = time.Now()

	c.logger.Info("Token refreshed successfully",
		zap.Time("expiresAt", actualExpiration),
		zap.Time("refreshedAt", c.lastRefreshTime))

	return nil
}

func (c *metadataServiceCredentialProvider) recentlyRefreshedAndValid() bool {
	if time.Since(c.lastRefreshTime) < metadataTokenCappedExpirationDuration {
		_, err := c.getValidCachedToken()
		return err == nil
	}

	return false
}

func (c *metadataServiceCredentialProvider) iamToken(ctx context.Context) (string, time.Time, error) {
	req, err := retryablehttp.NewRequest("GET", c.url(), nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("request make failed: %w", err)
	}
	req.Header.Set("Metadata-Flavor", "Google")
	reqDump, _ := httputil.DumpRequestOut(req.Request, false)
	c.logger.Debug("Requesting instance SA token", zap.String("request", string(reqDump)))

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		c.logger.Error("Metadata service call failed", zap.Error(err))
		return "", time.Time{}, fmt.Errorf("compute instance metadata service call failed.\n" +
			"Are you inside compute instance?\n" +
			"Details")
	}

	defer resp.Body.Close()
	respDump, _ := httputil.DumpResponse(resp, false)
	c.logger.Debug("Received metadata service response", zap.String("response", string(respDump)))

	if resp.StatusCode == http.StatusNotFound {
		c.logger.Error("Service account not found")
		return "", time.Time{}, fmt.Errorf("%s.\n"+
			"Is this compute instance running using Service Account? That is, Instance.service_account_id should not be empty.",
			resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if err != nil {
			body = []byte(fmt.Sprintf("Failed response body read failed: %s", err.Error()))
		}

		c.logger.Error("Failed to get SA token",
			zap.String("status", resp.Status),
			zap.String("body", string(body)))

		return "", time.Time{}, fmt.Errorf("%s", resp.Status)
	}

	if err != nil {
		return "", time.Time{}, fmt.Errorf("response read failed: %s", err)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		c.logger.Error("Failed to unmarshal token response",
			zap.Error(err))
		return "", time.Time{}, fmt.Errorf("body unmarshal failed: %w", err)
	}

	actualExpirationTime := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	c.logger.Debug("Successfully parsed token response",
		zap.Time("expiresAt", actualExpirationTime),
		zap.String("tokenType", tokenResponse.TokenType))

	return tokenResponse.AccessToken, actualExpirationTime, nil
}

func (c *metadataServiceCredentialProvider) url() string {
	return fmt.Sprintf("http://%s/computeMetadata/v1/instance/service-accounts/default/token", c.metadataServiceAddr)
}

func (c *metadataServiceCredentialProvider) Addr() string {
	return c.metadataServiceAddr
}

func (c *metadataServiceCredentialProvider) Available(ctx context.Context) bool {
	_, err := c.getValidCachedToken()
	if err == nil {
		c.logger.Debug("Cached token is available and valid")
		return true
	}

	c.logger.Debug("Checking metadata service availability")
	dialer := net.Dialer{Timeout: time.Second}

	conn, err := dialer.DialContext(ctx, "tcp", c.metadataServiceAddr)
	if err != nil {
		// Debug, not Error: underlay metadata is expected to be missing on
		// most hosts, and overlay misses are surfaced to the user via a
		// dedicated message in cli-v2 creds.go. Avoid scaring the logs.
		c.logger.Debug("Failed to connect to metadata service", zap.Error(err))
		return false
	}

	_ = conn.Close()

	return c.refreshToken(ctx) == nil
}

func (c *metadataServiceCredentialProvider) getValidCachedToken() (string, error) {
	c.currentTokenMutex.RLock()
	defer c.currentTokenMutex.RUnlock()

	if c.currentToken != "" && !c.currentTokenRealExpiration.IsZero() {
		if c.currentTokenRealExpiration.After(time.Now().Add(-metadataIamTokenValidityThreshold)) {
			return c.currentToken, nil
		} else {
			c.logger.Debug("Cached token is expired",
				zap.Time("expiresAt", c.currentTokenRealExpiration))
			return "", errors.New("current token is expired")
		}
	}

	c.logger.Debug("No cached token available")
	return "", errors.New("current token is unavailable")
}

func (c *metadataServiceCredentialProvider) IAMToken(ctx context.Context) (*CredentialsToken, error) {
	_, err := c.getValidCachedToken()
	if err == nil {
		c.logger.Debug("Using cached token")
		return c.toTokenResponse(), nil
	}

	c.logger.Debug("Cached token invalid or expired, refreshing token")
	// It is expected that the token is updated at the background, but we do this as the last resort measure here
	// (basically, the legacy behavior reproduction)
	err = c.refreshToken(ctx)
	if err != nil {
		c.logger.Error("Failed to refresh token", zap.Error(err))
		return nil, err
	}

	return c.toTokenResponse(), nil
}

func (c *metadataServiceCredentialProvider) toTokenResponse() *CredentialsToken {
	c.currentTokenMutex.RLock()
	defer c.currentTokenMutex.RUnlock()

	expiresAt := getCappedExpiresAt(c.currentTokenRealExpiration)
	c.logger.Debug("Creating token response", zap.Time("expiresAt", expiresAt))

	return &CredentialsToken{
		Token:     c.currentToken,
		ExpiresAt: expiresAt,
	}
}

// getCappedExpiresAt limits the token expiration duration by the metadataTokenCappedExpirationDuration to handle the
// following use-case: instance sa is stripped of all permissions (or even worse, deleted), instance is updated to use
// another sa with valid permissions; however, since the token is not expired it is still used for subsequent requests
// resulting in Permission Denied. Shorter expiration time is a workaround
func getCappedExpiresAt(actualExpirationTime time.Time) time.Time {
	result := time.Now()
	if actualExpirationTime.Sub(result) > metadataTokenCappedExpirationDuration {
		result = result.Add(metadataTokenCappedExpirationDuration)
	}

	return result
}

func (c *metadataServiceCredentialProvider) YandexCloudAPICredentials() {}

// GetMetadataServiceAddr returns the address of Metadata Service, gets the value from InstanceMetadataOverrideEnvVar
// env variable if it is set, otherwise uses the default address from InstanceMetadataAddr joined with port 80.
// InstanceMetadataAddr is host-only by convention (mirrors legacy sdk/instance_metadata.go); the port must be
// appended here so that Available()'s raw TCP dial succeeds — net.Dial requires a host:port form, unlike net/http
// which would default to :80 for the http scheme.
func GetMetadataServiceAddr() string {
	if nonDefaultAddr := os.Getenv(InstanceMetadataOverrideEnvVar); nonDefaultAddr != "" {
		return nonDefaultAddr
	}

	return net.JoinHostPort(InstanceMetadataAddr, "80")
}
