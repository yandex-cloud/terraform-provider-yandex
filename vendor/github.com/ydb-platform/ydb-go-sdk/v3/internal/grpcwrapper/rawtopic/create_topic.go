package rawtopic

import (
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
)

type CreateTopicRequest struct {
	OperationParams rawydb.OperationParams

	Path                              string
	PartitioningSettings              PartitioningSettings
	RetentionPeriod                   time.Duration
	RetentionStorageMB                int64
	SupportedCodecs                   rawtopiccommon.SupportedCodecs
	PartitionWriteSpeedBytesPerSecond int64
	PartitionWriteBurstBytes          int64
	Attributes                        map[string]string
	Consumers                         []Consumer
	MeteringMode                      MeteringMode
}

func (req *CreateTopicRequest) ToProto() *Ydb_Topic.CreateTopicRequest {
	proto := &Ydb_Topic.CreateTopicRequest{
		Path:                 req.Path,
		PartitioningSettings: req.PartitioningSettings.ToProto(),
	}

	if req.RetentionPeriod != 0 {
		proto.RetentionPeriod = durationpb.New(req.RetentionPeriod)
	}
	proto.RetentionStorageMb = req.RetentionStorageMB
	proto.SupportedCodecs = req.SupportedCodecs.ToProto()
	proto.PartitionWriteSpeedBytesPerSecond = req.PartitionWriteSpeedBytesPerSecond
	proto.PartitionWriteBurstBytes = req.PartitionWriteBurstBytes
	proto.Attributes = req.Attributes

	proto.Consumers = make([]*Ydb_Topic.Consumer, len(req.Consumers))
	for i := range proto.GetConsumers() {
		proto.Consumers[i] = req.Consumers[i].ToProto()
	}

	proto.MeteringMode = Ydb_Topic.MeteringMode(req.MeteringMode)

	return proto
}

type CreateTopicResult struct {
	Operation rawydb.Operation
}

func (r *CreateTopicResult) FromProto(proto *Ydb_Topic.CreateTopicResponse) error {
	return r.Operation.FromProtoWithStatusCheck(proto.GetOperation())
}
