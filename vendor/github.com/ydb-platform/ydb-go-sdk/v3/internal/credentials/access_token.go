package credentials

import (
	"context"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/secret"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

var (
	_ Credentials                  = (*AccessToken)(nil)
	_ fmt.Stringer                 = (*AccessToken)(nil)
	_ AccessTokenCredentialsOption = SourceInfoOption("")
)

type AccessTokenCredentialsOption interface {
	ApplyAccessTokenCredentialsOption(c *AccessToken)
}

// AccessToken implements Credentials interface with static
// authorization parameters.
type AccessToken struct {
	token      string
	sourceInfo string
}

func NewAccessTokenCredentials(token string, opts ...AccessTokenCredentialsOption) *AccessToken {
	c := &AccessToken{
		token:      token,
		sourceInfo: stack.Record(1),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyAccessTokenCredentialsOption(c)
		}
	}

	return c
}

// Token implements Credentials.
func (c AccessToken) Token(_ context.Context) (string, error) {
	return c.token, nil
}

// Token implements Credentials.
func (c AccessToken) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString("AccessToken{Token:")
	fmt.Fprintf(buffer, "%q", secret.Token(c.token))
	if c.sourceInfo != "" {
		buffer.WriteString(",From:")
		fmt.Fprintf(buffer, "%q", c.sourceInfo)
	}
	buffer.WriteByte('}')

	return buffer.String()
}
