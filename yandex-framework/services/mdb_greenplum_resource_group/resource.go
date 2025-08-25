package mdb_greenplum_resource_group

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	yandexMDBGreenplumResourceGroupDefaultTimeout = 120 * time.Minute
	yandexMDBGreenplumResourceGroupUpdateTimeout  = 120 * time.Minute
)

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_greenplum_resource_group"
}

func (r *bindingResource) Configure(_ context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func getSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a Greenplum resource group within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-greenplum/).",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to which resource group belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the resource group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_user_defined": schema.BoolAttribute{
				MarkdownDescription: "If false, the resource group is immutable and controlled by yandex",
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"concurrency": schema.Int64Attribute{
				Description: "The maximum number of concurrent transactions, including active and idle transactions, that are permitted in the resource group.",
				Optional:    true,
			},
			"cpu_rate_limit": schema.Int64Attribute{
				Description: "The percentage of CPU resources available to this resource group.",
				Optional:    true,
			},
			"memory_limit": schema.Int64Attribute{
				Description: "The percentage of reserved memory resources available to this resource group.",
				Optional:    true,
			},
			"memory_shared_quota": schema.Int64Attribute{
				Description: "The percentage of reserved memory to share across transactions submitted in this resource group.",
				Optional:    true,
			},
			"memory_spill_ratio": schema.Int64Attribute{
				Description: "The memory usage threshold for memory-intensive transactions. When a transaction reaches this threshold, it spills to disk.",
				Optional:    true,
			},
		},
	}
}

func (r *bindingResource) Schema(ctx context.Context,
	_ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema(ctx)
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceGroup
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cid := state.ClusterID.ValueString()
	rgName := state.Name.ValueString()
	rg := readResourceGroup(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, rgName)
	if resp.Diagnostics.HasError() || rg == nil {
		return
	}

	resourceGroupToState(rg, &state)

	state.Id = types.StringValue(resourceid.Construct(cid, rgName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceGroup
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBGreenplumResourceGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	rgPlan := resourceGroupFromState(ctx, &plan)

	createResourceGroup(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, rgPlan)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, rgPlan.Name))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func isWrapperInt64NotEqual(state, plan *wrapperspb.Int64Value) bool {
	return ((state != nil) != (plan != nil)) || (state != nil && plan != nil && state.GetValue() != plan.GetValue())
}

func isWrapperBoolNotEqual(state, plan *wrapperspb.BoolValue) bool {
	return ((state != nil) != (plan != nil)) || (state != nil && plan != nil && state.GetValue() != plan.GetValue())
}

func getUpdatePaths(plan, state *greenplum.ResourceGroup) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updatePaths []string
	if isWrapperBoolNotEqual(state.IsUserDefined, plan.IsUserDefined) {
		diags.AddError(
			"is_user_defined is immutable and controlled by yandex",
			"is_user_defined is immutable and controlled by yandex",
		)
	}
	if isWrapperInt64NotEqual(state.Concurrency, plan.Concurrency) {
		updatePaths = append(updatePaths, "resource_group.concurrency")
	}
	if isWrapperInt64NotEqual(state.CpuRateLimit, plan.CpuRateLimit) {
		updatePaths = append(updatePaths, "resource_group.cpu_rate_limit")
	}
	if isWrapperInt64NotEqual(state.MemoryLimit, plan.MemoryLimit) {
		updatePaths = append(updatePaths, "resource_group.memory_limit")
	}
	if isWrapperInt64NotEqual(state.MemorySharedQuota, plan.MemorySharedQuota) {
		updatePaths = append(updatePaths, "resource_group.memory_shared_quota")
	}
	if isWrapperInt64NotEqual(state.MemorySpillRatio, plan.MemorySpillRatio) {
		updatePaths = append(updatePaths, "resource_group.memory_spill_ratio")
	}
	return updatePaths, diags
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceGroup
	var state ResourceGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBGreenplumResourceGroupUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	rgState := resourceGroupFromState(ctx, &state)
	rgPlan := resourceGroupFromState(ctx, &plan)
	updatePaths, diags := getUpdatePaths(rgPlan, rgState)

	if len(updatePaths) > 0 {
		updateResourceGroup(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, rgPlan, updatePaths)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, rgPlan.Name))
	diags.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceGroup
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBGreenplumResourceGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	deleteResourceGroup(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbName)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, rgName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	rg := readResourceGroup(ctx, r.providerConfig.SDK, &resp.Diagnostics, clusterId, rgName)
	if resp.Diagnostics.HasError() || rg == nil {
		return
	}
	var state ResourceGroup
	resourceGroupToState(rg, &state)
	state.Id = types.StringValue(req.ID)
	state.ClusterID = types.StringValue(clusterId)

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"delete": types.StringType,
			"update": types.StringType,
		}),
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
