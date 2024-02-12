package yandex_mdb_mongodb_database

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider-config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/utils"
)

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mongodb_database"
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

func (r *bindingResource) Schema(_ context.Context,
	_ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Database
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	db := readDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbName)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ClusterID = types.StringValue(db.ClusterId)
	state.Name = types.StringValue(db.Name)

	state.Id = types.StringValue(utils.ConstructResourceId(cid, dbName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Database
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := plan.ClusterID.ValueString()
	dbName := plan.Name.ValueString()
	createDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbName)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(utils.ConstructResourceId(cid, dbName))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Update when cluster_id changed
func (r *bindingResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	panic("method not implemented")
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Database
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	deleteDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbName)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, dbName, err := utils.DeconstructResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	db := readDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, clusterId, dbName)
	if resp.Diagnostics.HasError() {
		return
	}
	var state Database
	state.ClusterID = types.StringValue(db.ClusterId)
	state.Name = types.StringValue(db.Name)
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
