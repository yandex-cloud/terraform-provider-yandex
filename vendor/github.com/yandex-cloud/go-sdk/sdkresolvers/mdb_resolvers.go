package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	redis "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1alpha"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func PostgreSQLClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &postgreSQLClusterResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

type postgreSQLClusterResolver struct {
	BaseResolver
}

func (r *postgreSQLClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.MDB().PostgreSQL().Cluster().List(ctx, &postgresql.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName("cluster", resp.GetClusters(), err)
}

func MongoDBClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &mongodbClusterResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

type mongodbClusterResolver struct {
	BaseResolver
}

func (r *mongodbClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.MDB().MongoDB().Cluster().List(ctx, &mongodb.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName("cluster", resp.GetClusters(), err)
}

func ClickhouseClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &clickhouseClusterResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

type clickhouseClusterResolver struct {
	BaseResolver
}

func (r *clickhouseClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.MDB().Clickhouse().Cluster().List(ctx, &clickhouse.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName("cluster", resp.GetClusters(), err)
}

func RedisClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &redisClusterResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

type redisClusterResolver struct {
	BaseResolver
}

func (r *redisClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.MDB().Redis().Cluster().List(ctx, &redis.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName("cluster", resp.GetClusters(), err)
}
