package credentials

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/secret"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
)

const (
	defaultRequestTimeout      = time.Second * 10
	defaultSyncExchangeTimeout = time.Second * 20
	defaultJWTTokenTTL         = 3600 * time.Second
	updateTimeDivider          = 2
	retryJitterLimit           = 0.5
	syncRetryFastSlot          = time.Millisecond * 100
	syncRetrySlowSlot          = time.Millisecond * 300
	syncRetryCeiling           = 1
	backgroundRetryFastSlot    = time.Millisecond * 10
	backgroundRetrySlowSlot    = time.Millisecond * 300
	backgroundRetryFastCeiling = 12
	backgroundRetrySlowCeiling = 7
)

var (
	syncRetryFastBackoff = backoff.New(
		backoff.WithSlotDuration(syncRetryFastSlot),
		backoff.WithCeiling(syncRetryCeiling),
		backoff.WithJitterLimit(retryJitterLimit),
	)
	syncRetrySlowBackoff = backoff.New(
		backoff.WithSlotDuration(syncRetrySlowSlot),
		backoff.WithCeiling(syncRetryCeiling),
		backoff.WithJitterLimit(retryJitterLimit),
	)
	backgroundRetryFastBackoff = backoff.New(
		backoff.WithSlotDuration(backgroundRetryFastSlot),
		backoff.WithCeiling(backgroundRetryFastCeiling),
		backoff.WithJitterLimit(retryJitterLimit),
	)
	backgroundRetrySlowBackoff = backoff.New(
		backoff.WithSlotDuration(backgroundRetrySlowSlot),
		backoff.WithCeiling(backgroundRetrySlowCeiling),
		backoff.WithJitterLimit(retryJitterLimit),
	)
)

var (
	errCouldNotReadFile           = errors.New("could not read file")
	errCouldNotParseHomeDir       = errors.New("could not parse home dir")
	errEmptyTokenEndpointError    = errors.New("OAuth2 token exchange: empty token endpoint")
	errCouldNotParseResponse      = errors.New("OAuth2 token exchange: could not parse response")
	errCouldNotExchangeToken      = errors.New("OAuth2 token exchange: could not exchange token")
	errUnsupportedTokenType       = errors.New("OAuth2 token exchange: unsupported token type")
	errIncorrectExpirationTime    = errors.New("OAuth2 token exchange: incorrect expiration time")
	errDifferentScope             = errors.New("OAuth2 token exchange: got different scope")
	errEmptyAccessToken           = errors.New("OAuth2 token exchange: got empty access token")
	errCouldNotMakeHTTPRequest    = errors.New("OAuth2 token exchange: could not make http request")
	errCouldNotApplyOption        = errors.New("OAuth2 token exchange: could not apply option")
	errCouldNotCreateTokenSource  = errors.New("OAuth2 token exchange: could not create TokenSource")
	errNoSigningMethodError       = errors.New("JWT token source: no signing method")
	errNoPrivateKeyError          = errors.New("JWT token source: no private key")
	errCouldNotSignJWTToken       = errors.New("JWT token source: could not sign jwt token")
	errCouldNotApplyJWTOption     = errors.New("JWT token source: could not apply option")
	errCouldNotparsePrivateKey    = errors.New("JWT token source: could not parse private key from PEM")
	errCouldNotReadPrivateKeyFile = errors.New("JWT token source: could not read from private key file")
	errCouldNotParseBase64Secret  = errors.New("JWT token source: could not parse base64 secret")
	errCouldNotReadConfigFile     = errors.New("OAuth2 token exchange file: could not read from config file")
	errCouldNotUnmarshalJSON      = errors.New("OAuth2 token exchange file: could not unmarshal json config file")
	errUnknownTokenSourceType     = errors.New("OAuth2 token exchange file: incorrect \"type\" parameter: only \"JWT\" and \"FIXED\" are supported") //nolint:lll
	errTokenAndTokenTypeRequired  = errors.New("OAuth2 token exchange file: \"token\" and \"token-type\" are required")
	errAlgAndKeyRequired          = errors.New("OAuth2 token exchange file: \"alg\" and \"private-key\" are required")
	errUnsupportedSigningMethod   = errors.New("OAuth2 token exchange file: signing method not supported")
	errTTLMustBePositive          = errors.New("OAuth2 token exchange file: \"ttl\" must be positive value")
)

func readFileContent(filePath string) ([]byte, error) {
	if len(filePath) > 0 && filePath[0] == '~' {
		usr, err := user.Current()
		if err != nil {
			return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotParseHomeDir, err))
		}
		filePath = filepath.Join(usr.HomeDir, filePath[1:])
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("%w %s: %w", errCouldNotReadFile, filePath, err))
	}

	return bytes, nil
}

type Oauth2TokenExchangeCredentialsOption interface {
	ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error
}

// TokenEndpoint
type tokenEndpointOption string

func (endpoint tokenEndpointOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.tokenEndpoint = string(endpoint)

	return nil
}

func WithTokenEndpoint(endpoint string) tokenEndpointOption {
	return tokenEndpointOption(endpoint)
}

// GrantType
type grantTypeOption string

func (grantType grantTypeOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.grantType = string(grantType)

	return nil
}

func WithGrantType(grantType string) grantTypeOption {
	return grantTypeOption(grantType)
}

