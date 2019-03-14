// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type cloudResolver struct {
	BaseResolver
}

func CloudResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &cloudResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *cloudResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.ResourceManager().Cloud().List(ctx, &resourcemanager.ListCloudsRequest{
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("cloud", resp.GetClouds(), err)
}
