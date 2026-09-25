package authentication

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	iampb "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/errors"
	"github.com/yandex-cloud/go-sdk/v2/pkg/transport"
	"github.com/yandex-cloud/go-sdk/v2/pkg/transport/middleware/grpcdebug"
	"github.com/yandex-cloud/go-sdk/v2/pkg/transport/middleware/requestid"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
)

// Authenticator provides methods for generating IAM tokens for an authenticated entity or service account.
type Authenticator interface {
	CreateIAMToken(ctx context.Context) (IamToken, error)
	CreateIAMTokenForServiceAccount(ctx context.Context, serviceAccountID string) (IamToken, error)
}

// AuthenticatorImpl provides functionality for generating and managing IAM tokens using supplied credentials and IAM client.
type AuthenticatorImpl struct {
	creds          credentials.Credentials
	iamTokenClient iamsdk.IamTokenClient
	logger         *zap.Logger
}

// NewAuthenticatorFromEndpoint creates a new AuthenticatorImpl using provided credentials and endpoint configuration.
// Returns the constructed AuthenticatorImpl instance or an error if the connector initialization fails.
// The IAM token client gets the requestid interceptor so token-creation failures carry x-request-id /
// x-server-trace-id all the way up to error reporting, and the grpcdebug interceptor so the call is
// logged when the supplied logger is at Debug level (i.e. when --debug is on).
func NewAuthenticatorFromEndpoint(logger *zap.Logger, creds credentials.Credentials, endpoint *endpoints.Endpoint) (*AuthenticatorImpl, error) {
	dialOpts := append([]grpc.DialOption{}, endpoint.DialOptions...)
	dialOpts = append(dialOpts, grpc.WithChainUnaryInterceptor(
		requestid.Interceptor(),
		grpcdebug.UnaryClientInterceptor(logger),
	))
	return NewAuthenticator(logger, creds, iamsdk.NewIamTokenClient(transport.NewSingleConnector(endpoint.Addr, dialOpts...))), nil
}

// NewAuthenticator creates and returns a new instance of AuthenticatorImpl using the provided credentials and IamTokenClient.
func NewAuthenticator(logger *zap.Logger, creds credentials.Credentials, iamTokenClient iamsdk.IamTokenClient) *AuthenticatorImpl {
	return &AuthenticatorImpl{
		creds:          creds,
		iamTokenClient: iamTokenClient,
		logger:         logger,
	}
}

// CreateIAMToken generates an IAM token using the provided credentials in the `AuthenticatorImpl` instance.
func (a *AuthenticatorImpl) CreateIAMToken(ctx context.Context) (IamToken, error) {
	a.logger.Debug("Creating IAM token")
	switch creds := a.creds.(type) {
	case credentials.ExchangeableCredentials:
		a.logger.Debug("Using exchangeable credentials")
		return a.createTokenFromExchangeable(ctx, creds)
	case credentials.NonExchangeableCredentials:
		a.logger.Debug("Using non-exchangeable credentials")
		return a.createTokenFromNonExchangeable(ctx, creds)
	default:
		err := fmt.Errorf("unsupported credentials type %T", creds)
		a.logger.Error("Failed to create IAM token", zap.Error(err))
		return nil, &errors.AuthError{Err: err}
	}
}

// createTokenFromExchangeable generates an IAM token using exchangeable credentials, handling request creation and errors.
func (a *AuthenticatorImpl) createTokenFromExchangeable(ctx context.Context, creds credentials.ExchangeableCredentials) (IamToken, error) {
	a.logger.Debug("Generating IAM token request")
	req, err := creds.IAMTokenRequest()
	if err != nil {
		a.logger.Error("Failed to create IAM token request", zap.Error(err))
		return nil, &errors.AuthError{Err: err}
	}

	tokenReq := a.createIamTokenRequestFromCredential(req)
	if tokenReq == nil {
		err := fmt.Errorf("invalid identity type")
		a.logger.Error("Failed to create token request from credentials", zap.Error(err))
		return nil, &errors.AuthError{Err: err}
	}

	a.logger.Debug("Creating IAM token from request")
	tokenResp, err := a.iamTokenClient.Create(ctx, tokenReq)
	if err != nil {
		a.logger.Error("Failed to create IAM token", zap.Error(err))
		return nil, &errors.AuthError{Err: err}
	}

	a.logger.Debug("Successfully created IAM token")
	return NewIamToken(tokenResp.IamToken, tokenResp.ExpiresAt.AsTime()), nil
}

// createTokenFromNonExchangeable generates an IAM token using NonExchangeableCredentials without calling the token service.
// Returns the generated IamToken or an error if token retrieval fails.
func (a *AuthenticatorImpl) createTokenFromNonExchangeable(ctx context.Context, creds credentials.NonExchangeableCredentials) (IamToken, error) {
	a.logger.Debug("Retrieving IAM token from non-exchangeable credentials")
	tokenResp, err := creds.IAMToken(ctx)
	if err != nil {
		a.logger.Error("Failed to get IAM token from non-exchangeable credentials", zap.Error(err))
		return nil, &errors.AuthError{Err: err}
	}
	a.logger.Debug("Successfully retrieved IAM token")
	return NewIamToken(tokenResp.Token, tokenResp.ExpiresAt), nil
}

// CreateIAMTokenForServiceAccount generates a new IAM token for the provided service account ID using the IAM token client.
func (a *AuthenticatorImpl) CreateIAMTokenForServiceAccount(ctx context.Context, serviceAccountID string) (IamToken, error) {
	a.logger.Debug("Creating IAM token for service account", zap.String("serviceAccountID", serviceAccountID))
	token, err := a.CreateIAMToken(ctx)
	if err != nil {
		a.logger.Error("Failed to create IAM token", zap.Error(err))
		return nil, &errors.AuthError{Err: err}
	}
	a.logger.Debug("Successfully created IAM token")
	authctx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.GetIamToken())
	tokenResp, err := a.iamTokenClient.CreateForServiceAccount(authctx,
		&iampb.CreateIamTokenForServiceAccountRequest{
			ServiceAccountId: serviceAccountID,
		},
	)
	if err != nil {
		a.logger.Error("Failed to create IAM token for service account",
			zap.String("serviceAccountID", serviceAccountID),
			zap.Error(err))
		return nil, &errors.AuthError{Err: err}
	}
	a.logger.Debug("Successfully created IAM token for service account",
		zap.String("serviceAccountID", serviceAccountID))
	return NewIamToken(tokenResp.IamToken, tokenResp.ExpiresAt.AsTime()), nil
}

// createIamTokenRequestFromCredential converts a CredentialsTokenRequest into an iampb.CreateIamTokenRequest.
func (a *AuthenticatorImpl) createIamTokenRequestFromCredential(req *credentials.CredentialsTokenRequest) *iampb.CreateIamTokenRequest {
	switch req.Identity {
	case credentials.CredentialsIdentityYandexPassportOauthToken:
		a.logger.Warn("OAuthToken credetial provider is deprecated. Please consider to use another credential provider.")
		return &iampb.CreateIamTokenRequest{
			Identity: &iampb.CreateIamTokenRequest_YandexPassportOauthToken{
				YandexPassportOauthToken: req.Token},
		}
	case credentials.CredentialsIdentityJWT:
		return &iampb.CreateIamTokenRequest{
			Identity: &iampb.CreateIamTokenRequest_Jwt{Jwt: req.Token}}
	}
	return nil
}
