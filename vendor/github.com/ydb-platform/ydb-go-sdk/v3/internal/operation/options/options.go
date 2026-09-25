package options

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"

type (
	ListOperationsRequest struct {
		Ydb_Operations.ListOperationsRequest
	}
	List func(r *ListOperationsRequest)
)
