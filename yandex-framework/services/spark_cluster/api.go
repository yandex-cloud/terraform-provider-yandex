package spark_cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *spark.CreateClusterRequest) (string, diag.Diagnostic) {
	op, err := sdk.WrapOperation(sdk.Spark().Cluster().Create(ctx, req))
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Spark cluster",
			"Error while requesting API to create Spark cluster: "+err.Error(),
		)
	}

	err = op.WaitInterval(ctx, 5*time.Second)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Spark cluster",
			"Error while requesting API to create Spark cluster. Failed to wait: "+err.Error(),
		)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Spark cluster",
			"Failed to unmarshal metadata: "+err.Error(),
		)
	}

	md, ok := protoMetadata.(*spark.CreateClusterMetadata)
	if !ok {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Spark cluster",
			"Failed to convert response metadata to CreateClusterMetadata",
		)
	}

	return md.ClusterId, nil
}

func GetClusterByID(ctx context.Context, sdk *ycsdk.SDK, cid string) (*spark.Cluster, diag.Diagnostic) {
	cluster, err := sdk.Spark().Cluster().Get(ctx, &spark.GetClusterRequest{
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

	return waitOperation(ctx, sdk, "update Spark cluster", func() (*operation.Operation, error) {
		return sdk.Spark().Cluster().Update(ctx, req)
	})
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &spark.DeleteClusterRequest{
		ClusterId: cid,
	}

	return waitOperation(ctx, sdk, "delete Spark cluster", func() (*operation.Operation, error) {
		return sdk.Spark().Cluster().Delete(ctx, req)
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
