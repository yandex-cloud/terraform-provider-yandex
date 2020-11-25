package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type diskPlacementGroupResolver struct {
	BaseNameResolver
}

func DiskPlacementGroupResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &diskPlacementGroupResolver{
		BaseNameResolver: NewBaseNameResolver(name, "disk placement group", opts...),
	}
}

func (r *diskPlacementGroupResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	if err := r.ensureFolderID(); err != nil {
		return err
	}

	resp, err := sdk.Compute().DiskPlacementGroup().List(ctx, &compute.ListDiskPlacementGroupsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)

	return r.findName(resp.GetDiskPlacementGroups(), err)
}
