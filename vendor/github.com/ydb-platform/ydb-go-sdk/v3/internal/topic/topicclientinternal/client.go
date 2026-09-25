package topicclientinternal

import (
	"context"
	"errors"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Topic_V1"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topiclistenerinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreadercommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreaderinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicwriterinternal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topiclistener"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var errUnsupportedTransactionType = xerrors.Wrap(errors.New("ydb: unsuppotred transaction type. Use transaction from Driver().Query().DoTx(...)")) //nolint:lll

type Client struct {
	cfg                    topic.Config
	cred                   credentials.Credentials
	defaultOperationParams rawydb.OperationParams
	rawClient              rawtopic.Client
}

func New(
	ctx context.Context,
	conn grpc.ClientConnInterface,
	cred credentials.Credentials,
	opts ...topicoptions.TopicOption,
) *Client {
	rawClient := rawtopic.NewClient(Ydb_Topic_V1.NewTopicServiceClient(conn))

	cfg := newTopicConfig(opts...)

	var defaultOperationParams rawydb.OperationParams
	topic.OperationParamsFromConfig(&defaultOperationParams, &cfg.Common)

	return &Client{
		cfg:                    cfg,
		cred:                   cred,
		defaultOperationParams: defaultOperationParams,
		rawClient:              rawClient,
	}
}

func newTopicConfig(opts ...topicoptions.TopicOption) topic.Config {
	c := topic.Config{
		Trace:              &trace.Topic{},
		MaxGrpcMessageSize: config.DefaultGRPCMsgSize,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}

	return c
}

// Close the client
func (c *Client) Close(_ context.Context) error {
	return nil
}

// Alter topic options
func (c *Client) Alter(ctx context.Context, path string, opts ...topicoptions.AlterOption) error {
	req := &rawtopic.AlterTopicRequest{}
	req.OperationParams = c.defaultOperationParams
	req.Path = path
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyAlterOption(req)
		}
	}

	call := func(ctx context.Context) error {
		_, alterErr := c.rawClient.AlterTopic(ctx, req)

		return alterErr
	}

	if c.cfg.AutoRetry() {
		return retry.Retry(ctx, call,
			retry.WithIdempotent(true),
			retry.WithTrace(c.cfg.TraceRetry()),
			retry.WithBudget(c.cfg.RetryBudget()),
		)
	}

	return call(ctx)
}

// Create new topic
func (c *Client) Create(
	ctx context.Context,
	path string,
	opts ...topicoptions.CreateOption,
) error {
	req := &rawtopic.CreateTopicRequest{}
	req.OperationParams = c.defaultOperationParams
	req.Path = path

	for _, opt := range opts {
		if opt != nil {
			opt.ApplyCreateOption(req)
		}
	}

	call := func(ctx context.Context) error {
		_, createErr := c.rawClient.CreateTopic(ctx, req)

		return createErr
	}

	if c.cfg.AutoRetry() {
		return retry.Retry(ctx, call,
			retry.WithIdempotent(true),
			retry.WithTrace(c.cfg.TraceRetry()),
			retry.WithBudget(c.cfg.RetryBudget()),
		)
	}

	return call(ctx)
}

// Describe topic
func (c *Client) Describe(
	ctx context.Context,
	path string,
	opts ...topicoptions.DescribeOption,
) (res topictypes.TopicDescription, _ error) {
	req := rawtopic.DescribeTopicRequest{
		OperationParams: c.defaultOperationParams,
		Path:            path,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&req)
		}
	}

	var rawRes rawtopic.DescribeTopicResult

	call := func(ctx context.Context) (describeErr error) {
		rawRes, describeErr = c.rawClient.DescribeTopic(ctx, req)

		return describeErr
	}

	var err error

	if c.cfg.AutoRetry() {
		err = retry.Retry(ctx, call,
			retry.WithIdempotent(true),
			retry.WithTrace(c.cfg.TraceRetry()),
			retry.WithBudget(c.cfg.RetryBudget()),
		)
	} else {
		err = call(ctx)
	}

	if err != nil {
		return res, err
	}

	res.FromRaw(&rawRes)

	return res, nil
}

// Describe topic consumer
func (c *Client) DescribeTopicConsumer(
	ctx context.Context,
	path string,
	consumer string,
	opts ...topicoptions.DescribeConsumerOption,
) (res topictypes.TopicConsumerDescription, _ error) {
	req := rawtopic.DescribeConsumerRequest{
		OperationParams: c.defaultOperationParams,
		Path:            path,
		Consumer:        consumer,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&req)
		}
	}

	var rawRes rawtopic.DescribeConsumerResult

	call := func(ctx context.Context) (describeErr error) {
		rawRes, describeErr = c.rawClient.DescribeConsumer(ctx, req)

		return describeErr
	}

	var err error

	if c.cfg.AutoRetry() {
		err = retry.Retry(ctx, call,
			retry.WithIdempotent(true),
			retry.WithTrace(c.cfg.TraceRetry()),
			retry.WithBudget(c.cfg.RetryBudget()),
		)
	} else {
		err = call(ctx)
	}

	if err != nil {
		return res, err
	}

	res.FromRaw(&rawRes)

	return res, nil
}

