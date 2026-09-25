package credentials

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Auth_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Auth"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/secret"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

const TokenRefreshDivisor = 10

var (
	_ Credentials             = (*Static)(nil)
	_ fmt.Stringer            = (*Static)(nil)
	_ StaticCredentialsOption = grpcDialOptionsOption(nil)
)

type grpcDialOptionsOption []grpc.DialOption

func (opts grpcDialOptionsOption) ApplyStaticCredentialsOption(c *Static) {
	c.opts = opts
}

type StaticCredentialsOption interface {
	ApplyStaticCredentialsOption(c *Static)
}

func WithGrpcDialOptions(opts ...grpc.DialOption) grpcDialOptionsOption {
	return opts
}

func NewStaticCredentials(user, password, endpoint string, opts ...StaticCredentialsOption) *Static {
	c := &Static{
		user:       user,
		password:   password,
		endpoint:   endpoint,
		sourceInfo: stack.Record(1),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyStaticCredentialsOption(c)
		}
	}

	return c
}

var (
	_ Credentials  = (*Static)(nil)
	_ fmt.Stringer = (*Static)(nil)
)

// Static implements Credentials interface with static
// authorization parameters.
type Static struct {
	user       string
	password   string
	endpoint   string
	opts       []grpc.DialOption
	token      string
	requestAt  time.Time
	mu         sync.Mutex
	sourceInfo string
}

func (c *Static) Token(ctx context.Context) (token string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Until(c.requestAt) > 0 {
		return c.token, nil
	}
	cc, err := grpc.DialContext(ctx, c.endpoint, c.opts...) //nolint:staticcheck,nolintlint
	if err != nil {
		return "", xerrors.WithStackTrace(
			fmt.Errorf("dial failed: %w", err),
		)
	}
	defer cc.Close()

	client := Ydb_Auth_V1.NewAuthServiceClient(cc)

	response, err := client.Login(ctx, &Ydb_Auth.LoginRequest{
		OperationParams: &Ydb_Operations.OperationParams{
			OperationMode:    0,
			OperationTimeout: nil,
			CancelAfter:      nil,
			Labels:           nil,
			ReportCostInfo:   0,
		},
		User:     c.user,
		Password: c.password,
	})
	if err != nil {
		return "", xerrors.WithStackTrace(err)
	}

	switch {
	case !response.GetOperation().GetReady():
		return "", xerrors.WithStackTrace(
			fmt.Errorf("operation '%s' not ready: %v",
				response.GetOperation().GetId(),
				response.GetOperation().GetIssues(),
			),
		)

	case response.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS:
		return "", xerrors.WithStackTrace(
			xerrors.Operation(
				xerrors.FromOperation(response.GetOperation()),
				xerrors.WithAddress(c.endpoint),
			),
		)
	}
	var result Ydb_Auth.LoginResult
	if err = response.GetOperation().GetResult().UnmarshalTo(&result); err != nil {
		return "", xerrors.WithStackTrace(err)
	}

	expiresAt, err := parseExpiresAt(result.GetToken())
	if err != nil {
		return "", xerrors.WithStackTrace(err)
	}

	c.requestAt = time.Now().Add(time.Until(expiresAt) / TokenRefreshDivisor)
	c.token = result.GetToken()

	return c.token, nil
}

func parseExpiresAt(raw string) (expiresAt time.Time, err error) {
	var claims jwt.RegisteredClaims
	if _, _, err = jwt.NewParser().ParseUnverified(raw, &claims); err != nil {
		return expiresAt, xerrors.WithStackTrace(err)
	}

	if claims.ExpiresAt == nil {
		return expiresAt, xerrors.WithStackTrace(errNilExpiresAt)
	}

	return claims.ExpiresAt.Time, nil
}

func (c *Static) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString("Static{User:")
	fmt.Fprintf(buffer, "%q", c.user)
	buffer.WriteString(",Password:")
	fmt.Fprintf(buffer, "%q", secret.Password(c.password))
	buffer.WriteString(",Token:")
	fmt.Fprintf(buffer, "%q", secret.Token(c.token))
	if c.sourceInfo != "" {
		buffer.WriteString(",From:")
		fmt.Fprintf(buffer, "%q", c.sourceInfo)
	}
	buffer.WriteByte('}')

	return buffer.String()
}
