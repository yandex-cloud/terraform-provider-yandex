package trino_cluster

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	trinosdk "github.com/yandex-cloud/go-sdk/services/trino/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *trino.CreateClusterRequest) (string, diag.Diagnostic) {
	op, err := trinosdk.NewClusterClient(sdk).Create(ctx, req)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino cluster",
			"Error while requesting API to create Trino cluster: "+err.Error(),
		)
	}

	_, err = op.WaitInterval(ctx, func(int) time.Duration { return 5 * time.Second })
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino cluster",
			"Error while requesting API to create Trino cluster. Failed to wait: "+err.Error(),
		)
	}

	return op.Metadata().ClusterId, nil
}

func GetClusterByID(ctx context.Context, sdk *ycsdk.SDK, cid string) (*trino.Cluster, diag.Diagnostic) {
	cluster, err := trinosdk.NewClusterClient(sdk).Get(ctx, &trino.GetClusterRequest{
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

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*trinosdk.ClusterUpdateOperation, error) {
		return trinosdk.NewClusterClient(sdk).Update(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to update Trino cluster", err.Error())
	}
	return nil
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &trino.DeleteClusterRequest{
		ClusterId: cid,
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*trinosdk.ClusterDeleteOperation, error) {
		return trinosdk.NewClusterClient(sdk).Delete(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to delete Trino cluster", err.Error())
	}
	return nil
}
