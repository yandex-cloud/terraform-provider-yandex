package pluginsdk

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkoperationv2 "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

type OperationRetryConfig struct {
	retriableCodes []codes.Code
	retryCount     int
	retryInterval  time.Duration
	pollInterval   time.Duration
}

func waitOperationWithRetryV2(ctx context.Context, config *providerconfig.Config, retryConfig *OperationRetryConfig, action func() (sdkcommon.SdkV2Operation, error)) error {
	var err error
	for step := 0; step < retryConfig.retryCount; step++ {
		var op sdkoperationv2.AbstractOperation
		op, err = sdkcommon.RetryConflictingOperationV2(ctx, config, action)
		if err != nil {
			return err
		}

		err = WaitOperationV2(ctx, op, retryConfig.pollInterval)
		if shouldRetryOperationV2(op, retryConfig.retriableCodes, err) {
			time.Sleep(retryConfig.retryInterval)
			continue
		}
		break
	}
	return err
}

func WaitOperationV2(ctx context.Context, op sdkoperationv2.AbstractOperation, pollInterval time.Duration) error {
	concrete, ok := op.(*sdkoperationv2.Operation)
	if !ok {
		return fmt.Errorf("unsupported SDK v2 operation type %T", op)
	}
	_, err := concrete.WaitInterval(ctx, func(int) time.Duration { return pollInterval })
	return err
}

func shouldRetryOperationV2(op sdkoperationv2.AbstractOperation, retriableCodes []codes.Code, err error) bool {
	if err == nil {
		return op.Done() && op.Error() != nil
	}
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return false
	}
	for _, code := range retriableCodes {
		if grpcStatus.Code() == code {
			return true
		}
	}
	return false
}
