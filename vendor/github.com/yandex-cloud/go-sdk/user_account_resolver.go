// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdk

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

type userAccountByLoginResolver struct {
	BaseResolver
}

func UserAccountByLoginResolver(login string, opts ...ResolveOption) Resolver {
	return &userAccountByLoginResolver{
		BaseResolver: NewBaseResolver(login, opts...),
	}
}

func (r *userAccountByLoginResolver) Run(ctx context.Context, sdk *SDK, opts ...grpc.CallOption) error {
	return r.Set(sdk.IAM().YandexPassportUserAccount().GetByLogin(ctx, &iam.GetUserAccountByLoginRequest{
		Login: r.Name,
	}, opts...))
}
