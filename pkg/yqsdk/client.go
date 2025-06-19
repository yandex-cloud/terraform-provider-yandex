package yqsdk

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
	ModifyConnection(context.Context, *Ydb_FederatedQuery.ModifyConnectionRequest) error
	DeleteConnection(context.Context, *Ydb_FederatedQuery.DeleteConnectionRequest) error

	CreateBinding(context.Context, *Ydb_FederatedQuery.CreateBindingRequest) (*Ydb_FederatedQuery.CreateBindingResult, error)
	DescribeBinding(context.Context, *Ydb_FederatedQuery.DescribeBindingRequest) (*Ydb_FederatedQuery.DescribeBindingResult, error)
	ModifyBinding(context.Context, *Ydb_FederatedQuery.ModifyBindingRequest) error
	DeleteBinding(context.Context, *Ydb_FederatedQuery.DeleteBindingRequest) error
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
	if err != nil {
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
	if err != nil {
		return nil, fmt.Errorf("describe connection: %+v; unmarshal: %w", r, err)
	}

	return &result, nil
}

func (c yqClient) ModifyConnection(
	ctx context.Context,
	req *Ydb_FederatedQuery.ModifyConnectionRequest,
) error {

	r, err := c.client.ModifyConnection(ctx, req)
	if err != nil {
		return fmt.Errorf("modify connection: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return fmt.Errorf("modify connection: %+v", r)
	}

	return nil
}

func (c yqClient) DeleteConnection(ctx context.Context,
	req *Ydb_FederatedQuery.DeleteConnectionRequest,
) error {

	r, err := c.client.DeleteConnection(ctx, req)
	if err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return fmt.Errorf("delete connection: %+v", r)
	}

	return nil
}

func (c yqClient) CreateBinding(
	ctx context.Context,
	req *Ydb_FederatedQuery.CreateBindingRequest,
) (*Ydb_FederatedQuery.CreateBindingResult, error) {

	r, err := c.client.CreateBinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create binding: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("create binding: %+v", r)
	}

	var result Ydb_FederatedQuery.CreateBindingResult

	err = r.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return nil, fmt.Errorf("create binding: %+v; unmarshal: %w", r, err)
	}

	return &result, nil
}

func (c yqClient) DescribeBinding(ctx context.Context,
	req *Ydb_FederatedQuery.DescribeBindingRequest,
) (*Ydb_FederatedQuery.DescribeBindingResult, error) {

	r, err := c.client.DescribeBinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("describe binding: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return nil, fmt.Errorf("describe binding: %+v", r)
	}

	var result Ydb_FederatedQuery.DescribeBindingResult

	err = r.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return nil, fmt.Errorf("describe binding: %+v; unmarshal: %w", r, err)
	}

	return &result, nil
}

func (c yqClient) ModifyBinding(
	ctx context.Context,
	req *Ydb_FederatedQuery.ModifyBindingRequest,
) error {

	r, err := c.client.ModifyBinding(ctx, req)
	if err != nil {
		return fmt.Errorf("modify binding: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return fmt.Errorf("modify binding: %+v", r)
	}

	return nil
}

func (c yqClient) DeleteBinding(ctx context.Context,
	req *Ydb_FederatedQuery.DeleteBindingRequest,
) error {

	r, err := c.client.DeleteBinding(ctx, req)
	if err != nil {
		return fmt.Errorf("delete binding: %w", err)
	}

	if r.GetOperation().GetStatus() != Ydb.StatusIds_SUCCESS {
		return fmt.Errorf("delete binding: %+v", r)
	}

	return nil
}

func NewYQClient(ctx context.Context, clientConn *grpc.ClientConn) *yqClient {
	return &yqClient{
		client: Ydb_FederatedQuery_V1.NewFederatedQueryServiceClient(clientConn),
	}
}
