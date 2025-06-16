package trino_cluster

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *trino.CreateClusterRequest) (string, diag.Diagnostic) {
	op, err := sdk.WrapOperation(sdk.Trino().Cluster().Create(ctx, req))
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino cluster",
			"Error while requesting API to create Trino cluster: "+err.Error(),
		)
	}

	err = op.WaitInterval(ctx, 5*time.Second)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino cluster",
			"Error while requesting API to create Trino cluster. Failed to wait: "+err.Error(),
		)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino cluster",
			"Failed to unmarshal metadata: "+err.Error(),
		)
	}

	md, ok := protoMetadata.(*trino.CreateClusterMetadata)
	if !ok {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino cluster",
			"Failed to convert response metadata to CreateClusterMetadata",
		)
	}

	return md.ClusterId, nil
}

func GetClusterByID(ctx context.Context, sdk *ycsdk.SDK, cid string) (*trino.Cluster, diag.Diagnostic) {
	cluster, err := sdk.Trino().Cluster().Get(ctx, &trino.GetClusterRequest{
		ClusterId: cid,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}

		return nil, diag.NewErrorDiagnostic(
			"Failed to read Trino cluster",
			"Error while requesting API to get Trino cluster: "+err.Error(),
		)
	}
	return cluster, nil
}

func UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, req *trino.UpdateClusterRequest) diag.Diagnostic {
	if req == nil || req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil
	}

	return waitOperation(ctx, sdk, "update Trino cluster", func() (*operation.Operation, error) {
		return sdk.Trino().Cluster().Update(ctx, req)
	})
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &trino.DeleteClusterRequest{
		ClusterId: cid,
	}

	return waitOperation(ctx, sdk, "delete Trino cluster", func() (*operation.Operation, error) {
		return sdk.Trino().Cluster().Delete(ctx, req)
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
