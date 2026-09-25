package authentication

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-sdk/v2/pkg/authentication"
)

// Option defines a functional option for configuring an IamTokenMiddleware instance.
type Option func(*IamTokenMiddleware)

// WithTimeFunc allows overriding the time function used by IamTokenMiddleware, useful for testing.
func WithTimeFunc(timeFunc func() time.Time) Option {
	return func(m *IamTokenMiddleware) {
		m.now = timeFunc
	}
}

// WithLogger allows setting a custom logger for IamTokenMiddleware.
func WithLogger(logger *zap.Logger) Option {
	return func(m *IamTokenMiddleware) {
		m.logger = logger
	}
}

// NewIAMTokenMiddleware initializes and returns a new instance of IamTokenMiddleware for managing IAM token authentication.
func NewIAMTokenMiddleware(authenticator authentication.Authenticator, opts ...Option) *IamTokenMiddleware {
	m := &IamTokenMiddleware{
		now:            time.Now,
		authenticator:  authenticator,
		subjectToState: map[authSubject]iamTokenState{},
		logger:         zap.NewNop(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// IamTokenMiddleware is a middleware for managing IAM token generation and caching for authenticated entities.
// It uses an Authenticator to create and retrieve tokens and ensures thread-safe access and updates with a mutex.
// The middleware also employs a singleflight pattern to prevent duplicate token generation for the same subject.
// Customizable `now` function allows time manipulation for testing purposes.
type IamTokenMiddleware struct {
	logger        *zap.Logger
	authenticator authentication.Authenticator
	// now may be replaced in tests
	now func() time.Time

	singleFlight singleflight.Group

	// mutex guards conn and currentState, and excludes multiple simultaneous Token updates
	mutex          sync.RWMutex
	subjectToState map[authSubject]iamTokenState
}

// iamTokenState represents the state of an IAM token including its value, expiration time, and version for tracking updates.
type iamTokenState struct {
	token     string
	expiresAt time.Time
	version   int
}

// InterceptUnary intercepts unary gRPC calls to inject IAM Token into the context as authorization metadata.
func (c *IamTokenMiddleware) InterceptUnary(ctx context.Context, method string, req, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	c.logger.Debug("Intercepting unary RPC call", zap.String("method", method))
	ctx, err := c.contextWithAuthMetadata(ctx, method, opts)
	if err != nil {
		c.logger.Error("Failed to get auth metadata for unary call", zap.Error(err))
		return err
	}

	return invoker(ctx, method, req, reply, conn, opts...)
}

// InterceptStream intercepts a gRPC streaming client call to add authentication metadata to the context.
func (c *IamTokenMiddleware) InterceptStream(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	c.logger.Debug("Intercepting stream RPC call", zap.String("method", method))
	ctx, err := c.contextWithAuthMetadata(ctx, method, opts)
	if err != nil {
		c.logger.Error("Failed to get auth metadata for stream call", zap.Error(err))
		return nil, err
	}

	return streamer(ctx, desc, conn, method, opts...)
}

// contextWithAuthMetadata adds authentication metadata to the provided context if required by the given method.
// It retrieves an IAM token if necessary and appends it to the outgoing context's authorization header.
func (c *IamTokenMiddleware) contextWithAuthMetadata(ctx context.Context, method string, opts []grpc.CallOption) (context.Context, error) {
	if !needAuth(method) {
		c.logger.Debug("Authentication not required for method", zap.String("method", method))
		return ctx, nil
	}
	// User can add WithAuthAsServiceAccount to default call options and we will
	// always try to issue Token for service account. That results in a deadlock.
	// Here we check for methods that always require original authentication and
	// not delegated mode.
	c.logger.Debug("Getting IAM Token for method", zap.String("method", method))

	token, err := c.GetIAMToken(ctx, needOriginalSubject(method), opts...)
	if err != nil {
		c.logger.Error("Failed to get IAM token", zap.Error(err))
		return nil, err
	}

	c.logger.Debug("Successfully obtained IAM token and set authorization header")

	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token), nil
}

// GetIAMToken retrieves an IAM token for a given subject, either from cache or by creating a new token if expired or absent.
func (c *IamTokenMiddleware) GetIAMToken(ctx context.Context, originalSubject bool, opts ...grpc.CallOption) (string, error) {
	subject, err := callAuthSubject(ctx, originalSubject, opts)
	if err != nil {
		c.logger.Error("Failed to get auth subject", zap.Error(err))
		return "", err
	}

	if subject_, ok := subject.(serviceAccount); ok {
		c.logger.Info("Getting IAM Token for Service Account", zap.String("serviceAccountID", subject_.id))
	}

	c.mutex.RLock()
	state := c.subjectToState[subject]
	c.mutex.RUnlock()

	token := state.token

	expiresIn := state.expiresAt.Sub(c.now())
	if expiresIn > 0 {
		c.logger.Debug("Using cached IAM Token",
			zap.Duration("expiresIn", expiresIn),
			zap.Time("expiresAt", state.expiresAt))
		return token, nil
	}

	if token == "" {
		c.logger.Debug("No IAM Token cached, creating new token")
	} else {
		c.logger.Debug("IAM Token expired, updating token", zap.Time("expiredAt", state.expiresAt))
	}

	token, err = c.updateTokenSingleFlight(ctx, subject, state)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			c.logger.Error("Authentication failed", zap.Error(err))
			return "", err
		}

		c.logger.Error("Failed to update token", zap.Error(err))
		return "", status.Errorf(codes.Unauthenticated, "%v", err)
	}

	return token, nil
}

