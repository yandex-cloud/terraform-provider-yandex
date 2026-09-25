package topiclistener

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topiclistenerinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
)

// EventHandler methods will be called sequentially by partition,
// but can be called in parallel for different partitions.
// You should include topiclistener.BaseHandler into your struct for the interface implementation
// It allows to extend the interface in the future without broke compatibility.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type EventHandler interface {
	// topicReaderHandler needs for guarantee inherits from base struct with default implementations of new methods
	topicReaderHandler()

	topiclistenerinternal.EventHandler

	// OnReaderCreated called once at the reader complete internal initialization
	// It not mean that reader is connected to a server.
	// Allow easy initialize your handler with the reader without sync with return of topic.Client StartListener method
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	OnReaderCreated(event *ReaderReady) error
}

type ReaderReady struct {
	Listener *TopicListener
}

type ReadMessages = topiclistenerinternal.PublicReadMessages

// BaseHandler implements default behavior for EventHandler interface
// you must embed the structure to your own implementation of the interface.
//
// # It allows to extend the interface in the future without broke compatibility
//
// Temporary restrictions: all method should be work fast, because is it call from main read loop message and block
// handle messages loop
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type BaseHandler struct{}

func (b BaseHandler) topicReaderHandler() {}

func (b BaseHandler) OnReaderCreated(event *ReaderReady) error {
	return nil
}

func (b BaseHandler) OnStartPartitionSessionRequest(
	ctx context.Context,
	event *EventStartPartitionSession,
) error {
	event.Confirm()

	return nil
}

// OnStopPartitionSessionRequest called when server want to stop send messages for the partition
// the method may be called more than once for partition session: with graceful shutdown and without
// no guarantee to call with graceful=false after graceful true
// it called with partition context if partition exists and with cancelled background context if
// not
func (b BaseHandler) OnStopPartitionSessionRequest(
	ctx context.Context,
	event *EventStopPartitionSession,
) error {
	event.Confirm()

	return nil
}

func (b BaseHandler) OnReadMessages(
	ctx context.Context,
	event *ReadMessages,
) error {
	return nil
}

type (
	EventStartPartitionSession = topiclistenerinternal.PublicEventStartPartitionSession
	EventStopPartitionSession  = topiclistenerinternal.PublicEventStopPartitionSession
)

type PartitionSession = topicreadercommon.PublicPartitionSession

type OffsetsRange = topiclistenerinternal.PublicOffsetsRange

type StartPartitionSessionConfirm = topiclistenerinternal.PublicStartPartitionSessionConfirm
