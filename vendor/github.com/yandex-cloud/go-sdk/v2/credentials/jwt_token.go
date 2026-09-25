package credentials

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/yandex-cloud/go-sdk/v2/pkg/iamkey"
)

// newServiceAccountJWTBuilder creates a JWT builder for a given IAM service account key.
// Returns an error if the key validation or parsing of the private key fails.
func newServiceAccountJWTBuilder(key *iamkey.Key) (*serviceAccountJWTBuilder, error) {
	err := validateServiceAccountKey(key)
	if err != nil {
		return nil, fmt.Errorf("key validation failed: %w", err)
	}

	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("private key parsing failed: %w", err)
	}

	return &serviceAccountJWTBuilder{
		key:           key,
		rsaPrivateKey: rsaPrivateKey,
	}, nil
}

// validateServiceAccountKey validates an IAM key to ensure it has an ID and is issued for a service account.
func validateServiceAccountKey(key *iamkey.Key) error {
	if key.Id == "" {
		return errors.New("key id is missing")
	}

	if key.GetServiceAccountId() == "" {
		return fmt.Errorf("key should be issued for service account, but subject is %#v", key.Subject)
	}

	return nil
}

// serviceAccountJWTBuilder is used to generate signed JWT tokens for service account authorization.
// It relies on a service account key and an associated RSA private key.
type serviceAccountJWTBuilder struct {
	key           *iamkey.Key
	rsaPrivateKey *rsa.PrivateKey
}

// SignedToken generates a signed JWT token using the service account's private RSA key and returns it as a string.
func (b *serviceAccountJWTBuilder) SignedToken() (string, error) {
	return b.issueToken().SignedString(b.rsaPrivateKey)
}

// issueToken creates a new JWT token with claims using the service account's ID and sets the token's header information.
func (b *serviceAccountJWTBuilder) issueToken() *jwt.Token {
	issuedAt := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, jwt.RegisteredClaims{
		Issuer:    b.key.GetServiceAccountId(),
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(time.Hour)),
		Audience:  jwt.ClaimStrings{"https://iam.api.cloud.yandex.net/iam/v1/tokens"},
	})
	token.Header["kid"] = b.key.Id

	return token
}
