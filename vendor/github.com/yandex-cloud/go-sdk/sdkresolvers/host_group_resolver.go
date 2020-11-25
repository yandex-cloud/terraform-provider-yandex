package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type hostGroupResolver struct {
	BaseNameResolver
}

func HostGroupResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &hostGroupResolver{
		BaseNameResolver: NewBaseNameResolver(name, "host group", opts...),
	}
}

func (r *hostGroupResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	if err := r.ensureFolderID(); err != nil {
		return err
	}

	resp, err := sdk.Compute().HostGroup().List(ctx, &compute.ListHostGroupsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)

	return r.findName(resp.GetHostGroups(), err)
}
