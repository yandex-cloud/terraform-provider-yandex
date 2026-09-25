package ydb

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/coordination"
	"github.com/ydb-platform/ydb-go-sdk/v3/discovery"
	"github.com/ydb-platform/ydb-go-sdk/v3/ratelimiter"
	"github.com/ydb-platform/ydb-go-sdk/v3/scheme"
	"github.com/ydb-platform/ydb-go-sdk/v3/scripting"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic"
)

// Connection interface provide access to YDB service clients
// Interface and list of clients may be changed in the future
//
// Deprecated: use Driver instance instead.
// Will be removed at next major release.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
type Connection interface { //nolint:interfacebloat
	// Endpoint returns initial endpoint
	Endpoint() string

	// Name returns database name
	Name() string

	// Secure returns true if database connection is secure
	Secure() bool

	// Close closes connection and clear resources
	Close(ctx context.Context) error

	// Table returns table client
	Table() table.Client

	// Scheme returns scheme client
	Scheme() scheme.Client

	// Coordination returns coordination client
	Coordination() coordination.Client

	// Ratelimiter returns ratelimiter client
	Ratelimiter() ratelimiter.Client

	// Discovery returns discovery client
	Discovery() discovery.Client

	// Scripting returns scripting client
	Scripting() scripting.Client

	// Topic returns topic client
	Topic() topic.Client
}
