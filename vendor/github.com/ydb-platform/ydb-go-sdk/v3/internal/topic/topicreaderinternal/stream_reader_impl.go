package topicreaderinternal

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"runtime/pprof"
	"sync/atomic"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/background"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var (
	errCommitWithNilPartitionSession = xerrors.Wrap(errors.New("ydb: commit with nil partition session"))
	errUnexpectedEmptyConsumerName   = xerrors.Wrap(errors.New("ydb: create ydb reader with empty consumer name. Set one of: consumer name or option WithReaderWithoutConsumer")) //nolint:lll
	errCantCommitWithoutConsumer     = xerrors.Wrap(errors.New("ydb: reader can't commit messages without consumer"))
	errBufferSize                    = xerrors.Wrap(errors.New("ydb: buffer of topic reader must be greater than zero, see option topicoptions.WithReaderBufferSizeBytes")) //nolint:lll
	errTopicSelectorsEmpty           = xerrors.Wrap(errors.New("ydb: topic selector for topic reader is empty, see arguments on topic starts"))                             //nolint:lll
)

var clientSessionCounter atomic.Int64

type partitionSessionID = rawtopicreader.PartitionSessionID

type topicStreamReaderImpl struct {
	cfg    topicStreamReaderConfig
	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc

	topicClient         TopicClient
	freeBytes           chan int
	restBufferSizeBytes atomic.Int64
	sessionController   topicreadercommon.PartitionSessionStorage
	backgroundWorkers   background.Worker

	rawMessagesFromBuffer chan rawtopicreader.ServerMessage

	batcher   *batcher
	committer *topicreadercommon.Committer

	stream           topicreadercommon.RawTopicReaderStream
	readConnectionID string
	readerID         int64

	m       xsync.RWMutex
	err     error
	started bool
	closed  bool
}

type topicStreamReaderConfig struct {
	CommitterBatchTimeLag           time.Duration
	CommitterBatchCounterTrigger    int
	BaseContext                     context.Context //nolint:containedctx
	BufferSizeProtoBytes            int
	Cred                            credentials.Credentials
	CredUpdateInterval              time.Duration
	Consumer                        string
	ReadWithoutConsumer             bool
	ReadSelectors                   []*topicreadercommon.PublicReadSelector
	Trace                           *trace.Topic
	GetPartitionStartOffsetCallback PublicGetPartitionStartOffsetFunc
	CommitMode                      topicreadercommon.PublicCommitMode
	Decoders                        topicreadercommon.DecoderMap
	EnableSplitMergeSupport         bool
}

func newTopicStreamReaderConfig() topicStreamReaderConfig {
	return topicStreamReaderConfig{
		BaseContext:             context.Background(),
		BufferSizeProtoBytes:    topicreadercommon.DefaultBufferSize,
		Cred:                    credentials.NewAnonymousCredentials(),
		CredUpdateInterval:      time.Hour,
		CommitMode:              topicreadercommon.CommitModeAsync,
		CommitterBatchTimeLag:   time.Second,
		Decoders:                topicreadercommon.NewDecoderMap(),
		Trace:                   &trace.Topic{},
		EnableSplitMergeSupport: true,
	}
}

func (cfg *topicStreamReaderConfig) Validate() []error {
	var validateErrors []error

	if cfg.Consumer != "" && cfg.ReadWithoutConsumer {
		validateErrors = append(validateErrors, errSetConsumerAndNoConsumer)
	}
	if cfg.Consumer == "" && !cfg.ReadWithoutConsumer {
		validateErrors = append(validateErrors, errUnexpectedEmptyConsumerName)
	}
	if cfg.ReadWithoutConsumer && cfg.CommitMode != topicreadercommon.CommitModeNone {
		validateErrors = append(validateErrors, errCantCommitWithoutConsumer)
	}
	if cfg.BufferSizeProtoBytes <= 0 {
		validateErrors = append(validateErrors, errBufferSize)
	}
	if len(cfg.ReadSelectors) == 0 {
		validateErrors = append(validateErrors, errTopicSelectorsEmpty)
	}

	return validateErrors
}

