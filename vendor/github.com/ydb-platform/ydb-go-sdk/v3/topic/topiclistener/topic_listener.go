package topiclistener

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topiclistenerinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type TopicListener struct {
	listenerReconnector *topiclistenerinternal.TopicListenerReconnector
}

func NewTopicListener(
	client *rawtopic.Client,
	config *topiclistenerinternal.StreamListenerConfig,
	handler EventHandler,
) (*TopicListener, error) {
	reconnector, err := topiclistenerinternal.NewTopicListenerReconnector(client, config, handler)
	if err != nil {
		return nil, err
	}

	res := &TopicListener{listenerReconnector: reconnector}
	if err = handler.OnReaderCreated(&ReaderReady{Listener: res}); err != nil {
		_ = res.Close(context.Background())

		return nil, err
	}

	return res, nil
}

func (cr *TopicListener) WaitInit(ctx context.Context) error {
	return cr.listenerReconnector.WaitInit(ctx)
}

func (cr *TopicListener) WaitStop(ctx context.Context) error {
	return cr.listenerReconnector.WaitStop(ctx)
}

func (cr *TopicListener) Close(ctx context.Context) error {
	return cr.listenerReconnector.Close(ctx, xerrors.WithStackTrace(topiclistenerinternal.ErrUserCloseTopic))
}