// Resource
type resourceOption []string

func (resource resourceOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.resource = append(c.resource, resource...)

	return nil
}

func WithResource(resource string, resources ...string) resourceOption {
	return append([]string{resource}, resources...)
}

// RequestedTokenType
type requestedTokenTypeOption string

func (requestedTokenType requestedTokenTypeOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.requestedTokenType = string(requestedTokenType)

	return nil
}

func WithRequestedTokenType(requestedTokenType string) requestedTokenTypeOption {
	return requestedTokenTypeOption(requestedTokenType)
}

// Audience
type audienceOption []string

func (audience audienceOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.audience = append(c.audience, audience...)

	return nil
}

func WithAudience(audience string, audiences ...string) audienceOption {
	return append([]string{audience}, audiences...)
}

// Scope
type scopeOption []string

func (scope scopeOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.scope = append(c.scope, scope...)

	return nil
}

func WithScope(scope string, scopes ...string) scopeOption {
	return append([]string{scope}, scopes...)
}

// RequestTimeout
type requestTimeoutOption time.Duration

func (timeout requestTimeoutOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.requestTimeout = time.Duration(timeout)

	return nil
}

func WithRequestTimeout(timeout time.Duration) requestTimeoutOption {
	return requestTimeoutOption(timeout)
}

// SyncExchangeTimeout
type syncExchangeTimeoutOption time.Duration

func (timeout syncExchangeTimeoutOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	c.syncExchangeTimeout = time.Duration(timeout)

	return nil
}

func WithSyncExchangeTimeout(timeout time.Duration) syncExchangeTimeoutOption {
	return syncExchangeTimeoutOption(timeout)
}

const (
	SubjectTokenSourceType = 1
	ActorTokenSourceType   = 2
)

// SubjectTokenSource/ActorTokenSource
type tokenSourceOption struct {
	source          TokenSource
	createFunc      func() (TokenSource, error)
	tokenSourceType int
}

func (tokenSource *tokenSourceOption) ApplyOauth2CredentialsOption(c *oauth2TokenExchange) error {
	src := tokenSource.source
	var err error
	if src == nil {
		src, err = tokenSource.createFunc()
		if err != nil {
			return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotCreateTokenSource, err))
		}
	}
	switch tokenSource.tokenSourceType {
	case SubjectTokenSourceType:
		c.subjectTokenSource = src
	case ActorTokenSourceType:
		c.actorTokenSource = src
	}

	return nil
}

func WithSubjectToken(subjectToken TokenSource) *tokenSourceOption {
	return &tokenSourceOption{
		source:          subjectToken,
		tokenSourceType: SubjectTokenSourceType,
	}
}

func WithFixedSubjectToken(token, tokenType string) *tokenSourceOption {
	return &tokenSourceOption{
		createFunc: func() (TokenSource, error) {
			return NewFixedTokenSource(token, tokenType), nil
		},
		tokenSourceType: SubjectTokenSourceType,
	}
}

func WithJWTSubjectToken(opts ...JWTTokenSourceOption) *tokenSourceOption {
	return &tokenSourceOption{
		createFunc: func() (TokenSource, error) {
			return NewJWTTokenSource(opts...)
		},
		tokenSourceType: SubjectTokenSourceType,
	}
}

// ActorTokenSource
func WithActorToken(actorToken TokenSource) *tokenSourceOption {
	return &tokenSourceOption{
		source:          actorToken,
		tokenSourceType: ActorTokenSourceType,
	}
}

func WithFixedActorToken(token, tokenType string) *tokenSourceOption {
	return &tokenSourceOption{
		createFunc: func() (TokenSource, error) {
			return NewFixedTokenSource(token, tokenType), nil
		},
		tokenSourceType: ActorTokenSourceType,
	}
}

func WithJWTActorToken(opts ...JWTTokenSourceOption) *tokenSourceOption {
	return &tokenSourceOption{
		createFunc: func() (TokenSource, error) {
			return NewJWTTokenSource(opts...)
		},
		tokenSourceType: ActorTokenSourceType,
	}
}

type oauth2TokenExchange struct {
	tokenEndpoint string

	// grant_type parameter
	// urn:ietf:params:oauth:grant-type:token-exchange by default
	grantType string

	resource []string
	audience []string
	scope    []string

	// requested_token_type parameter
	// urn:ietf:params:oauth:token-type:access_token by default
	requestedTokenType string

	subjectTokenSource TokenSource

	actorTokenSource TokenSource

	// Http request timeout
	// 10 by default
	requestTimeout time.Duration

	// Timeout when performing synchronous token exchange
	// It is used when getting token for the first time
	// or when it is already expired
	syncExchangeTimeout time.Duration

	// Received data
	receivedToken           string
	updateTokenTime         time.Time
	receivedTokenExpireTime time.Time

	mutex    sync.RWMutex
	updating atomic.Bool // true if separate goroutine is run and updates token in background

	sourceInfo string
}