func newTopicStreamReader(
	client TopicClient,
	readerID int64,
	stream topicreadercommon.RawTopicReaderStream,
	cfg topicStreamReaderConfig, //nolint:gocritic
) (_ *topicStreamReaderImpl, err error) {
	defer func() {
		if err != nil {
			_ = stream.CloseSend()
		}
	}()

	reader := newTopicStreamReaderStopped(client, readerID, stream, cfg)
	if err = reader.initSession(); err != nil {
		return nil, err
	}
	if err = reader.startBackgroundWorkers(); err != nil {
		return nil, err
	}

	return reader, nil
}

func newTopicStreamReaderStopped(
	client TopicClient,
	readerID int64,
	stream topicreadercommon.RawTopicReaderStream,
	cfg topicStreamReaderConfig, //nolint:gocritic
) *topicStreamReaderImpl {
	labeledContext := pprof.WithLabels(cfg.BaseContext, pprof.Labels("base-context", "topic-stream-reader"))
	stopPump, cancel := xcontext.WithCancel(labeledContext)

	readerConnectionID, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		readerConnectionID = big.NewInt(-1)
	}

	res := &topicStreamReaderImpl{
		cfg:                   cfg,
		ctx:                   stopPump,
		topicClient:           client,
		freeBytes:             make(chan int, 1),
		stream:                topicreadercommon.NewSyncedStream(stream),
		cancel:                cancel,
		batcher:               newBatcher(),
		readConnectionID:      "preinitID-" + readerConnectionID.String(),
		readerID:              readerID,
		rawMessagesFromBuffer: make(chan rawtopicreader.ServerMessage, 1),
	}

	res.backgroundWorkers = *background.NewWorker(stopPump, "topic-reader-stream-background")

	res.committer = topicreadercommon.NewCommitterStopped(cfg.Trace, labeledContext, cfg.CommitMode, res.send)
	res.committer.BufferTimeLagTrigger = cfg.CommitterBatchTimeLag
	res.committer.BufferCountTrigger = cfg.CommitterBatchCounterTrigger
	res.freeBytes <- cfg.BufferSizeProtoBytes

	return res
}

func (r *topicStreamReaderImpl) WaitInit(_ context.Context) error {
	if !r.started {
		return errors.New("not started: can be started only after initialize from constructor")
	}

	return nil
}

func (r *topicStreamReaderImpl) TopicOnReaderStart(consumer string, err error) {
	logCtx := r.cfg.BaseContext
	trace.TopicOnReaderStart(r.cfg.Trace, &logCtx, r.readerID, consumer, err)
}

