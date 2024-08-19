package client

import (
	"context"
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/draft/Ydb_FederatedQuery_V1"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
)

type YQClient struct {
	conf *YQConfig

	client Ydb_FederatedQuery_V1.FederatedQueryServiceClient
}

func (c YQClient) CreateStorageConnection(
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

func (c YQClient) DescribeStorageConnection(ctx context.Context,
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

func (c YQClient) ModifyStorageConnection(
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

func (c YQClient) DeleteStorageConnection(ctx context.Context,
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

func NewYQClient(ctx context.Context, conf *YQConfig) (*YQClient, error) {
	// the TLS, folderId and authToken parameters were set when creating the YQSDK with interceptors
	conn, err := conf.dialer.GetConnection(ctx, conf.grpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("yq dial connection: %w", err)
	}

	c := Ydb_FederatedQuery_V1.NewFederatedQueryServiceClient(conn)

	return &YQClient{
		conf:   conf,
		client: c,
	}, nil
}
