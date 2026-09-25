package rawtopic

import (
	"fmt"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/clone"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawscheme"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type DescribeTopicRequest struct {
	OperationParams rawydb.OperationParams
	Path            string
	IncludeStats    bool
}

func (req *DescribeTopicRequest) ToProto() *Ydb_Topic.DescribeTopicRequest {
	return &Ydb_Topic.DescribeTopicRequest{
		OperationParams: req.OperationParams.ToProto(),
		Path:            req.Path,
		IncludeStats:    req.IncludeStats,
	}
}

type DescribeTopicResult struct {
	Operation rawydb.Operation

	Self                              rawscheme.Entry
	PartitioningSettings              PartitioningSettings
	Partitions                        []PartitionInfo
	RetentionPeriod                   time.Duration
	RetentionStorageMB                int64
	SupportedCodecs                   rawtopiccommon.SupportedCodecs
	PartitionWriteSpeedBytesPerSecond int64
	PartitionWriteBurstBytes          int64
	Attributes                        map[string]string
	Consumers                         []Consumer
	MeteringMode                      MeteringMode
}

func (res *DescribeTopicResult) FromProto(response operation.Response) error {
	if err := res.Operation.FromProtoWithStatusCheck(response.GetOperation()); err != nil {
		return err
	}

	protoResult := &Ydb_Topic.DescribeTopicResult{}
	if err := response.GetOperation().GetResult().UnmarshalTo(protoResult); err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: describe topic result failed on unmarshal grpc result: %w", err))
	}

	if err := res.Self.FromProto(protoResult.GetSelf()); err != nil {
		return err
	}

	if err := res.PartitioningSettings.FromProto(protoResult.GetPartitioningSettings()); err != nil {
		return err
	}

	protoPartitions := protoResult.GetPartitions()
	res.Partitions = make([]PartitionInfo, len(protoPartitions))
	for i, protoPartition := range protoPartitions {
		err := res.Partitions[i].FromProto(protoPartition)
		if err != nil {
			return err
		}
	}

	res.RetentionPeriod = protoResult.GetRetentionPeriod().AsDuration()
	res.RetentionStorageMB = protoResult.GetRetentionStorageMb()

	for _, v := range protoResult.GetSupportedCodecs().GetCodecs() {
		res.SupportedCodecs = append(res.SupportedCodecs, rawtopiccommon.Codec(v))
	}

	res.PartitionWriteSpeedBytesPerSecond = protoResult.GetPartitionWriteSpeedBytesPerSecond()
	res.PartitionWriteBurstBytes = protoResult.GetPartitionWriteBurstBytes()

	res.Attributes = protoResult.GetAttributes()

	res.Consumers = make([]Consumer, len(protoResult.GetConsumers()))
	for i := range res.Consumers {
		res.Consumers[i].MustFromProto(protoResult.GetConsumers()[i])
	}

	res.MeteringMode = MeteringMode(protoResult.GetMeteringMode())

	return nil
}

type PartitionInfo struct {
	PartitionID        int64
	Active             bool
	ChildPartitionIDs  []int64
	ParentPartitionIDs []int64
	PartitionStats     PartitionStats
}

func (pi *PartitionInfo) FromProto(proto *Ydb_Topic.DescribeTopicResult_PartitionInfo) error {
	pi.PartitionID = proto.GetPartitionId()
	pi.Active = proto.GetActive()

	pi.ChildPartitionIDs = clone.Int64Slice(proto.GetChildPartitionIds())
	pi.ParentPartitionIDs = clone.Int64Slice(proto.GetParentPartitionIds())

	return pi.PartitionStats.FromProto(proto.GetPartitionStats())
}
