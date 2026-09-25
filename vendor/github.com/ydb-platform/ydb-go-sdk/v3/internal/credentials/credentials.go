package credentials

import (
	"context"
)

// Credentials is an interface of YDB credentials required for connect with YDB
type Credentials interface {
	// Token must return actual token or error
	Token(ctx context.Context) (string, error)
}
