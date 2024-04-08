package utils

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	sdkoperation "github.com/yandex-cloud/go-sdk/operation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const DefaultTimeout = 1 * time.Minute
const DefaultPageSize = 1000
const defaultTimeFormat = time.RFC3339

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetTimestamp(timestamp *timestamppb.Timestamp) string {
	if timestamp == nil {
		return ""
	}
	return timestamp.AsTime().Format(defaultTimeFormat)
}

func RetryConflictingOperation(ctx context.Context, sdk *ycsdk.SDK, action func() (*operation.Operation, error)) (*sdkoperation.Operation, error) {
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

// isStatusWithCode checks if any nested error matches provided code
func IsStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	check := ok && grpcStatus.Code() == code

	nestedErr := errors.Unwrap(err)
	if nestedErr != nil {
		return IsStatusWithCode(nestedErr, code) || check
	}
	return check
}

func ConstructResourceId(clusterID string, resourceName string) string {
	return fmt.Sprintf("%s:%s", clusterID, resourceName)
}

func DeconstructResourceId(resourceID string) (string, string, error) {
	parts := strings.SplitN(resourceID, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid resource id format: %q", resourceID)
	}

	clusterID := parts[0]
	resourceName := parts[1]
	return clusterID, resourceName, nil
}
