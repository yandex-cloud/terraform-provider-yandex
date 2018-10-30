// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package sdk

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

type networkResolver struct {
	BaseResolver
}

func NetworkResolver(name string, opts ...ResolveOption) Resolver {
	return &networkResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *networkResolver) Run(ctx context.Context, sdk *SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.VPC().Network().List(ctx, &vpc.ListNetworksRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("network", resp.GetNetworks(), err)
}

type subnetResolver struct {
	BaseResolver
}

func SubnetResolver(name string, opts ...ResolveOption) Resolver {
	return &subnetResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *subnetResolver) Run(ctx context.Context, sdk *SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.VPC().Subnet().List(ctx, &vpc.ListSubnetsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("subnet", resp.GetSubnets(), err)
}
