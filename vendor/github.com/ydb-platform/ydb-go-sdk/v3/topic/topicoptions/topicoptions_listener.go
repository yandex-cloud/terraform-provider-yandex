package topicoptions

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topiclistenerinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
)

// ListenerOption set settings for topic listener struct
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type ListenerOption func(cfg *topiclistenerinternal.StreamListenerConfig)

// WithListenerAddDecoder add decoder for a codec.
// It allows to set decoders fabric for custom codec and replace internal decoders.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithListenerAddDecoder(codec topictypes.Codec, decoderCreate CreateDecoderFunc) ListenerOption {
	return func(cfg *topiclistenerinternal.StreamListenerConfig) {
		cfg.Decoders.AddDecoder(rawtopiccommon.Codec(codec), decoderCreate)
	}
}
