// Copyright (c) 2019 Yandex LLC. All rights reserved.
// Author: Vasiliy Briginets <0x40@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	loadbalancer "github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type networkLoadBalancerResolver struct {
	BaseResolver
}

func NetworkLoadBalancerResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &networkLoadBalancerResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *networkLoadBalancerResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.LoadBalancer().NetworkLoadBalancer().List(ctx, &loadbalancer.ListNetworkLoadBalancersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("network load balancer", resp.GetNetworkLoadBalancers(), err)
}

type targetGroupResolver struct {
	BaseResolver
}

func TargetGroupResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &targetGroupResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *targetGroupResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.LoadBalancer().TargetGroup().List(ctx, &loadbalancer.ListTargetGroupsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("target group", resp.GetTargetGroups(), err)
}
