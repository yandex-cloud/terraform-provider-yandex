package rawtopic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Topic_V1"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicwriter"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Client struct {
	service Ydb_Topic_V1.TopicServiceClient
}

func NewClient(service Ydb_Topic_V1.TopicServiceClient) Client {
	return Client{service: service}
}

func (c *Client) AlterTopic(ctx context.Context, req *AlterTopicRequest) (res AlterTopicResult, err error) {
	resp, err := c.service.AlterTopic(ctx, req.ToProto())
	if err != nil {
		return res, xerrors.WithStackTrace(fmt.Errorf("ydb: alter topic grpc failed: %w", err))
	}
	err = res.FromProto(resp)

	return res, err
}

func (c *Client) CreateTopic(
	ctx context.Context,
	req *CreateTopicRequest,
) (res CreateTopicResult, err error) {
	resp, err := c.service.CreateTopic(ctx, req.ToProto())
	if err != nil {
		return res, xerrors.WithStackTrace(fmt.Errorf("ydb: create topic grpc failed: %w", err))
	}
	err = res.FromProto(resp)

	return res, err
}

func (c *Client) DescribeTopic(ctx context.Context, req DescribeTopicRequest) (res DescribeTopicResult, err error) {
	resp, err := c.service.DescribeTopic(ctx, req.ToProto())
	if err != nil {
		return DescribeTopicResult{}, xerrors.WithStackTrace(
			xerrors.Wrap(
				fmt.Errorf("ydb: describe topic grpc failed: %w", err),
			),
		)
	}
	err = res.FromProto(resp)

	return res, err
}

func (c *Client) DescribeConsumer(ctx context.Context, req DescribeConsumerRequest) (
	res DescribeConsumerResult, err error,
) {
	resp, err := c.service.DescribeConsumer(ctx, req.ToProto())
	if err != nil {
		return DescribeConsumerResult{}, xerrors.WithStackTrace(
			xerrors.Wrap(
				fmt.Errorf("ydb: describe topic consumer grpc failed: %w", err),
			),
		)
	}
	err = res.FromProto(resp)

	return res, err
}

func (c *Client) DropTopic(
	ctx context.Context,
	req DropTopicRequest,
) (res DropTopicResult, err error) {
	resp, err := c.service.DropTopic(ctx, req.ToProto())
	if err != nil {
		return res, xerrors.WithStackTrace(fmt.Errorf("ydb: drop topic grpc failed: %w", err))
	}
	err = res.FromProto(resp)

	return res, err
}

func (c *Client) StreamRead(
	ctxStreamLifeTime context.Context,
	readerID int64,
	tracer *trace.Topic,
) (rawtopicreader.StreamReader, error) {
	protoResp, err := c.service.StreamRead(ctxStreamLifeTime)
	if err != nil {
		return rawtopicreader.StreamReader{}, xerrors.WithStackTrace(
			xerrors.Wrap(
				fmt.Errorf("ydb: failed start grpc topic stream read: %w", err),
			),
		)
	}

	return rawtopicreader.StreamReader{
		Stream:   protoResp,
		ReaderID: readerID,
		Tracer:   tracer,
	}, nil
}

func (c *Client) StreamWrite(
	ctxStreamLifeTime context.Context,
	tracer *trace.Topic,
) (*rawtopicwriter.StreamWriter, error) {
	protoResp, err := c.service.StreamWrite(ctxStreamLifeTime)
	if err != nil {
		return nil, xerrors.WithStackTrace(
			xerrors.Wrap(
				fmt.Errorf("ydb: failed start grpc topic stream write: %w", err),
			),
		)
	}

	return &rawtopicwriter.StreamWriter{
		Stream:           protoResp,
		Tracer:           tracer,
		InternalStreamID: uuid.New().String(),
		LogContext:       &ctxStreamLifeTime,
	}, nil
}

func (c *Client) UpdateOffsetsInTransaction(
	ctx context.Context,
	req *UpdateOffsetsInTransactionRequest,
) error {
	protoResp, err := c.service.UpdateOffsetsInTransaction(ctx, req.ToProto())
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: update offsets in transaction failed: %w", err))
	}

	var operation rawydb.Operation

	return operation.FromProtoWithStatusCheck(protoResp.GetOperation())
}