func (r *topicStreamReaderImpl) PopMessagesBatchTx(
	ctx context.Context,
	tx tx.Transaction,
	opts ReadMessageBatchOptions,
) (_ *topicreadercommon.PublicBatch, resErr error) {
	defer func() {
		logCtx := r.cfg.BaseContext
		trace.TopicOnReaderStreamPopBatchTx(
			r.cfg.Trace,
			&logCtx,
			r.readerID,
			r.readConnectionID,
			tx.SessionID(),
			tx,
		)(resErr)
	}()

	batch, err := r.ReadMessageBatch(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err = r.commitWithTransaction(ctx, tx, batch); err == nil {
		return batch, nil
	}

	return nil, err
}

func (r *topicStreamReaderImpl) commitWithTransaction(
	ctx context.Context,
	tx tx.Transaction,
	batch *topicreadercommon.PublicBatch,
) error {
	if err := tx.UnLazy(ctx); err != nil {
		return fmt.Errorf("ydb: failed to materialize transaction: %w", err)
	}

	req := r.createUpdateOffsetRequest(ctx, batch, tx)
	updateOffsetInTransactionErr := retry.Retry(ctx, func(ctx context.Context) (err error) {
		logCtx := r.cfg.BaseContext
		onDone := trace.TopicOnReaderUpdateOffsetsInTransaction(
			r.cfg.Trace,
			&logCtx,
			r.readerID,
			r.readConnectionID,
			tx.SessionID(),
			tx,
		)
		defer func() {
			onDone(err)
		}()

		err = r.topicClient.UpdateOffsetsInTransaction(ctx, req)

		return err
	})
	if updateOffsetInTransactionErr == nil {
		r.addOnTransactionCompletedHandler(ctx, tx, batch, updateOffsetInTransactionErr)
	} else {
		_ = retry.Retry(ctx, func(ctx context.Context) (err error) {
			defer func() {
				logCtx := r.cfg.BaseContext
				trace.TopicOnReaderTransactionRollback(
					r.cfg.Trace,
					&logCtx,
					r.readerID,
					r.readConnectionID,
					tx.SessionID(),
					tx,
				)(err)
			}()

			return tx.Rollback(ctx)
		})

		_ = r.CloseWithError(xcontext.ValueOnly(ctx), xerrors.WithStackTrace(xerrors.Retryable(
			fmt.Errorf("ydb: failed add topic offsets in transaction: %w", updateOffsetInTransactionErr),
		)))

		return updateOffsetInTransactionErr
	}

	return nil
}

func (r *topicStreamReaderImpl) addOnTransactionCompletedHandler(
	ctx context.Context,
	tx tx.Transaction,
	batch *topicreadercommon.PublicBatch,
	updateOffesetInTransactionErr error,
) {
	commitRange := topicreadercommon.GetCommitRange(batch)
	tx.OnCompleted(func(transactionResult error) {
		defer func() {
			logCtx := r.cfg.BaseContext
			trace.TopicOnReaderTransactionCompleted(
				r.cfg.Trace,
				&logCtx,
				r.readerID,
				r.readConnectionID,
				tx.SessionID(),
				tx,
				transactionResult,
			)
		}()

		if transactionResult == nil {
			topicreadercommon.BatchGetPartitionSession(batch).SetCommittedOffsetForward(commitRange.CommitOffsetEnd)
		} else {
			_ = r.CloseWithError(xcontext.ValueOnly(ctx), xerrors.WithStackTrace(xerrors.Retryable(
				fmt.Errorf("ydb: failed batch commit because transaction doesn't committed: %w", updateOffesetInTransactionErr),
			)))
		}
	})
}

func (r *topicStreamReaderImpl) createUpdateOffsetRequest(
	ctx context.Context,
	batch *topicreadercommon.PublicBatch,
	tx tx.Transaction,
) *rawtopic.UpdateOffsetsInTransactionRequest {
	commitRange := topicreadercommon.GetCommitRange(batch)

	return &rawtopic.UpdateOffsetsInTransactionRequest{
		OperationParams: rawydb.NewRawOperationParamsFromProto(operation.Params(ctx, 0, 0, operation.ModeSync)),
		Tx: rawtopiccommon.TransactionIdentity{
			ID:      tx.ID(),
			Session: tx.SessionID(),
		},
		Topics: []rawtopic.UpdateOffsetsInTransactionRequest_TopicOffsets{
			{
				Path: batch.Topic(),
				Partitions: []rawtopic.UpdateOffsetsInTransactionRequest_PartitionOffsets{
					{
						PartitionID: batch.PartitionID(),
						PartitionOffsets: []rawtopiccommon.OffsetRange{
							{
								Start: commitRange.CommitOffsetStart,
								End:   commitRange.CommitOffsetEnd,
							},
						},
					},
				},
			},
		},
		Consumer: r.cfg.Consumer,
	}
}

func (r *topicStreamReaderImpl) ReadMessageBatch(
	ctx context.Context,
	opts ReadMessageBatchOptions,
) (batch *topicreadercommon.PublicBatch, err error) {
	defer func() {
		traceFunc := func(
			messagesCount int,
			topic string,
			partitionID int64,
			partitionSessionID int64,
			offsetStart int64,
			offsetEnd int64,
			freeBufferCapacity int,
		) {
			mergeCtx := xcontext.MergeContexts(ctx, r.cfg.BaseContext)
			onDone := trace.TopicOnReaderReadMessages(
				r.cfg.Trace,
				&mergeCtx,
				opts.MinCount,
				opts.MaxCount,
				r.getRestBufferBytes(),
			)
			onDone(messagesCount, topic, partitionID, partitionSessionID, offsetStart, offsetEnd, freeBufferCapacity, nil)
		}
		if batch == nil {
			traceFunc(0, "", -1, -1, -1, -1, r.getRestBufferBytes())
		} else {
			commitRange := topicreadercommon.GetCommitRange(batch)
			traceFunc(
				len(batch.Messages),
				batch.Topic(),
				batch.PartitionID(),
				topicreadercommon.BatchGetPartitionSession(batch).StreamPartitionSessionID.ToInt64(),
				commitRange.CommitOffsetStart.ToInt64(),
				commitRange.CommitOffsetEnd.ToInt64(),
				r.getRestBufferBytes(),
			)
		}
	}()

	if err = ctx.Err(); err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			r.freeBufferFromMessages(batch)
		}
	}()

	return r.consumeMessagesUntilBatch(ctx, opts)
}

