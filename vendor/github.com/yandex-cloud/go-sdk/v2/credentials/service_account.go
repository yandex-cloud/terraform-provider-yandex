package credentials

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/yandex-cloud/go-sdk/v2/pkg/iamkey"
)

// ServiceAccountKey returns credentials for the given IAM Key. The key is used to sign JWT tokens.
// JWT tokens are exchanged for IAM Tokens used to authorize API calls.
// This authorization method is not supported for IAM Keys issued for User Accounts.
func ServiceAccountKey(key *iamkey.Key) (ExchangeableCredentials, error) {
	jwtBuilder, err := newServiceAccountJWTBuilder(key)
	if err != nil {
		return nil, err
	}

	return exchangeableCredentialsFunc(func() (*CredentialsTokenRequest, error) {
		signedJWT, err := jwtBuilder.SignedToken()
		if err != nil {
			return nil, fmt.Errorf("JWT sign failed : %w", err)
		}
		return &CredentialsTokenRequest{
			Token:    signedJWT,
			Identity: CredentialsIdentityJWT,
		}, nil
	}), nil
}

// ServiceAccountKeyFile creates Credentials using a service account key file specified by the keyFilePath.
// It reads and parses the key file to build exchangeable credentials for API authorization.
func ServiceAccountKeyFile(keyFilePath string) (Credentials, error) {
	key, err := iamkey.ReadFromJSONFile(keyFilePath)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed to load service account key from %s", keyFilePath))
	}

	return ServiceAccountKey(key)
}
