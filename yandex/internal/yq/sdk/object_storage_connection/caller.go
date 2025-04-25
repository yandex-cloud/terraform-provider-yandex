package object_storage_connection

import (
	"context"

	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ObjectStorageClient interface {
	CreateStorageConnection(context.Context, *Ydb_FederatedQuery.CreateConnectionRequest) (*Ydb_FederatedQuery.CreateConnectionResult, error)
	DescribeStorageConnection(context.Context, *Ydb_FederatedQuery.DescribeConnectionRequest) (*Ydb_FederatedQuery.DescribeConnectionResult, error)
	ModifyStorageConnection(context.Context, *Ydb_FederatedQuery.ModifyConnectionRequest) (*Ydb_FederatedQuery.ModifyConnectionResult, error)
	DeleteStorageConnection(context.Context, *Ydb_FederatedQuery.DeleteConnectionRequest) (*Ydb_FederatedQuery.DeleteConnectionResult, error)
}
