package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk"
)

type diskResolver struct {
	BaseResolver
}

func DiskResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &diskResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *diskResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.Compute().Disk().List(ctx, &compute.ListDisksRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("disk", resp.GetDisks(), err)
}
