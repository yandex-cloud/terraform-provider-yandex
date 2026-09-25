package topicoptions

import "github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"

// DropOption type for drop options. Not used now.
type DropOption interface {
	ApplyDropOption(request *rawtopic.DropTopicRequest)
}
