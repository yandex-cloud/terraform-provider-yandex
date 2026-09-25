package rawtopic

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
)

type DropTopicRequest struct {
	OperationParams rawydb.OperationParams
	Path            string
}

func (req *DropTopicRequest) ToProto() *Ydb_Topic.DropTopicRequest {
	return &Ydb_Topic.DropTopicRequest{
		OperationParams: req.OperationParams.ToProto(),
		Path:            req.Path,
	}
}

type DropTopicResult struct {
	Operation rawydb.Operation
}

func (r *DropTopicResult) FromProto(proto *Ydb_Topic.DropTopicResponse) error {
	return r.Operation.FromProtoWithStatusCheck(proto.GetOperation())
}
