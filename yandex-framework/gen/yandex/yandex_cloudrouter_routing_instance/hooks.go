package yandex_cloudrouter_routing_instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cloudrouter "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudrouter/v1"
	cloudrouterv1sdk "github.com/yandex-cloud/go-sdk/services/cloudrouter/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ resource.ResourceWithModifyPlan = (*yandexCloudrouterRoutingInstanceResource)(nil)

func (r *yandexCloudrouterRoutingInstanceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state yandexCloudrouterRoutingInstanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) ||
		!plan.FolderId.Equal(state.FolderId) || !plan.Labels.Equal(state.Labels) ||
		!plan.Timeouts.Equal(state.Timeouts) ||
		!plan.DeletionProtection.Equal(state.DeletionProtection) ||
		!plan.VpcInfo.Equal(state.VpcInfo) || !plan.CicPrivateConnectionInfo.Equal(state.CicPrivateConnectionInfo) {
		return
	}

	plan.CreatedAt = state.CreatedAt
	plan.RegionId = state.RegionId
	plan.Status = state.Status
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func beforeUpdateHook(ctx context.Context, config *config.Config, updateReq *cloudrouter.UpdateRoutingInstanceRequest, plan *yandexCloudrouterRoutingInstanceModel, state *yandexCloudrouterRoutingInstanceModel) diag.Diagnostics {
	dg := diag.Diagnostics{}
	if plan.FolderId.IsUnknown() || plan.FolderId.IsNull() || plan.FolderId.Equal(state.FolderId) {
		return dg
	}

	req := &cloudrouter.MoveRoutingInstanceRequest{
		RoutingInstanceId:   updateReq.GetRoutingInstanceId(),
		DestinationFolderId: plan.FolderId.ValueString(),
	}
	md := new(metadata.MD)
	op, err := cloudrouterv1sdk.NewRoutingInstanceClient(config.SDKv2).Move(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Move routing_instance x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Move routing_instance x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		dg.AddError("Failed to Move resource", "Error while requesting API to move routing_instance: "+err.Error())
		return dg
	}

	moveRes, err := op.Wait(ctx)
	if err != nil {
		dg.AddError("Unable to Move Resource", "Error while waiting for routing_instance move: "+err.Error())
		return dg
	}
	tflog.Debug(ctx, fmt.Sprintf("Move routing_instance response: %s", validate.ProtoDump(moveRes)))
	return dg
}
