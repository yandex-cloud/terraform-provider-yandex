package credentials

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/credentials"
)

type OAuth2Config = credentials.OAuth2Config

type OAuth2StringOrArrayConfig = credentials.StringOrArrayConfig

type OAuth2TokenSourceConfig = credentials.OAuth2TokenSourceConfig

type Oauth2TokenExchangeCredentialsOption = credentials.Oauth2TokenExchangeCredentialsOption

type TokenSource = credentials.TokenSource

type Token = credentials.Token

// WithSourceInfo option append to credentials object the source info for reporting source info details on error case
func WithSourceInfo(sourceInfo string) credentials.SourceInfoOption {
	return credentials.WithSourceInfo(sourceInfo)
}

// WithGrpcDialOptions option append to static credentials object GRPC dial options
func WithGrpcDialOptions(opts ...grpc.DialOption) credentials.StaticCredentialsOption {
	return credentials.WithGrpcDialOptions(opts...)
}

// TokenEndpoint
func WithTokenEndpoint(endpoint string) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithTokenEndpoint(endpoint)
}

// GrantType
func WithGrantType(grantType string) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithGrantType(grantType)
}

// Resource
func WithResource(resource string, resources ...string) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithResource(resource, resources...)
}

// RequestedTokenType
func WithRequestedTokenType(requestedTokenType string) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithRequestedTokenType(requestedTokenType)
}

// Scope
func WithScope(scope string, scopes ...string) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithScope(scope, scopes...)
}

// RequestTimeout
func WithRequestTimeout(timeout time.Duration) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithRequestTimeout(timeout)
}

// SyncExchangeTimeout
func WithSyncExchangeTimeout(timeout time.Duration) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithSyncExchangeTimeout(timeout)
}

// SubjectTokenSource
func WithSubjectToken(subjectToken credentials.TokenSource) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithSubjectToken(subjectToken)
}

// SubjectTokenSource
func WithFixedSubjectToken(token, tokenType string) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithFixedSubjectToken(token, tokenType)
}

// SubjectTokenSource
func WithJWTSubjectToken(opts ...credentials.JWTTokenSourceOption) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithJWTSubjectToken(opts...)
}

// ActorTokenSource
func WithActorToken(actorToken credentials.TokenSource) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithActorToken(actorToken)
}

// ActorTokenSource
func WithFixedActorToken(token, tokenType string) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithFixedActorToken(token, tokenType)
}

// ActorTokenSource
func WithJWTActorToken(opts ...credentials.JWTTokenSourceOption) Oauth2TokenExchangeCredentialsOption {
	return credentials.WithJWTActorToken(opts...)
}

// Audience
type oauthCredentialsAndJWTCredentialsOption interface {
	credentials.Oauth2TokenExchangeCredentialsOption
	credentials.JWTTokenSourceOption
}

func WithAudience(audience string, audiences ...string) oauthCredentialsAndJWTCredentialsOption {
	return credentials.WithAudience(audience, audiences...)
}

// Issuer
func WithIssuer(issuer string) credentials.JWTTokenSourceOption {
	return credentials.WithIssuer(issuer)
}

// Subject
func WithSubject(subject string) credentials.JWTTokenSourceOption {
	return credentials.WithSubject(subject)
}

// ID
func WithID(id string) credentials.JWTTokenSourceOption {
	return credentials.WithID(id)
}

// TokenTTL
func WithTokenTTL(ttl time.Duration) credentials.JWTTokenSourceOption {
	return credentials.WithTokenTTL(ttl)
}

// KeyID
func WithKeyID(id string) credentials.JWTTokenSourceOption {
	return credentials.WithKeyID(id)
}

// SigningMethod
func WithSigningMethod(method jwt.SigningMethod) credentials.JWTTokenSourceOption {
	return credentials.WithSigningMethod(method)
}

// SigningMethod
func WithSigningMethodName(method string) credentials.JWTTokenSourceOption {
	return credentials.WithSigningMethodName(method)
}

// PrivateKey
func WithPrivateKey(key interface{}) credentials.JWTTokenSourceOption {
	return credentials.WithPrivateKey(key)
}

// PrivateKey
// For RSA signing methods: RS256, RS384, RS512, PS256, PS384, PS512
func WithRSAPrivateKeyPEMContent(key []byte) credentials.JWTTokenSourceOption {
	return credentials.WithRSAPrivateKeyPEMContent(key)
}

// PrivateKey
// For RSA signing methods: RS256, RS384, RS512, PS256, PS384, PS512
func WithRSAPrivateKeyPEMFile(path string) credentials.JWTTokenSourceOption {
	return credentials.WithRSAPrivateKeyPEMFile(path)
}

// PrivateKey
// For EC signing methods: ES256, ES384, ES512
func WithECPrivateKeyPEMContent(key []byte) credentials.JWTTokenSourceOption {
	return credentials.WithECPrivateKeyPEMContent(key)
}

// PrivateKey
// For EC signing methods: ES256, ES384, ES512
func WithECPrivateKeyPEMFile(path string) credentials.JWTTokenSourceOption {
	return credentials.WithECPrivateKeyPEMFile(path)
}

// Key
// For HMAC signing methods: HS256, HS384, HS512
func WithHMACSecretKey(key []byte) credentials.JWTTokenSourceOption {
	return credentials.WithHMACSecretKey(key)
}

// Key
// For HMAC signing methods: HS256, HS384, HS512
func WithHMACSecretKeyBase64Content(base64KeyContent string) credentials.JWTTokenSourceOption {
	return credentials.WithHMACSecretKeyBase64Content(base64KeyContent)
}

// Key
// For HMAC signing methods: HS256, HS384, HS512
func WithHMACSecretKeyFile(path string) credentials.JWTTokenSourceOption {
	return credentials.WithHMACSecretKeyFile(path)
}

// Key
// For HMAC signing methods: HS256, HS384, HS512
func WithHMACSecretKeyBase64File(path string) credentials.JWTTokenSourceOption {
	return credentials.WithHMACSecretKeyBase64File(path)
}
