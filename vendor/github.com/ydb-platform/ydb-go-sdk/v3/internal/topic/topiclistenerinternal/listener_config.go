package topiclistenerinternal

import (
	"errors"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type StreamListenerConfig struct {
	BufferSize             int
	Decoders               topicreadercommon.DecoderMap
	Selectors              []*topicreadercommon.PublicReadSelector
	Consumer               string
	ConnectWithoutConsumer bool
	readerID               int64
	Tracer                 *trace.Topic
}

func NewStreamListenerConfig() StreamListenerConfig {
	return StreamListenerConfig{
		BufferSize: topicreadercommon.DefaultBufferSize,
		Decoders:   topicreadercommon.NewDecoderMap(),
		Selectors:  nil,
		Consumer:   "",
		readerID:   topicreadercommon.NextReaderID(),
		Tracer:     &trace.Topic{},
	}
}

func (cfg *StreamListenerConfig) Validate() error {
	var errs []error
	if cfg.Consumer == "" && !cfg.ConnectWithoutConsumer {
		errs = append(errs, errors.New(
			"empty consumer without ConnectWithoutConsumer flag. Set the consumer or  the flag",
		))
	}
	if cfg.Consumer != "" && cfg.ConnectWithoutConsumer {
		errs = append(errs, errors.New(
			"non empty consumer, but ConnectWithoutConsumer flag set. Clear the consumer or the flag",
		))
	}
	if len(cfg.Selectors) == 0 {
		errs = append(errs, errors.New("topic selectors are empty"))
	}
	if cfg.BufferSize <= 0 {
		errs = append(errs, fmt.Errorf(
			"buffer size of the topic listener should be greater then 0, now: %v",
			cfg.BufferSize,
		))
	}

	if len(errs) > 0 {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: topic listener config validation failed: %w",
			errors.Join(errs...),
		)))
	}

	return nil
}
