package topicreaderinternal

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xslices"
)

var errBatcherPopConcurency = xerrors.Wrap(errors.New("ydb: batch pop concurency, internal state error"))

type batcher struct {
	popInFlight    int64
	closeErr       error
	hasNewMessages empty.Chan

	m xsync.Mutex

	forceIgnoreMinRestrictionsOnNextMessagesBatch bool
	closed                                        bool
	closeChan                                     empty.Chan
	messages                                      batcherMessagesMap
	sessionsForFlush                              []*topicreadercommon.PartitionSession
}

func newBatcher() *batcher {
	return &batcher{
		messages:       make(batcherMessagesMap),
		closeChan:      make(empty.Chan),
		hasNewMessages: make(empty.Chan, 1),
	}
}

func (b *batcher) Close(err error) error {
	b.m.Lock()
	defer b.m.Unlock()

	if b.closed {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: batch closed already: %w", err))
	}

	b.closed = true
	b.closeErr = err
	close(b.closeChan)

	return nil
}

func (b *batcher) PushBatches(batches ...*topicreadercommon.PublicBatch) error {
	b.m.Lock()
	defer b.m.Unlock()
	if b.closed {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: push batch to closed batcher :%w", b.closeErr))
	}

	for _, batch := range batches {
		if err := b.addNeedLock(
			topicreadercommon.GetCommitRange(batch).PartitionSession,
			newBatcherItemBatch(batch),
		); err != nil {
			return err
		}
	}

	return nil
}

func (b *batcher) PushRawMessage(session *topicreadercommon.PartitionSession, m rawtopicreader.ServerMessage) error {
	b.m.Lock()
	defer b.m.Unlock()

	if b.closed {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: push raw message to closed batcher: %w", b.closeErr))
	}

	return b.addNeedLock(session, newBatcherItemRawMessage(m))
}

func (b *batcher) FlushPartitionSession(session *topicreadercommon.PartitionSession) {
	b.m.WithLock(func() {
		if b.closed {
			return
		}

		if _, ok := b.messages[session]; !ok {
			return
		}

		b.sessionsForFlush = append(b.sessionsForFlush, session)
	})
}

func (b *batcher) addNeedLock(session *topicreadercommon.PartitionSession, item batcherMessageOrderItem) error {
	var currentItems batcherMessageOrderItems
	var ok bool
	var err error
	if currentItems, ok = b.messages[session]; ok {
		if currentItems, err = currentItems.Append(item); err != nil {
			return err
		}
	} else {
		currentItems = batcherMessageOrderItems{item}
	}

	b.messages[session] = currentItems

	b.notifyAboutNewMessages()

	return nil
}

type batcherGetOptions struct {
	MinCount        int
	MaxCount        int
	rawMessagesOnly bool
}

func (o batcherGetOptions) cutBatchItemsHead(items batcherMessageOrderItems) (
	head batcherMessageOrderItem,
	rest batcherMessageOrderItems,
	ok bool,
) {
	notFound := func() (batcherMessageOrderItem, batcherMessageOrderItems, bool) {
		return batcherMessageOrderItem{}, batcherMessageOrderItems{}, false
	}
	if len(items) == 0 {
		return notFound()
	}

	if items[0].IsBatch() {
		if o.rawMessagesOnly {
			return notFound()
		}

		batchHead, batchRest, ok := o.splitBatch(items[0].Batch)

		if !ok {
			return notFound()
		}

		head = newBatcherItemBatch(batchHead)
		rest = items.ReplaceHeadItem(newBatcherItemBatch(batchRest))

		return head, rest, true
	}

	return items[0], items[1:], true
}

func (o batcherGetOptions) splitBatch(batch *topicreadercommon.PublicBatch) (
	head, rest *topicreadercommon.PublicBatch,
	ok bool,
) {
	notFound := func() (*topicreadercommon.PublicBatch, *topicreadercommon.PublicBatch, bool) {
		return nil, nil, false
	}

	if len(batch.Messages) < o.MinCount {
		return notFound()
	}

	if o.MaxCount == 0 {
		return batch, nil, true
	}

	head, rest = topicreadercommon.BatchCutMessages(batch, o.MaxCount)

	return head, rest, true
}

func (b *batcher) Pop(ctx context.Context, opts batcherGetOptions) (_ batcherMessageOrderItem, err error) {
	counter := atomic.AddInt64(&b.popInFlight, 1)
	defer atomic.AddInt64(&b.popInFlight, -1)

	if counter != 1 {
		return batcherMessageOrderItem{}, xerrors.WithStackTrace(errBatcherPopConcurency)
	}

	if err = ctx.Err(); err != nil {
		return batcherMessageOrderItem{}, err
	}

	for {
		var findRes batcherResultCandidate
		var closed bool

		b.m.WithLock(func() {
			closed = b.closed
			if closed {
				return
			}

			findRes = b.findNeedLock(opts)
			if findRes.Ok {
				b.applyNeedLock(&findRes)

				return
			}
		})
		if closed {
			return batcherMessageOrderItem{},
				xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf("ydb: try pop messages from closed batcher: %w", b.closeErr)))
		}
		if findRes.Ok {
			return findRes.Result, nil
		}

		// wait new messages for next iteration
		select {
		case <-b.hasNewMessages:
			// new iteration
		case <-b.closeChan:
			return batcherMessageOrderItem{},
				xerrors.WithStackTrace(
					fmt.Errorf(
						"ydb: batcher close while pop wait new messages: %w",
						b.closeErr,
					),
				)
		case <-ctx.Done():
			return batcherMessageOrderItem{}, ctx.Err()
		}
	}
}

