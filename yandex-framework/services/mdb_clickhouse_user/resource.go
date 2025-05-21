package mdb_clickhouse_user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	yandexMDBClickhouseUserCreateTimeout = 15 * time.Minute
	yandexMDBClickhouseUserDeleteTimeout = 10 * time.Minute
	yandexMDBClickhouseUserUpdateTimeout = 30 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &bindingResource{}
var _ resource.ResourceWithImportState = &bindingResource{}

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_clickhouse_user"
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

func (r *bindingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = UserSchema(ctx)
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceUser
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

	// diagnostics don't have errors and user is nil => user not found
	if user == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(userToState(ctx, user, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, userName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceUser
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBClickhouseUserCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	userName := plan.Name.ValueString()
	log.Printf("[DEBUG] User state: %v\n", plan)
	userSpec, diags := userFromState(ctx, &plan)
	log.Printf("[DEBUG] User spec from state: %v\n", userSpec)

	if !isValidPasswordConfiguration(userSpec) {
		resp.Diagnostics.AddError(
			"Invalid user configuration",
			"must specify either password or generate_password",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, userSpec)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, userName))
	r.refreshResourceState(ctx, &plan, &diags)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func getUpdatePaths(plan, state *ResourceUser) []string {
	log.Printf("[DEBUG] Calculate update paths plan: %v state: %v\n", plan, state)
	var updatePaths []string
	if state.Password != plan.Password {
		updatePaths = append(updatePaths, "password")
	}
	if !plan.Permissions.Equal(state.Permissions) {
		updatePaths = append(updatePaths, "permissions")
	}
	if !plan.Quotas.Equal(state.Quotas) {
		updatePaths = append(updatePaths, "quotas")
	}
	if !plan.Settings.Equal(state.Settings) {
		updatePaths = append(updatePaths, "settings")
	}
	return updatePaths
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceUser
	var state ResourceUser
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := state.Timeouts.Update(ctx, yandexMDBClickhouseUserUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	userPlan, diags := userFromState(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if !isValidPasswordConfiguration(userPlan) {
		resp.Diagnostics.AddError(
			"Invalid user configuration",
			"must specify either password or generate_password",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	updatePaths := getUpdatePaths(&plan, &state)

	if len(updatePaths) == 0 {
		return
	}

	userName := plan.Name.ValueString()
	log.Printf("[DEBUG] Updating user %v with update_mask %v", userName, updatePaths)

	updateUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, userPlan, updatePaths)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[TRACE] mdb_clickhouse_user: refresh state settings: %+v\n", state.GetSettings())
	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	plan.Id = types.StringValue(resourceid.Construct(cid, userName))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceUser
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBClickhouseUserDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()
	deleteUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, userName)
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

	var state ResourceUser
	// default settings object for correct import unchanged settings
	state.SetSettings(types.ObjectNull(settingsType))

	resp.Diagnostics.Append(userToState(ctx, user, &state)...)
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

func (r *bindingResource) refreshResourceState(ctx context.Context, state *ResourceUser, respDiagnostics *diag.Diagnostics) {
	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()
	user := readUser(ctx, r.providerConfig.SDK, respDiagnostics, cid, userName)
	if respDiagnostics.HasError() {
		return
	}

	respDiagnostics.Append(userToState(ctx, user, state)...)
	if respDiagnostics.HasError() {
		return
	}
}

func isValidPasswordConfiguration(userSpec *clickhouse.UserSpec) bool {
	passwordSpecified := len(userSpec.Password) > 0

	isBothFieldNotSpecified := !passwordSpecified && !userSpec.GeneratePassword.GetValue()
	isBothFieldSpecified := passwordSpecified && userSpec.GeneratePassword.GetValue()
	return !isBothFieldNotSpecified && !isBothFieldSpecified
}
