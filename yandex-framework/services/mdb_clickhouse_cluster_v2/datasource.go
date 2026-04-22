package mdb_clickhouse_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
)

type bindingDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &bindingDataSource{}
}

func (d *bindingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_clickhouse_cluster_v2"
}

func (d *bindingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}

func (d *bindingDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataSourceClusterSchema(ctx)
}

func (d *bindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read config into state. Important for default values ​​(e.g. "timeouts").
	var state models.ClusterDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := ""
	if !state.ClusterId.IsNull() && !state.ClusterId.IsUnknown() {
		clusterID = state.ClusterId.ValueString()
	}
	name := ""
	if !state.Name.IsNull() && !state.Name.IsUnknown() {
		name = state.Name.ValueString()
	}

	if clusterID == "" && name == "" {
		resp.Diagnostics.AddError(
			"At least one of cluster_id or name is required",
			"The cluster ID or Name must be specified in the configuration",
		)
		return
	}

	// Get cluster id by name and folder.
	if clusterID == "" {
		folderID, diags := validate.FolderID(state.FolderId, &d.providerConfig.ProviderState)
		resp.Diagnostics.Append(diags)
		if resp.Diagnostics.HasError() {
			return
		}

		resolvedID, diags := objectid.ResolveByNameAndFolderID(
			ctx,
			d.providerConfig.SDK,
			folderID,
			name,
			sdkresolvers.ClickhouseClusterResolver,
		)
		resp.Diagnostics.Append(diags)
		if resp.Diagnostics.HasError() {
			return
		}

		clusterID = resolvedID
		state.ClusterId = types.StringValue(clusterID)
	}

	state.Id = types.StringValue(clusterID)
	prevState := state

	refreshDataSourceState(ctx, &prevState, &state, d.providerConfig.SDK, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func refreshDataSourceState(ctx context.Context, prevState, state *models.ClusterDataSource, sdk *ycsdk.SDK, diags *diag.Diagnostics) {
	cid := state.Id.ValueString()
	cluster := clickhouseApi.GetCluster(ctx, sdk, diags, cid)
	if diags.HasError() {
		return
	}

	entityIdToApiHosts := mdbcommon.ReadHosts(ctx, sdk, diags, clickhouseHostService, &clickhouseApi, state.HostSpecs, cid)
	if diags.HasError() {
		return
	}

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
	newSecurityGroupIds := mdbcommon.FlattenSetString(ctx, cluster.SecurityGroupIds, diags)
	if !utils.SetsAreEqual(state.SecurityGroupIds, newSecurityGroupIds) {
		state.SecurityGroupIds = newSecurityGroupIds
	}
	state.DiskEncryptionKeyId = mdbcommon.FlattenStringWrapper(ctx, cluster.DiskEncryptionKeyId, diags)

	state.Version = types.StringValue(cluster.Config.Version)
	state.ClickHouse = models.FlattenClickHouse(ctx, prevState.ClickHouse, cluster.Config.Clickhouse, diags)
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

	currentExtensions := clickhouseApi.ListExtensions(ctx, sdk, diags, cid)
	state.Extension = models.FlattenListExtensions(ctx, currentExtensions, diags)
}
