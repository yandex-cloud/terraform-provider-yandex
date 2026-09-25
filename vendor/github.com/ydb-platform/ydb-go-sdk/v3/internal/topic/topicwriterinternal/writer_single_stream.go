package topicwriterinternal

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/background"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicwriter"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var errSingleStreamWriterDoubleClose = xerrors.Wrap(errors.New("ydb: single stream writer impl double closed"))

type SingleStreamWriterConfig struct {
	WritersCommonConfig

	stream                RawTopicWriterStream
	queue                 *messageQueue
	encodersMap           *MultiEncoder
	getLastSeqNum         bool
	reconnectorInstanceID string
}

func newSingleStreamWriterConfig(
	common WritersCommonConfig, //nolint:gocritic
	stream RawTopicWriterStream,
	queue *messageQueue,
	encodersMap *MultiEncoder,
	getLastSeqNum bool,
	reconnectorID string,
) SingleStreamWriterConfig {
	return SingleStreamWriterConfig{
		WritersCommonConfig:   common,
		stream:                stream,
		queue:                 queue,
		encodersMap:           encodersMap,
		getLastSeqNum:         getLastSeqNum,
		reconnectorInstanceID: reconnectorID,
	}
}

type SingleStreamWriter struct {
	cfg                 SingleStreamWriterConfig
	Encoder             EncoderSelector
	background          background.Worker
	CodecsFromServer    rawtopiccommon.SupportedCodecs
	allowedCodecs       rawtopiccommon.SupportedCodecs
	SessionID           string
	closeReason         error
	ReceivedLastSeqNum  int64
	PartitionID         int64
	closeCompleted      empty.Chan
	closed              atomic.Bool
	LastSeqNumRequested bool
}

func NewSingleStreamWriter(
	ctxForPProfLabelsOnly context.Context,
	cfg SingleStreamWriterConfig, //nolint:gocritic
) (*SingleStreamWriter, error) {
	res := newSingleStreamWriterStopped(ctxForPProfLabelsOnly, cfg)

	if err := res.initStream(); err != nil {
		_ = res.close(context.Background(), err)

		return nil, err
	}

	res.start()

	return res, nil
}

func newSingleStreamWriterStopped(
	ctxForPProfLabelsOnly context.Context,
	cfg SingleStreamWriterConfig, //nolint:gocritic
) *SingleStreamWriter {
	return &SingleStreamWriter{
		cfg: cfg,
		background: *background.NewWorker(
			xcontext.ValueOnly(ctxForPProfLabelsOnly),
			"ydb-topic-stream-writer-background",
		),
		closeCompleted: make(empty.Chan),
	}
}

func (w *SingleStreamWriter) close(ctx context.Context, reason error) error {
	if !w.closed.CompareAndSwap(false, true) {
		return xerrors.WithStackTrace(errSingleStreamWriterDoubleClose)
	}

	defer close(w.closeCompleted)
	w.closeReason = reason

	resErr := w.cfg.stream.CloseSend()
	bgWaitErr := w.background.Close(ctx, reason)
	if resErr == nil {
		resErr = bgWaitErr
	}

	return resErr
}

func (w *SingleStreamWriter) WaitClose(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.closeCompleted:
		return w.closeReason
	}
}

func (w *SingleStreamWriter) start() {
	w.background.Start("topic writer update token", w.updateTokenLoop)
	w.background.Start("topic writer send messages", w.sendMessagesFromQueueToStreamLoop)
	w.background.Start("topic writer receive messages", w.receiveMessagesLoop)
}

func (w *SingleStreamWriter) initStream() (err error) {
	defer func() {
		logCtx := w.cfg.LogContext
		traceOnDone := trace.TopicOnWriterInitStream(
			w.cfg.Tracer,
			&logCtx,
			w.cfg.reconnectorInstanceID,
			w.cfg.topic,
			w.cfg.producerID,
		)
		traceOnDone(w.SessionID, err)
	}()

	req := w.createInitRequest()
	if err = w.cfg.stream.Send(&req); err != nil {
		return err
	}
	recvMessage, err := w.cfg.stream.Recv()
	if err != nil {
		return err
	}
	result, ok := recvMessage.(*rawtopicwriter.InitResult)
	if !ok {
		return xerrors.WithStackTrace(
			fmt.Errorf("ydb: failed init response message type: %v", reflect.TypeOf(recvMessage)),
		)
	}

	w.allowedCodecs = calculateAllowedCodecs(w.cfg.forceCodec, w.cfg.encodersMap, result.SupportedCodecs)
	if len(w.allowedCodecs) == 0 {
		return xerrors.WithStackTrace(errNoAllowedCodecs)
	}

	w.Encoder = NewEncoderSelector(
		w.cfg.LogContext,
		w.cfg.encodersMap,
		w.allowedCodecs,
		w.cfg.compressorCount,
		w.cfg.Tracer,
		w.cfg.reconnectorInstanceID,
		w.SessionID,
	)

	w.SessionID = result.SessionID
	w.LastSeqNumRequested = req.GetLastSeqNo
	w.ReceivedLastSeqNum = result.LastSeqNo
	w.PartitionID = result.PartitionID
	w.CodecsFromServer = result.SupportedCodecs

	return nil
}

