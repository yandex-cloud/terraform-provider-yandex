package gitlab_instance

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"
	gitlabsdk "github.com/yandex-cloud/go-sdk/services/gitlab/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateInstance(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *gitlab.CreateInstanceRequest) (string, diag.Diagnostic) {
	op, err := gitlabsdk.NewInstanceClient(sdk).Create(ctx, req)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Gitlab instance",
			"Error while requesting API to create Gitalb instance: "+err.Error(),
		)
	}

	_, err = op.WaitInterval(ctx, func(int) time.Duration { return 5 * time.Second })
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Gitlab instance",
			"Error while requesting API to create Gitlab instance. Failed to wait: "+err.Error(),
		)
	}

	return op.Metadata().InstanceId, nil
}

func GetInstanceByID(ctx context.Context, sdk *ycsdk.SDK, id string) (*gitlab.Instance, diag.Diagnostic) {
	instance, err := gitlabsdk.NewInstanceClient(sdk).Get(ctx, &gitlab.GetInstanceRequest{
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

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*gitlabsdk.InstanceDeleteOperation, error) {
		return gitlabsdk.NewInstanceClient(sdk).Delete(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to delete Gitlab instance", err.Error())
	}
	return nil
}

func UpdateInstance(ctx context.Context, sdk *ycsdk.SDK, req *gitlab.UpdateInstanceRequest) diag.Diagnostic {
	if req == nil || req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*gitlabsdk.InstanceUpdateOperation, error) {
		return gitlabsdk.NewInstanceClient(sdk).Update(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to update Gitlab instance", err.Error())
	}
	return nil
}
