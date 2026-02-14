package mdb_clickhouse_cluster_v2

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
)

const (
	yandexMDBClickHouseClusterCreateTimeout = 60 * time.Minute
	yandexMDBClickHouseClusterDeleteTimeout = 30 * time.Minute
	yandexMDBClickHouseClusterUpdateTimeout = 90 * time.Minute
	yandexMDBClickHouseClusterPollInterval  = 10 * time.Second
)

var _ resource.ResourceWithModifyPlan = &clusterResource{}

type clusterResource struct {
	providerConfig *provider_config.Config
}

func NewClickHouseClusterResourceV2() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_clickhouse_cluster_v2"
}

func (r *clusterResource) Configure(_ context.Context,
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

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Load the current state of the resource
	var state models.Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Resource State
	prevState := state
	refreshState(ctx, &prevState, &state, r.providerConfig.SDK, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	d := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(d...)
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.Cluster
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBClickHouseClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	hostSpecsSlice, diags := mdbcommon.CreateClusterHosts(ctx, clickhouseHostService, plan.HostSpecs)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	// Create cluster
	r.createCluster(ctx, &plan, hostSpecsSlice, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create format schemas
	r.createFormatSchemas(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create ml models
	r.createMlModels(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create shard groups
	r.createShardGroups(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update state
	prevState := plan
	refreshState(ctx, &prevState, &plan, r.providerConfig.SDK, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.Cluster
	var state models.Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBClickHouseClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, "Updating ClickHouse Cluster", map[string]interface{}{"id": plan.Id.ValueString()})
	tflog.Debug(ctx, fmt.Sprintf("Update ClickHouse Cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update ClickHouse Cluster plan: %+v", plan))

	if !state.FolderId.Equal(plan.FolderId) {
		// Update folder id
		tflog.Debug(ctx, "Updating ClickHouse folder id")
		updateFolderIdRequest := prepareFolderIdUpdateRequest(&state, &plan)

		clickhouseApi.MoveCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateFolderIdRequest)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !state.Version.Equal(plan.Version) {
		// Update ClickHouse version
		tflog.Debug(ctx, "Updating ClickHouse version")
		updateVersionRequest := prepareVersionUpdateRequest(&state, &plan)

		clickhouseApi.UpdateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateVersionRequest)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update cluster settings
	tflog.Debug(ctx, "Updating ClickHouse cluster settings")
	updateRequest := prepareClusterUpdateRequest(ctx, &state, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	clickhouseApi.UpdateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateRequest)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update hosts
	planChHostSpecs, planKeeperHostSpecs := splitHostSpecsByType(ctx, plan.HostSpecs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	stateChHostSpecs, stateKeeperHostSpecs := splitHostSpecsByType(ctx, state.HostSpecs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Opts generate
	copySchema := true
	if !plan.CopySchemaOnNewHosts.IsNull() && !plan.CopySchemaOnNewHosts.IsUnknown() {
		copySchema = plan.CopySchemaOnNewHosts.ValueBool()
	}

	shardSpecs := models.ExpandListShard(ctx, plan.Shards, plan.Id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	mapShardNameShardSpec := map[string]*clickhouse.ShardConfigSpec{}
	for _, shardSpec := range shardSpecs {
		configSpec := shardSpec.ConfigSpec
		if configSpec != nil {
			mapShardNameShardSpec[shardSpec.Name] = configSpec
		}
	}

	planZooKeeper := models.ExpandZooKeeper(ctx, plan.ZooKeeper, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var planZooKeeperResources *clickhouse.Resources
	if planZooKeeper != nil {
		planZooKeeperResources = planZooKeeper.Resources
	}

	opts := ClickHouseOpts{
		HasCoordinator:           len(stateKeeperHostSpecs.Elements()) > 0,
		CoordinatorResources:     planZooKeeperResources,
		CopySchema:               copySchema,
		PlanShardSpecByShardName: mapShardNameShardSpec,
	}

	// Update ZooKeeper/Keeper hosts
	tflog.Debug(ctx, "Updating ZooKeeper/Keeper hosts")
	mdbcommon.UpdateClusterHosts(
		ctx,
		r.providerConfig.SDK,
		&resp.Diagnostics,
		clickhouseHostService,
		&clickhouseApi,
		plan.Id.ValueString(),
		opts,
		planKeeperHostSpecs,
		stateKeeperHostSpecs,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update ClickHouse hosts and shards
	tflog.Debug(ctx, "Updating ClickHouse hosts and shards")
	mdbcommon.UpdateClusterHostsWithShards(
		ctx,
		r.providerConfig.SDK,
		&resp.Diagnostics,
		clickhouseHostService,
		&clickhouseApi,
		plan.Id.ValueString(),
		opts,
		planChHostSpecs,
		stateChHostSpecs,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.Shards.Equal(plan.Shards) {
		// Update shards
		tflog.Debug(ctx, "Updating Clickhouse shards")
		updateShards(ctx, plan, r.providerConfig.SDK, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !state.FormatSchema.Equal(plan.FormatSchema) {
		// Update format schemas
		tflog.Debug(ctx, "Updating Clickhouse format schemas")
		updateFormatSchemas(ctx, plan, r.providerConfig.SDK, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !state.MLModel.Equal(plan.MLModel) {
		// Update ml models
		tflog.Debug(ctx, "Updating Clickhouse ml models")
		updateMlModels(ctx, plan, r.providerConfig.SDK, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !state.ShardGroup.Equal(plan.ShardGroup) {
		// Update shard groups
		tflog.Debug(ctx, "Updating Clickhouse shard groups")
		updateShardGroups(ctx, plan, r.providerConfig.SDK, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update state
	prevState := plan
	refreshState(ctx, &prevState, &plan, r.providerConfig.SDK, &resp.Diagnostics)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBClickHouseClusterDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid := state.Id.ValueString()
	clickhouseApi.DeleteCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid)
	if resp.Diagnostics.HasError() {
		return
	}
}

// YC API behavior: updating clickhouse.<resources|disk_size_autoscaling> rewrites ALL shard resources/disk_size_autoscaling, and updating
// shard resources/disk_size_autoscaling may change the observed clickhouse.<resources|disk_size_autoscaling>.
// With UseStateForUnknown this may cause "inconsistent result after apply".
// Here we "unpin" the opposite branch by setting it to Unknown in the plan:
// - clickhouse.<resources|disk_size_autoscaling> changed => shards[*].<resources|disk_size_autoscaling>  = Unknown
// - shards[*].<resources|disk_size_autoscaling> changed  => clickhouse.<resources|disk_size_autoscaling> = Unknown
func (r *clusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.Config.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var config, plan, state models.Cluster
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// resources
	resourcesAttrName := "resources"
	configCHRes, configCHResSet := models.GetClickHouseResources(ctx, config, &resp.Diagnostics)
	configShardsRes, configShardsResSet := getShardOverrides(ctx, config, &resp.Diagnostics, models.ShardResourcesGetter)

	// disk_size_autoscaling
	diskSizeAutoscalingAttrName := "disk_size_autoscaling"
	configCHDsa, configCHDsaSet := models.GetClickHouseDiskSizeAutoscaling(ctx, config, &resp.Diagnostics)
	configShardsDsa, configShardsDsaSet := getShardOverrides(ctx, config, &resp.Diagnostics, models.ShardDiskSizeAutoscalingGetter)

	if resp.Diagnostics.HasError() {
		return
	}

	// clickhouse subcluster change resources management mode
	if configCHResSet {
		stateCHRes, stateCHResSet := models.GetClickHouseResources(ctx, state, &resp.Diagnostics)
		if resp.Diagnostics.HasError() || !stateCHResSet {
			return
		}

		if !configCHRes.Equal(stateCHRes) {
			setShardsAttrUnknown(ctx, &plan, resourcesAttrName, models.ResourcesAttrTypes, resp)
		}
	}

	// clickhouse subcluster change disk size autoscaling settings management mode
	if configCHDsaSet {
		stateCHDsa, stateCHDsaSet := models.GetClickHouseDiskSizeAutoscaling(ctx, state, &resp.Diagnostics)
		if resp.Diagnostics.HasError() || !stateCHDsaSet {
			return
		}

		if !configCHDsa.Equal(stateCHDsa) {
			setShardsAttrUnknown(ctx, &plan, diskSizeAutoscalingAttrName, models.DiskSizeAutoscalingAttrTypes, resp)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// shards change resources management mode
	if configShardsResSet {
		if shardOverridesChanged(ctx, configShardsRes, state, &resp.Diagnostics, models.ShardResourcesGetter) {
			resp.Diagnostics.Append(
				resp.Plan.SetAttribute(
					ctx,
					path.Root("clickhouse").AtName(resourcesAttrName),
					types.ObjectUnknown(models.ResourcesAttrTypes),
				)...,
			)
		}
	}

	// shards change disk size autoscaling settings management mode
	if configShardsDsaSet {
		if shardOverridesChanged(ctx, configShardsDsa, state, &resp.Diagnostics, models.ShardDiskSizeAutoscalingGetter) {
			resp.Diagnostics.Append(
				resp.Plan.SetAttribute(
					ctx,
					path.Root("clickhouse").AtName(diskSizeAutoscalingAttrName),
					types.ObjectUnknown(models.DiskSizeAutoscalingAttrTypes),
				)...,
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("clickhouse").AtName("resources"),
			path.MatchRoot("shards").AtAnyMapKey().AtName("resources"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("clickhouse").AtName("disk_size_autoscaling"),
			path.MatchRoot("shards").AtAnyMapKey().AtName("disk_size_autoscaling"),
		),
	}
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func refreshState(ctx context.Context, prevState, state *models.Cluster, sdk *ycsdk.SDK, diags *diag.Diagnostics) {
	cid := state.Id.ValueString()
	cluster := clickhouseApi.GetCluster(ctx, sdk, diags, cid)
	if diags.HasError() {
		return
	}

	entityIdToApiHosts := mdbcommon.ReadHosts(ctx, sdk, diags, clickhouseHostService, &clickhouseApi, state.HostSpecs, cid)

	var d diag.Diagnostics
	state.HostSpecs, d = types.MapValueFrom(ctx, models.HostType, entityIdToApiHosts)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	state.Id = types.StringValue(cluster.Id)
	state.ClusterId = state.Id
	state.FolderId = types.StringValue(cluster.FolderId)
	state.CreatedAt = types.StringValue(timestamp.Get(cluster.GetCreatedAt()))
	state.Name = types.StringValue(cluster.Name)
	state.Description = types.StringValue(cluster.Description)
	state.Labels = mdbcommon.FlattenMapString(ctx, cluster.Labels, diags)
	state.Environment = types.StringValue(cluster.Environment.String())
	state.NetworkId = types.StringValue(cluster.NetworkId)

	saId := cluster.GetServiceAccountId()
	if saId == "" {
		state.ServiceAccountId = types.StringNull()
	} else {
		state.ServiceAccountId = types.StringValue(saId)
	}

	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())
	state.MaintenanceWindow = mdbcommon.FlattenMaintenanceWindow[
		clickhouse.MaintenanceWindow,
		clickhouse.WeeklyMaintenanceWindow,
		clickhouse.AnytimeMaintenanceWindow,
		clickhouse.WeeklyMaintenanceWindow_WeekDay,
	](ctx, cluster.MaintenanceWindow, diags)
	state.SecurityGroupIds = mdbcommon.FlattenSetString(ctx, cluster.SecurityGroupIds, diags)
	state.DiskEncryptionKeyId = mdbcommon.FlattenStringWrapper(ctx, cluster.DiskEncryptionKeyId, diags)

	state.Version = types.StringValue(cluster.Config.Version)
	state.ClickHouse = models.FlattenClickHouse(ctx, prevState, cluster.Config.Clickhouse, diags)
	state.ZooKeeper = models.FlattenZooKeeper(ctx, cluster.Config.Zookeeper, diags)
	state.BackupWindowStart = mdbcommon.FlattenBackupWindowStart(ctx, cluster.Config.BackupWindowStart, diags)
	state.Access = models.FlattenAccess(ctx, cluster.Config.Access, diags)
	state.CloudStorage = models.FlattenCloudStorage(ctx, cluster.Config.CloudStorage, diags)
	state.AdminPassword = prevState.AdminPassword
	state.SqlDatabaseManagement = mdbcommon.FlattenBoolWrapper(ctx, cluster.Config.SqlDatabaseManagement, diags)
	state.SqlUserManagement = mdbcommon.FlattenBoolWrapper(ctx, cluster.Config.SqlUserManagement, diags)
	state.EmbeddedKeeper = mdbcommon.FlattenBoolWrapper(ctx, cluster.Config.EmbeddedKeeper, diags)
	state.BackupRetainPeriodDays = mdbcommon.FlattenInt64Wrapper(ctx, cluster.Config.BackupRetainPeriodDays, diags)

	currentFormatSchemas := clickhouseApi.ListFormatSchemas(ctx, sdk, diags, cid)
	state.FormatSchema = models.FlattenListFormatSchema(ctx, currentFormatSchemas, diags)

	currentMlModels := clickhouseApi.ListMlModels(ctx, sdk, diags, cid)
	state.MLModel = models.FlattenListMLModel(ctx, currentMlModels, diags)

	currentShards := clickhouseApi.ListShards(ctx, sdk, diags, cid)
	state.Shards = models.FlattenListShard(ctx, currentShards, diags)

	currentShardGroups := clickhouseApi.ListShardGroups(ctx, sdk, diags, cid)
	state.ShardGroup = models.FlattenListShardGroup(ctx, currentShardGroups, diags)
}

func shardOverridesChanged(
	ctx context.Context,
	configOverrides map[string]types.Object,
	state models.Cluster,
	diags *diag.Diagnostics,
	getter func(models.Shard) types.Object,
) bool {
	if state.Shards.IsNull() || state.Shards.IsUnknown() {
		return true
	}

	stateShards := map[string]models.Shard{}
	diags.Append(state.Shards.ElementsAs(ctx, &stateShards, false)...)
	if diags.HasError() {
		return true
	}

	for shardName, configObject := range configOverrides {
		st, ok := stateShards[shardName]
		if !ok {
			return true
		}
		currentObject := getter(st)
		if currentObject.IsNull() || currentObject.IsUnknown() || !configObject.Equal(currentObject) {
			return true
		}
	}

	return false
}

func setShardsAttrUnknown(
	ctx context.Context,
	plan *models.Cluster,
	attrName string,
	attrTypes map[string]attr.Type,
	resp *resource.ModifyPlanResponse,
) {
	if plan.Shards.IsNull() || plan.Shards.IsUnknown() {
		return
	}

	planShards := map[string]types.Object{}
	resp.Diagnostics.Append(plan.Shards.ElementsAs(ctx, &planShards, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for shardName := range planShards {
		resp.Diagnostics.Append(
			resp.Plan.SetAttribute(
				ctx,
				path.Root("shards").AtMapKey(shardName).AtName(attrName),
				types.ObjectUnknown(attrTypes),
			)...,
		)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func getShardOverrides(
	ctx context.Context,
	cluster models.Cluster,
	diags *diag.Diagnostics,
	getter func(models.Shard) types.Object,
) (map[string]types.Object, bool) {
	if cluster.Shards.IsNull() || cluster.Shards.IsUnknown() {
		return nil, false
	}

	shards := map[string]models.Shard{}
	diags.Append(cluster.Shards.ElementsAs(ctx, &shards, false)...)
	if diags.HasError() {
		return nil, false
	}

	overrides := make(map[string]types.Object)
	for name, s := range shards {
		v := getter(s)
		if v.IsNull() || v.IsUnknown() {
			continue
		}
		overrides[name] = v
	}

	return overrides, len(overrides) > 0
}
