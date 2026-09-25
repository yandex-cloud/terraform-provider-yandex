package rawtopic

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
)

type UpdateOffsetsInTransactionRequest struct {
	OperationParams rawydb.OperationParams
	Tx              rawtopiccommon.TransactionIdentity
	Topics          []UpdateOffsetsInTransactionRequest_TopicOffsets
	Consumer        string
}

func (r *UpdateOffsetsInTransactionRequest) ToProto() *Ydb_Topic.UpdateOffsetsInTransactionRequest {
	req := &Ydb_Topic.UpdateOffsetsInTransactionRequest{
		OperationParams: r.OperationParams.ToProto(),
		Tx:              r.Tx.ToProto(),
		Consumer:        r.Consumer,
	}

	req.Topics = make([]*Ydb_Topic.UpdateOffsetsInTransactionRequest_TopicOffsets, len(r.Topics))
	for topicIndex := range r.Topics {
		topic := &r.Topics[topicIndex]
		offsets := &Ydb_Topic.UpdateOffsetsInTransactionRequest_TopicOffsets{
			Path: topic.Path,
		}

		offsets.Partitions = make(
			[]*Ydb_Topic.UpdateOffsetsInTransactionRequest_TopicOffsets_PartitionOffsets,
			len(topic.Partitions),
		)
		for partitionIndex := range topic.Partitions {
			partition := &topic.Partitions[partitionIndex]
			protoPartition := &Ydb_Topic.UpdateOffsetsInTransactionRequest_TopicOffsets_PartitionOffsets{
				PartitionId: partition.PartitionID,
			}

			protoPartition.PartitionOffsets = make([]*Ydb_Topic.OffsetsRange, len(partition.PartitionOffsets))

			for offsetIndex, offset := range partition.PartitionOffsets {
				protoPartition.PartitionOffsets[offsetIndex] = offset.ToProto()
			}

			offsets.Partitions[partitionIndex] = protoPartition
		}

		req.Topics[topicIndex] = offsets
	}

	return req
}

type UpdateOffsetsInTransactionRequest_TopicOffsets struct { //nolint:revive
	Path       string // Topic path
	Partitions []UpdateOffsetsInTransactionRequest_PartitionOffsets
}

type UpdateOffsetsInTransactionRequest_PartitionOffsets struct { //nolint:revive
	PartitionID      int64
	PartitionOffsets []rawtopiccommon.OffsetRange
}
