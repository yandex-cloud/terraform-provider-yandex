package operation

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation/options"
)

func WithPageSize(pageSize uint64) options.List {
	return func(r *options.ListOperationsRequest) {
		r.PageSize = pageSize
	}
}

func WithPageToken(pageToken string) options.List {
	return func(r *options.ListOperationsRequest) {
		r.PageToken = pageToken
	}
}
