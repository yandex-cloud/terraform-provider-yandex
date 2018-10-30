// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package sdk

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

// OperationServiceClient is a operation.OperationServiceClient with
// lazy GRPC connection initialization.
type OperationServiceClient struct {
	getConn lazyConn
}

// Get implements operation.OperationServiceClient
func (o *OperationServiceClient) Get(ctx context.Context, in *operation.GetOperationRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	conn, err := o.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return operation.NewOperationServiceClient(conn).Get(ctx, in, opts...)
}

// Cancel implements operation.OperationServiceClient
func (o *OperationServiceClient) Cancel(ctx context.Context, in *operation.CancelOperationRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	conn, err := o.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return operation.NewOperationServiceClient(conn).Cancel(ctx, in, opts...)
}