// updateTokenSingleFlight ensures that a token update is performed only once per subject using a single-flight mechanism.
// It attempts to update the IAM token and returns the updated token string or an error if the operation fails.
func (c *IamTokenMiddleware) updateTokenSingleFlight(ctx context.Context, subject authSubject, state iamTokenState) (string, error) {
	c.logger.Debug("Starting token update with singleflight")
	rawResult, err, _ := c.singleFlight.Do(subject.hash(), func() (interface{}, error) {
		token, err := c.updateToken(ctx, subject, state.version)
		return token, err
	})
	if err != nil {
		c.logger.Error("Token update failed in singleflight", zap.Error(err))
		return "", err
	}

	c.logger.Debug("Token update completed successfully")
	return rawResult.(string), nil
}

// updateToken generates or updates an IAM token for the provided subject if the current version matches the stored version.
// It ensures thread safety during token updates and caches the token state for reuse.
// If the version differs, it returns the cached token without creating a new one.
func (c *IamTokenMiddleware) updateToken(ctx context.Context, subject authSubject, currentVersion int) (string, error) {
	c.mutex.RLock()
	state := c.subjectToState[subject]
	c.mutex.RUnlock()

	if state.version != currentVersion {
		c.logger.Debug("Token already updated by another goroutine",
			zap.Int("currentVersion", currentVersion),
			zap.Int("stateVersion", state.version))
		return state.token, nil
	}

	c.logger.Debug("Creating new IAM token")
	token, err := subject.createIAMToken(ctx, c.authenticator)
	if err != nil {
		c.logger.Error("Failed to create IAM token", zap.Error(err))
		return "", fmt.Errorf("iam Token create failed: %w", err)
	}

	expiresAt := token.GetExpiresAt()

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.subjectToState[subject] = iamTokenState{
		token:     token.GetIamToken(),
		expiresAt: expiresAt,
		version:   currentVersion + 1,
	}

	c.logger.Debug("Successfully updated token state",
		zap.Time("expiresAt", expiresAt),
		zap.Int("newVersion", currentVersion+1))

	return token.GetIamToken(), nil
}

// needAuth determines if authentication is required for the given method based on its name suffix.
func needAuth(method string) bool {
	switch {
	case strings.HasSuffix(method, "iam.v1.IamTokenService/Create"),
		strings.HasSuffix(method, "endpoint.ApiEndpointService/List"):
		return false
	default:
		return true
	}
}

// needOriginalSubject determines whether the original subject is required for the specified gRPC method.
func needOriginalSubject(method string) bool {
	switch {
	case strings.HasSuffix(method, "iam.v1.IamTokenService/CreateForServiceAccount"):
		return false
	default:
		return true
	}
}
