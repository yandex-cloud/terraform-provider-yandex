package topicwriterinternal

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicwriter"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
)

var (
	errCloseClosedMessageQueue   = xerrors.Wrap(errors.New("ydb: close closed message queue"))
	errAckOnClosedMessageQueue   = xerrors.Wrap(errors.New("ydb: ack on closed message queue"))
	errGetMessageFromClosedQueue = xerrors.Wrap(errors.New("ydb: get message from closed message queue"))
	errAddUnorderedMessages      = xerrors.Wrap(errors.New("ydb: add unordered messages"))
	errAckUnexpectedMessage      = xerrors.Wrap(errors.New("ydb: ack unexpected message"))
)

const (
	//nolint:mnd
	intSize = 32 << (^uint(0) >> 63) // copy from math package for use in go <= 1.16
	maxInt  = 1<<(intSize-1) - 1     // copy from math package for use in go <= 1.16
	minInt  = -1 << (intSize - 1)    // copy from math package for use in go <= 1.16

	minPositiveIndexWhichOrderLessThenNegative = maxInt / 2
)

type messageQueue struct {
	OnAckReceived func(count int)

	hasNewMessages    empty.Chan
	closedErr         error
	acksReceivedEvent xsync.EventBroadcast

	m                         xsync.RWMutex
	stopReceiveMessagesReason error
	closed                    bool
	closedChan                empty.Chan
	lastWrittenIndex          int
	lastSentIndex             int
	lastSeqNo                 int64

	messagesByOrder map[int]messageWithDataContent
	seqNoToOrderID  map[int64]int
}

func newMessageQueue() messageQueue {
	return messageQueue{
		messagesByOrder: make(map[int]messageWithDataContent),
		seqNoToOrderID:  make(map[int64]int),
		hasNewMessages:  make(empty.Chan, 1),
		closedChan:      make(empty.Chan),
		lastSeqNo:       -1,
	}
}

func (q *messageQueue) AddMessages(messages []messageWithDataContent) error {
	_, err := q.addMessages(messages, false)

	return err
}

func (q *messageQueue) AddMessagesWithWaiter(messages []messageWithDataContent) (
	waiter MessageQueueAckWaiter,
	err error,
) {
	return q.addMessages(messages, true)
}

func (q *messageQueue) addMessages(messages []messageWithDataContent, needWaiter bool) (
	waiter MessageQueueAckWaiter,
	err error,
) {
	q.m.Lock()
	defer q.m.Unlock()

	if q.stopReceiveMessagesReason != nil {
		return waiter, xerrors.WithStackTrace(
			fmt.Errorf("ydb: add message to closed message queue: %w", q.stopReceiveMessagesReason),
		)
	}

	if err := q.checkNewMessagesBeforeAddNeedLock(messages); err != nil {
		return waiter, err
	}

	for i := range messages {
		messageIndex := q.addMessageNeedLock(messages[i])

		if needWaiter {
			waiter.AddWaitIndex(messageIndex)
		}
	}

	q.notifyNewMessages()

	return waiter, nil
}

func (q *messageQueue) notifyNewMessages() {
	select {
	case q.hasNewMessages <- empty.Struct{}:
		// pass
	default:
	}
}

func (q *messageQueue) checkNewMessagesBeforeAddNeedLock(messages []messageWithDataContent) error {
	if len(messages) == 0 {
		return nil
	}

	checkedSeqNo := q.lastSeqNo
	for i := range messages {
		if messages[i].SeqNo <= checkedSeqNo {
			return xerrors.WithStackTrace(errAddUnorderedMessages)
		}
		checkedSeqNo = messages[i].SeqNo
	}

	return nil
}

func (q *messageQueue) addMessageNeedLock(
	mess messageWithDataContent, //nolint:gocritic
) (messageIndex int) {
	q.lastWrittenIndex++
	messageIndex = q.lastWrittenIndex

	if messageIndex == minInt {
		q.ensureNoSmallIntIndexes()
	}

	if _, ok := q.messagesByOrder[messageIndex]; ok {
		panic(fmt.Errorf("ydb: bad internal state os message queue - already exists with index: %v", messageIndex))
	}

	q.messagesByOrder[messageIndex] = mess
	q.seqNoToOrderID[mess.SeqNo] = messageIndex
	q.lastSeqNo = mess.SeqNo

	return messageIndex
}

func (q *messageQueue) AcksReceived(acks []rawtopicwriter.WriteAck) error {
	ackReceivedCounter := 0
	q.m.Lock()
	defer func() {
		q.m.Unlock()

		if q.OnAckReceived != nil {
			q.OnAckReceived(ackReceivedCounter)
		}
	}()
	if q.closed {
		return xerrors.WithStackTrace(errAckOnClosedMessageQueue)
	}

	for i := range acks {
		if err := q.ackReceivedNeedLock(acks[i].SeqNo); err != nil {
			return err
		}
		ackReceivedCounter++
	}

	q.acksReceivedEvent.Broadcast()

	return nil
}

func (q *messageQueue) ackReceivedNeedLock(seqNo int64) error {
	orderID, ok := q.seqNoToOrderID[seqNo]
	if !ok {
		return xerrors.WithStackTrace(errAckUnexpectedMessage)
	}

	delete(q.seqNoToOrderID, seqNo)
	delete(q.messagesByOrder, orderID)

	return nil
}

