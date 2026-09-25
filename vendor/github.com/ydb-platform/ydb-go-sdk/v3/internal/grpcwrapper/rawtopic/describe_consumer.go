package rawtopic

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/clone"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawoptional"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawscheme"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type DescribeConsumerRequest struct {
	OperationParams rawydb.OperationParams
	Path            string
	Consumer        string
	IncludeStats    bool
}

func (req *DescribeConsumerRequest) ToProto() *Ydb_Topic.DescribeConsumerRequest {
	return &Ydb_Topic.DescribeConsumerRequest{
		OperationParams: req.OperationParams.ToProto(),
		Path:            req.Path,
		Consumer:        req.Consumer,
		IncludeStats:    req.IncludeStats,
	}
}

type DescribeConsumerResult struct {
	Operation  rawydb.Operation
	Self       rawscheme.Entry
	Consumer   Consumer
	Partitions []DescribeConsumerResultPartitionInfo
}

func (res *DescribeConsumerResult) FromProto(response operation.Response) error {
	if err := res.Operation.FromProtoWithStatusCheck(response.GetOperation()); err != nil {
		return err
	}
	protoResult := &Ydb_Topic.DescribeConsumerResult{}
	if err := response.GetOperation().GetResult().UnmarshalTo(protoResult); err != nil {
		return xerrors.WithStackTrace(
			fmt.Errorf(
				"ydb: describe consumer result failed on unmarshal grpc result: %w", err,
			),
		)
	}

	if err := res.Self.FromProto(protoResult.GetSelf()); err != nil {
		return err
	}

	res.Consumer.MustFromProto(protoResult.GetConsumer())

	protoPartitions := protoResult.GetPartitions()
	res.Partitions = make([]DescribeConsumerResultPartitionInfo, len(protoPartitions))
	for i, protoPartition := range protoPartitions {
		if err := res.Partitions[i].FromProto(protoPartition); err != nil {
			return err
		}
	}

	return nil
}

type MultipleWindowsStat struct {
	PerMinute int64
	PerHour   int64
	PerDay    int64
}

func (stat *MultipleWindowsStat) MustFromProto(proto *Ydb_Topic.MultipleWindowsStat) {
	stat.PerMinute = proto.GetPerMinute()
	stat.PerHour = proto.GetPerHour()
	stat.PerDay = proto.GetPerDay()
}

type PartitionStats struct {
	PartitionsOffset rawtopiccommon.OffsetRange
	StoreSizeBytes   int64
	LastWriteTime    rawoptional.Time
	MaxWriteTimeLag  rawoptional.Duration
	BytesWritten     MultipleWindowsStat
}

func (ps *PartitionStats) FromProto(proto *Ydb_Topic.PartitionStats) error {
	if proto == nil {
		return nil
	}
	if err := ps.PartitionsOffset.FromProto(proto.GetPartitionOffsets()); err != nil {
		return err
	}
	ps.StoreSizeBytes = proto.GetStoreSizeBytes()
	ps.LastWriteTime.MustFromProto(proto.GetLastWriteTime())
	ps.MaxWriteTimeLag.ToProto()
	ps.BytesWritten.MustFromProto(proto.GetBytesWritten())

	return nil
}

type PartitionConsumerStats struct {
	LastReadOffset                 int64
	CommittedOffset                int64
	ReadSessionID                  string
	PartitionReadSessionCreateTime rawoptional.Time
	LastReadTime                   rawoptional.Time
	MaxReadTimeLag                 rawoptional.Duration
	MaxWriteTimeLag                rawoptional.Duration
	BytesRead                      MultipleWindowsStat
	ReaderName                     string
}

func (stats *PartitionConsumerStats) FromProto(proto *Ydb_Topic.DescribeConsumerResult_PartitionConsumerStats) error {
	if proto == nil {
		return nil
	}
	stats.LastReadOffset = proto.GetLastReadOffset()
	stats.CommittedOffset = proto.GetCommittedOffset()
	stats.ReadSessionID = proto.GetReadSessionId()
	stats.PartitionReadSessionCreateTime.MustFromProto(proto.GetPartitionReadSessionCreateTime())
	stats.LastReadTime.MustFromProto(proto.GetLastReadTime())
	stats.MaxReadTimeLag.MustFromProto(proto.GetMaxReadTimeLag())
	stats.MaxWriteTimeLag.MustFromProto(proto.GetMaxWriteTimeLag())
	stats.BytesRead.MustFromProto(proto.GetBytesRead())
	stats.ReaderName = proto.GetReaderName()

	return nil
}

type DescribeConsumerResultPartitionInfo struct {
	PartitionID            int64
	Active                 bool
	ChildPartitionIDs      []int64
	ParentPartitionIDs     []int64
	PartitionStats         PartitionStats
	PartitionConsumerStats PartitionConsumerStats
}

func (pi *DescribeConsumerResultPartitionInfo) FromProto(proto *Ydb_Topic.DescribeConsumerResult_PartitionInfo) error {
	pi.PartitionID = proto.GetPartitionId()
	pi.Active = proto.GetActive()

	pi.ChildPartitionIDs = clone.Int64Slice(proto.GetChildPartitionIds())
	pi.ParentPartitionIDs = clone.Int64Slice(proto.GetParentPartitionIds())

	if err := pi.PartitionConsumerStats.FromProto(proto.GetPartitionConsumerStats()); err != nil {
		return err
	}

	return pi.PartitionStats.FromProto(proto.GetPartitionStats())
}
