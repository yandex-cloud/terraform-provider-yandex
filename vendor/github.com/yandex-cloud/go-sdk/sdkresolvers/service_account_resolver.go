// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-sdk"
)

type serviceAccountResolver struct {
	BaseResolver
}

func ServiceAccountResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &serviceAccountResolver{
		BaseResolver: NewBaseResolver(name, opts...),
	}
}

func (r *serviceAccountResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.IAM().ServiceAccount().List(ctx, &iam.ListServiceAccountsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName("service account", resp.GetServiceAccounts(), err)
}
