package authentication

import "time"

// IamToken represents an interface for accessing an IAM token and its expiry information.
// GetIamToken retrieves the IAM token string.
// GetExpiresAt returns the expiration time of the IAM token.
type IamToken interface {
	GetIamToken() string
	GetExpiresAt() time.Time
}

// IamTokenImpl is an implementation of the IamToken interface, representing an IAM token with its value and expiration time.
type IamTokenImpl struct {
	Token     string
	ExpiresAt time.Time
}

// NewIamToken initializes and returns a new IamToken instance with the provided token value and expiration time.
func NewIamToken(token string, expiresAt time.Time) IamToken {
	return &IamTokenImpl{
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// GetIamToken returns the IAM token stored in the IamTokenImpl instance.
func (token *IamTokenImpl) GetIamToken() string {
	return token.Token
}

// GetExpiresAt returns the expiration time of the IAM token as a time.Time value.
func (token *IamTokenImpl) GetExpiresAt() time.Time {
	return token.ExpiresAt
}
