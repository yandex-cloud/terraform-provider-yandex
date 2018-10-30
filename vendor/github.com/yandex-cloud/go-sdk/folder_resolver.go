// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdk

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
)

type folderResolver struct {
	BaseResolver
}

func FolderResolver(name string, opts ...ResolveOption) Resolver {
	return &folderResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *folderResolver) Run(ctx context.Context, sdk *SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.ResourceManager().Folder().List(ctx, &resourcemanager.ListFoldersRequest{
		CloudId:  r.CloudID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("folder", resp.GetFolders(), err)
}
