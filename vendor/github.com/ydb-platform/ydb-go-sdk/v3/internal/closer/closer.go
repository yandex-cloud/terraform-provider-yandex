package closer

import (
	"context"
)

// Closer is the interface that for ydb variang of close with context.
type Closer interface {
	Close(ctx context.Context) error
}
