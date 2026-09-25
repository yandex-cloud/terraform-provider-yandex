package credentials

import (
	"context"
	"time"
)

// CredentialsToken represents a token with an associated expiration time for authentication purposes.
type CredentialsToken struct {
	Token     string
	ExpiresAt time.Time
}

// CredentialsIdentity represents the identity type used for credential-based operations or authentication scenarios.
type CredentialsIdentity int

const (
	// CredentialsIdentityUnknown represents an unknown credentials identity.
	CredentialsIdentityUnknown CredentialsIdentity = iota
	// CredentialsIdentityYandexPassportOauthToken represents a Yandex Passport OAuth Token identity.
	CredentialsIdentityYandexPassportOauthToken
	// CredentialsIdentityJWT represents a JWT identity.
	CredentialsIdentityJWT
)

// CredentialsTokenRequest represents a request containing credentials-related identity and token.
type CredentialsTokenRequest struct {
	Identity CredentialsIdentity
	Token    string
}

// Credentials is an abstraction of API authorization credentials.
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/ for details.
// Note that functions that return Credentials may return different Credentials implementation
// in next SDK version, and this is not considered breaking change.
type Credentials interface {
	// YandexCloudAPICredentials is a marker method. All compatible Credentials implementations have it
	YandexCloudAPICredentials()
}

// ExchangeableCredentials can be exchanged for IAM Token in IAM Token Service, that can be used
// to authorize API calls.
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/iam-token for details.
type ExchangeableCredentials interface {
	Credentials
	// IAMTokenRequest returns request for fresh IAM token or error.
	IAMTokenRequest() (*CredentialsTokenRequest, error)
}

// NonExchangeableCredentials allows to get IAM Token without calling IAM Token Service.
type NonExchangeableCredentials interface {
	Credentials
	// IAMToken returns IAM Token.
	IAMToken(ctx context.Context) (*CredentialsToken, error)
}

// exchangeableCredentialsFunc is a type representing a function that returns a CredentialsTokenRequest or an error.
type exchangeableCredentialsFunc func() (iamTokenReq *CredentialsTokenRequest, err error)

var _ ExchangeableCredentials = (exchangeableCredentialsFunc)(nil)

// YandexCloudAPICredentials retrieves API credentials for Yandex Cloud by invoking the exchangeableCredentialsFunc.
func (exchangeableCredentialsFunc) YandexCloudAPICredentials() {}

// IAMTokenRequest obtains a new IAM token request using the exchangeable credentials function.
func (f exchangeableCredentialsFunc) IAMTokenRequest() (iamTokenReq *CredentialsTokenRequest, err error) {
	return f()
}
