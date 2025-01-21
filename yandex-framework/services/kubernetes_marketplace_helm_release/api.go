package kubernetes_marketplace_helm_release

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	marketplace "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/marketplace/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func installHelmRelease(ctx context.Context, sdk *ycsdk.SDK, req *marketplace.InstallHelmReleaseRequest) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	op, err := sdk.WrapOperation(sdk.KubernetesMarketplace().HelmRelease().Install(ctx, req))
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Failed to install Helm Release",
			"Error while requesting API to install Helm Release: "+err.Error(),
		))
	}

	if op == nil {
		return "", diags
	}

	err = op.WaitInterval(ctx, 5*time.Second)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Failed to install Helm Release",
			"Error while requesting API to install Helm Release. Failed to wait: "+err.Error(),
		))
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Failed to install Helm Release",
			"Failed to unmarshal metadata: "+err.Error(),
		))
	}

	meta, ok := protoMetadata.(*marketplace.InstallHelmReleaseMetadata)
	if !ok {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Failed to install Helm Release",
			"Failed to convert response metadata to InstallHelmReleaseMetadata",
		))
		return "", diags
	}

	return meta.GetHelmReleaseId(), diags
}

func getHelmRelease(ctx context.Context, sdk *ycsdk.SDK, id string) (*marketplace.HelmRelease, diag.Diagnostic) {
	helmRelease, err := sdk.KubernetesMarketplace().HelmRelease().Get(ctx, &marketplace.GetHelmReleaseRequest{
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

	return waitOperation(ctx, sdk, "update Helm release", func() (*operation.Operation, error) {
		return sdk.KubernetesMarketplace().HelmRelease().Update(ctx, req)
	})
}

func uninstallHelmRelease(ctx context.Context, sdk *ycsdk.SDK, req *marketplace.UninstallHelmReleaseRequest) diag.Diagnostic {
	if req == nil {
		return nil
	}

	return waitOperation(ctx, sdk, "delete Helm release", func() (*operation.Operation, error) {
		return sdk.KubernetesMarketplace().HelmRelease().Uninstall(ctx, req)
	})
}

func waitOperation(ctx context.Context, sdk *ycsdk.SDK, action string, callback func() (*operation.Operation, error)) diag.Diagnostic {
	op, err := retry.ConflictingOperation(ctx, sdk, callback)

	if err == nil {
		err = op.Wait(ctx)
	}

	if err != nil {
		return diag.NewErrorDiagnostic(fmt.Sprintf("Failed to %s", action), err.Error())
	}

	return nil
}