func (w *SingleStreamWriter) createInitRequest() rawtopicwriter.InitRequest {
	return rawtopicwriter.InitRequest{
		Path:             w.cfg.topic,
		ProducerID:       w.cfg.producerID,
		WriteSessionMeta: w.cfg.writerMeta,
		Partitioning:     w.cfg.defaultPartitioning,
		GetLastSeqNo:     w.cfg.getLastSeqNum,
	}
}

func (w *SingleStreamWriter) receiveMessagesLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		mess, err := w.cfg.stream.Recv()
		if err != nil {
			err = xerrors.WithStackTrace(fmt.Errorf("ydb: failed to receive message from write stream: %w", err))
			_ = w.close(ctx, err)

			return
		}

		switch m := mess.(type) {
		case *rawtopicwriter.WriteResult:
			if err = w.cfg.queue.AcksReceived(m.Acks); err != nil && !errors.Is(err, errCloseClosedMessageQueue) {
				reason := xerrors.WithStackTrace(err)
				closeCtx, closeCtxCancel := xcontext.WithCancel(ctx)
				closeCtxCancel()
				_ = w.close(closeCtx, reason)

				return
			}
		case *rawtopicwriter.UpdateTokenResponse:
			// pass
		default:
			{
				logCtx := w.cfg.LogContext
				trace.TopicOnWriterReadUnknownGrpcMessage(
					w.cfg.Tracer,
					&logCtx,
					w.cfg.reconnectorInstanceID,
					w.SessionID,
					xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
						"ydb: unexpected message type in stream reader: %v",
						reflect.TypeOf(m),
					))),
				)
			}
		}
	}
}

func (w *SingleStreamWriter) sendMessagesFromQueueToStreamLoop(ctx context.Context) {
	for {
		messages, err := w.cfg.queue.GetMessagesForSend(ctx)
		if err != nil {
			_ = w.close(ctx, err)

			return
		}

		targetCodec, err := w.Encoder.CompressMessages(messages)
		if err != nil {
			_ = w.close(ctx, err)

			return
		}

		err = sendMessagesToStream(w.cfg.stream, w.cfg.maxBytesPerMessage, targetCodec, messages)

		logCtx := w.cfg.LogContext
		onSentComplete := trace.TopicOnWriterSendMessages(
			w.cfg.Tracer,
			&logCtx,
			w.cfg.reconnectorInstanceID,
			w.SessionID,
			targetCodec.ToInt32(),
			messages[0].SeqNo,
			len(messages),
		)
		onSentComplete(err)

		if err != nil {
			err = xerrors.WithStackTrace(fmt.Errorf("ydb: error send message to topic stream: %w", err))
			_ = w.close(ctx, err)

			return
		}
	}
}

func (w *SingleStreamWriter) updateTokenLoop(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	ticker := w.cfg.clock.NewTicker(w.cfg.credUpdateInterval)
	defer ticker.Stop()

	ctxDone := ctx.Done()
	tickerChan := ticker.Chan()
	for {
		select {
		case <-ctxDone:
			return
		case <-tickerChan:
			_ = w.sendUpdateToken(ctx)
		}
	}
}

func (w *SingleStreamWriter) sendUpdateToken(ctx context.Context) (err error) {
	token, err := w.cfg.cred.Token(ctx)
	if err != nil {
		return err
	}

	stream := w.cfg.stream
	if stream == nil {
		// not connected yet
		return nil
	}

	req := &rawtopicwriter.UpdateTokenRequest{}
	req.Token = token

	return stream.Send(req)
}

//go:generate mockgen -destination raw_topic_writer_stream_mock_test.go --typed -package topicwriterinternal -write_package_comment=false github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicwriterinternal RawTopicWriterStream

type RawTopicWriterStream interface {
	Recv() (rawtopicwriter.ServerMessage, error)
	Send(mess rawtopicwriter.ClientMessage) error
	CloseSend() error
}

func sendMessagesToStream(
	stream RawTopicWriterStream,
	maxBytesPerMessage int,
	targetCodec rawtopiccommon.Codec,
	messages []messageWithDataContent,
) error {
	if len(messages) == 0 {
		return nil
	}

	request, err := createWriteRequest(messages, targetCodec)
	if err != nil {
		return err
	}

	var rest *rawtopicwriter.WriteRequest
	for request != nil {
		request, rest = cutRequestBytes(request, maxBytesPerMessage)
		err = stream.Send(request)
		if err != nil {
			return xerrors.WithStackTrace(fmt.Errorf("ydb: failed send write request: %w", err))
		}
		request = rest
	}

	return nil
}

func cutRequestBytes(req *rawtopicwriter.WriteRequest, maxBytes int) (head, rest *rawtopicwriter.WriteRequest) {
	requestSize := req.Size()
	requestMessagesCount := len(req.Messages)

	// it needs 1 message minimum for request
	// reverse order reason:
	// 1. Fast process messages less, than maxBytes, without special way
	// 2. Prevent difficult for account bytes of encode messages count
	for requestSize > maxBytes && requestMessagesCount > 1 {
		requestMessagesCount--
		requestSize -= req.Messages[requestMessagesCount].ProtoWireSizeBytes()
	}

	return req.Cut(requestMessagesCount)
}
