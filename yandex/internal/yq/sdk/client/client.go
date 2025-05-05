package client

import (
	"context"
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/draft/Ydb_FederatedQuery_V1"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"google.golang.org/grpc"
)

type YQClient interface {
	CreateConnection(context.Context, *Ydb_FederatedQuery.CreateConnectionRequest) (*Ydb_FederatedQuery.CreateConnectionResult, error)
	DescribeConnection(context.Context, *Ydb_FederatedQuery.DescribeConnectionRequest) (*Ydb_FederatedQuery.DescribeConnectionResult, error)
	ModifyConnection(context.Context, *Ydb_FederatedQuery.ModifyConnectionRequest) (*Ydb_FederatedQuery.ModifyConnectionResult, error)
	DeleteConnection(context.Context, *Ydb_FederatedQuery.DeleteConnectionRequest) (*Ydb_FederatedQuery.DeleteConnectionResult, error)
}

type yqClient struct {
	client Ydb_FederatedQuery_V1.FederatedQueryServiceClient
}

func (c yqClient) CreateConnection(
	ctx context.Context,
	req *Ydb_FederatedQuery.CreateConnectionRequest,
) (*Ydb_FederatedQuery.CreateConnectionResult, error) {

	r, err := c.client.CreateConnection(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("create connection: %+v", r)
	}

	var result Ydb_FederatedQuery.CreateConnectionResult

	err = r.GetOperation().GetResult().UnmarshalTo(&result)
	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("create connection: %+v; unmarshal: %w", r, err)
	}

	return &result, nil
}

func (c yqClient) DescribeConnection(ctx context.Context,
	req *Ydb_FederatedQuery.DescribeConnectionRequest,
) (*Ydb_FederatedQuery.DescribeConnectionResult, error) {

	r, err := c.client.DescribeConnection(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("describe connection: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("describe connection: %+v", r)
	}

	var result Ydb_FederatedQuery.DescribeConnectionResult

	err = r.GetOperation().GetResult().UnmarshalTo(&result)
	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("describe connection: %+v; unmarshal: %w", r, err)
	}

	return &result, nil
}

func (c yqClient) ModifyConnection(
	ctx context.Context,
	req *Ydb_FederatedQuery.ModifyConnectionRequest,
) (*Ydb_FederatedQuery.ModifyConnectionResult, error) {

	r, err := c.client.ModifyConnection(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("modify connection: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("modify connection: %+v", r)
	}

	return nil, nil
}

func (c yqClient) DeleteConnection(ctx context.Context,
	req *Ydb_FederatedQuery.DeleteConnectionRequest,
) (*Ydb_FederatedQuery.DeleteConnectionResult, error) {

	r, err := c.client.DeleteConnection(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("delete connection: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("delete connection: %+v", r)
	}

	return nil, nil
}

func NewYQClient(ctx context.Context, clientConn *grpc.ClientConn) *yqClient {
	return &yqClient{
		client: Ydb_FederatedQuery_V1.NewFederatedQueryServiceClient(clientConn),
	}
}
