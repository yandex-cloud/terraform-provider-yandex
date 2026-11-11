package mdb_clickhouse_database

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	yandexMDBClickhouseDatabaseCreateTimeout = 15 * time.Minute
	yandexMDBClickhouseDatabaseUpdateTimeout = 10 * time.Minute
	yandexMDBClickhouseDatabaseDeleteTimeout = 10 * time.Minute
)

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_clickhouse_database"
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

func (r *bindingResource) Schema(ctx context.Context,
	_ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A database of the ClickHouse cluster.",
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
				MarkdownDescription: "ID of the ClickHouse cluster. Provided by the client when the database is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the database.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "Database engine, possible values are: atomic,replicated.",
				Optional:            true,
				Computed:            true,
				Validators:          Database_Engine_validator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deletion_protection_mode": schema.StringAttribute{
				MarkdownDescription: "Deletion protection mode of the database. Possible values: `DELETION_PROTECTION_MODE_INHERITED` (default) - the database is protected if `deletion_protection` is enabled for the cluster, `DELETION_PROTECTION_MODE_ENABLED` - the database is protected regardless of the cluster setting, `DELETION_PROTECTION_MODE_DISABLED` - the database can be deleted even if the cluster is protected. The mode does not prevent deletion of the cluster.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultDeletionProtectionMode.String()),
				Validators:          DeletionProtectionMode_validator,
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
	r.refreshState(ctx, &resp.Diagnostics, &state, cid, dbName)
	if resp.Diagnostics.HasError() {
		return
	}

	// database not found
	if state.Id.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

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

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBClickhouseDatabaseCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	var dbSpec clickhouse.DatabaseSpec
	stateToSpec(&plan, &dbSpec)
	createDatabase(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, cid, &dbSpec)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshState(ctx, &resp.Diagnostics, &plan, cid, dbSpec.GetName())
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBClickhouseDatabaseUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	updateReq := getUpdateRequest(&plan, &state)
	if len(updateReq.GetUpdateMask().GetPaths()) > 0 {
		updateDatabase(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, updateReq)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func getUpdateRequest(plan, state *Database) *clickhouse.UpdateDatabaseRequest {
	req := &clickhouse.UpdateDatabaseRequest{
		ClusterId:    state.ClusterID.ValueString(),
		DatabaseName: state.Name.ValueString(),
		UpdateMask:   &fieldmaskpb.FieldMask{},
	}
	if !plan.DeletionProtectionMode.Equal(state.DeletionProtectionMode) {
		req.DeletionProtectionMode = getDeletionProtectionModeValue(plan.DeletionProtectionMode)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection_mode")
	}
	return req
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Database
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBClickhouseDatabaseDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	deleteDatabase(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, cid, dbName)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, dbName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	var state Database
	r.refreshState(ctx, &resp.Diagnostics, &state, clusterId, dbName)

	if resp.Diagnostics.HasError() {
		return
	}

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) refreshState(ctx context.Context, diag *diag.Diagnostics, state *Database, clusterId, dbName string) {
	dbSpec := readDatabase(ctx, r.providerConfig.SDKv2, diag, clusterId, dbName)
	if diag.HasError() {
		return
	}
	if dbSpec == nil {
		state.Id = types.StringNull()
		return
	}
	specToState(dbSpec, state)
}
