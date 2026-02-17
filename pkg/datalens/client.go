package datalens

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	DefaultEndpoint = "https://api.datalens.tech"

	apiVersionHeaderName  = "x-dl-api-version"
	apiVersionHeaderValue = "1"
	orgIDHeaderName       = "x-dl-org-id"
	authHeaderName        = "x-yacloud-subjecttoken"
)

type TokenProvider func(ctx context.Context) (string, error)

type cachedToken struct {
	token     string
	expiresAt time.Time
}

func (t *cachedToken) isValid() bool {
	return t != nil && t.token != "" && time.Now().Add(time.Minute).Before(t.expiresAt)
}

type CachedTokenProvider struct {
	mu       sync.Mutex
	provider TokenProvider
	cached   *cachedToken
	ttl      time.Duration
}

const defaultTokenTTL = 11 * time.Hour

func NewCachedTokenProvider(provider TokenProvider) *CachedTokenProvider {
	return &CachedTokenProvider{provider: provider, ttl: defaultTokenTTL}
}

func (c *CachedTokenProvider) Token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cached.isValid() {
		return c.cached.token, nil
	}

	token, err := c.provider(ctx)
	if err != nil {
		return "", err
	}

	c.cached = &cachedToken{
		token:     token,
		expiresAt: time.Now().Add(c.ttl),
	}
	return token, nil
}

type Config struct {
	Endpoint      string
	TokenProvider TokenProvider
	HTTPClient    *http.Client
}

type Client struct {
	endpoint      string
	tokenProvider TokenProvider
	httpClient    *http.Client
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.TokenProvider == nil {
		return nil, fmt.Errorf("datalens: TokenProvider is required")
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	cached := NewCachedTokenProvider(cfg.TokenProvider)

	return &Client{
		endpoint:      endpoint,
		tokenProvider: cached.Token,
		httpClient:    httpClient,
	}, nil
}

func (c *Client) Do(ctx context.Context, path string, orgID string, reqBody any, result any) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	token, err := c.tokenProvider(ctx)
	if err != nil {
		return fmt.Errorf("get IAM token: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(authHeaderName, token)
	httpReq.Header.Set(apiVersionHeaderName, apiVersionHeaderValue)
	if orgID != "" {
		httpReq.Header.Set(orgIDHeaderName, orgID)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &APIError{
			StatusCode: httpResp.StatusCode,
			Body:       string(respBody),
			RPC:        path,
		}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

type APIError struct {
	StatusCode int
	Body       string
	RPC        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("datalens API %s returned status %d: %s", e.RPC, e.StatusCode, e.Body)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