func NewOauth2TokenExchangeCredentials(
	opts ...Oauth2TokenExchangeCredentialsOption,
) (*oauth2TokenExchange, error) {
	c := &oauth2TokenExchange{
		grantType:           "urn:ietf:params:oauth:grant-type:token-exchange",
		requestedTokenType:  "urn:ietf:params:oauth:token-type:access_token",
		requestTimeout:      defaultRequestTimeout,
		syncExchangeTimeout: defaultSyncExchangeTimeout,
		sourceInfo:          stack.Record(1),
	}

	var err error
	for _, opt := range opts {
		if opt != nil {
			err = opt.ApplyOauth2CredentialsOption(c)
			if err != nil {
				return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotApplyOption, err))
			}
		}
	}

	if c.tokenEndpoint == "" {
		return nil, xerrors.WithStackTrace(errEmptyTokenEndpointError)
	}

	return c, nil
}

type privateKeyLoadOptionFunc func(string) JWTTokenSourceOption

type signingMethodDescription struct {
	method        jwt.SigningMethod
	keyLoadOption privateKeyLoadOptionFunc
}

func getHMACPrivateKeyOption(privateKey string) JWTTokenSourceOption {
	return WithHMACSecretKeyBase64Content(privateKey)
}

func getRSAPrivateKeyOption(privateKey string) JWTTokenSourceOption {
	return WithRSAPrivateKeyPEMContent([]byte(privateKey))
}

func getECPrivateKeyOption(privateKey string) JWTTokenSourceOption {
	return WithECPrivateKeyPEMContent([]byte(privateKey))
}

var signingMethodsRegistry = map[string]signingMethodDescription{
	"HS256": {
		method:        jwt.SigningMethodHS256,
		keyLoadOption: getHMACPrivateKeyOption,
	},
	"HS384": {
		method:        jwt.SigningMethodHS384,
		keyLoadOption: getHMACPrivateKeyOption,
	},
	"HS512": {
		method:        jwt.SigningMethodHS512,
		keyLoadOption: getHMACPrivateKeyOption,
	},
	"RS256": {
		method:        jwt.SigningMethodRS256,
		keyLoadOption: getRSAPrivateKeyOption,
	},
	"RS384": {
		method:        jwt.SigningMethodRS384,
		keyLoadOption: getRSAPrivateKeyOption,
	},
	"RS512": {
		method:        jwt.SigningMethodRS512,
		keyLoadOption: getRSAPrivateKeyOption,
	},
	"PS256": {
		method:        jwt.SigningMethodPS256,
		keyLoadOption: getRSAPrivateKeyOption,
	},
	"PS384": {
		method:        jwt.SigningMethodPS384,
		keyLoadOption: getRSAPrivateKeyOption,
	},
	"PS512": {
		method:        jwt.SigningMethodPS512,
		keyLoadOption: getRSAPrivateKeyOption,
	},
	"ES256": {
		method:        jwt.SigningMethodES256,
		keyLoadOption: getECPrivateKeyOption,
	},
	"ES384": {
		method:        jwt.SigningMethodES384,
		keyLoadOption: getECPrivateKeyOption,
	},
	"ES512": {
		method:        jwt.SigningMethodES512,
		keyLoadOption: getECPrivateKeyOption,
	},
}

func GetSupportedOauth2TokenExchangeJwtAlgorithms() []string {
	algs := make([]string, len(signingMethodsRegistry))
	i := 0
	for alg := range signingMethodsRegistry {
		algs[i] = alg
		i++
	}
	sort.Strings(algs)

	return algs
}

type StringOrArrayConfig struct {
	Values []string
}

func (a *StringOrArrayConfig) UnmarshalJSON(data []byte) error {
	// Case 1: string
	var s string
	err := json.Unmarshal(data, &s)
	if err == nil {
		a.Values = []string{s}

		return nil
	}

	var arr []string
	err = json.Unmarshal(data, &arr)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}
	a.Values = arr

	return nil
}

type prettyTTL struct {
	Value time.Duration
}

func (d *prettyTTL) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}
	d.Value, err = time.ParseDuration(s)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}
	if d.Value <= 0 {
		return xerrors.WithStackTrace(fmt.Errorf("%w, but got: %q", errTTLMustBePositive, s))
	}

	return xerrors.WithStackTrace(err)
}

//nolint:tagliatelle
type OAuth2TokenSourceConfig struct {
	Type string `json:"type"`

	// Fixed
	Token     string `json:"token"`
	TokenType string `json:"token-type"`

	// JWT
	Algorithm  string               `json:"alg"`
	PrivateKey string               `json:"private-key"`
	KeyID      string               `json:"kid"`
	Issuer     string               `json:"iss"`
	Subject    string               `json:"sub"`
	Audience   *StringOrArrayConfig `json:"aud"`
	ID         string               `json:"jti"`
	TTL        *prettyTTL           `json:"ttl"`
}

func signingMethodNotSupportedError(method string) error {
	var supported string
	for i, alg := range GetSupportedOauth2TokenExchangeJwtAlgorithms() {
		if i != 0 {
			supported += ", "
		}
		supported += "\""
		supported += alg
		supported += "\""
	}

	return fmt.Errorf("%w: %q. Supported signing methods are %s", errUnsupportedSigningMethod, method, supported)
}

