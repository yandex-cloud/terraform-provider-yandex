package retry

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	sdkoperation "github.com/yandex-cloud/go-sdk/operation"
	ycsdkv2 "github.com/yandex-cloud/go-sdk/v2"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	operationServiceGetMethod = protoreflect.FullName("yandex.cloud.operation.OperationService.Get")
	conflictingOperationPoll  = time.Second
)

var (
	conflictingOperationGoAPIRe = regexp.MustCompile(`conflicting operation "(.+)" detected`)
	conflictingOperationPyAPIRe = regexp.MustCompile(`Conflicting operation (.+) detected`)
)

// conflictingOperationID extracts the ID of the operation that conflicts with the current
// request from the API error message, or returns an empty string if the error is unrelated.
func conflictingOperationID(err error) string {
	message := status.Convert(err).Message()
	if m := conflictingOperationGoAPIRe.FindStringSubmatch(message); len(m) > 0 {
		return m[1]
	}
	if m := conflictingOperationPyAPIRe.FindStringSubmatch(message); len(m) > 0 {
		return m[1]
	}
	return ""
}

// ConflictingOperation runs action and, when the API rejects it because another operation is
// already running on the same resource, waits for that operation to finish and retries.
func ConflictingOperation(ctx context.Context, sdk *ycsdk.SDK, action func() (*operation.Operation, error)) (*sdkoperation.Operation, error) {
	for {
		op, err := sdk.WrapOperation(action())
		if err == nil {
			return op, nil
		}

		operationID := conflictingOperationID(err)
		if operationID == "" {
			return op, err
		}

		tflog.Debug(ctx, fmt.Sprintf("Waiting for conflicting operation %q to complete", operationID))
		req := &operation.GetOperationRequest{OperationId: operationID}
		op, err = sdk.WrapOperation(sdk.Operation().Get(ctx, req))
		if err != nil {
			return nil, err
		}

		_ = op.Wait(ctx)
		tflog.Debug(ctx, fmt.Sprintf("Conflicting operation %q has completed. Going to retry initial action.", operationID))
	}
}

// ConflictingOperationV2 is the modular SDK (SDKv2) counterpart of ConflictingOperation. Its
// typed clients return their own operation wrappers, so it is generic over the returned
// operation type and only retries action after waiting for the conflicting operation.
func ConflictingOperationV2[O any](ctx context.Context, sdk *ycsdkv2.SDK, action func() (O, error)) (O, error) {
	for {
		op, err := action()
		if err == nil {
			return op, nil
		}

		operationID := conflictingOperationID(err)
		if operationID == "" {
			return op, err
		}

		tflog.Debug(ctx, fmt.Sprintf("Waiting for conflicting operation %q to complete", operationID))
		if waitErr := waitOperationV2(ctx, sdk, operationID); waitErr != nil {
			var zero O
			return zero, waitErr
		}
		tflog.Debug(ctx, fmt.Sprintf("Conflicting operation %q has completed. Going to retry initial action.", operationID))
	}
}

// waitOperationV2 polls the operation with the given ID via the modular SDK until it is done.
func waitOperationV2(ctx context.Context, sdk *ycsdkv2.SDK, operationID string) error {
	conn, err := sdk.GetConnection(ctx, operationServiceGetMethod)
	if err != nil {
		return err
	}
	client := operation.NewOperationServiceClient(conn)
	for {
		op, err := client.Get(ctx, &operation.GetOperationRequest{OperationId: operationID})
		if err != nil {
			return err
		}
		if op.GetDone() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(conflictingOperationPoll):
		}
	}
}
