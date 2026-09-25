package topicreaderinternal

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var (
	errUnconnected = xerrors.Retryable(xerrors.Wrap(
		errors.New("ydb: first connection attempt not finished"),
	))
	errReaderClosed                 = xerrors.Wrap(errors.New("ydb: reader closed"))
	errSetConsumerAndNoConsumer     = xerrors.Wrap(errors.New("ydb: reader has non empty consumer name and set option WithReaderWithoutConsumer. Only one of them must be set")) //nolint:lll
	errCommitSessionFromOtherReader = xerrors.Wrap(errors.New("ydb: commit with session from other reader"))
)

// TopicSteamReaderConnect connect to grpc stream
// when connectionCtx closed stream must stop work and return errors for all methods
type TopicSteamReaderConnect func(
	connectionCtx context.Context,
	readerID int64,
	tracer *trace.Topic,
) (topicreadercommon.RawTopicReaderStream, error)

type Reader struct {
	reader             batchedStreamReader
	defaultBatchConfig ReadMessageBatchOptions
	tracer             *trace.Topic
	readerID           int64
}

func (r *Reader) TopicOnReaderStart(consumer string, err error) {
	r.reader.TopicOnReaderStart(consumer, err)
}

type ReadMessageBatchOptions struct {
	batcherGetOptions
}

func (o ReadMessageBatchOptions) clone() ReadMessageBatchOptions {
	return o
}

func newReadMessageBatchOptions() ReadMessageBatchOptions {
	return ReadMessageBatchOptions{}
}

// PublicReadBatchOption для различных пожеланий к батчу вроде WithMaxMessages(int)
type PublicReadBatchOption interface {
	Apply(options ReadMessageBatchOptions) ReadMessageBatchOptions
}

type readExplicitMessagesCount int

// Apply implements PublicReadBatchOption
func (count readExplicitMessagesCount) Apply(options ReadMessageBatchOptions) ReadMessageBatchOptions {
	options.MinCount = int(count)
	options.MaxCount = int(count)

	return options
}

func NewReader(
	client TopicClient,
	connector TopicSteamReaderConnect,
	consumer string,
	readSelectors []topicreadercommon.PublicReadSelector,
	opts ...PublicReaderOption,
) (Reader, error) {
	cfg := convertNewParamsToStreamConfig(consumer, readSelectors, opts...)

	if errs := cfg.Validate(); len(errs) > 0 {
		return Reader{}, xerrors.WithStackTrace(fmt.Errorf(
			"ydb: failed to start topic reader, because is contains error in config: %w",
			errors.Join(errs...),
		))
	}

	readerID := topicreadercommon.NextReaderID()

	readerConnector := func(ctx context.Context) (batchedStreamReader, error) {
		stream, err := connector(ctx, readerID, cfg.Trace)
		if err != nil {
			return nil, err
		}

		return newTopicStreamReader(client, readerID, stream, cfg.topicStreamReaderConfig)
	}

	reader := newReaderReconnector(
		cfg.BaseContext,
		readerID,
		readerConnector,
		cfg.OperationTimeout(),
		cfg.RetrySettings,
		cfg.Trace,
	)

	res := Reader{
		reader:             reader,
		defaultBatchConfig: cfg.DefaultBatchConfig,
		tracer:             cfg.Trace,
		readerID:           readerID,
	}

	return res, nil
}

func (r *Reader) WaitInit(ctx context.Context) error {
	return r.reader.WaitInit(ctx)
}

func (r *Reader) ID() int64 {
	return r.readerID
}

func (r *Reader) Tracer() *trace.Topic {
	return r.tracer
}

func (r *Reader) Close(ctx context.Context) error {
	return r.reader.CloseWithError(ctx, xerrors.WithStackTrace(errReaderClosed))
}

func (r *Reader) PopBatchTx(
	ctx context.Context,
	tx tx.Transaction,
	opts ...PublicReadBatchOption,
) (*topicreadercommon.PublicBatch, error) {
	batchOptions := r.getBatchOptions(opts)

	return r.reader.PopMessagesBatchTx(ctx, tx, batchOptions)
}

