// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdk

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

const DefaultResolverPageSize = 100

func CreateResolverFilter(nameField string, value string) string {
	// TODO(novikoff): should we add escaping or value validation?
	return fmt.Sprintf(`%s = "%s"`, nameField, value)
}

type Resolver interface {
	ID() string
	Err() error

	Run(context.Context, *SDK, ...grpc.CallOption) error
}
