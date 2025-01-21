package request

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	sdkoperation "github.com/yandex-cloud/go-sdk/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultMDBPageSize      = 1000
	operationsRetryCount    = 5
	operationsRetryInterval = 2 * time.Minute
)

func GetCusterByID(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) *opensearch.Cluster {
	cluster, err := sdk.MDB().OpenSearch().Cluster().Get(ctx, &opensearch.GetClusterRequest{
		ClusterId: cid,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}

		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get OpenSearch cluster: "+err.Error(),
		)
		return nil
	}
	return cluster
}

func GetHostsList(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) []*opensearch.Host {
	hosts := []*opensearch.Host{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().OpenSearch().Cluster().ListHosts(ctx, &opensearch.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to get OpenSearch hosts: "+err.Error(),
			)
			return nil
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts
}

func CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *opensearch.CreateClusterRequest) string {
	op, err := sdk.WrapOperation(sdk.MDB().OpenSearch().Cluster().Create(ctx, req))
	if err != nil {
		// if validate.IsStatusWithCode(err, codes.AlreadyExists) {
		// 	TODO: maybe get list clusters, and find cid by name
		// }

		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create OpenSearch cluster: "+err.Error(),
		)
		return ""
	}

	//Notice: in old version we didn't wait for result, but in new one we have to wait for result. Otherwise we will miss some data in Get request
	err = op.WaitInterval(ctx, 5*time.Second)
	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create OpenSearch cluster. Failed to wait: "+err.Error(),
		)
		return ""
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create OpenSearch cluster. Failed to unmarshal metadata: "+err.Error(),
		)
		return ""
	}

	md, ok := protoMetadata.(*opensearch.CreateClusterMetadata)
	if !ok {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create OpenSearch cluster. Failed to cast proto metadata",
		)
		return ""
	}

	return md.ClusterId
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Cluster Delete", func() (*operation.Operation, error) {
		op, err := sdk.MDB().OpenSearch().Cluster().Delete(ctx, &opensearch.DeleteClusterRequest{
			ClusterId: cid,
		})
		if err != nil {
			if validate.IsStatusWithCode(err, codes.NotFound) {
				// to prevent panic on underlying levels
				return &operation.Operation{Done: true}, nil
			}

			return nil, err
		}

		return op, nil
	}))
}

func UpdateClusterSpec(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.UpdateClusterRequest) {
	if req == nil || req.UpdateMask == nil {
		return
	}

	diags.Append(waitOperationWithRetry(ctx, sdk, "Cluster Update", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().Update(ctx, req)
	}))
}

func AddOpenSearchNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.AddOpenSearchNodeGroupRequest) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Add OpenSearch nodegroup", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().AddOpenSearchNodeGroup(ctx, req)
	}))
}

func UpdateOpenSearchNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.UpdateOpenSearchNodeGroupRequest) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Update OpenSearch nodegroup", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().UpdateOpenSearchNodeGroup(ctx, req)
	}))
}

func DeleteOpenSearchNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.DeleteOpenSearchNodeGroupRequest) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Delete OpenSearch nodegroup", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().DeleteOpenSearchNodeGroup(ctx, req)
	}))
}

func AddDashboardsNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.AddDashboardsNodeGroupRequest) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Add Dashboards nodegroup", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().AddDashboardsNodeGroup(ctx, req)
	}))
}

func UpdateDashboardsNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.UpdateDashboardsNodeGroupRequest) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Update Dashboards nodegroup", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().UpdateDashboardsNodeGroup(ctx, req)
	}))
}

func DeleteDashboardsNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.DeleteDashboardsNodeGroupRequest) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Delete Dashboards nodegroup", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().DeleteDashboardsNodeGroup(ctx, req)
	}))
}

func GetAuthSettings(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) *opensearch.AuthSettings {
	resp, err := sdk.MDB().OpenSearch().Cluster().GetAuthSettings(ctx, &opensearch.GetAuthSettingsRequest{
		ClusterId: cid,
	})
	if err != nil {
		diags.AddError(
			"Failed to Read resource",
			"Error while requesting API to get OpenSearch Auth Settings: "+err.Error(),
		)
		return nil
	}

	return resp
}

func UpdateAuthSettings(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.UpdateAuthSettingsRequest) {
	diags.Append(waitOperationWithRetry(ctx, sdk, "Update Auth Settings", func() (*operation.Operation, error) {
		return sdk.MDB().OpenSearch().Cluster().UpdateAuthSettings(ctx, req)
	}))
}

func PrepareAndExecute[T any, V any](
	ctx context.Context,
	sdk *ycsdk.SDK,
	clusterID string,
	plan, state []T,
	generator func(string, []T, []T) ([]V, diag.Diagnostics),
	executor func(context.Context, *ycsdk.SDK, *diag.Diagnostics, V)) diag.Diagnostics {
	requests, diags := generator(clusterID, plan, state)
	if diags.HasError() {
		return diags
	}

	for _, req := range requests {
		executor(ctx, sdk, &diags, req)
		if diags.HasError() {
			return diags
		}
	}

	return diag.Diagnostics{}
}

func waitOperationWithRetry(ctx context.Context, sdk *ycsdk.SDK, caller string, action func() (*operation.Operation, error)) diag.Diagnostic {
	var err error
	for retryCount := 0; retryCount < operationsRetryCount; retryCount++ {
		var op *sdkoperation.Operation
		op, err = retry.ConflictingOperation(ctx, sdk, action)
		if err != nil {
			return diag.NewErrorDiagnostic(
				"Failed to Wait for operation",
				fmt.Sprintf("Error while requesting API for %s: %s", caller, err.Error()),
			)
		}

		err = op.Wait(ctx)
		if shouldRetry(op, err) {
			time.Sleep(operationsRetryInterval)
			continue
		}
		_, err = op.Response()
		if shouldRetry(op, err) {
			time.Sleep(operationsRetryInterval)
			continue
		}
		break
	}

	if err != nil {
		return diag.NewErrorDiagnostic(
			"Failed to Wait for operation",
			fmt.Sprintf("Error while waiting for %s: %s", caller, err.Error()),
		)
	}

	return nil
}

func shouldRetry(op *sdkoperation.Operation, err error) bool {
	if err != nil {
		status, ok := status.FromError(err)
		if ok && status.Code() == codes.Internal {
			return true
		}
		return false
	}

	return op.Failed()
}
