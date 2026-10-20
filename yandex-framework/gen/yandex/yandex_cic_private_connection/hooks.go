package yandex_cic_private_connection

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cic "github.com/yandex-cloud/go-genproto/yandex/cloud/cic/v1"
	cicv1sdk "github.com/yandex-cloud/go-sdk/services/cic/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ resource.ResourceWithModifyPlan = (*yandexCicPrivateConnectionResource)(nil)

func (r *yandexCicPrivateConnectionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state yandexCicPrivateConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Ipv4Peering.IsNull() && !plan.Ipv4Peering.IsUnknown() &&
		!state.Ipv4Peering.IsNull() && !state.Ipv4Peering.IsUnknown() {
		attributes := plan.Ipv4Peering.Attributes()
		attributes["cloud_bgp_asn"] = state.Ipv4Peering.Attributes()["cloud_bgp_asn"]
		peering, diags := types.ObjectValue(plan.Ipv4Peering.AttributeTypes(ctx), attributes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Ipv4Peering = peering
	}
	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) ||
		!plan.FolderId.Equal(state.FolderId) || !plan.Labels.Equal(state.Labels) ||
		!plan.Timeouts.Equal(state.Timeouts) ||
		!plan.TrunkConnectionId.Equal(state.TrunkConnectionId) || !plan.VlanId.Equal(state.VlanId) ||
		!plan.Ipv4Peering.Equal(state.Ipv4Peering) || !plan.Ipv4StaticRoutes.Equal(state.Ipv4StaticRoutes) {
		return
	}

	plan.CreatedAt = state.CreatedAt
	plan.Status = state.Status
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func beforeUpdateHook(ctx context.Context, config *config.Config, updateReq *cic.UpdatePrivateConnectionRequest, plan *yandexCicPrivateConnectionModel, state *yandexCicPrivateConnectionModel) diag.Diagnostics {
	dg := diag.Diagnostics{}
	if plan.FolderId.IsUnknown() || plan.FolderId.IsNull() || plan.FolderId.Equal(state.FolderId) {
		return dg
	}

	req := &cic.MovePrivateConnectionRequest{
		PrivateConnectionId: updateReq.GetPrivateConnectionId(),
		DestinationFolderId: plan.FolderId.ValueString(),
	}
	md := new(metadata.MD)
	op, err := cicv1sdk.NewPrivateConnectionClient(config.SDKv2).Move(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Move private_connection x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Move private_connection x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		dg.AddError("Failed to Move resource", "Error while requesting API to move private_connection: "+err.Error())
		return dg
	}

	_, err = op.Wait(ctx)
	if err != nil {
		dg.AddError("Unable to Move Resource", "Error while waiting for private_connection move: "+err.Error())
		return dg
	}
	return dg
}