func (cfg *OAuth2TokenSourceConfig) applyConfigFixed(tokenSrcType int) (*tokenSourceOption, error) {
	if cfg.Token == "" || cfg.TokenType == "" {
		return nil, xerrors.WithStackTrace(errTokenAndTokenTypeRequired)
	}

	return &tokenSourceOption{
		createFunc: func() (TokenSource, error) {
			return NewFixedTokenSource(cfg.Token, cfg.TokenType), nil
		},
		tokenSourceType: tokenSrcType,
	}, nil
}

func (cfg *OAuth2TokenSourceConfig) applyConfigFixedJWT(tokenSrcType int) (*tokenSourceOption, error) {
	var opts []JWTTokenSourceOption

	if cfg.Algorithm == "" || cfg.PrivateKey == "" {
		return nil, xerrors.WithStackTrace(errAlgAndKeyRequired)
	}

	signingMethodDesc, signingMethodFound := signingMethodsRegistry[strings.ToUpper(cfg.Algorithm)]
	if !signingMethodFound {
		return nil, xerrors.WithStackTrace(signingMethodNotSupportedError(cfg.Algorithm))
	}

	opts = append(opts,
		WithSigningMethod(signingMethodDesc.method),
		signingMethodDesc.keyLoadOption(cfg.PrivateKey),
	)

	if cfg.KeyID != "" {
		opts = append(opts, WithKeyID(cfg.KeyID))
	}

	if cfg.Issuer != "" {
		opts = append(opts, WithIssuer(cfg.Issuer))
	}

	if cfg.Subject != "" {
		opts = append(opts, WithSubject(cfg.Subject))
	}

	if cfg.Audience != nil && len(cfg.Audience.Values) > 0 {
		opts = append(opts, WithAudience(cfg.Audience.Values[0], cfg.Audience.Values[1:]...))
	}

	if cfg.ID != "" {
		opts = append(opts, WithID(cfg.ID))
	}

	if cfg.TTL != nil {
		opts = append(opts, WithTokenTTL(cfg.TTL.Value))
	}

	return &tokenSourceOption{
		createFunc: func() (TokenSource, error) {
			return NewJWTTokenSource(opts...)
		},
		tokenSourceType: tokenSrcType,
	}, nil
}

func (cfg *OAuth2TokenSourceConfig) applyConfig(tokenSrcType int) (*tokenSourceOption, error) {
	if strings.EqualFold(cfg.Type, "FIXED") {
		return cfg.applyConfigFixed(tokenSrcType)
	}

	if strings.EqualFold(cfg.Type, "JWT") {
		return cfg.applyConfigFixedJWT(tokenSrcType)
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %q", errUnknownTokenSourceType, cfg.Type))
}

//nolint:tagliatelle
type OAuth2Config struct {
	GrantType          string               `json:"grant-type"`
	Resource           *StringOrArrayConfig `json:"res"`
	Audience           *StringOrArrayConfig `json:"aud"`
	Scope              *StringOrArrayConfig `json:"scope"`
	RequestedTokenType string               `json:"requested-token-type"`
	TokenEndpoint      string               `json:"token-endpoint"`

	SubjectCreds *OAuth2TokenSourceConfig `json:"subject-credentials"`
	ActorCreds   *OAuth2TokenSourceConfig `json:"actor-credentials"`
}

func (cfg *OAuth2Config) AsOptions() ([]Oauth2TokenExchangeCredentialsOption, error) {
	var fullOptions []Oauth2TokenExchangeCredentialsOption
	if err := cfg.applyConfig(&fullOptions); err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return fullOptions, nil
}

func (cfg *OAuth2Config) applyConfig(opts *[]Oauth2TokenExchangeCredentialsOption) error {
	if cfg.GrantType != "" {
		*opts = append(*opts, WithGrantType(cfg.GrantType))
	}

	if cfg.Resource != nil && len(cfg.Resource.Values) > 0 {
		*opts = append(*opts, WithResource(cfg.Resource.Values[0], cfg.Resource.Values[1:]...))
	}

	if cfg.Audience != nil && len(cfg.Audience.Values) > 0 {
		*opts = append(*opts, WithAudience(cfg.Audience.Values[0], cfg.Audience.Values[1:]...))
	}

	if cfg.Scope != nil && len(cfg.Scope.Values) > 0 {
		*opts = append(*opts, WithScope(cfg.Scope.Values[0], cfg.Scope.Values[1:]...))
	}

	if cfg.RequestedTokenType != "" {
		*opts = append(*opts, WithRequestedTokenType(cfg.RequestedTokenType))
	}

	if cfg.TokenEndpoint != "" {
		*opts = append(*opts, WithTokenEndpoint(cfg.TokenEndpoint))
	}

	if cfg.SubjectCreds != nil {
		opt, err := cfg.SubjectCreds.applyConfig(SubjectTokenSourceType)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		*opts = append(*opts, opt)
	}

	if cfg.ActorCreds != nil {
		opt, err := cfg.ActorCreds.applyConfig(ActorTokenSourceType)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		*opts = append(*opts, opt)
	}

	return nil
}

func NewOauth2TokenExchangeCredentialsFile(
	configFilePath string,
	opts ...Oauth2TokenExchangeCredentialsOption,
) (*oauth2TokenExchange, error) {
	configFileData, err := readFileContent(configFilePath)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotReadConfigFile, err))
	}

	var cfg OAuth2Config
	if err = json.Unmarshal(configFileData, &cfg); err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotUnmarshalJSON, err))
	}

	var fullOptions []Oauth2TokenExchangeCredentialsOption
	err = cfg.applyConfig(&fullOptions)
	if err != nil {
		return nil, err
	}

	// Add additional options
	for _, opt := range opts {
		if opt != nil {
			fullOptions = append(fullOptions, opt)
		}
	}

	return NewOauth2TokenExchangeCredentials(fullOptions...)
}

