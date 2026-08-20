package spark_cluster

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	sparksdk "github.com/yandex-cloud/go-sdk/services/spark/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/operationcompat"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func requestUpdateCluster(ctx context.Context, sdk *ycsdk.SDK, req *spark.UpdateClusterRequest) (*operationpb.Operation, error) {
	conn, err := sdk.GetConnection(ctx, sparksdk.ClusterUpdate)
	if err != nil {
		return nil, err
	}
	return spark.NewClusterServiceClient(conn).Update(ctx, req)
}

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *spark.CreateClusterRequest) (string, diag.Diagnostic) {
	op, err := sparksdk.NewClusterClient(sdk).Create(ctx, req)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Spark cluster",
			"Error while requesting API to create Spark cluster: "+err.Error(),
		)
	}

	_, err = op.WaitInterval(ctx, func(int) time.Duration { return 5 * time.Second })
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Spark cluster",
			"Error while requesting API to create Spark cluster. Failed to wait: "+err.Error(),
		)
	}

	return op.Metadata().ClusterId, nil
}

func GetClusterByID(ctx context.Context, sdk *ycsdk.SDK, cid string) (*spark.Cluster, diag.Diagnostic) {
	cluster, err := sparksdk.NewClusterClient(sdk).Get(ctx, &spark.GetClusterRequest{
		ClusterId: cid,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}

		return nil, diag.NewErrorDiagnostic(
			"Failed to read Spark cluster",
			"Error while requesting API to get Spark cluster: "+err.Error(),
		)
	}
	return cluster, nil
}

func UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, req *spark.UpdateClusterRequest) diag.Diagnostic {
	if req == nil || req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*operationpb.Operation, error) {
		return requestUpdateCluster(ctx, sdk, req)
	})
	if err == nil {
		err = operationcompat.Wait(ctx, sdk, op.GetId())
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to update Spark cluster", err.Error())
	}
	return nil
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &spark.DeleteClusterRequest{
		ClusterId: cid,
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*sparksdk.ClusterDeleteOperation, error) {
		return sparksdk.NewClusterClient(sdk).Delete(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to delete Spark cluster", err.Error())
	}
	return nil
}
