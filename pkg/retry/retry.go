package retry

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	sdkoperation "github.com/yandex-cloud/go-sdk/operation"
	"google.golang.org/grpc/status"
)

func ConflictingOperation(ctx context.Context, sdk *ycsdk.SDK, action func() (*operation.Operation, error)) (*sdkoperation.Operation, error) {
	for {
		op, err := sdk.WrapOperation(action())
		if err == nil {
			return op, nil
		}

		operationID := ""
		message := status.Convert(err).Message()
		submatchGoApi := regexp.MustCompile(`conflicting operation "(.+)" detected`).FindStringSubmatch(message)
		submatchPyApi := regexp.MustCompile(`Conflicting operation (.+) detected`).FindStringSubmatch(message)
		if len(submatchGoApi) > 0 {
			operationID = submatchGoApi[1]
		} else if len(submatchPyApi) > 0 {
			operationID = submatchPyApi[1]
		} else {
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
