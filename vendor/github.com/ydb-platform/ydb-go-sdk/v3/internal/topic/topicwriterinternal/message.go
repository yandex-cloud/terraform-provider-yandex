package topicwriterinternal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicwriter"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var errNoRawContent = xerrors.Wrap(errors.New("ydb: internal state error - no raw message content"))

type PublicMessage struct {
	SeqNo     int64
	CreatedAt time.Time
	Data      io.Reader
	Metadata  map[string][]byte

	tx tx.Transaction

	// partitioning at level message available by protocol, but doesn't available by current server implementation
	// the field hidden from public access for prevent runtime errors.
	// it will be published after implementation on server side.
	futurePartitioning PublicFuturePartitioning
}

// PublicFuturePartitioning will be published in feature, after server implementation completed.
type PublicFuturePartitioning struct {
	messageGroupID string
	partitionID    int64
	hasPartitionID bool
}

func (p PublicFuturePartitioning) ToRaw() rawtopicwriter.Partitioning {
	if p.hasPartitionID {
		return rawtopicwriter.NewPartitioningPartitionID(p.partitionID)
	}

	return rawtopicwriter.NewPartitioningMessageGroup(p.messageGroupID)
}

func NewPartitioningWithMessageGroupID(id string) PublicFuturePartitioning {
	return PublicFuturePartitioning{
		messageGroupID: id,
	}
}

func NewPartitioningWithPartitionID(id int64) PublicFuturePartitioning {
	return PublicFuturePartitioning{
		partitionID:    id,
		hasPartitionID: true,
	}
}

type messageWithDataContent struct {
	PublicMessage

	dataWasRead         bool
	hasRawContent       bool
	hasEncodedContent   bool
	metadataCached      bool
	bufCodec            rawtopiccommon.Codec
	bufEncoded          bytes.Buffer
	rawBuf              bytes.Buffer
	encoders            *MultiEncoder
	BufUncompressedSize int
}

func (m *messageWithDataContent) GetEncodedBytes(codec rawtopiccommon.Codec) ([]byte, error) {
	if codec == rawtopiccommon.CodecRaw {
		return m.getRawBytes()
	}

	return m.getEncodedBytes(codec)
}

func (m *messageWithDataContent) cacheMetadata() {
	if m.metadataCached {
		return
	}

	// ensure message metadata can't be changed by external code
	if len(m.Metadata) > 0 {
		ownCopy := make(map[string][]byte, len(m.Metadata))
		for key, val := range m.Metadata {
			ownCopy[key] = bytes.Clone(val)
		}
		m.Metadata = ownCopy
	} else {
		m.Metadata = nil
	}
	m.metadataCached = true
}

func (m *messageWithDataContent) CacheMessageData(codec rawtopiccommon.Codec) error {
	m.cacheMetadata()
	_, err := m.GetEncodedBytes(codec)

	return err
}

func (m *messageWithDataContent) encodeRawContent(codec rawtopiccommon.Codec) ([]byte, error) {
	if !m.hasRawContent {
		return nil, xerrors.WithStackTrace(errNoRawContent)
	}

	m.bufEncoded.Reset()

	_, err := m.encoders.EncodeBytes(codec, &m.bufEncoded, m.rawBuf.Bytes())
	if err != nil {
		return nil, xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: failed to compress message, codec '%v': %w",
			codec,
			err,
		)))
	}

	m.bufCodec = codec

	return m.bufEncoded.Bytes(), nil
}

func (m *messageWithDataContent) readDataToRawBuf() error {
	m.rawBuf.Reset()
	m.hasRawContent = true
	if m.Data != nil {
		writtenBytes, err := io.Copy(&m.rawBuf, m.Data)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		m.BufUncompressedSize = int(writtenBytes)
		m.Data = nil
	}

	return nil
}

func (m *messageWithDataContent) readDataToTargetCodec(codec rawtopiccommon.Codec) error {
	m.dataWasRead = true
	m.hasEncodedContent = true
	m.bufCodec = codec
	m.bufEncoded.Reset()

	reader := m.Data
	if reader == nil {
		reader = &bytes.Reader{}
	}

	bytesCount, err := m.encoders.Encode(codec, &m.bufEncoded, reader)
	if err != nil {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: failed compress message with codec '%v': %w",
			codec,
			err,
		)))
	}
	m.BufUncompressedSize = bytesCount
	m.Data = nil

	return nil
}

func (m *messageWithDataContent) getRawBytes() ([]byte, error) {
	if m.hasRawContent {
		return m.rawBuf.Bytes(), nil
	}
	if m.dataWasRead {
		return nil, xerrors.WithStackTrace(errNoRawContent)
	}

	if err := m.readDataToRawBuf(); err != nil {
		return nil, err
	}

	return m.rawBuf.Bytes(), nil
}

func (m *messageWithDataContent) getEncodedBytes(codec rawtopiccommon.Codec) ([]byte, error) {
	switch {
	case m.hasEncodedContent && m.bufCodec == codec:
		return m.bufEncoded.Bytes(), nil
	case m.hasRawContent:
		return m.encodeRawContent(codec)
	case m.dataWasRead:
		return nil, errNoRawContent
	default:
		err := m.readDataToTargetCodec(codec)
		if err != nil {
			return nil, err
		}

		return m.bufEncoded.Bytes(), nil
	}
}

func newMessageDataWithContent(
	message PublicMessage, //nolint:gocritic
	encoders *MultiEncoder,
) messageWithDataContent {
	return messageWithDataContent{
		PublicMessage: message,
		encoders:      encoders,
	}
}
