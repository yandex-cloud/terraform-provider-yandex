package trino_access_control

import (
	"context"
	"fmt"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const (
	operationWaitInterval = 5 * time.Second
	baseUpdatePath        = "trino.access_control"
)

func DeleteClusterAccessControl(ctx context.Context, sdk *ycsdk.SDK, clusterID string) diag.Diagnostics {
	return UpdateClusterAccessControl(ctx, sdk, clusterID, nil)
}

func UpdateClusterAccessControl(ctx context.Context, sdk *ycsdk.SDK, clusterID string, cfg *trino.AccessControlConfig) diag.Diagnostics {
	req := &trino.UpdateClusterRequest{
		ClusterId: clusterID,
		Trino: &trino.UpdateTrinoConfigSpec{
			AccessControl: cfg,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{baseUpdatePath}},
	}
	tflog.Debug(ctx, fmt.Sprintf("Update Trino cluster request: %+v", req))

	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.Trino().Cluster().Update(ctx, req)
	})
	var diags diag.Diagnostics
	if err != nil {
		diags.AddError("Failed to update Trino access control", "API request finished with error: "+err.Error())
		return diags
	}

	err = op.WaitInterval(ctx, operationWaitInterval)
	if err != nil {
		diags.AddError("Failed to update Trino access control", "Waiting operation to complete finished with error: "+err.Error())
		return diags
	}

	return nil
}

func GetClusterAccessControl(ctx context.Context, sdk *ycsdk.SDK, clusterID string) (*trino.AccessControlConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cluster, err := sdk.Trino().Cluster().Get(ctx, &trino.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}
		diags.AddError("Failed to read Trino access control config",
			"Error while requesting API to get Trino access control config: "+err.Error())
		return nil, diags
	}
	return cluster.Trino.AccessControl, nil
}
