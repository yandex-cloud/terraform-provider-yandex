package topicreaderinternal

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
)

//go:generate mockgen -source topic_client_interface.go -destination topic_client_interface_mock_test.go -package topicreaderinternal -write_package_comment=false --typed

// TopicClient is part of rawtopic.Client
type TopicClient interface {
	UpdateOffsetsInTransaction(ctx context.Context, req *rawtopic.UpdateOffsetsInTransactionRequest) error
}
