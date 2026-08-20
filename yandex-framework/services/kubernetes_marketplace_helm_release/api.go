package kubernetes_marketplace_helm_release

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	marketplace "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/marketplace/v1"
	operation "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	marketplacesdk "github.com/yandex-cloud/go-sdk/services/k8s/marketplace/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	sdkop "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func installHelmRelease(ctx context.Context, sdk *ycsdk.SDK, req *marketplace.InstallHelmReleaseRequest) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	op, err := marketplacesdk.NewHelmReleaseClient(sdk).Install(ctx, req)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Failed to install Helm Release",
			"Error while requesting API to install Helm Release: "+err.Error(),
		))
	}

	if op == nil {
		return "", diags
	}

	err = waitHelmReleaseOperation(ctx, sdk, op.ID())
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Failed to install Helm Release",
			"Error while requesting API to install Helm Release. Failed to wait: "+err.Error(),
		))
	}

	return op.Metadata().GetHelmReleaseId(), diags
}

func waitHelmReleaseOperation(ctx context.Context, sdk *ycsdk.SDK, operationID string) error {
	conn, err := sdk.GetConnection(ctx, marketplacesdk.HelmReleaseOperationPoller)
	if err != nil {
		return err
	}

	client := operation.NewOperationServiceClient(conn)
	poll := func(ctx context.Context, operationID string, opts ...grpc.CallOption) (sdkop.YCOperation, error) {
		return client.Get(ctx, &operation.GetOperationRequest{OperationId: operationID}, opts...)
	}

	return waitOperationWithoutResponse(ctx, operationID, poll, func(int) time.Duration { return 5 * time.Second })
}

func waitOperationWithoutResponse(
	ctx context.Context,
	operationID string,
	poll sdkop.PollFunc,
	pollInterval sdkop.PollIntervalFunc,
) error {
	op, err := sdkop.PollUntilDone(ctx, operationID, poll, pollInterval)
	if err != nil {
		return err
	}
	if err := sdkop.Error(op); err != nil {
		return fmt.Errorf("operation (id=%s) failed: %w", operationID, err)
	}
	return nil
}

func getHelmRelease(ctx context.Context, sdk *ycsdk.SDK, id string) (*marketplace.HelmRelease, diag.Diagnostic) {
	helmRelease, err := marketplacesdk.NewHelmReleaseClient(sdk).Get(ctx, &marketplace.GetHelmReleaseRequest{
		Id: id,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}

		return nil, diag.NewErrorDiagnostic(
			"Failed to read Helm Release",
			"Error while requesting API to get Helm Release: "+err.Error(),
		)
	}
	return helmRelease, nil
}

func updateHelmRelease(ctx context.Context, sdk *ycsdk.SDK, req *marketplace.UpdateHelmReleaseRequest) diag.Diagnostic {
	if req == nil {
		return nil
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*marketplacesdk.HelmReleaseUpdateOperation, error) {
		return marketplacesdk.NewHelmReleaseClient(sdk).Update(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to update Helm release", err.Error())
	}
	return nil
}

func uninstallHelmRelease(ctx context.Context, sdk *ycsdk.SDK, req *marketplace.UninstallHelmReleaseRequest) diag.Diagnostic {
	if req == nil {
		return nil
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*marketplacesdk.HelmReleaseUninstallOperation, error) {
		return marketplacesdk.NewHelmReleaseClient(sdk).Uninstall(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to delete Helm release", err.Error())
	}
	return nil
}
