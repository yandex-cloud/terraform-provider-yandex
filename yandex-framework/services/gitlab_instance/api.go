package gitlab_instance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateInstance(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *gitlab.CreateInstanceRequest) (string, diag.Diagnostic) {
	op, err := sdk.WrapOperation(sdk.Gitlab().Instance().Create(ctx, req))
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Gitlab instance",
			"Error while requesting API to create Gitalb instance: "+err.Error(),
		)
	}

	err = op.WaitInterval(ctx, 5*time.Second)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Gitlab instance",
			"Error while requesting API to create Gitlab instance. Failed to wait: "+err.Error(),
		)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Gitlab instance",
			"Failed to unmarshal metadata: "+err.Error(),
		)
	}

	md, ok := protoMetadata.(*gitlab.CreateInstanceMetadata)
	if !ok {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Gitlab instance",
			"Failed to convert response metadata to CreateInstanceMetadata",
		)
	}

	return md.InstanceId, nil
}

func GetInstanceByID(ctx context.Context, sdk *ycsdk.SDK, id string) (*gitlab.Instance, diag.Diagnostic) {
	instance, err := sdk.Gitlab().Instance().Get(ctx, &gitlab.GetInstanceRequest{
		InstanceId: id,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}

		return nil, diag.NewErrorDiagnostic(
			"Failed to read Gitlab instance",
			"Error while requesting API to get Gitlab instance: "+err.Error(),
		)
	}
	return instance, nil
}

func DeleteInstance(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &gitlab.DeleteInstanceRequest{
		InstanceId: cid,
	}

	return waitOperation(ctx, sdk, "delete Gitlab instance", func() (*operation.Operation, error) {
		return sdk.Gitlab().Instance().Delete(ctx, req)
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
