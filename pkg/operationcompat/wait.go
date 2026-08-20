// Package operationcompat contains provider-side compatibility helpers for
// long-running operations whose actual payload differs from the SDK v2 type.
package operationcompat

import (
	"context"
	"fmt"
	"time"

	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	sdkop "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const operationServiceGetMethod = protoreflect.FullName("yandex.cloud.operation.OperationService.Get")

// wait waits for an operation and checks its service error without decoding a
// successful response. This preserves SDK v1 behavior for operations whose
// service response does not match the stricter SDK v2 declaration.
func wait(ctx context.Context, operationID string, poll sdkop.PollFunc) (sdkop.YCOperation, error) {
	op, err := sdkop.PollUntilDone(ctx, operationID, poll, func(int) time.Duration { return time.Second })
	if err != nil {
		return nil, err
	}
	if err := sdkop.Error(op); err != nil {
		return nil, fmt.Errorf("operation (id=%s) failed: %w", operationID, err)
	}
	return op, nil
}

func waitResponse(ctx context.Context, operationID string, poll sdkop.PollFunc, response proto.Message) error {
	op, err := wait(ctx, operationID, poll)
	if err != nil {
		return err
	}
	if op.GetResponse() == nil {
		return fmt.Errorf("operation (id=%s) returned no response", operationID)
	}
	return op.GetResponse().UnmarshalTo(response)
}

func operationPoll(ctx context.Context, sdk *ycsdk.SDK) (sdkop.PollFunc, error) {
	conn, err := sdk.GetConnection(ctx, operationServiceGetMethod)
	if err != nil {
		return nil, err
	}
	client := operationpb.NewOperationServiceClient(conn)
	return func(ctx context.Context, operationID string, opts ...grpc.CallOption) (sdkop.YCOperation, error) {
		return client.Get(ctx, &operationpb.GetOperationRequest{OperationId: operationID}, opts...)
	}, nil
}

// Wait polls an operation through the common OperationService without
// decoding its successful response or metadata.
func Wait(ctx context.Context, sdk *ycsdk.SDK, operationID string) error {
	poll, err := operationPoll(ctx, sdk)
	if err != nil {
		return err
	}
	_, err = wait(ctx, operationID, poll)
	return err
}

// WaitResponse polls an operation without decoding metadata and unmarshals its
// successful response into response.
func WaitResponse(ctx context.Context, sdk *ycsdk.SDK, operationID string, response proto.Message) error {
	poll, err := operationPoll(ctx, sdk)
	if err != nil {
		return err
	}
	return waitResponse(ctx, operationID, poll, response)
}
