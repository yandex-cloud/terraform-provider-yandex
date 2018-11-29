// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	"github.com/yandex-cloud/go-sdk"
)

type folderResolver struct {
	BaseResolver
}

func FolderResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &folderResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *folderResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.ResourceManager().Folder().List(ctx, &resourcemanager.ListFoldersRequest{
		CloudId:  r.CloudID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("folder", resp.GetFolders(), err)
}
