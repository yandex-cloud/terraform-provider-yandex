package topicreadercommon

import (
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
)

type RawTopicReaderStream interface {
	Recv() (rawtopicreader.ServerMessage, error)
	Send(msg rawtopicreader.ClientMessage) error
	CloseSend() error
}

var _ RawTopicReaderStream = &SyncedStream{}

type SyncedStream struct {
	m      sync.Mutex
	stream RawTopicReaderStream
}

func NewSyncedStream(stream RawTopicReaderStream) *SyncedStream {
	return &SyncedStream{
		stream: stream,
	}
}

func (s *SyncedStream) Recv() (rawtopicreader.ServerMessage, error) {
	// not need synced
	return s.stream.Recv()
}

func (s *SyncedStream) Send(msg rawtopicreader.ClientMessage) error {
	s.m.Lock()
	defer s.m.Unlock()

	return s.stream.Send(msg)
}

func (s *SyncedStream) CloseSend() error {
	s.m.Lock()
	defer s.m.Unlock()

	return s.stream.CloseSend()
}