func (r *topicStreamReaderImpl) consumeMessagesUntilBatch(
	ctx context.Context,
	opts ReadMessageBatchOptions,
) (*topicreadercommon.PublicBatch, error) {
	for {
		item, err := r.batcher.Pop(ctx, opts.batcherGetOptions)
		if err != nil {
			return nil, err
		}

		switch {
		case item.IsBatch():
			return item.Batch, nil
		case item.IsRawMessage():
			r.sendRawMessageToChannelUnblocked(item.RawMessage)
		default:
			return nil, xerrors.WithStackTrace(fmt.Errorf("ydb: unexpected item type from batcher: %#v", item))
		}
	}
}

func (r *topicStreamReaderImpl) sendRawMessageToChannelUnblocked(msg rawtopicreader.ServerMessage) {
	select {
	case r.rawMessagesFromBuffer <- msg:
		return
	default:
		// send in goroutine, without block caller
		r.backgroundWorkers.Start("sendMessageToRawChannel", func(ctx context.Context) {
			select {
			case r.rawMessagesFromBuffer <- msg:
			case <-ctx.Done():
			}
		})
	}
}

func (r *topicStreamReaderImpl) consumeRawMessageFromBuffer(ctx context.Context) {
	doneChan := ctx.Done()

	for {
		var msg rawtopicreader.ServerMessage
		select {
		case <-doneChan:
			return
		case msg = <-r.rawMessagesFromBuffer:
			// pass
		}

		switch m := msg.(type) {
		case *rawtopicreader.StartPartitionSessionRequest:
			if err := r.onStartPartitionSessionRequestFromBuffer(m); err != nil {
				_ = r.CloseWithError(ctx, err)

				return
			}
		case *rawtopicreader.StopPartitionSessionRequest:
			if err := r.onStopPartitionSessionRequestFromBuffer(m); err != nil {
				_ = r.CloseWithError(ctx, xerrors.WithStackTrace(
					fmt.Errorf("ydb: unexpected error on stop partition handler: %w", err),
				))

				return
			}
		case *rawtopicreader.PartitionSessionStatusResponse:
			r.onPartitionSessionStatusResponseFromBuffer(ctx, m)
		default:
			_ = r.CloseWithError(ctx, xerrors.WithStackTrace(
				fmt.Errorf("ydb: unexpected server message from buffer: %v", reflect.TypeOf(msg))),
			)
		}
	}
}

func (r *topicStreamReaderImpl) onStopPartitionSessionRequestFromBuffer(
	msg *rawtopicreader.StopPartitionSessionRequest,
) (err error) {
	session, err := r.sessionController.Get(msg.PartitionSessionID)
	if err != nil {
		return err
	}
	onDone := trace.TopicOnReaderPartitionReadStopResponse(
		r.cfg.Trace,
		r.readConnectionID,
		session.Context(),
		session.Topic,
		session.PartitionID,
		session.StreamPartitionSessionID.ToInt64(),
		msg.CommittedOffset.ToInt64(),
		msg.Graceful,
	)
	defer func() {
		onDone(err)
	}()

	if msg.Graceful {
		session.Close()
		resp := &rawtopicreader.StopPartitionSessionResponse{
			PartitionSessionID: session.StreamPartitionSessionID,
		}
		if err = r.send(resp); err != nil {
			return err
		}
	}

	if _, err = r.sessionController.Remove(session.StreamPartitionSessionID); err != nil {
		if msg.Graceful {
			return err
		} else { //nolint:revive,staticcheck
			// double message with graceful=false is ok.
			// It may be received after message with graceful=true and session was removed while process that.

			// pass
		}
	}

	return nil
}

