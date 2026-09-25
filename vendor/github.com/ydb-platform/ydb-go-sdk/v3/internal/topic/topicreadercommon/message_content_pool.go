package topicreadercommon

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

// Pool is interface for sync.Pool and may be extended by follow to original type
//
//go:generate mockgen -destination=pool_interface_mock_test.go --typed -write_package_comment=false -package=topicreadercommon github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon Pool
type Pool interface {
	Get() interface{}
	Put(x interface{})
}

// CallbackWithMessageContentFunc is callback function for work with message content
// data bytes MUST NOT be used after f returned
// if you need content longer - copy content to other slice
type CallbackWithMessageContentFunc func(data []byte) error

const (
	minInitializeBufferSize = bytes.MinRead * 2
	maxInitialBufferSize    = 1024 * 1024 * 50 // protection from bad UncompressedSize from stream
)

var globalReadMessagePool = &sync.Pool{}

func callbackOnReaderContent(
	p Pool,
	reader io.Reader,
	estinamedSize int,
	consumer PublicMessageContentUnmarshaler,
) error {
	var buf *bytes.Buffer
	var ok bool
	if buf, ok = p.Get().(*bytes.Buffer); !ok {
		buf = &bytes.Buffer{}
	}
	defer func() {
		buf.Reset()
		p.Put(buf)
	}()

	// + bytes.MinRead need for prevent additional allocation for read io.EOF after read message content
	targetSize := estinamedSize + bytes.MinRead
	switch {
	case targetSize < minInitializeBufferSize:
		buf.Grow(minInitializeBufferSize)
	case targetSize > maxInitialBufferSize:
		buf.Grow(maxInitialBufferSize)
	default:
		buf.Grow(targetSize)
	}

	if _, err := buf.ReadFrom(reader); err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: error read from data reader: %w", err))
	}

	if err := consumer.UnmarshalYDBTopicMessage(buf.Bytes()); err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: error unmarshal data: %w", err))
	}

	return nil
}
