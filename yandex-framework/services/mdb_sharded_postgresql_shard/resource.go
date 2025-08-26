package mdb_sharded_postgresql_shard

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
	yandexMDBShardedPostgreSQLShardCreateTimeout = 15 * time.Minute
	yandexMDBShardedPostgreSQLShardDeleteTimeout = 10 * time.Minute
	yandexMDBShardedPostgreSQLShardUpdateTimeout = 30 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &bindingResource{}
var _ resource.ResourceWithImportState = &bindingResource{}

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewShardedPostgreSQLShardResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_sharded_postgresql_shard"
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
	resp.Schema = ShardSchema(ctx)
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Shard
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cid := state.ClusterID.ValueString()
	shardName := state.Name.ValueString()

	shard := shardedPostgreSQLAPI.ReadShard(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, shardName)
	if resp.Diagnostics.HasError() {
		return
	}

	// diagnostics don't have errors and shard is nil => shard not found
	if shard == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(shardToState(shard, &state, cid)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, shardName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Shard
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, yandexMDBShardedPostgreSQLShardCreateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	shardName := plan.Name.ValueString()
	shardspec, diags := shardFromState(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	shardedPostgreSQLAPI.CreateShard(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, shardspec)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, shardName))
	r.refreshResourceState(ctx, &plan, &diags)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func getUpdatePaths(plan, state *Shard) []string {
	log.Printf("[DEBUG] Calculate update paths plan: %v state: %v\n", plan, state)
	var updatePaths []string
	return updatePaths
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Shard
	var state Shard
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, yandexMDBShardedPostgreSQLShardUpdateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	shardPlan, diags := shardFromState(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	updatePaths := getUpdatePaths(&plan, &state)

	if len(updatePaths) == 0 {
		return
	}

	shardName := plan.Name.ValueString()
	log.Printf("[DEBUG] Updating shard %v with update_mask %v", shardName, updatePaths)

	shardedPostgreSQLAPI.UpdateShard(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, shardPlan, updatePaths)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	plan.Id = types.StringValue(resourceid.Construct(cid, shardName))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Shard
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, yandexMDBShardedPostgreSQLShardDeleteTimeout)
	defer cancel()

	cid := state.ClusterID.ValueString()
	shardname := state.Name.ValueString()
	shardedPostgreSQLAPI.DeleteShard(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, shardname)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, shardName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	shard := shardedPostgreSQLAPI.ReadShard(ctx, r.providerConfig.SDK, &resp.Diagnostics, clusterId, shardName)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Shard

	resp.Diagnostics.Append(shardToState(shard, &state, clusterId)...)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) refreshResourceState(ctx context.Context, state *Shard, respDiagnostics *diag.Diagnostics) {
	cid := state.ClusterID.ValueString()
	shardName := state.Name.ValueString()
	shard := shardedPostgreSQLAPI.ReadShard(ctx, r.providerConfig.SDK, respDiagnostics, cid, shardName)
	if respDiagnostics.HasError() {
		return
	}

	respDiagnostics.Append(shardToState(shard, state, cid)...)
	if respDiagnostics.HasError() {
		return
	}
}