func (r *topicStreamReaderImpl) onPartitionSessionStatusResponseFromBuffer(
	ctx context.Context,
	m *rawtopicreader.PartitionSessionStatusResponse,
) {
	panic("not implemented")
}

func (r *topicStreamReaderImpl) onUpdateTokenResponse(m *rawtopicreader.UpdateTokenResponse) {
}

func (r *topicStreamReaderImpl) Commit(ctx context.Context, commitRange topicreadercommon.CommitRange) (err error) {
	defer func() {
		if errors.Is(
			err,
			topicreadercommon.ErrPublicCommitSessionToExpiredSession,
		) && r.cfg.CommitMode == topicreadercommon.CommitModeAsync {
			err = nil
		}
	}()

	if commitRange.PartitionSession == nil {
		return xerrors.WithStackTrace(errCommitWithNilPartitionSession)
	}

	defer func() {
		session := commitRange.PartitionSession
		mergeCtx := xcontext.MergeContexts(ctx, r.cfg.BaseContext)
		trace.TopicOnReaderCommit(
			r.cfg.Trace,
			&mergeCtx,
			session.Topic,
			session.PartitionID,
			session.StreamPartitionSessionID.ToInt64(),
			commitRange.CommitOffsetStart.ToInt64(),
			commitRange.CommitOffsetEnd.ToInt64(),
		)
	}()

	if err = r.checkCommitRange(commitRange); err != nil {
		return err
	}

	return r.committer.Commit(ctx, commitRange)
}

func (r *topicStreamReaderImpl) checkCommitRange(commitRange topicreadercommon.CommitRange) error {
	if r.cfg.CommitMode == topicreadercommon.CommitModeNone {
		return topicreadercommon.ErrCommitDisabled
	}
	session := commitRange.PartitionSession

	if session == nil {
		return xerrors.WithStackTrace(errCommitWithNilPartitionSession)
	}

	if session.Context().Err() != nil {
		return xerrors.WithStackTrace(topicreadercommon.ErrPublicCommitSessionToExpiredSession)
	}

	ownSession, err := r.sessionController.Get(session.StreamPartitionSessionID)
	if err != nil || session != ownSession {
		return xerrors.WithStackTrace(topicreadercommon.ErrPublicCommitSessionToExpiredSession)
	}
	if session.CommittedOffset() != commitRange.CommitOffsetStart && r.cfg.CommitMode == topicreadercommon.CommitModeSync {
		return topicreadercommon.ErrWrongCommitOrderInSyncMode
	}

	return nil
}

func (r *topicStreamReaderImpl) send(msg rawtopicreader.ClientMessage) error {
	err := r.stream.Send(msg)
	if err != nil {
		logCtx := r.cfg.BaseContext
		trace.TopicOnReaderError(r.cfg.Trace, &logCtx, r.readConnectionID, err)

		_ = r.CloseWithError(r.ctx, err)
	}

	return err
}

func (r *topicStreamReaderImpl) startBackgroundWorkers() error {
	if err := r.setStarted(); err != nil {
		return err
	}

	r.committer.Start()

	r.backgroundWorkers.Start("readMessagesLoop", r.readMessagesLoop)
	r.backgroundWorkers.Start("dataRequestLoop", r.dataRequestLoop)
	r.backgroundWorkers.Start("updateTokenLoop", r.updateTokenLoop)

	r.backgroundWorkers.Start("consumeRawMessageFromBuffer", r.consumeRawMessageFromBuffer)

	return nil
}

func (r *topicStreamReaderImpl) setStarted() error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.started {
		return xerrors.WithStackTrace(errors.New("already started"))
	}

	r.started = true

	return nil
}

