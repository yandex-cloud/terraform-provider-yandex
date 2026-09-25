package operation

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
)

type Mode uint

const (
	ModeUnknown Mode = iota
	ModeSync
	ModeAsync
)

func (m Mode) String() string {
	switch m {
	case ModeSync:
		return "sync"
	case ModeAsync:
		return "async"
	default:
		return "unknown"
	}
}

func (m Mode) toYDB() Ydb_Operations.OperationParams_OperationMode {
	switch m {
	case ModeSync:
		return Ydb_Operations.OperationParams_SYNC
	case ModeAsync:
		return Ydb_Operations.OperationParams_ASYNC
	default:
		return Ydb_Operations.OperationParams_OPERATION_MODE_UNSPECIFIED
	}
}