func (provider *oauth2TokenExchange) getScopeParam() string {
	var scope string
	if len(provider.scope) != 0 {
		for _, s := range provider.scope {
			if s != "" {
				if scope != "" {
					scope += " "
				}
				scope += s
			}
		}
	}

	return scope
}

func (provider *oauth2TokenExchange) addTokenSrc(params *url.Values, src TokenSource, tName, tTypeName string) error {
	if src != nil {
		token, err := src.Token()
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		params.Set(tName, token.Token)
		params.Set(tTypeName, token.TokenType)
	}

	return nil
}

func (provider *oauth2TokenExchange) getRequestParams() (string, error) {
	params := url.Values{}
	params.Set("grant_type", provider.grantType)
	for _, res := range provider.resource {
		if res != "" {
			params.Add("resource", res)
		}
	}
	for _, aud := range provider.audience {
		if aud != "" {
			params.Add("audience", aud)
		}
	}
	scope := provider.getScopeParam()
	if scope != "" {
		params.Set("scope", scope)
	}

	params.Set("requested_token_type", provider.requestedTokenType)

	err := provider.addTokenSrc(&params, provider.subjectTokenSource, "subject_token", "subject_token_type")
	if err != nil {
		return "", xerrors.WithStackTrace(err)
	}

	err = provider.addTokenSrc(&params, provider.actorTokenSource, "actor_token", "actor_token_type")
	if err != nil {
		return "", xerrors.WithStackTrace(err)
	}

	return params.Encode(), nil
}

func (provider *oauth2TokenExchange) processTokenExchangeResponse(
	result *http.Response,
	now time.Time,
	retryAllErrors bool,
) (*tokenResponse, error) {
	data, err := readResponseBody(result)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != http.StatusOK {
		return nil, provider.handleErrorResponse(result, data, retryAllErrors)
	}

	parsedResponse, err := parseTokenResponse(data, retryAllErrors)
	if err != nil {
		return nil, err
	}

	if err := validateTokenResponse(parsedResponse, provider); err != nil {
		return nil, err
	}

	parsedResponse.Now = now

	return parsedResponse, nil
}

func readResponseBody(result *http.Response) ([]byte, error) {
	if result.Body != nil {
		data, err := io.ReadAll(result.Body)
		if err != nil {
			return nil, xerrors.WithStackTrace(xerrors.Retryable(err,
				xerrors.WithBackoff(retry.TypeFastBackoff),
			))
		}

		return data, nil
	}

	return make([]byte, 0), nil
}

func makeError(result *http.Response, err error, retryAllErrors bool) error {
	if result != nil {
		if result.StatusCode == http.StatusRequestTimeout ||
			result.StatusCode == http.StatusGatewayTimeout ||
			result.StatusCode == http.StatusTooManyRequests ||
			result.StatusCode == http.StatusInternalServerError ||
			result.StatusCode == http.StatusBadGateway ||
			result.StatusCode == http.StatusServiceUnavailable {
			return xerrors.WithStackTrace(xerrors.Retryable(err,
				xerrors.WithBackoff(retry.TypeSlowBackoff),
			))
		}
	}
	if retryAllErrors {
		return xerrors.WithStackTrace(xerrors.Retryable(err,
			xerrors.WithBackoff(retry.TypeFastBackoff),
		))
	}

	return xerrors.WithStackTrace(err)
}

func (provider *oauth2TokenExchange) handleErrorResponse(
	result *http.Response,
	data []byte,
	retryAllErrors bool,
) error {
	description := result.Status

	//nolint:tagliatelle
	type errorResponse struct {
		ErrorName        string `json:"error"`
		ErrorDescription string `json:"error_description"`
		ErrorURI         string `json:"error_uri"`
	}
	var parsedErrorResponse errorResponse
	if err := json.Unmarshal(data, &parsedErrorResponse); err != nil {
		description += ", could not parse response: " + err.Error()

		return makeError(
			result,
			xerrors.WithStackTrace(fmt.Errorf("%w: %s", errCouldNotExchangeToken, description)),
			retryAllErrors)
	}

	if parsedErrorResponse.ErrorName != "" {
		description += ", error: " + parsedErrorResponse.ErrorName
	}
	if parsedErrorResponse.ErrorDescription != "" {
		description += fmt.Sprintf(", description: %q", parsedErrorResponse.ErrorDescription)
	}
	if parsedErrorResponse.ErrorURI != "" {
		description += ", error_uri: " + parsedErrorResponse.ErrorURI
	}

	return makeError(
		result,
		xerrors.WithStackTrace(fmt.Errorf("%w: %s", errCouldNotExchangeToken, description)),
		retryAllErrors)
}

//nolint:tagliatelle
type tokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	Scope       string    `json:"scope"`
	Now         time.Time `json:"-"`
}

