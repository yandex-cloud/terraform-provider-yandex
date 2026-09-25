package topicreaderinternal

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
)

// PublicGetPartitionStartOffsetResponse allow to set start offset for read messages for the partition
type PublicGetPartitionStartOffsetResponse struct {
	startOffset     rawtopiccommon.Offset
	startOffsetUsed bool
}

// StartFrom set start offset for read the partition
func (r *PublicGetPartitionStartOffsetResponse) StartFrom(offset int64) {
	r.startOffset.FromInt64(offset)
	r.startOffsetUsed = true
}

// PublicGetPartitionStartOffsetRequest info about partition
type PublicGetPartitionStartOffsetRequest struct {
	Topic       string
	PartitionID int64

	// ExampleOnly
	PartitionSessionID int64
}

// PublicGetPartitionStartOffsetFunc callback function for optional manage read progress store at own side
type PublicGetPartitionStartOffsetFunc func(
	ctx context.Context,
	req PublicGetPartitionStartOffsetRequest,
) (res PublicGetPartitionStartOffsetResponse, err error)
