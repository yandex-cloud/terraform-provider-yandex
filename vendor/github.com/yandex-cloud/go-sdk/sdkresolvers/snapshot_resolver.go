package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	compute "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type snapshotResolver struct {
	BaseResolver
}

func SnapshotResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &snapshotResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *snapshotResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.Compute().Snapshot().List(ctx, &compute.ListSnapshotsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("snapshot", resp.GetSnapshots(), err)
}
