package operation

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"

type Response interface {
	GetOperation() *Ydb_Operations.Operation
}
