// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/pkg/sdkerrors"
)

type imageResolver struct {
	BaseResolver
}

func ImageResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &imageResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *imageResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.Compute().Image().List(ctx, &compute.ListImagesRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("image", resp.GetImages(), err)
}

type imageByFamilyResolver struct {
	BaseResolver
}

func ImageByFamilyResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &imageByFamilyResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *imageByFamilyResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	img, err := sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
		FolderId: r.FolderID(),
		Family:   r.Name,
	})
	if err != nil {
		err = sdkerrors.WithMessagef(err, "failed to find image with family \"%v\"", r.Name)
	}
	return r.Set(img, err)
}
