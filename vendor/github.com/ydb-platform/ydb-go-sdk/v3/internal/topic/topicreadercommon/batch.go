package topicreadercommon

import (
	"context"
	"errors"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	errBadSessionWhileMessageBatchCreate       = xerrors.Wrap(errors.New("ydb: bad session while messages batch create"))        //nolint:lll
	errBadMessageOffsetWhileMessageBatchCreate = xerrors.Wrap(errors.New("ydb: bad message offset while messages batch create")) //nolint:lll
)

// PublicBatch is ordered group of message from one partition
type PublicBatch struct {
	empty.DoNotCopy

	Messages []*PublicMessage

	commitRange CommitRange // от всех сообщений батча
}

func NewBatch(session *PartitionSession, messages []*PublicMessage) (*PublicBatch, error) {
	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		if msg.commitRange.PartitionSession == nil {
			msg.commitRange.PartitionSession = session
		}
		if session != msg.commitRange.PartitionSession {
			return nil, xerrors.WithStackTrace(errBadSessionWhileMessageBatchCreate)
		}

		if i == 0 {
			continue
		}

		prev := messages[i-1]
		if prev.commitRange.CommitOffsetEnd != msg.commitRange.CommitOffsetStart {
			return nil, xerrors.WithStackTrace(errBadMessageOffsetWhileMessageBatchCreate)
		}
	}

	commitRange := CommitRange{
		PartitionSession: session,
	}
	if len(messages) > 0 {
		commitRange.CommitOffsetStart = messages[0].commitRange.CommitOffsetStart
		commitRange.CommitOffsetEnd = messages[len(messages)-1].commitRange.CommitOffsetEnd
	}

	return &PublicBatch{
		Messages:    messages,
		commitRange: commitRange,
	}, nil
}

func NewBatchFromStream(
	decoders DecoderMap,
	session *PartitionSession,
	sb rawtopicreader.Batch, //nolint:gocritic
) (*PublicBatch, error) {
	messages := make([]*PublicMessage, len(sb.MessageData))
	prevOffset := session.LastReceivedMessageOffset()
	for i := range sb.MessageData {
		sMess := &sb.MessageData[i]

		dstMess := &PublicMessage{}
		messages[i] = dstMess
		dstMess.CreatedAt = sMess.CreatedAt
		dstMess.MessageGroupID = sMess.MessageGroupID
		dstMess.Offset = sMess.Offset.ToInt64()
		dstMess.SeqNo = sMess.SeqNo
		dstMess.WrittenAt = sb.WrittenAt
		dstMess.ProducerID = sb.ProducerID
		dstMess.WriteSessionMetadata = sb.WriteSessionMeta

		dstMess.rawDataLen = len(sMess.Data)
		dstMess.data = createReader(decoders, sb.Codec, sMess.Data)
		dstMess.UncompressedSize = int(sMess.UncompressedSize)

		dstMess.commitRange.PartitionSession = session
		dstMess.commitRange.CommitOffsetStart = prevOffset + 1
		dstMess.commitRange.CommitOffsetEnd = sMess.Offset + 1

		if len(sMess.MetadataItems) > 0 {
			dstMess.Metadata = make(map[string][]byte, len(sMess.MetadataItems))
			for metadataIndex := range sMess.MetadataItems {
				dstMess.Metadata[sMess.MetadataItems[metadataIndex].Key] = sMess.MetadataItems[metadataIndex].Value
			}
		}

		prevOffset = sMess.Offset
	}

	session.SetLastReceivedMessageOffset(prevOffset)

	return NewBatch(session, messages)
}

// Context is cancelled when code should stop to process messages batch
// for example - lost connection to server or receive stop partition signal without graceful flag
func (m *PublicBatch) Context() context.Context {
	return m.commitRange.PartitionSession.Context()
}

// Topic is path of source topic of the messages in the batch
func (m *PublicBatch) Topic() string {
	return m.partitionSession().Topic
}

// PartitionID of messages in the batch
func (m *PublicBatch) PartitionID() int64 {
	return m.partitionSession().PartitionID
}

func (m *PublicBatch) partitionSession() *PartitionSession {
	return m.commitRange.PartitionSession
}

func (m *PublicBatch) getCommitRange() PublicCommitRange {
	return m.commitRange.getCommitRange()
}

func splitBytesByMessagesInBatches(batches []*PublicBatch, totalBytesCount int) error {
	restBytes := totalBytesCount

	cutBytes := func(want int) int {
		switch {
		case restBytes == 0:
			return 0
		case want >= restBytes:
			res := restBytes
			restBytes = 0

			return res
		default:
			restBytes -= want

			return want
		}
	}

	messagesCount := 0
	for batchIndex := range batches {
		messagesCount += len(batches[batchIndex].Messages)
		for messageIndex := range batches[batchIndex].Messages {
			message := batches[batchIndex].Messages[messageIndex]
			message.bufferBytesAccount = cutBytes(batches[batchIndex].Messages[messageIndex].rawDataLen)
		}
	}

	if messagesCount == 0 {
		if totalBytesCount == 0 {
			return nil
		}

		return xerrors.NewWithIssues("ydb: can't split bytes to zero length messages count")
	}

	overheadPerMessage := restBytes / messagesCount
	overheadRemind := restBytes % messagesCount

	for batchIndex := range batches {
		for messageIndex := range batches[batchIndex].Messages {
			msg := batches[batchIndex].Messages[messageIndex]

			msg.bufferBytesAccount += cutBytes(overheadPerMessage)
			if overheadRemind > 0 {
				msg.bufferBytesAccount += cutBytes(1)
			}
		}
	}

	return nil
}

func BatchAppend(original, appended *PublicBatch) (*PublicBatch, error) {
	var res *PublicBatch
	if original == nil {
		res = &PublicBatch{}
	} else {
		res = original
	}

	if res.commitRange.PartitionSession != appended.commitRange.PartitionSession {
		return nil, xerrors.WithStackTrace(errors.New("ydb: bad partition session for merge"))
	}

	if res.commitRange.CommitOffsetEnd != appended.commitRange.CommitOffsetStart {
		return nil, xerrors.WithStackTrace(errors.New("ydb: bad offset interval for merge"))
	}

	res.Messages = append(res.Messages, appended.Messages...)
	res.commitRange.CommitOffsetEnd = appended.commitRange.CommitOffsetEnd

	return res, nil
}

func BatchCutMessages(b *PublicBatch, count int) (head, rest *PublicBatch) {
	switch {
	case count == 0:
		return nil, b
	case count >= len(b.Messages):
		return b, nil
	default:
		// slice[0:count:count] - limit slice capacity and prevent overwrite rest by append messages to head
		// explicit 0 need for prevent typos, when type slice[count:count] instead of slice[:count:count]
		head, _ = NewBatch(b.commitRange.PartitionSession, b.Messages[0:count:count])
		rest, _ = NewBatch(b.commitRange.PartitionSession, b.Messages[count:])

		return head, rest
	}
}

func BatchIsEmpty(b *PublicBatch) bool {
	return b == nil || len(b.Messages) == 0
}

func BatchGetPartitionSession(item *PublicBatch) *PartitionSession {
	return item.partitionSession()
}

func BatchSetCommitRangeForTest(b *PublicBatch, commitRange CommitRange) *PublicBatch {
	b.commitRange = commitRange

	return b
}
