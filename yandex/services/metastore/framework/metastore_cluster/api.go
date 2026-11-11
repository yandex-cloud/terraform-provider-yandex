package metastore_cluster

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
	metastoresdk "github.com/yandex-cloud/go-sdk/services/metastore/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *metastore.CreateClusterRequest) (string, diag.Diagnostic) {
	op, err := metastoresdk.NewClusterClient(sdk).Create(ctx, req)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Metastore cluster",
			"Error while requesting API to create Metastore cluster: "+err.Error(),
		)
	}

	_, err = op.WaitInterval(ctx, func(int) time.Duration { return 5 * time.Second })
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Metastore cluster",
			"Error while requesting API to create Metastore cluster. Failed to wait: "+err.Error(),
		)
	}

	return op.Metadata().ClusterId, nil
}

func GetClusterByID(ctx context.Context, sdk *ycsdk.SDK, cid string) (*metastore.Cluster, diag.Diagnostic) {
	cluster, err := metastoresdk.NewClusterClient(sdk).Get(ctx, &metastore.GetClusterRequest{
		ClusterId: cid,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}

		return nil, diag.NewErrorDiagnostic(
			"Failed to read Metastore cluster",
			"Error while requesting API to get Metastore cluster: "+err.Error(),
		)
	}
	return cluster, nil
}

func UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, req *metastore.UpdateClusterRequest) diag.Diagnostic {
	if req == nil || req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*metastoresdk.ClusterUpdateOperation, error) {
		return metastoresdk.NewClusterClient(sdk).Update(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to update Metastore cluster", err.Error())
	}
	return nil
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &metastore.DeleteClusterRequest{
		ClusterId: cid,
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*metastoresdk.ClusterDeleteOperation, error) {
		return metastoresdk.NewClusterClient(sdk).Delete(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to delete Metastore cluster", err.Error())
	}
	return nil
}
