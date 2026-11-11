package request

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	opensearchsdk "github.com/yandex-cloud/go-sdk/services/mdb/opensearch/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
)

const (
	defaultMDBPageSize = 1000
)

func GetCusterByID(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) *opensearch.Cluster {
	cluster, err := opensearchsdk.NewClusterClient(sdk).Get(ctx, &opensearch.GetClusterRequest{
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
		resp, err := opensearchsdk.NewClusterClient(sdk).ListHosts(ctx, &opensearch.ListClusterHostsRequest{
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
	op, err := opensearchsdk.NewClusterClient(sdk).Create(ctx, req)
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
	_, err = op.WaitInterval(ctx, func(int) time.Duration { return 5 * time.Second })
	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create OpenSearch cluster. Failed to wait: "+err.Error(),
		)
		return ""
	}

	md := op.Metadata()

	return md.ClusterId
}

func DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterDeleteOperation, error) {
		op, err := opensearchsdk.NewClusterClient(sdk).Delete(ctx, &opensearch.DeleteClusterRequest{
			ClusterId: cid,
		})
		if err != nil {
			if validate.IsStatusWithCode(err, codes.NotFound) {
				return nil, nil
			}
			return nil, err
		}
		return op, nil
	})
	if err == nil && op != nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Cluster Delete", err.Error())
	}
}

func UpdateClusterSpec(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.UpdateClusterRequest) {
	if req == nil || req.UpdateMask == nil {
		return
	}

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterUpdateOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).Update(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Cluster Update", err.Error())
	}
}

func AddOpenSearchNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.AddOpenSearchNodeGroupRequest) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterAddOpenSearchNodeGroupOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).AddOpenSearchNodeGroup(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Add OpenSearch nodegroup", err.Error())
	}
}

func UpdateOpenSearchNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.UpdateOpenSearchNodeGroupRequest) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterUpdateOpenSearchNodeGroupOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).UpdateOpenSearchNodeGroup(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Update OpenSearch nodegroup", err.Error())
	}
}

func DeleteOpenSearchNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.DeleteOpenSearchNodeGroupRequest) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterDeleteOpenSearchNodeGroupOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).DeleteOpenSearchNodeGroup(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Delete OpenSearch nodegroup", err.Error())
	}
}

func AddDashboardsNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.AddDashboardsNodeGroupRequest) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterAddDashboardsNodeGroupOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).AddDashboardsNodeGroup(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Add Dashboards nodegroup", err.Error())
	}
}

func UpdateDashboardsNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.UpdateDashboardsNodeGroupRequest) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterUpdateDashboardsNodeGroupOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).UpdateDashboardsNodeGroup(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Update Dashboards nodegroup", err.Error())
	}
}

func DeleteDashboardsNodeGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *opensearch.DeleteDashboardsNodeGroupRequest) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterDeleteDashboardsNodeGroupOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).DeleteDashboardsNodeGroup(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Delete Dashboards nodegroup", err.Error())
	}
}

func GetAuthSettings(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) *opensearch.AuthSettings {
	resp, err := opensearchsdk.NewClusterClient(sdk).GetAuthSettings(ctx, &opensearch.GetAuthSettingsRequest{
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
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*opensearchsdk.ClusterUpdateAuthSettingsOperation, error) {
		return opensearchsdk.NewClusterClient(sdk).UpdateAuthSettings(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		diags.AddError("Failed to Update Auth Settings", err.Error())
	}
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
