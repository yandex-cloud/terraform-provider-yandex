package metastore_cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *metastore.CreateClusterRequest) (string, diag.Diagnostic) {
	op, err := sdk.WrapOperation(sdk.Metastore().Cluster().Create(ctx, req))
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Metastore cluster",
			"Error while requesting API to create Metastore cluster: "+err.Error(),
		)
	}

	err = op.WaitInterval(ctx, 5*time.Second)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Metastore cluster",
			"Error while requesting API to create Metastore cluster. Failed to wait: "+err.Error(),
		)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Metastore cluster",
			"Failed to unmarshal metadata: "+err.Error(),
		)
	}

	md, ok := protoMetadata.(*metastore.CreateClusterMetadata)
	if !ok {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Metastore cluster",
			"Failed to convert response metadata to CreateClusterMetadata",
		)
	}

	return md.ClusterId, nil
}

func GetClusterByID(ctx context.Context, sdk *ycsdk.SDK, cid string) (*metastore.Cluster, diag.Diagnostic) {
	cluster, err := sdk.Metastore().Cluster().Get(ctx, &metastore.GetClusterRequest{
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

	return waitOperation(ctx, sdk, "update Metastore cluster", func() (*operation.Operation, error) {
		return sdk.Metastore().Cluster().Update(ctx, req)
	})
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, cid string) diag.Diagnostic {
	req := &metastore.DeleteClusterRequest{
		ClusterId: cid,
	}

	return waitOperation(ctx, sdk, "delete Metastore cluster", func() (*operation.Operation, error) {
		return sdk.Metastore().Cluster().Delete(ctx, req)
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
