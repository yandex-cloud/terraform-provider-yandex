package airflow_cluster

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/airflow/v1"
	airflowsdk "github.com/yandex-cloud/go-sdk/services/airflow/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *airflow.CreateClusterRequest) (string, diag.Diagnostic) {
	op, err := airflowsdk.NewClusterClient(sdk).Create(ctx, req)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Airflow cluster",
			"Error while requesting API to create Airflow cluster: "+err.Error(),
		)
	}

	_, err = op.WaitInterval(ctx, func(int) time.Duration { return 5 * time.Second })
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Airflow cluster",
			"Error while requesting API to create Airflow cluster. Failed to wait: "+err.Error(),
		)
	}

	return op.Metadata().ClusterId, nil
}

func GetClusterByID(ctx context.Context, sdk *ycsdk.SDK, cid string) (*airflow.Cluster, diag.Diagnostic) {
	cluster, err := airflowsdk.NewClusterClient(sdk).Get(ctx, &airflow.GetClusterRequest{
		ClusterId: cid,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}

		return nil, diag.NewErrorDiagnostic(
			"Failed to read Airflow cluster",
			"Error while requesting API to get Airflow cluster: "+err.Error(),
		)
	}
	return cluster, nil
}

func UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, req *airflow.UpdateClusterRequest) diag.Diagnostic {
	if req == nil || req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*airflowsdk.ClusterUpdateOperation, error) {
		return airflowsdk.NewClusterClient(sdk).Update(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to update Airflow cluster", err.Error())
	}
	return nil
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &airflow.DeleteClusterRequest{
		ClusterId: cid,
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*airflowsdk.ClusterDeleteOperation, error) {
		return airflowsdk.NewClusterClient(sdk).Delete(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to delete Airflow cluster", err.Error())
	}
	return nil
}
