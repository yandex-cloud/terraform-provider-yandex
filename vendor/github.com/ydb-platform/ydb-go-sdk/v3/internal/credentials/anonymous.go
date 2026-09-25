package credentials

import (
	"context"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

var (
	_ Credentials                = (*Anonymous)(nil)
	_ fmt.Stringer               = (*Anonymous)(nil)
	_ AnonymousCredentialsOption = SourceInfoOption("")
)

type AnonymousCredentialsOption interface {
	ApplyAnonymousCredentialsOption(c *Anonymous)
}

// Anonymous implements Credentials interface with Anonymous access
type Anonymous struct {
	sourceInfo string
}

func NewAnonymousCredentials(opts ...AnonymousCredentialsOption) *Anonymous {
	c := &Anonymous{
		sourceInfo: stack.Record(1),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyAnonymousCredentialsOption(c)
		}
	}

	return c
}

// Token implements Credentials.
func (c Anonymous) Token(_ context.Context) (string, error) {
	return "", nil
}

// Token implements Credentials.
func (c Anonymous) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString("Anonymous{")
	if c.sourceInfo != "" {
		buffer.WriteString("From:")
		fmt.Fprintf(buffer, "%q", c.sourceInfo)
	}
	buffer.WriteByte('}')

	return buffer.String()
}