func (r *topicStreamReaderImpl) initSession() (err error) {
	initMessage := topicreadercommon.CreateInitMessage(r.cfg.Consumer, r.cfg.EnableSplitMergeSupport, r.cfg.ReadSelectors)

	defer func() {
		logCtx := r.cfg.BaseContext
		onDone := trace.TopicOnReaderInit(r.cfg.Trace, &logCtx, r.readConnectionID, initMessage)
		onDone(r.readConnectionID, err)
	}()

	if err = r.send(initMessage); err != nil {
		return err
	}

	resp, err := r.stream.Recv()
	if err != nil {
		return err
	}

	if status := resp.StatusData(); !status.Status.IsSuccess() {
		// Need wrap status to common ydb operational error
		// https://github.com/ydb-platform/ydb-go-sdk/issues/1361
		return xerrors.WithStackTrace(fmt.Errorf("bad status on initial error: %v (%v)", status.Status, status.Issues))
	}

	initResp, ok := resp.(*rawtopicreader.InitResponse)
	if !ok {
		return xerrors.WithStackTrace(fmt.Errorf("bad message type on session init: %v (%v)", resp, reflect.TypeOf(resp)))
	}

	r.readConnectionID = initResp.SessionID

	return nil
}

func (r *topicStreamReaderImpl) addRestBufferBytes(delta int) int {
	val := r.restBufferSizeBytes.Add(int64(delta))
	if val <= 0 {
		r.batcher.IgnoreMinRestrictionsOnNextPop()
	}

	return int(val)
}

func (r *topicStreamReaderImpl) getRestBufferBytes() int {
	return int(r.restBufferSizeBytes.Load())
}

//nolint:funlen
func (r *topicStreamReaderImpl) readMessagesLoop(ctx context.Context) {
	ctx, cancel := xcontext.WithCancel(ctx)
	defer cancel()

	for {
		serverMessage, err := r.stream.Recv()
		if err != nil {
			logCtx := r.cfg.BaseContext
			trace.TopicOnReaderError(r.cfg.Trace, &logCtx, r.readConnectionID, err)
			if errors.Is(err, rawtopicreader.ErrUnexpectedMessageType) {
				trace.TopicOnReaderUnknownGrpcMessage(r.cfg.Trace, &logCtx, r.readConnectionID, err)
				// new messages can be added to protocol, it must be backward compatible to old programs
				// and skip message is safe
				continue
			}
			_ = r.CloseWithError(ctx, err)

			return
		}

		status := serverMessage.StatusData()
		if !status.Status.IsSuccess() {
			_ = r.CloseWithError(ctx,
				xerrors.WithStackTrace(
					fmt.Errorf("ydb: bad status from pq grpc stream: %v, %v", status.Status, status.Issues.String()),
				),
			)
		}

		switch m := serverMessage.(type) {
		case *rawtopicreader.ReadResponse:
			if err = r.onReadResponse(m); err != nil {
				_ = r.CloseWithError(ctx, err)
			}
		case *rawtopicreader.StartPartitionSessionRequest:
			if err = r.onStartPartitionSessionRequest(m); err != nil {
				_ = r.CloseWithError(ctx, err)

				return
			}
		case *rawtopicreader.StopPartitionSessionRequest:
			if err = r.onStopPartitionSessionRequest(m); err != nil {
				_ = r.CloseWithError(ctx, err)

				return
			}
		case *rawtopicreader.EndPartitionSession:
			if err = r.onEndPartitionSession(m); err != nil {
				_ = r.CloseWithError(ctx, err)
			}
		case *rawtopicreader.CommitOffsetResponse:
			if err = r.onCommitResponse(m); err != nil {
				_ = r.CloseWithError(ctx, err)

				return
			}

		case *rawtopicreader.UpdateTokenResponse:
			r.onUpdateTokenResponse(m)
		default:
			{
				logCtx := r.cfg.BaseContext
				trace.TopicOnReaderUnknownGrpcMessage(
					r.cfg.Trace,
					&logCtx,
					r.readConnectionID,
					xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
						"ydb: unexpected message type in stream reader: %v",
						reflect.TypeOf(serverMessage),
					))),
				)
			}
		}
	}
}