func (b *batcher) notifyAboutNewMessages() {
	select {
	case b.hasNewMessages <- empty.Struct{}:
		// sent signal
	default:
		// signal already in progress
	}
}

type batcherResultCandidate struct {
	Key         *topicreadercommon.PartitionSession
	Result      batcherMessageOrderItem
	Rest        batcherMessageOrderItems
	WaiterIndex int
	Ok          bool
}

func newBatcherResultCandidate(
	key *topicreadercommon.PartitionSession,
	result batcherMessageOrderItem,
	rest batcherMessageOrderItems,
	ok bool,
) batcherResultCandidate {
	return batcherResultCandidate{
		Key:    key,
		Result: result,
		Rest:   rest,
		Ok:     ok,
	}
}

func (b *batcher) findNeedLock(filter batcherGetOptions) batcherResultCandidate {
	if len(b.messages) == 0 {
		return batcherResultCandidate{}
	}

	var batchResult batcherResultCandidate
	needBatchResult := true

	for _, session := range b.sessionsForFlush {
		items := b.messages[session]
		candidate := b.findNeedLockProcessItem(filter, session, items, true)
		if !candidate.Result.IsEmpty() {
			return candidate
		}
	}

	for session, items := range b.messages {
		candidate := b.findNeedLockProcessItem(filter, session, items, needBatchResult)

		if candidate.Result.IsRawMessage() {
			return candidate
		}

		if candidate.Result.IsBatch() {
			batchResult = candidate
			needBatchResult = false
		}
	}

	return batchResult
}

func (b *batcher) findNeedLockProcessItem(
	filter batcherGetOptions,
	k *topicreadercommon.PartitionSession,
	items batcherMessageOrderItems,
	needBatchResult bool,
) (
	result batcherResultCandidate,
) {
	rawMessageOpts := batcherGetOptions{rawMessagesOnly: true}
	head, rest, ok := rawMessageOpts.cutBatchItemsHead(items)
	if ok {
		return newBatcherResultCandidate(k, head, rest, true)
	}

	if needBatchResult {
		head, rest, ok = b.applyForceFlagToOptions(filter).cutBatchItemsHead(items)
		if !ok {
			return batcherResultCandidate{}
		}

		return newBatcherResultCandidate(k, head, rest, true)
	}

	return batcherResultCandidate{}
}

func (b *batcher) applyForceFlagToOptions(options batcherGetOptions) batcherGetOptions {
	if !b.forceIgnoreMinRestrictionsOnNextMessagesBatch {
		return options
	}

	res := options
	res.MinCount = 1

	return res
}

func (b *batcher) applyNeedLock(res *batcherResultCandidate) {
	if res.Rest.IsEmpty() && res.WaiterIndex >= 0 {
		delete(b.messages, res.Key)
		if len(b.sessionsForFlush) > 0 && b.sessionsForFlush[0] == res.Key {
			b.sessionsForFlush = xslices.Delete(b.sessionsForFlush, 0, 1)
		}
	} else {
		b.messages[res.Key] = res.Rest
	}

	if res.Result.IsBatch() {
		b.forceIgnoreMinRestrictionsOnNextMessagesBatch = false
	}
}

func (b *batcher) IgnoreMinRestrictionsOnNextPop() {
	b.m.Lock()
	defer b.m.Unlock()

	b.forceIgnoreMinRestrictionsOnNextMessagesBatch = true
	b.notifyAboutNewMessages()
}

type batcherMessagesMap map[*topicreadercommon.PartitionSession]batcherMessageOrderItems

type batcherMessageOrderItems []batcherMessageOrderItem

func (items batcherMessageOrderItems) Append(item batcherMessageOrderItem) (batcherMessageOrderItems, error) {
	if len(items) == 0 {
		return append(items, item), nil
	}

	lastItem := &items[len(items)-1]
	if item.IsBatch() && lastItem.IsBatch() {
		if resBatch, err := topicreadercommon.BatchAppend(lastItem.Batch, item.Batch); err == nil {
			lastItem.Batch = resBatch
		} else {
			return nil, err
		}

		return items, nil
	}

	return append(items, item), nil
}

func (items batcherMessageOrderItems) IsEmpty() bool {
	return len(items) == 0
}

func (items batcherMessageOrderItems) ReplaceHeadItem(item batcherMessageOrderItem) batcherMessageOrderItems {
	if item.IsEmpty() {
		return items[1:]
	}

	res := make(batcherMessageOrderItems, len(items))
	res[0] = item
	copy(res[1:], items[1:])

	return res
}

type batcherMessageOrderItem struct {
	Batch      *topicreadercommon.PublicBatch
	RawMessage rawtopicreader.ServerMessage
}

func newBatcherItemBatch(b *topicreadercommon.PublicBatch) batcherMessageOrderItem {
	return batcherMessageOrderItem{Batch: b}
}

func newBatcherItemRawMessage(b rawtopicreader.ServerMessage) batcherMessageOrderItem {
	return batcherMessageOrderItem{RawMessage: b}
}

func (item *batcherMessageOrderItem) IsBatch() bool {
	return !topicreadercommon.BatchIsEmpty(item.Batch)
}

func (item *batcherMessageOrderItem) IsRawMessage() bool {
	return item.RawMessage != nil
}

func (item *batcherMessageOrderItem) IsEmpty() bool {
	return item.RawMessage == nil && topicreadercommon.BatchIsEmpty(item.Batch)
}
