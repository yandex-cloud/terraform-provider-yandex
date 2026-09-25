package trace

import "context"

// tool gtrace used from ./internal/cmd/gtrace

//go:generate gtrace

type (
	// Discovery specified trace of discovery client activity.
	// gtrace:gen
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	Discovery struct {
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnDiscover func(DiscoveryDiscoverStartInfo) func(DiscoveryDiscoverDoneInfo)
		// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
		OnWhoAmI func(DiscoveryWhoAmIStartInfo) func(DiscoveryWhoAmIDoneInfo)
	}
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	DiscoveryDiscoverStartInfo struct {
		// Context make available context in trace callback function.
		// Pointer to context provide replacement of context in trace callback function.
		// Warning: concurrent access to pointer on client side must be excluded.
		// Safe replacement of context are provided only inside callback function
		Context  *context.Context
		Call     call
		Address  string
		Database string
	}
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	DiscoveryDiscoverDoneInfo struct {
		Location  string
		Endpoints []EndpointInfo
		Error     error
	}
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	DiscoveryWhoAmIStartInfo struct {
		// Context make available context in trace callback function.
		// Pointer to context provide replacement of context in trace callback function.
		// Warning: concurrent access to pointer on client side must be excluded.
		// Safe replacement of context are provided only inside callback function
		Context *context.Context
		Call    call
	}
	// Internals: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#internals
	DiscoveryWhoAmIDoneInfo struct {
		User   string
		Groups []string
		Error  error
	}
)