// Drop topic
func (c *Client) Drop(ctx context.Context, path string, opts ...topicoptions.DropOption) error {
	req := rawtopic.DropTopicRequest{}
	req.OperationParams = c.defaultOperationParams
	req.Path = path

	for _, opt := range opts {
		if opt != nil {
			opt.ApplyDropOption(&req)
		}
	}

	call := func(ctx context.Context) error {
		_, removeErr := c.rawClient.DropTopic(ctx, req)

		return removeErr
	}

	if c.cfg.AutoRetry() {
		return retry.Retry(ctx, call,
			retry.WithIdempotent(true),
			retry.WithTrace(c.cfg.TraceRetry()),
			retry.WithBudget(c.cfg.RetryBudget()),
		)
	}

	return call(ctx)
}

// StartListener starts read listen topic with the handler
// it is fast non block call, connection starts in background
func (c *Client) StartListener(
	consumer string,
	handler topiclistener.EventHandler,
	readSelectors topicoptions.ReadSelectors,
	opts ...topicoptions.ListenerOption,
) (*topiclistener.TopicListener, error) {
	cfg := topiclistenerinternal.NewStreamListenerConfig()

	cfg.Consumer = consumer
	cfg.Tracer = c.cfg.Trace // Set tracer from client config

	cfg.Selectors = make([]*topicreadercommon.PublicReadSelector, len(readSelectors))
	for i := range readSelectors {
		cfg.Selectors[i] = readSelectors[i].Clone()
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(&cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return topiclistener.NewTopicListener(&c.rawClient, &cfg, handler)
}

// StartReader create new topic reader and start pull messages from server
// it is fast non block call, connection will start in background
func (c *Client) StartReader(
	consumer string,
	readSelectors topicoptions.ReadSelectors,
	opts ...topicoptions.ReaderOption,
) (*topicreader.Reader, error) {
	var connector topicreaderinternal.TopicSteamReaderConnect = func(
		ctx context.Context,
		readerID int64,
		tracer *trace.Topic,
	) (
		topicreadercommon.RawTopicReaderStream, error,
	) {
		return c.rawClient.StreamRead(ctx, readerID, tracer)
	}

	defaultOpts := []topicoptions.ReaderOption{
		topicoptions.WithCommonConfig(c.cfg.Common),
		topicreaderinternal.WithCredentials(c.cred),
		topicreaderinternal.WithTrace(c.cfg.Trace),
		topicoptions.WithReaderStartTimeout(topic.DefaultStartTimeout),
	}
	opts = append(defaultOpts, opts...)

	internalReader, err := topicreaderinternal.NewReader(&c.rawClient, connector, consumer, readSelectors, opts...)
	if err != nil {
		return nil, err
	}

	internalReader.TopicOnReaderStart(consumer, err)

	return topicreader.NewReader(internalReader), nil
}

// StartWriter create new topic writer wrapper
func (c *Client) StartWriter(topicPath string, opts ...topicoptions.WriterOption) (*topicwriter.Writer, error) {
	cfg := c.createWriterConfig(topicPath, append(opts, topicwriterinternal.WithMaxGrpcMessageBytes(
		c.cfg.MaxGrpcMessageSize,
	)))
	writer, err := topicwriterinternal.NewWriterReconnector(cfg)
	if err != nil {
		return nil, err
	}

	return topicwriter.NewWriter(writer), nil
}

func (c *Client) StartTransactionalWriter(
	transaction tx.Identifier,
	topicpath string,
	opts ...topicoptions.WriterOption,
) (*topicwriter.TxWriter, error) {
	internalTx, ok := transaction.(tx.Transaction)
	if !ok {
		return nil, xerrors.WithStackTrace(errUnsupportedTransactionType)
	}

	cfg := c.createWriterConfig(topicpath, opts)
	writer, err := topicwriterinternal.NewWriterReconnector(cfg)
	if err != nil {
		return nil, err
	}

	txWriter := topicwriterinternal.NewTopicWriterTransaction(writer, internalTx, cfg.Tracer)

	return topicwriter.NewTxWriterInternal(txWriter), nil
}

func (c *Client) createWriterConfig(
	topicPath string,
	opts []topicoptions.WriterOption,
) topicwriterinternal.WriterReconnectorConfig {
	options := []topicoptions.WriterOption{
		topicwriterinternal.WithRawClient(&c.rawClient),
		topicwriterinternal.WithTopic(topicPath),
		topicwriterinternal.WithCommonConfig(c.cfg.Common),
		topicwriterinternal.WithTrace(c.cfg.Trace),
		topicwriterinternal.WithCredentials(c.cred),
	}

	options = append(options, opts...)

	return topicwriterinternal.NewWriterReconnectorConfig(options...)
}
