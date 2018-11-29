// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	compute "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk"
)

type instanceResolver struct {
	BaseResolver
}

func InstanceResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &instanceResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *instanceResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.Compute().Instance().List(ctx, &compute.ListInstancesRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("instance", resp.GetInstances(), err)
}
