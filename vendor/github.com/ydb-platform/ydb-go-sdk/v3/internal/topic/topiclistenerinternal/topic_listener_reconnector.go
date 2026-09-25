package topiclistenerinternal

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/background"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
)

var (
	ErrUserCloseTopic      = errors.New("ydb: user closed topic listener")
	errTopicListenerClosed = errors.New("ydb: the topic listener already closed")
)

type TopicListenerReconnector struct {
	streamConfig *StreamListenerConfig
	client       TopicClient
	handler      EventHandler

	background background.Worker

	connectionResult    error
	connectionCompleted empty.Chan
	connectionIDCounter atomic.Int64
	closing             atomic.Bool

	m              sync.Mutex
	streamListener *streamListener
}

func NewTopicListenerReconnector(
	client TopicClient,
	streamConfig *StreamListenerConfig,
	handler EventHandler,
) (*TopicListenerReconnector, error) {
	res := &TopicListenerReconnector{
		streamConfig:        streamConfig,
		client:              client,
		handler:             handler,
		connectionCompleted: make(empty.Chan),
	}

	res.background.Start("connection", res.connect)

	return res, nil
}

func (lr *TopicListenerReconnector) Close(ctx context.Context, reason error) error {
	if !lr.closing.CompareAndSwap(false, true) {
		return errTopicListenerClosed
	}
	var closeErrors []error
	err := lr.background.Close(ctx, reason)
	closeErrors = append(closeErrors, err)

	lr.m.Lock()
	sl := lr.streamListener
	lr.m.Unlock()

	if sl != nil {
		err = sl.Close(ctx, reason)
		if !errors.Is(err, context.Canceled) {
			closeErrors = append(closeErrors, err)
		}
	}

	return errors.Join(closeErrors...)
}

func (lr *TopicListenerReconnector) connect(connectionCtx context.Context) {
	sl, connRes := newStreamListener(
		connectionCtx,
		lr.client,
		lr.handler,
		lr.streamConfig,
		&lr.connectionIDCounter,
	)
	lr.m.Lock()
	lr.streamListener = sl
	lr.connectionResult = connRes
	lr.m.Unlock()

	close(lr.connectionCompleted)
}

func (lr *TopicListenerReconnector) WaitInit(ctx context.Context) error {
	select {
	case <-ctx.Done():
		// pass
	case <-lr.connectionCompleted:
		// pass
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return lr.connectionResult
}

func (lr *TopicListenerReconnector) WaitStop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-lr.background.StopDone():
		err := lr.background.CloseReason()
		if errors.Is(err, ErrUserCloseTopic) {
			return nil
		}

		return err
	}
}