func parseTokenResponse(data []byte, retryAllErrors bool) (*tokenResponse, error) {
	var parsedResponse tokenResponse
	if err := json.Unmarshal(data, &parsedResponse); err != nil {
		return nil, makeError(
			nil,
			xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotParseResponse, err)),
			retryAllErrors)
	}

	return &parsedResponse, nil
}

func validateTokenResponse(parsedResponse *tokenResponse, provider *oauth2TokenExchange) error {
	if !strings.EqualFold(parsedResponse.TokenType, "bearer") {
		return xerrors.WithStackTrace(
			fmt.Errorf("%w: %q", errUnsupportedTokenType, parsedResponse.TokenType))
	}

	if parsedResponse.ExpiresIn <= 0 {
		return xerrors.WithStackTrace(
			fmt.Errorf("%w: %d", errIncorrectExpirationTime, parsedResponse.ExpiresIn))
	}

	if parsedResponse.Scope != "" {
		scope := provider.getScopeParam()
		if parsedResponse.Scope != scope {
			return xerrors.WithStackTrace(
				fmt.Errorf("%w. Expected %q, but got %q", errDifferentScope, scope, parsedResponse.Scope))
		}
	}

	if parsedResponse.AccessToken == "" {
		return xerrors.WithStackTrace(errEmptyAccessToken)
	}

	return nil
}

func (provider *oauth2TokenExchange) updateToken(parsedResponse *tokenResponse) {
	provider.receivedToken = "Bearer " + parsedResponse.AccessToken

	expireDelta := time.Duration(parsedResponse.ExpiresIn) * time.Second
	provider.receivedTokenExpireTime = parsedResponse.Now.Add(expireDelta)

	updateDelta := time.Duration(parsedResponse.ExpiresIn/updateTimeDivider) * time.Second
	provider.updateTokenTime = parsedResponse.Now.Add(updateDelta)
}

// performExchangeTokenRequest is a read only func that performs request. Can be used without lock
func (provider *oauth2TokenExchange) performExchangeTokenRequest(
	ctx context.Context,
	retryAllErrors bool,
) (*tokenResponse, error) {
	body, err := provider.getRequestParams()
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotMakeHTTPRequest, err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.tokenEndpoint, strings.NewReader(body))
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotMakeHTTPRequest, err))
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	req.Close = true

	client := http.Client{
		Transport: http.DefaultTransport,
		Timeout:   provider.requestTimeout,
	}

	now := time.Now()
	result, err := client.Do(req)
	if err != nil {
		return nil, xerrors.WithStackTrace(xerrors.Retryable(
			fmt.Errorf("%w: %w", errCouldNotExchangeToken, err),
			xerrors.WithBackoff(retry.TypeFastBackoff),
		))
	}

	defer result.Body.Close()

	return provider.processTokenExchangeResponse(result, now, retryAllErrors)
}

// exchangeTokenSync exchanges token synchronously, must be called under lock
func (provider *oauth2TokenExchange) exchangeTokenSync(ctx context.Context) error {
	retryAllErrors := provider.receivedToken != "" // already received token => all params are correct, can retry

	ctx, cancelFunc := context.WithTimeout(ctx, provider.syncExchangeTimeout)
	defer cancelFunc()

	response, err := retry.RetryWithResult[*tokenResponse](
		ctx,
		func(ctx context.Context) (*tokenResponse, error) {
			return provider.performExchangeTokenRequest(ctx, retryAllErrors)
		},
		retry.WithFastBackoff(syncRetryFastBackoff),
		retry.WithSlowBackoff(syncRetrySlowBackoff),
	)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	provider.updateToken(response)

	return nil
}

func (provider *oauth2TokenExchange) exchangeTokenInBackground() {
	defer provider.updating.Store(false)

	provider.mutex.RLock()
	ctx, cancelFunc := context.WithDeadline(context.Background(), provider.receivedTokenExpireTime)
	provider.mutex.RUnlock()
	defer cancelFunc()

	response, err := retry.RetryWithResult[*tokenResponse](
		ctx,
		func(ctx context.Context) (*tokenResponse, error) {
			return provider.performExchangeTokenRequest(ctx, true)
		},
		retry.WithFastBackoff(backgroundRetryFastBackoff),
		retry.WithSlowBackoff(backgroundRetrySlowBackoff),
	)
	if err != nil {
		return
	}

	provider.mutex.Lock()
	defer provider.mutex.Unlock()
	provider.updateToken(response)
}

func (provider *oauth2TokenExchange) checkBackgroundUpdate(now time.Time) {
	if provider.needUpdate(now) && !provider.updating.Load() {
		if provider.updating.CompareAndSwap(false, true) {
			go provider.exchangeTokenInBackground()
		}
	}
}

func (provider *oauth2TokenExchange) expired(now time.Time) bool {
	return now.Compare(provider.receivedTokenExpireTime) > 0
}

func (provider *oauth2TokenExchange) needUpdate(now time.Time) bool {
	return now.Compare(provider.updateTokenTime) > 0
}

func (provider *oauth2TokenExchange) fastCheck(now time.Time) string {
	provider.mutex.RLock()
	defer provider.mutex.RUnlock()

	if !provider.expired(now) {
		provider.checkBackgroundUpdate(now)

		return provider.receivedToken
	}

	return ""
}