// ReadMessage read exactly one message
func (r *Reader) ReadMessage(ctx context.Context) (*topicreadercommon.PublicMessage, error) {
	res, err := r.ReadMessageBatch(ctx, readExplicitMessagesCount(1))
	if err != nil {
		return nil, err
	}

	return res.Messages[0], nil
}

// ReadMessageBatch read batch of messages.
// Batch is collection of messages, which can be atomically committed
func (r *Reader) ReadMessageBatch(
	ctx context.Context,
	opts ...PublicReadBatchOption,
) (
	batch *topicreadercommon.PublicBatch,
	err error,
) {
	batchOptions := r.getBatchOptions(opts)

	for {
		if err = ctx.Err(); err != nil {
			return nil, err
		}

		batch, err = r.reader.ReadMessageBatch(ctx, batchOptions)
		if err != nil {
			return nil, err
		}

		// if batch context is canceled - do not return it to client
		// and read next batch
		if batch.Context().Err() == nil {
			return batch, nil
		}
	}
}

func (r *Reader) getBatchOptions(opts []PublicReadBatchOption) ReadMessageBatchOptions {
	readOptions := r.defaultBatchConfig.clone()

	for _, opt := range opts {
		if opt != nil {
			readOptions = opt.Apply(readOptions)
		}
	}

	return readOptions
}

func (r *Reader) Commit(ctx context.Context, offsets topicreadercommon.PublicCommitRangeGetter) (err error) {
	cr := topicreadercommon.GetCommitRange(offsets)
	if cr.PartitionSession.ReaderID != r.readerID {
		return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
			"ydb: messages session reader id (%v) != current reader id (%v): %w",
			cr.PartitionSession.ReaderID, r.readerID, errCommitSessionFromOtherReader,
		)))
	}

	return r.reader.Commit(ctx, cr)
}

func (r *Reader) CommitRanges(ctx context.Context, ranges []topicreadercommon.PublicCommitRange) error {
	for i := range ranges {
		commitRange := topicreadercommon.GetCommitRange(ranges[i])
		if commitRange.PartitionSession.ReaderID != r.readerID {
			return xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
				"ydb: commit ranges (range item %v) "+
					"messages session reader id (%v) != current reader id (%v): %w",
				i, commitRange.PartitionSession.ReaderID, r.readerID, errCommitSessionFromOtherReader,
			)))
		}
	}

	commitRanges := topicreadercommon.NewCommitRangesFromPublicCommits(ranges)
	commitRanges.Optimize()

	commitErrors := make(chan error, commitRanges.Len())

	var wg sync.WaitGroup

	commit := func(cr topicreadercommon.CommitRange) {
		defer wg.Done()
		commitErrors <- r.Commit(ctx, &cr)
	}

	wg.Add(commitRanges.Len())
	for _, cr := range commitRanges.Ranges {
		go commit(cr)
	}
	wg.Wait()
	close(commitErrors)

	// return first error
	for err := range commitErrors {
		if err != nil {
			return err
		}
	}

	return nil
}

type ReaderConfig struct {
	config.Common

	RetrySettings      topic.RetrySettings
	DefaultBatchConfig ReadMessageBatchOptions
	topicStreamReaderConfig
}

type PublicReaderOption func(cfg *ReaderConfig)

func WithCredentials(cred credentials.Credentials) PublicReaderOption {
	return func(cfg *ReaderConfig) {
		if cred == nil {
			cred = credentials.NewAnonymousCredentials()
		}
		cfg.Cred = cred
	}
}

func WithTrace(tracer *trace.Topic) PublicReaderOption {
	return func(cfg *ReaderConfig) {
		cfg.Trace = cfg.Trace.Compose(tracer)
	}
}

func convertNewParamsToStreamConfig(
	consumer string,
	readSelectors []topicreadercommon.PublicReadSelector,
	opts ...PublicReaderOption,
) (cfg ReaderConfig) {
	cfg.topicStreamReaderConfig = newTopicStreamReaderConfig()
	cfg.Consumer = consumer

	// make own copy, for prevent changing internal states if readSelectors will change outside
	cfg.ReadSelectors = make([]*topicreadercommon.PublicReadSelector, len(readSelectors))
	for i := range readSelectors {
		cfg.ReadSelectors[i] = readSelectors[i].Clone()
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return cfg
}
