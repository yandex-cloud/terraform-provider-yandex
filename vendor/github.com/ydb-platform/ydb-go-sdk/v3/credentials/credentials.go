package credentials

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/credentials"
)

// Credentials is an interface of YDB credentials required for connect with YDB
type Credentials interface {
	// Token must return actual token or error
	Token(ctx context.Context) (string, error)
}

// NewAccessTokenCredentials makes access token credentials object
// Passed options redefines default values of credentials object internal fields
func NewAccessTokenCredentials(
	accessToken string, opts ...credentials.AccessTokenCredentialsOption,
) *credentials.AccessToken {
	return credentials.NewAccessTokenCredentials(accessToken, opts...)
}

// NewAnonymousCredentials makes anonymous credentials object
// Passed options redefines default values of credentials object internal fields
func NewAnonymousCredentials(
	opts ...credentials.AnonymousCredentialsOption,
) *credentials.Anonymous {
	return credentials.NewAnonymousCredentials(opts...)
}

// NewStaticCredentials makes static credentials object
func NewStaticCredentials(
	user, password, authEndpoint string, opts ...credentials.StaticCredentialsOption,
) *credentials.Static {
	return credentials.NewStaticCredentials(user, password, authEndpoint, opts...)
}

// NewOauth2TokenExchangeCredentials makes OAuth 2.0 token exchange protocol credentials object
// https://www.rfc-editor.org/rfc/rfc8693
func NewOauth2TokenExchangeCredentials(
	opts ...credentials.Oauth2TokenExchangeCredentialsOption,
) (Credentials, error) {
	return credentials.NewOauth2TokenExchangeCredentials(opts...)
}

/*
NewOauth2TokenExchangeCredentialsFile makes OAuth 2.0 token exchange protocol credentials object from config file
https://www.rfc-editor.org/rfc/rfc8693
Config file must be a valid json file

Fields of json file

	grant-type:           [string] Grant type option (default: "urn:ietf:params:oauth:grant-type:token-exchange")
	res:                  [string | list of strings] Resource option (optional)
	aud:                  [string | list of strings] Audience option for token exchange request (optional)
	scope:                [string | list of strings] Scope option (optional)
	requested-token-type: [string] Requested token type option (default: "urn:ietf:params:oauth:token-type:access_token")
	subject-credentials:  [creds_json] Subject credentials options (optional)
	actor-credentials:    [creds_json] Actor credentials options (optional)
	token-endpoint:       [string] Token endpoint

Fields of creds_json (JWT):

	type:                 [string] Token source type. Set JWT
	alg:                  [string] Algorithm for JWT signature.
								   Supported algorithms can be listed
								   with GetSupportedOauth2TokenExchangeJwtAlgorithms()
	private-key:          [string] (Private) key in PEM format (RSA, EC) or Base64 format (HMAC) for JWT signature
	kid:                  [string] Key id JWT standard claim (optional)
	iss:                  [string] Issuer JWT standard claim (optional)
	sub:                  [string] Subject JWT standard claim (optional)
	aud:                  [string | list of strings] Audience JWT standard claim (optional)
	jti:                  [string] JWT ID JWT standard claim (optional)
	ttl:                  [string] Token TTL (default: 1h)

Fields of creds_json (FIXED):

	type:                 [string] Token source type. Set FIXED
	token:                [string] Token value
	token-type:           [string] Token type value. It will become
								   subject_token_type/actor_token_type parameter
								   in token exchange request (https://www.rfc-editor.org/rfc/rfc8693)
*/
func NewOauth2TokenExchangeCredentialsFile(
	configFilePath string,
	opts ...credentials.Oauth2TokenExchangeCredentialsOption,
) (Credentials, error) {
	return credentials.NewOauth2TokenExchangeCredentialsFile(configFilePath, opts...)
}

// GetSupportedOauth2TokenExchangeJwtAlgorithms returns supported algorithms for
// initializing OAuth 2.0 token exchange protocol credentials from config file
func GetSupportedOauth2TokenExchangeJwtAlgorithms() []string {
	return credentials.GetSupportedOauth2TokenExchangeJwtAlgorithms()
}

// NewJWTTokenSource makes JWT token source for OAuth 2.0 token exchange credentials
func NewJWTTokenSource(opts ...credentials.JWTTokenSourceOption) (credentials.TokenSource, error) {
	return credentials.NewJWTTokenSource(opts...)
}

// NewFixedTokenSource makes fixed token source for OAuth 2.0 token exchange credentials
func NewFixedTokenSource(token, tokenType string) credentials.TokenSource {
	return credentials.NewFixedTokenSource(token, tokenType)
}
