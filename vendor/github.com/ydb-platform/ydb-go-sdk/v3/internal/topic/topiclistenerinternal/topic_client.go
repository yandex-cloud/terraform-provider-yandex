package topiclistenerinternal

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type TopicClient interface {
	StreamRead(ctxStreamLifeTime context.Context, readerID int64, tracer *trace.Topic) (rawtopicreader.StreamReader, error)
}
