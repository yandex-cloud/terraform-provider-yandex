package topicreaderinternal

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
)

//go:generate mockgen -source batched_stream_reader_interface.go --typed -destination batched_stream_reader_mock_test.go -package topicreaderinternal -write_package_comment=false

type batchedStreamReader interface {
	WaitInit(ctx context.Context) error
	ReadMessageBatch(ctx context.Context, opts ReadMessageBatchOptions) (*topicreadercommon.PublicBatch, error)
	Commit(ctx context.Context, commitRange topicreadercommon.CommitRange) error
	CloseWithError(ctx context.Context, err error) error
	PopMessagesBatchTx(ctx context.Context, tx tx.Transaction, opts ReadMessageBatchOptions) (*topicreadercommon.PublicBatch, error) //nolint:lll
	TopicOnReaderStart(consumer string, err error)
}
