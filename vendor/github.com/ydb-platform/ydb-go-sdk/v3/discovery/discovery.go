package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/endpoint"
)

type WhoAmI struct {
	User   string
	Groups []string
}

func (w WhoAmI) String() string {
	return fmt.Sprintf("{User: %s, Groups: [%s]}", w.User, strings.Join(w.Groups, ","))
}

type Client interface {
	Discover(ctx context.Context) ([]endpoint.Endpoint, error)
	WhoAmI(ctx context.Context) (*WhoAmI, error)
}