func (q *messageQueue) StopAddNewMessages(reason error) {
	q.m.Lock()
	defer q.m.Unlock()

	q.stopAddNewMessagesNeedLock(reason)
}

func (q *messageQueue) stopAddNewMessagesNeedLock(reason error) {
	if q.stopReceiveMessagesReason == nil {
		q.stopReceiveMessagesReason = reason
	}
}

func (q *messageQueue) Close(err error) error {
	isFirstTimeClosed := false
	q.m.Lock()
	defer func() {
		q.m.Unlock()

		// release all
		if isFirstTimeClosed && q.OnAckReceived != nil {
			q.OnAckReceived(len(q.seqNoToOrderID))
		}
	}()

	q.stopAddNewMessagesNeedLock(err)

	if q.closed {
		return xerrors.WithStackTrace(errCloseClosedMessageQueue)
	}
	isFirstTimeClosed = true

	q.closed = true
	q.closedErr = err
	close(q.closedChan)

	return nil
}

func (q *messageQueue) ensureNoSmallIntIndexes() {
	for k := range q.messagesByOrder {
		if k >= 0 && k < minPositiveIndexWhichOrderLessThenNegative {
			panic("ydb: message queue has bad state - overflow or has very old element")
		}
	}
}

// GetMessagesForSend one or more messages for send
// it blocked until context cancelled of have least one message for send
func (q *messageQueue) GetMessagesForSend(ctx context.Context) ([]messageWithDataContent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var closed bool
	q.m.WithLock(func() {
		closed = q.closed
	})
	if closed {
		return nil, xerrors.WithStackTrace(errGetMessageFromClosedQueue)
	}

	for {
		res := q.getMessagesForSendWithLock()
		if len(res) != 0 {
			return res, nil
		}

		select {
		case <-ctx.Done():
			return nil, xerrors.WithStackTrace(ctx.Err())
		case <-q.hasNewMessages:
			// pass
		case <-q.closedChan:
			return nil, xerrors.WithStackTrace(fmt.Errorf("ydb: message queue closed with: %w", q.closedErr))
		}
	}
}

func (q *messageQueue) ResetSentProgress() {
	q.m.Lock()
	defer q.m.Unlock()

	minKey := q.lastWrittenIndex
	for k := range q.messagesByOrder {
		if isFirstCycledIndexLess(k, minKey) {
			minKey = k
		}
	}

	q.lastSentIndex = minKey - 1
	q.notifyNewMessages()
}

func (q *messageQueue) getMessagesForSendWithLock() []messageWithDataContent {
	q.m.Lock()
	defer q.m.Unlock()

	if q.lastWrittenIndex == q.lastSentIndex {
		return nil
	}

	var res []messageWithDataContent

	// use  "!=" stop instead of  "<" - for work with negative indexes after overflow
	for q.lastWrittenIndex != q.lastSentIndex {
		q.lastSentIndex++

		// msg may be unexisted if it already has ack from server
		// pass
		if msg, ok := q.messagesByOrder[q.lastSentIndex]; ok {
			res = append(res, msg)
		}
	}

	return res
}

func (q *messageQueue) Wait(ctx context.Context, waiter MessageQueueAckWaiter) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ctxDone := ctx.Done()
	for {
		ackReceived := q.acksReceivedEvent.Waiter()

		hasWaited := false
		q.m.WithRLock(func() {
			for len(waiter.sequenseNumbers) > 0 {
				checkMessageIndex := waiter.sequenseNumbers[0]
				if _, ok := q.messagesByOrder[checkMessageIndex]; ok {
					hasWaited = true

					return
				}
				waiter.sequenseNumbers = waiter.sequenseNumbers[1:]
			}
		})

		if !hasWaited {
			return nil
		}

		select {
		case <-ctxDone:
			return ctx.Err()
		case <-q.closedChan:
			return q.closedErr
		case <-ackReceived.Done():
			// pass next iteration
		}
	}
}

// WaitLastWritten waits for last written message gets ack.
func (q *messageQueue) WaitLastWritten(ctx context.Context) error {
	var lastIndex int
	q.m.WithRLock(func() {
		lastIndex = q.lastWrittenIndex
	})

	return q.Wait(ctx, MessageQueueAckWaiter{sequenseNumbers: []int{lastIndex}})
}

type MessageQueueAckWaiter struct {
	sequenseNumbers []int
}

func (m *MessageQueueAckWaiter) AddWaitIndex(index int) {
	m.sequenseNumbers = append(m.sequenseNumbers, index)
}

// sortMessageQueueIndexes deprecated
func sortMessageQueueIndexes(keys []int) {
	sort.Ints(keys)
	// check index overflow
	if len(keys) > 0 && keys[0] < 0 && keys[len(keys)-1] > 0 {
		sort.Slice(keys, func(i, k int) bool {
			return isFirstCycledIndexLess(keys[i], keys[k])
		})
	}
}

func isFirstCycledIndexLess(first, second int) bool {
	switch {
	case first > minPositiveIndexWhichOrderLessThenNegative && second < 0:
		return true
	case first < 0 && second > minPositiveIndexWhichOrderLessThenNegative:
		return false
	default:
		return first < second
	}
}