func (r *topicStreamReaderImpl) dataRequestLoop(ctx context.Context) {
	if r.ctx.Err() != nil {
		return
	}

	doneChan := ctx.Done()

	for {
		select {
		case <-doneChan:
			_ = r.CloseWithError(ctx, r.ctx.Err())

			return

		case free := <-r.freeBytes:
			sum := free

			// consume all messages from order and compress it to one data request
		forConsumeRequests:
			for {
				select {
				case free = <-r.freeBytes:
					sum += free
				default:
					break forConsumeRequests
				}
			}

			resCapacity := r.addRestBufferBytes(sum)
			logCtx := r.cfg.BaseContext
			trace.TopicOnReaderSentDataRequest(r.cfg.Trace, &logCtx, r.readConnectionID, sum, resCapacity)
			if err := r.sendDataRequest(sum); err != nil {
				return
			}
		}
	}
}

func (r *topicStreamReaderImpl) sendDataRequest(size int) error {
	return r.send(&rawtopicreader.ReadRequest{BytesSize: size})
}

func (r *topicStreamReaderImpl) freeBufferFromMessages(batch *topicreadercommon.PublicBatch) {
	size := 0
	for messageIndex := range batch.Messages {
		size += topicreadercommon.MessageGetBufferBytesAccount(batch.Messages[messageIndex])
	}
	select {
	case r.freeBytes <- size:
	case <-r.ctx.Done():
	}
}

func (r *topicStreamReaderImpl) updateTokenLoop(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.CredUpdateInterval)
	defer ticker.Stop()

	readerCancel := ctx.Done()
	for {
		select {
		case <-readerCancel:
			return
		case <-ticker.C:
			r.updateToken(r.ctx)
		}
	}
}

func (r *topicStreamReaderImpl) onReadResponse(msg *rawtopicreader.ReadResponse) (err error) {
	resCapacity := r.addRestBufferBytes(-msg.BytesSize)
	defer func() {
		logCtx := r.cfg.BaseContext
		trace.TopicOnReaderReceiveDataResponse(r.cfg.Trace, &logCtx, r.readConnectionID, resCapacity, msg)
	}()

	batches, err2 := topicreadercommon.ReadRawBatchesToPublicBatches(msg, &r.sessionController, r.cfg.Decoders)
	if err2 != nil {
		return err2
	}

	for i := range batches {
		if err := r.batcher.PushBatches(batches[i]); err != nil {
			return err
		}
	}

	return nil
}

func (r *topicStreamReaderImpl) CloseWithError(ctx context.Context, reason error) (closeErr error) {
	defer func() {
		logCtx := r.cfg.BaseContext
		trace.TopicOnReaderClose(r.cfg.Trace, &logCtx, r.readConnectionID, reason)
	}()

	isFirstClose := false
	r.m.WithLock(func() {
		if r.closed {
			return
		}
		isFirstClose = true
		r.closed = true

		r.err = reason
		r.cancel()
	})
	if !isFirstClose {
		return nil
	}

	closeErr = r.committer.Close(ctx, reason)

	batcherErr := r.batcher.Close(reason)
	if closeErr == nil {
		closeErr = batcherErr
	}

	// close stream strong after committer close - for flush commits buffer
	streamCloseErr := r.stream.CloseSend()
	if closeErr == nil {
		closeErr = streamCloseErr
	}

	// close background workers after r.stream.CloseSend
	bgCloseErr := r.backgroundWorkers.Close(ctx, reason)
	if closeErr == nil {
		closeErr = bgCloseErr
	}

	return closeErr
}

func (r *topicStreamReaderImpl) onCommitResponse(msg *rawtopicreader.CommitOffsetResponse) error {
	for i := range msg.PartitionsCommittedOffsets {
		commit := &msg.PartitionsCommittedOffsets[i]
		partition, err := r.sessionController.Get(commit.PartitionSessionID)
		if err != nil {
			return fmt.Errorf("ydb: can't found session on commit response: %w", err)
		}
		partition.SetCommittedOffsetForward(commit.CommittedOffset)

		logCtx := r.cfg.BaseContext
		trace.TopicOnReaderCommittedNotify(
			r.cfg.Trace,
			&logCtx,
			r.readConnectionID,
			partition.Topic,
			partition.PartitionID,
			partition.StreamPartitionSessionID.ToInt64(),
			commit.CommittedOffset.ToInt64(),
		)

		r.committer.OnCommitNotify(partition, commit.CommittedOffset)
	}

	return nil
}

