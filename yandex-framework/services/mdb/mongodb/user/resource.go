package user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/resourceid"
)

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mongodb_user"
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
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"database_name": schema.StringAttribute{
							Required: true,
						},
						"roles": schema.SetAttribute{
							Optional:    true,
							ElementType: basetypes.StringType{},
						},
					},
				},
			},
		},
	}
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state User
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()
	user := readUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, userName)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(userToState(user, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, userName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan User
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := plan.ClusterID.ValueString()
	userPlan, diags := userFromState(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, userPlan)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, userPlan.Name))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func getUpdatePaths(plan, state *mongodb.UserSpec) []string {
	var updatePaths []string
	if state.Password != plan.Password {
		updatePaths = append(updatePaths, "password")
	}
	if fmt.Sprintf("%v", state.Permissions) != fmt.Sprintf("%v", plan.Permissions) {
		updatePaths = append(updatePaths, "permissions")
	}
	return updatePaths
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan User
	var state User
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := plan.ClusterID.ValueString()
	userState, diags := userFromState(ctx, &state)
	resp.Diagnostics.Append(diags...)
	userPlan, diags := userFromState(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updatePaths := getUpdatePaths(userPlan, userState)

	if len(updatePaths) > 0 {
		updateUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, userPlan, updatePaths)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, userPlan.Name))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state User
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	deleteUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbName)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, userName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	user := readUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, clusterId, userName)
	if resp.Diagnostics.HasError() {
		return
	}
	var state User
	resp.Diagnostics.Append(userToState(user, &state)...)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
