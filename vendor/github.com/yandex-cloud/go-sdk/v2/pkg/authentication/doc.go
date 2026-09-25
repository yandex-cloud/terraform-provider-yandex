// Package authentication provides a interface for IAM (Identity and Access Management)
// token operations in sdk services.
//
// Core Interfaces:
//
// The Authenticator interface defines the main contract for token operations:
//
//	type Authenticator interface {
//	    CreateIAMToken(ctx context.Context) (IamToken, error)
//	    CreateIAMTokenForServiceAccount(ctx context.Context, serviceAccountID string) (IamToken, error)
//	}
//
// Usage Examples:
//
// Creating an authenticator with endpoint:
//
//	auth, err := authentication.NewAuthenticatorFromEndpoint(credentials, endpoint)
//	if err != nil {
//	    // handle error
//	}
//
//	// Generate token
//	token, err := auth.CreateIAMToken(ctx)
//
// Creating an authenticator directly:
//
//	auth := authentication.NewAuthenticator(credentials, iamTokenClient)
//	token, err := auth.CreateIAMToken(ctx)
//
// Error Handling:
// The package uses AuthError type for detailed error reporting:
//
//	type AuthError struct {
//	    Op  string  // Operation where error occurred
//	    Err error   // Underlying error
//	}
//
// Credential Types:
// The authenticator supports two main types of credentials:
//   - ExchangeableCredentials: Credentials that can be exchanged for IAM tokens
//   - NonExchangeableCredentials: Credentials that directly provide IAM tokens
package authentication
