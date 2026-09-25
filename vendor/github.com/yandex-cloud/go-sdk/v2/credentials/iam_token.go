package credentials

import (
	"context"
	"time"
)

var _ NonExchangeableCredentials = (*IAMTokenCredentials)(nil)

// IAMTokenCredentials implements Credentials with IAM token as-is
// Read more on https://yandex.cloud/en-ru/docs/iam/concepts/authorization/iam-token
type IAMTokenCredentials struct {
	iamToken  string
	expiresAt time.Time
}

func (creds *IAMTokenCredentials) YandexCloudAPICredentials() {}

func (creds *IAMTokenCredentials) IAMToken(ctx context.Context) (*CredentialsToken, error) {
	return &CredentialsToken{Token: creds.iamToken, ExpiresAt: creds.expiresAt}, nil
}

func IAMToken(iamToken string) NonExchangeableCredentials {
	return &IAMTokenCredentials{
		iamToken: iamToken,
	}
}

// IAMTokenWithExpiry is like IAMToken but preserves the token's expiration time
// so that callers asking the authenticator for the current IAM token receive
// the original ExpiresAt instead of a zero timestamp.
func IAMTokenWithExpiry(iamToken string, expiresAt time.Time) NonExchangeableCredentials {
	return &IAMTokenCredentials{
		iamToken:  iamToken,
		expiresAt: expiresAt,
	}
}
