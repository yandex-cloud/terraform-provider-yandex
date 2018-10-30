// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdk

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
)

type cloudResolver struct {
	BaseResolver
}

func CloudResolver(name string, opts ...ResolveOption) Resolver {
	return &cloudResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *cloudResolver) Run(ctx context.Context, sdk *SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.ResourceManager().Cloud().List(ctx, &resourcemanager.ListCloudsRequest{
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("cloud", resp.GetClouds(), err)
}
