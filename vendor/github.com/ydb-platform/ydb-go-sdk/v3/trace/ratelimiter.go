package trace

// tool gtrace used from ./internal/cmd/gtrace

//go:generate gtrace

type (
	// Ratelimiter specified trace of ratelimiter client activity.
	// gtrace:gen
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	Ratelimiter struct{}
)
