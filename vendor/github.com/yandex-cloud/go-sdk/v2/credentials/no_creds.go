package credentials

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ NonExchangeableCredentials = (*NoCredentials)(nil)

// NoCredentials implements Credentials, it allows to create unauthenticated connections
type NoCredentials struct{}

// NoAuthentication returns an instance of NoCredentials, allowing the creation of unauthenticated connections.
// Authentication could be added later via injection of gRPC metadata authorization header
func NoAuthentication() *NoCredentials {
	return &NoCredentials{}
}
func (creds *NoCredentials) YandexCloudAPICredentials() {}

// IAMToken always returns gRPC error with status UNAUTHENTICATED
func (creds *NoCredentials) IAMToken(ctx context.Context) (*CredentialsToken, error) {
	return nil, status.Error(codes.Unauthenticated, "unauthenticated connection")
}