func (provider *oauth2TokenExchange) Token(ctx context.Context) (string, error) {
	now := time.Now()

	token := provider.fastCheck(now)
	if token != "" {
		return token, nil
	}

	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if !provider.expired(now) { // for the case of concurrent call
		return provider.receivedToken, nil
	}

	if err := provider.exchangeTokenSync(ctx); err != nil {
		return "", err
	}

	return provider.receivedToken, nil
}

func (provider *oauth2TokenExchange) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	fmt.Fprintf(
		buffer,
		"OAuth2TokenExchange{Endpoint:%q,GrantType:%s,Resource:%v,Audience:%v,Scope:%v,RequestedTokenType:%s",
		provider.tokenEndpoint,
		provider.grantType,
		provider.resource,
		provider.audience,
		provider.scope,
		provider.requestedTokenType,
	)
	if provider.subjectTokenSource != nil {
		fmt.Fprintf(buffer, ",SubjectToken:%s", provider.subjectTokenSource)
	}
	if provider.actorTokenSource != nil {
		fmt.Fprintf(buffer, ",ActorToken:%s", provider.actorTokenSource)
	}
	if provider.sourceInfo != "" {
		fmt.Fprintf(buffer, ",From:%q", provider.sourceInfo)
	}
	buffer.WriteByte('}')

	return buffer.String()
}

type Token struct {
	Token string

	// token type according to OAuth 2.0 token exchange protocol
	// https://www.rfc-editor.org/rfc/rfc8693#TokenTypeIdentifiers
	// for example urn:ietf:params:oauth:token-type:jwt
	TokenType string
}

type TokenSource interface {
	Token() (Token, error)
}

type fixedTokenSource struct {
	fixedToken Token
}

func (s *fixedTokenSource) Token() (Token, error) {
	return s.fixedToken, nil
}

func (s *fixedTokenSource) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	fmt.Fprintf(
		buffer,
		"FixedTokenSource{Token:%q,Type:%s}",
		secret.Token(s.fixedToken.Token),
		s.fixedToken.TokenType,
	)

	return buffer.String()
}

func NewFixedTokenSource(token, tokenType string) *fixedTokenSource {
	return &fixedTokenSource{
		fixedToken: Token{
			Token:     token,
			TokenType: tokenType,
		},
	}
}

type JWTTokenSourceOption interface {
	ApplyJWTTokenSourceOption(s *jwtTokenSource) error
}

// Issuer
type issuerOption string

func (issuer issuerOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.issuer = string(issuer)

	return nil
}

func WithIssuer(issuer string) issuerOption {
	return issuerOption(issuer)
}

// Subject
type subjectOption string

func (subject subjectOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.subject = string(subject)

	return nil
}

func WithSubject(subject string) subjectOption {
	return subjectOption(subject)
}

// Audience
func (audience audienceOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.audience = append(s.audience, audience...)

	return nil
}

// ID
type idOption string

func (id idOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.id = string(id)

	return nil
}

func WithID(id string) idOption {
	return idOption(id)
}

// TokenTTL
type tokenTTLOption time.Duration

func (ttl tokenTTLOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.tokenTTL = time.Duration(ttl)

	return nil
}

func WithTokenTTL(ttl time.Duration) tokenTTLOption {
	return tokenTTLOption(ttl)
}

// SigningMethod
type signingMethodOption struct {
	method jwt.SigningMethod
}

func (method *signingMethodOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.signingMethod = method.method

	return nil
}

func WithSigningMethod(method jwt.SigningMethod) *signingMethodOption {
	return &signingMethodOption{method}
}

// SigningMethodName
type signingMethodNameOption struct {
	method string
}

func (method *signingMethodNameOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	signingMethodDesc, signingMethodFound := signingMethodsRegistry[strings.ToUpper(method.method)]
	if !signingMethodFound {
		return xerrors.WithStackTrace(signingMethodNotSupportedError(method.method))
	}

	s.signingMethod = signingMethodDesc.method

	return nil
}

func WithSigningMethodName(method string) *signingMethodNameOption {
	return &signingMethodNameOption{method}
}

// KeyID
type keyIDOption string

func (id keyIDOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.keyID = string(id)

	return nil
}

func WithKeyID(id string) keyIDOption {
	return keyIDOption(id)
}

// PrivateKey
type privateKeyOption struct {
	key interface{}
}

func (key *privateKeyOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.privateKey = key.key

	return nil
}

func WithPrivateKey(key interface{}) *privateKeyOption {
	return &privateKeyOption{key}
}

// PrivateKey
type rsaPrivateKeyPemContentOption struct {
	keyContent []byte
}

func (key *rsaPrivateKeyPemContentOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(key.keyContent)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotparsePrivateKey, err))
	}
	s.privateKey = privateKey

	return nil
}

func WithRSAPrivateKeyPEMContent(key []byte) *rsaPrivateKeyPemContentOption {
	return &rsaPrivateKeyPemContentOption{key}
}

// PrivateKey
type rsaPrivateKeyPemFileOption struct {
	path string
}

func (key *rsaPrivateKeyPemFileOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	bytes, err := readFileContent(key.path)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotReadPrivateKeyFile, err))
	}

	o := rsaPrivateKeyPemContentOption{bytes}

	return o.ApplyJWTTokenSourceOption(s)
}

