package topicreadercommon

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/background"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var (
	ErrCommitDisabled             = xerrors.Wrap(errors.New("ydb: commits disabled"))
	ErrWrongCommitOrderInSyncMode = xerrors.Wrap(errors.New("ydb: wrong commit order in sync mode. It means you skipped committing some messages. Out-of-order commits are OK for async mode - you can commit the messages later. But im sync mode, it means deadlock: the code waits for a commit ack from the server, but the server waits for the commits of the skipped message. In sync mode, ensure that you commit messages/batches in the same order as you read them")) //nolint:lll
)

type SendMessageToServerFunc func(msg rawtopicreader.ClientMessage) error

type PublicCommitMode int

const (
	CommitModeAsync PublicCommitMode = iota // default
	CommitModeNone
	CommitModeSync
)

func (m PublicCommitMode) CommitsEnabled() bool {
	return m != CommitModeNone
}

type Committer struct {
	BufferTimeLagTrigger time.Duration // 0 mean no additional time lag
	BufferCountTrigger   int

	send SendMessageToServerFunc
	mode PublicCommitMode

	clock            clockwork.Clock
	commitLoopSignal empty.Chan
	backgroundWorker background.Worker
	tracer           *trace.Topic

	m       xsync.Mutex
	waiters []commitWaiter
	commits CommitRanges
}

func NewCommitterStopped(
	tracer *trace.Topic,
	lifeContext context.Context, //nolint:revive
	mode PublicCommitMode,
	send SendMessageToServerFunc,
) *Committer {
	res := &Committer{
		mode:             mode,
		clock:            clockwork.NewRealClock(),
		send:             send,
		backgroundWorker: *background.NewWorker(lifeContext, "ydb-topic-reader-committer"),
		tracer:           tracer,
	}
	res.initChannels()

	return res
}

func (c *Committer) initChannels() {
	c.commitLoopSignal = make(empty.Chan, 1)
}

func (c *Committer) Start() {
	c.backgroundWorker.Start("commit pusher", c.pushCommitsLoop)
}

func (c *Committer) Close(ctx context.Context, err error) error {
	return c.backgroundWorker.Close(ctx, err)
}

func (c *Committer) Commit(ctx context.Context, commitRange CommitRange) error {
	if !c.mode.CommitsEnabled() {
		return ErrCommitDisabled
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	waiter, err := c.pushCommit(commitRange)
	if err != nil {
		return err
	}

	return c.waitCommitAck(ctx, waiter)
}

func (c *Committer) pushCommit(commitRange CommitRange) (commitWaiter, error) {
	var resErr error
	waiter := newCommitWaiter(commitRange.PartitionSession, commitRange.CommitOffsetEnd)
	c.m.WithLock(func() {
		if err := c.backgroundWorker.Context().Err(); err != nil {
			resErr = err

			return
		}

		c.commits.Append(&commitRange)
		if c.mode == CommitModeSync {
			c.addWaiterNeedLock(waiter)
		}
	})

	select {
	case c.commitLoopSignal <- struct{}{}:
	default:
	}

	return waiter, resErr
}

func (c *Committer) pushCommitsLoop(ctx context.Context) {
	for {
		c.waitSendTrigger(ctx)

		var commits CommitRanges
		c.m.WithLock(func() {
			commits = c.commits
			c.commits = NewCommitRangesWithCapacity(commits.Len() * 2) //nolint:mnd
		})

		if commits.Len() == 0 && c.backgroundWorker.Context().Err() != nil {
			// committer closed with empty buffer - target close state
			return
		}

		// all ranges already committed of prev iteration
		if commits.Len() == 0 {
			continue
		}

		commits.Optimize()

		onDone := trace.TopicOnReaderSendCommitMessage(
			c.tracer,
			&ctx,
			&commits,
		)
		err := c.send(commits.ToRawMessage())
		onDone(err)

		if err != nil {
			_ = c.backgroundWorker.Close(ctx, err)
		}
	}
}

func (c *Committer) waitSendTrigger(ctx context.Context) {
	ctxDone := ctx.Done()
	select {
	case <-ctxDone:
		return
	case <-c.commitLoopSignal:
	}

	// In sync mode, ignore time lag trigger and send immediately,
	// because we need to wait for commit ack from the server and can't get
	// more commit messages until the ack is received.
	if c.mode == CommitModeSync {
		return
	}

	if c.BufferTimeLagTrigger == 0 {
		return
	}

	bufferTimeLagTriggerTimer := c.clock.NewTimer(c.BufferTimeLagTrigger)
	defer bufferTimeLagTriggerTimer.Stop()

	finish := bufferTimeLagTriggerTimer.Chan()
	if c.BufferCountTrigger == 0 {
		select {
		case <-ctxDone:
		case <-finish:
		}

		return
	}

	for {
		var commitsLen int
		c.m.WithLock(func() {
			commitsLen = c.commits.Len()
		})
		if commitsLen >= c.BufferCountTrigger {
			return
		}

		select {
		case <-ctxDone:
			return
		case <-finish:
			return
		case <-c.commitLoopSignal:
			// check count on next loop iteration
		}
	}
}

func (c *Committer) waitCommitAck(ctx context.Context, waiter commitWaiter) error {
	if c.mode != CommitModeSync {
		return nil
	}

	defer c.m.WithLock(func() {
		c.removeWaiterByIDNeedLock(waiter.ID)
	})
	if waiter.checkCondition(waiter.Session, waiter.Session.CommittedOffset()) {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-waiter.Session.Context().Done():
		return ErrPublicCommitSessionToExpiredSession
	case <-waiter.Committed:
		return nil
	}
}

func (c *Committer) OnCommitNotify(session *PartitionSession, offset rawtopiccommon.Offset) {
	c.m.WithLock(func() {
		for i := range c.waiters {
			waiter := c.waiters[i]
			if waiter.checkCondition(session, offset) {
				select {
				case waiter.Committed <- struct{}{}:
				default:
				}
			}
		}
	})
}

func (c *Committer) addWaiterNeedLock(waiter commitWaiter) {
	c.waiters = append(c.waiters, waiter)
}

func (c *Committer) removeWaiterByIDNeedLock(id int64) {
	newWaiters := c.waiters[:0]
	for i := range c.waiters {
		if c.waiters[i].ID == id {
			continue
		}

		newWaiters = append(newWaiters, c.waiters[i])
	}
	c.waiters = newWaiters
}

type commitWaiter struct {
	ID        int64
	Session   *PartitionSession
	EndOffset rawtopiccommon.Offset
	Committed empty.Chan
}

func (w *commitWaiter) checkCondition(
	session *PartitionSession,
	offset rawtopiccommon.Offset,
) (finished bool) {
	return session == w.Session && offset >= w.EndOffset
}

var commitWaiterLastID int64

func newCommitWaiter(session *PartitionSession, endOffset rawtopiccommon.Offset) commitWaiter {
	id := atomic.AddInt64(&commitWaiterLastID, 1)

	return commitWaiter{
		ID:        id,
		Session:   session,
		EndOffset: endOffset,
		Committed: make(empty.Chan, 1),
	}
}
