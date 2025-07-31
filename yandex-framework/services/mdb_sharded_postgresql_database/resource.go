package mdb_sharded_postgresql_database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	yandexMDBShardedPostgreSQLDatabaseCreateTimeout = 15 * time.Minute
	yandexMDBShardedPostgreSQLDatabaseDeleteTimeout = 10 * time.Minute
	yandexMDBShardedPostgreSQLDatabaseUpdateTimeout = 30 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &bindingResource{}
var _ resource.ResourceWithImportState = &bindingResource{}

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewShardedPostgreSQLDatabaseResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_sharded_postgresql_database"
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
	resp.Schema = DatabaseSchema(ctx)
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Database
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cid := state.ClusterID.ValueString()
	dbname := state.Name.ValueString()

	db := shardedPostgreSQLAPI.ReadDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbname)
	if resp.Diagnostics.HasError() {
		return
	}

	// diagnostics don't have errors and db is nil => db not found
	if db == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(dbToState(db, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, dbname))
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

	ctx, cancel := context.WithTimeout(ctx, yandexMDBShardedPostgreSQLDatabaseCreateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	dbname := plan.Name.ValueString()
	log.Printf("[DEBUG] Database state: %v\n", plan)
	dbspec, diags := dbFromState(&plan)
	log.Printf("[DEBUG] Database spec from state: %v\n", dbspec)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	shardedPostgreSQLAPI.CreateDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbspec)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, dbname))
	r.refreshResourceState(ctx, &plan, &diags)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func getUpdatePaths(plan, state *Database) []string {
	log.Printf("[DEBUG] Calculate update paths plan: %v state: %v\n", plan, state)
	var updatePaths []string
	return updatePaths
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Database
	var state Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, yandexMDBShardedPostgreSQLDatabaseUpdateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	dbPlan, diags := dbFromState(&plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	updatePaths := getUpdatePaths(&plan, &state)

	if len(updatePaths) == 0 {
		return
	}

	dbname := plan.Name.ValueString()
	log.Printf("[DEBUG] Updating database %v with update_mask %v", dbname, updatePaths)

	shardedPostgreSQLAPI.UpdateDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbPlan, updatePaths)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	plan.Id = types.StringValue(resourceid.Construct(cid, dbname))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Database
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, yandexMDBShardedPostgreSQLDatabaseDeleteTimeout)
	defer cancel()

	cid := state.ClusterID.ValueString()
	dbname := state.Name.ValueString()
	shardedPostgreSQLAPI.DeleteDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbname)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, dbname, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	db := shardedPostgreSQLAPI.ReadDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, clusterId, dbname)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Database

	resp.Diagnostics.Append(dbToState(db, &state)...)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) refreshResourceState(ctx context.Context, state *Database, respDiagnostics *diag.Diagnostics) {
	cid := state.ClusterID.ValueString()
	dbname := state.Name.ValueString()
	db := shardedPostgreSQLAPI.ReadDatabase(ctx, r.providerConfig.SDK, respDiagnostics, cid, dbname)
	if respDiagnostics.HasError() {
		return
	}

	respDiagnostics.Append(dbToState(db, state)...)
	if respDiagnostics.HasError() {
		return
	}
}