func WithRSAPrivateKeyPEMFile(path string) *rsaPrivateKeyPemFileOption {
	return &rsaPrivateKeyPemFileOption{path}
}

// PrivateKey
type ecPrivateKeyPemContentOption struct {
	keyContent []byte
}

func (key *ecPrivateKeyPemContentOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	privateKey, err := jwt.ParseECPrivateKeyFromPEM(key.keyContent)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotparsePrivateKey, err))
	}
	s.privateKey = privateKey

	return nil
}

func WithECPrivateKeyPEMContent(key []byte) *ecPrivateKeyPemContentOption {
	return &ecPrivateKeyPemContentOption{key}
}

// PrivateKey
type ecPrivateKeyPemFileOption struct {
	path string
}

func (key *ecPrivateKeyPemFileOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	bytes, err := readFileContent(key.path)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotReadPrivateKeyFile, err))
	}

	o := ecPrivateKeyPemContentOption{bytes}

	return o.ApplyJWTTokenSourceOption(s)
}

func WithECPrivateKeyPEMFile(path string) *ecPrivateKeyPemFileOption {
	return &ecPrivateKeyPemFileOption{path}
}

// Key
type hmacSecretKeyContentOption struct {
	keyContent []byte
}

func (key *hmacSecretKeyContentOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	s.privateKey = key.keyContent

	return nil
}

func WithHMACSecretKey(key []byte) *hmacSecretKeyContentOption {
	return &hmacSecretKeyContentOption{key}
}

// Key
type hmacSecretKeyBase64ContentOption struct {
	base64KeyContent string
}

func (key *hmacSecretKeyBase64ContentOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	keyData, err := base64.StdEncoding.DecodeString(key.base64KeyContent)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotParseBase64Secret, err))
	}
	s.privateKey = keyData

	return nil
}

func WithHMACSecretKeyBase64Content(base64KeyContent string) *hmacSecretKeyBase64ContentOption {
	return &hmacSecretKeyBase64ContentOption{base64KeyContent}
}

// Key
type hmacSecretKeyFileOption struct {
	path string
}

func (key *hmacSecretKeyFileOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	bytes, err := readFileContent(key.path)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotReadPrivateKeyFile, err))
	}

	s.privateKey = bytes

	return nil
}

func WithHMACSecretKeyFile(path string) *hmacSecretKeyFileOption {
	return &hmacSecretKeyFileOption{path}
}

// Key
type hmacSecretKeyBase64FileOption struct {
	path string
}

func (key *hmacSecretKeyBase64FileOption) ApplyJWTTokenSourceOption(s *jwtTokenSource) error {
	bytes, err := readFileContent(key.path)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotReadPrivateKeyFile, err))
	}

	o := hmacSecretKeyBase64ContentOption{string(bytes)}

	return o.ApplyJWTTokenSourceOption(s)
}

func WithHMACSecretKeyBase64File(path string) *hmacSecretKeyBase64FileOption {
	return &hmacSecretKeyBase64FileOption{path}
}

func NewJWTTokenSource(opts ...JWTTokenSourceOption) (*jwtTokenSource, error) {
	s := &jwtTokenSource{
		tokenTTL: defaultJWTTokenTTL,
	}

	var err error
	for _, opt := range opts {
		if opt != nil {
			err = opt.ApplyJWTTokenSourceOption(s)
			if err != nil {
				return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotApplyJWTOption, err))
			}
		}
	}

	if s.signingMethod == nil {
		return nil, xerrors.WithStackTrace(errNoSigningMethodError)
	}

	if s.privateKey == nil {
		return nil, xerrors.WithStackTrace(errNoPrivateKeyError)
	}

	return s, nil
}

type jwtTokenSource struct {
	signingMethod jwt.SigningMethod
	keyID         string
	privateKey    interface{} // symmetric key in case of symmetric algorithm

	// JWT claims
	issuer   string
	subject  string
	audience []string
	id       string
	tokenTTL time.Duration
}

func (s *jwtTokenSource) Token() (Token, error) {
	var (
		now    = time.Now()
		issued = jwt.NewNumericDate(now.UTC())
		expire = jwt.NewNumericDate(now.Add(s.tokenTTL).UTC())
		err    error
	)
	t := jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": s.signingMethod.Alg(),
			"kid": s.keyID,
		},
		Claims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   s.subject,
			IssuedAt:  issued,
			Audience:  s.audience,
			ExpiresAt: expire,
			ID:        s.id,
		},
		Method: s.signingMethod,
	}

	var token Token
	token.Token, err = t.SignedString(s.privateKey)
	if err != nil {
		return token, xerrors.WithStackTrace(fmt.Errorf("%w: %w", errCouldNotSignJWTToken, err))
	}
	token.TokenType = "urn:ietf:params:oauth:token-type:jwt"

	return token, nil
}

func (s *jwtTokenSource) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	fmt.Fprintf(
		buffer,
		"JWTTokenSource{Method:%s,KeyID:%s,Issuer:%q,Subject:%q,Audience:%v,ID:%s,TokenTTL:%s}",
		s.signingMethod.Alg(),
		s.keyID,
		s.issuer,
		s.subject,
		s.audience,
		s.id,
		s.tokenTTL,
	)

	return buffer.String()
}