func (r *topicStreamReaderImpl) updateToken(ctx context.Context) {
	logCtx := r.cfg.BaseContext
	onUpdateToken := trace.TopicOnReaderUpdateToken(
		r.cfg.Trace,
		&logCtx,
		r.readConnectionID,
	)
	token, err := r.cfg.Cred.Token(ctx)
	onSent := onUpdateToken(&ctx, len(token), err)
	if err != nil {
		return
	}

	err = r.send(&rawtopicreader.UpdateTokenRequest{UpdateTokenRequest: rawtopiccommon.UpdateTokenRequest{Token: token}})
	onSent(err)
}

func (r *topicStreamReaderImpl) onStartPartitionSessionRequest(m *rawtopicreader.StartPartitionSessionRequest) error {
	session := topicreadercommon.NewPartitionSession(
		r.ctx,
		m.PartitionSession.Path,
		m.PartitionSession.PartitionID,
		r.readerID,
		r.readConnectionID,
		m.PartitionSession.PartitionSessionID,
		clientSessionCounter.Add(1),
		m.CommittedOffset,
	)
	if err := r.sessionController.Add(session); err != nil {
		return err
	}

	return r.batcher.PushRawMessage(session, m)
}

func (r *topicStreamReaderImpl) onStartPartitionSessionRequestFromBuffer(
	m *rawtopicreader.StartPartitionSessionRequest,
) (err error) {
	session, err := r.sessionController.Get(m.PartitionSession.PartitionSessionID)
	if err != nil {
		return err
	}

	var (
		ctx    = session.Context()
		onDone = trace.TopicOnReaderPartitionReadStartResponse(
			r.cfg.Trace,
			r.readConnectionID,
			&ctx,
			session.Topic,
			session.PartitionID,
			session.StreamPartitionSessionID.ToInt64(),
		)
	)

	respMessage := &rawtopicreader.StartPartitionSessionResponse{
		PartitionSessionID: session.StreamPartitionSessionID,
	}

	var forceOffset *int64
	var commitOffset *int64

	defer func() {
		onDone(forceOffset, commitOffset, err)
	}()

	if r.cfg.GetPartitionStartOffsetCallback != nil {
		req := PublicGetPartitionStartOffsetRequest{
			Topic:       session.Topic,
			PartitionID: session.PartitionID,
		}
		resp, callbackErr := r.cfg.GetPartitionStartOffsetCallback(session.Context(), req)
		if callbackErr != nil {
			return callbackErr
		}
		if resp.startOffsetUsed {
			wantOffset := resp.startOffset.ToInt64()
			forceOffset = &wantOffset
		}
	}

	respMessage.ReadOffset.FromInt64Pointer(forceOffset)
	if r.cfg.CommitMode.CommitsEnabled() {
		commitOffset = forceOffset
		respMessage.CommitOffset.FromInt64Pointer(commitOffset)
	}
	if forceOffset != nil {
		session.SetInitialCommitOffset(rawtopiccommon.NewOffset(*forceOffset))
	}

	return r.send(respMessage)
}

func (r *topicStreamReaderImpl) onStopPartitionSessionRequest(m *rawtopicreader.StopPartitionSessionRequest) error {
	session, err := r.sessionController.Get(m.PartitionSessionID)
	if err != nil {
		return err
	}

	session.SetNoMoreMessages()
	if !m.Graceful {
		session.Close()
	}

	return r.batcher.PushRawMessage(session, m)
}

func (r *topicStreamReaderImpl) onEndPartitionSession(m *rawtopicreader.EndPartitionSession) error {
	// need err value in else block
	//nolint:revive
	if session, err := r.sessionController.Get(m.PartitionSessionID); err == nil {
		trace.TopicOnReaderEndPartitionSession(
			r.cfg.Trace,
			r.readConnectionID,
			session.Context(),
			session.Topic,
			session.PartitionID,
			m.PartitionSessionID.ToInt64(),
			m.AdjacentPartitionIDs,
			m.ChildPartitionIDs,
		)
		session.SetNoMoreMessages()
		r.batcher.FlushPartitionSession(session)

		return nil
	} else {
		return xerrors.Retryable(xerrors.Wrap(fmt.Errorf(
			"ydb: unknown partition for end partition session: %w", err,
		)))
	}
}
