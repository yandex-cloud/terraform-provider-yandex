package mdb_greenplum_cluster_v2

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"
)

type yandexMdbGreenplumClusterV2DatasourceModel struct {
	CloudStorage        types.Object   `tfsdk:"cloud_storage"`
	ClusterConfig       types.Object   `tfsdk:"cluster_config"`
	ID                  types.String   `tfsdk:"id"`
	Config              types.Object   `tfsdk:"config"`
	CreatedAt           types.String   `tfsdk:"created_at"`
	DeletionProtection  types.Bool     `tfsdk:"deletion_protection"`
	Description         types.String   `tfsdk:"description"`
	Environment         types.String   `tfsdk:"environment"`
	FolderId            types.String   `tfsdk:"folder_id"`
	HostGroupIds        types.Set      `tfsdk:"host_group_ids"`
	Labels              types.Map      `tfsdk:"labels"`
	Logging             types.Object   `tfsdk:"logging"`
	MaintenanceWindow   types.Object   `tfsdk:"maintenance_window"`
	MasterConfig        types.Object   `tfsdk:"master_config"`
	MasterHostCount     types.Int64    `tfsdk:"master_host_count"`
	MasterHostGroupIds  types.Set      `tfsdk:"master_host_group_ids"`
	Monitoring          types.Set      `tfsdk:"monitoring"`
	Name                types.String   `tfsdk:"name"`
	NetworkId           types.String   `tfsdk:"network_id"`
	PlannedOperation    types.Object   `tfsdk:"planned_operation"`
	SecurityGroupIds    types.Set      `tfsdk:"security_group_ids"`
	SegmentConfig       types.Object   `tfsdk:"segment_config"`
	SegmentHostCount    types.Int64    `tfsdk:"segment_host_count"`
	SegmentHostGroupIds types.Set      `tfsdk:"segment_host_group_ids"`
	SegmentInHost       types.Int64    `tfsdk:"segment_in_host"`
	ServiceAccountId    types.String   `tfsdk:"service_account_id"`
	UserName            types.String   `tfsdk:"user_name"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetCloudStorage() types.Object {
	return m.CloudStorage
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetClusterConfig() types.Object {
	return m.ClusterConfig
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetID() types.String {
	return m.ID
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetConfig() types.Object {
	return m.Config
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetCreatedAt() types.String {
	return m.CreatedAt
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetDeletionProtection() types.Bool {
	return m.DeletionProtection
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetDescription() types.String {
	return m.Description
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetEnvironment() types.String {
	return m.Environment
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetFolderId() types.String {
	return m.FolderId
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetHostGroupIds() types.Set {
	return m.HostGroupIds
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetLabels() types.Map {
	return m.Labels
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetLogging() types.Object {
	return m.Logging
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetMaintenanceWindow() types.Object {
	return m.MaintenanceWindow
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetMasterConfig() types.Object {
	return m.MasterConfig
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetMasterHostCount() types.Int64 {
	return m.MasterHostCount
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetMasterHostGroupIds() types.Set {
	return m.MasterHostGroupIds
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetMonitoring() types.Set {
	return m.Monitoring
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetName() types.String {
	return m.Name
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetNetworkId() types.String {
	return m.NetworkId
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetPlannedOperation() types.Object {
	return m.PlannedOperation
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetSecurityGroupIds() types.Set {
	return m.SecurityGroupIds
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetSegmentConfig() types.Object {
	return m.SegmentConfig
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetSegmentHostCount() types.Int64 {
	return m.SegmentHostCount
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetSegmentHostGroupIds() types.Set {
	return m.SegmentHostGroupIds
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetSegmentInHost() types.Int64 {
	return m.SegmentInHost
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetServiceAccountId() types.String {
	return m.ServiceAccountId
}
func (m *yandexMdbGreenplumClusterV2DatasourceModel) GetUserName() types.String {
	return m.UserName
}

func NewYandexMdbGreenplumClusterV2DatasourceModel() yandexMdbGreenplumClusterV2DatasourceModel {
	return yandexMdbGreenplumClusterV2DatasourceModel{
		CloudStorage:        types.ObjectNull(yandexMdbGreenplumClusterV2CloudStorageModelType.AttrTypes),
		ClusterConfig:       types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigModelType.AttrTypes),
		ID:                  types.StringNull(),
		Config:              types.ObjectNull(yandexMdbGreenplumClusterV2ConfigModelType.AttrTypes),
		CreatedAt:           types.StringNull(),
		DeletionProtection:  types.BoolNull(),
		Description:         types.StringNull(),
		Environment:         types.StringNull(),
		FolderId:            types.StringNull(),
		HostGroupIds:        types.SetNull(types.StringType),
		Labels:              types.MapNull(types.StringType),
		Logging:             types.ObjectNull(yandexMdbGreenplumClusterV2LoggingModelType.AttrTypes),
		MaintenanceWindow:   types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowModelType.AttrTypes),
		MasterConfig:        types.ObjectNull(yandexMdbGreenplumClusterV2MasterConfigModelType.AttrTypes),
		MasterHostCount:     types.Int64Null(),
		MasterHostGroupIds:  types.SetNull(types.StringType),
		Monitoring:          types.SetNull(yandexMdbGreenplumClusterV2MonitoringStructModelType),
		Name:                types.StringNull(),
		NetworkId:           types.StringNull(),
		PlannedOperation:    types.ObjectNull(yandexMdbGreenplumClusterV2PlannedOperationModelType.AttrTypes),
		SecurityGroupIds:    types.SetNull(types.StringType),
		SegmentConfig:       types.ObjectNull(yandexMdbGreenplumClusterV2SegmentConfigModelType.AttrTypes),
		SegmentHostCount:    types.Int64Null(),
		SegmentHostGroupIds: types.SetNull(types.StringType),
		SegmentInHost:       types.Int64Null(),
		ServiceAccountId:    types.StringNull(),
		UserName:            types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2DatasourceModelFillUnknown(target yandexMdbGreenplumClusterV2DatasourceModel) yandexMdbGreenplumClusterV2DatasourceModel {
	if target.CloudStorage.IsUnknown() || target.CloudStorage.IsNull() {
		target.CloudStorage = types.ObjectNull(yandexMdbGreenplumClusterV2CloudStorageModelType.AttrTypes)
	}
	if target.ClusterConfig.IsUnknown() || target.ClusterConfig.IsNull() {
		target.ClusterConfig = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigModelType.AttrTypes)
	}
	if target.ID.IsUnknown() || target.ID.IsNull() {
		target.ID = types.StringNull()
	}
	if target.Config.IsUnknown() || target.Config.IsNull() {
		target.Config = types.ObjectNull(yandexMdbGreenplumClusterV2ConfigModelType.AttrTypes)
	}
	if target.CreatedAt.IsUnknown() || target.CreatedAt.IsNull() {
		target.CreatedAt = types.StringNull()
	}
	if target.DeletionProtection.IsUnknown() || target.DeletionProtection.IsNull() {
		target.DeletionProtection = types.BoolNull()
	}
	if target.Description.IsUnknown() || target.Description.IsNull() {
		target.Description = types.StringNull()
	}
	if target.Environment.IsUnknown() || target.Environment.IsNull() {
		target.Environment = types.StringNull()
	}
	if target.FolderId.IsUnknown() || target.FolderId.IsNull() {
		target.FolderId = types.StringNull()
	}
	if target.HostGroupIds.IsUnknown() || target.HostGroupIds.IsNull() {
		target.HostGroupIds = types.SetNull(types.StringType)
	}
	if target.Labels.IsUnknown() || target.Labels.IsNull() {
		target.Labels = types.MapNull(types.StringType)
	}
	if target.Logging.IsUnknown() || target.Logging.IsNull() {
		target.Logging = types.ObjectNull(yandexMdbGreenplumClusterV2LoggingModelType.AttrTypes)
	}
	if target.MaintenanceWindow.IsUnknown() || target.MaintenanceWindow.IsNull() {
		target.MaintenanceWindow = types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowModelType.AttrTypes)
	}
	if target.MasterConfig.IsUnknown() || target.MasterConfig.IsNull() {
		target.MasterConfig = types.ObjectNull(yandexMdbGreenplumClusterV2MasterConfigModelType.AttrTypes)
	}
	if target.MasterHostCount.IsUnknown() || target.MasterHostCount.IsNull() {
		target.MasterHostCount = types.Int64Null()
	}
	if target.MasterHostGroupIds.IsUnknown() || target.MasterHostGroupIds.IsNull() {
		target.MasterHostGroupIds = types.SetNull(types.StringType)
	}
	if target.Monitoring.IsUnknown() || target.Monitoring.IsNull() {
		target.Monitoring = types.SetNull(yandexMdbGreenplumClusterV2MonitoringStructModelType)
	}
	if target.Name.IsUnknown() || target.Name.IsNull() {
		target.Name = types.StringNull()
	}
	if target.NetworkId.IsUnknown() || target.NetworkId.IsNull() {
		target.NetworkId = types.StringNull()
	}
	if target.PlannedOperation.IsUnknown() || target.PlannedOperation.IsNull() {
		target.PlannedOperation = types.ObjectNull(yandexMdbGreenplumClusterV2PlannedOperationModelType.AttrTypes)
	}
	if target.SecurityGroupIds.IsUnknown() || target.SecurityGroupIds.IsNull() {
		target.SecurityGroupIds = types.SetNull(types.StringType)
	}
	if target.SegmentConfig.IsUnknown() || target.SegmentConfig.IsNull() {
		target.SegmentConfig = types.ObjectNull(yandexMdbGreenplumClusterV2SegmentConfigModelType.AttrTypes)
	}
	if target.SegmentHostCount.IsUnknown() || target.SegmentHostCount.IsNull() {
		target.SegmentHostCount = types.Int64Null()
	}
	if target.SegmentHostGroupIds.IsUnknown() || target.SegmentHostGroupIds.IsNull() {
		target.SegmentHostGroupIds = types.SetNull(types.StringType)
	}
	if target.SegmentInHost.IsUnknown() || target.SegmentInHost.IsNull() {
		target.SegmentInHost = types.Int64Null()
	}
	if target.ServiceAccountId.IsUnknown() || target.ServiceAccountId.IsNull() {
		target.ServiceAccountId = types.StringNull()
	}
	if target.UserName.IsUnknown() || target.UserName.IsNull() {
		target.UserName = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2DatasourceModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"cloud_storage":          yandexMdbGreenplumClusterV2CloudStorageModelType,
		"cluster_config":         yandexMdbGreenplumClusterV2ClusterConfigModelType,
		"cluster_id":             types.StringType,
		"id":                     types.StringType,
		"config":                 yandexMdbGreenplumClusterV2ConfigModelType,
		"created_at":             types.StringType,
		"deletion_protection":    types.BoolType,
		"description":            types.StringType,
		"environment":            types.StringType,
		"folder_id":              types.StringType,
		"host_group_ids":         types.SetType{ElemType: types.StringType},
		"labels":                 types.MapType{ElemType: types.StringType},
		"logging":                yandexMdbGreenplumClusterV2LoggingModelType,
		"maintenance_window":     yandexMdbGreenplumClusterV2MaintenanceWindowModelType,
		"master_config":          yandexMdbGreenplumClusterV2MasterConfigModelType,
		"master_host_count":      types.Int64Type,
		"master_host_group_ids":  types.SetType{ElemType: types.StringType},
		"monitoring":             types.SetType{ElemType: yandexMdbGreenplumClusterV2MonitoringStructModelType},
		"name":                   types.StringType,
		"network_id":             types.StringType,
		"planned_operation":      yandexMdbGreenplumClusterV2PlannedOperationModelType,
		"security_group_ids":     types.SetType{ElemType: types.StringType},
		"segment_config":         yandexMdbGreenplumClusterV2SegmentConfigModelType,
		"segment_host_count":     types.Int64Type,
		"segment_host_group_ids": types.SetType{ElemType: types.StringType},
		"segment_in_host":        types.Int64Type,
		"service_account_id":     types.StringType,
		"user_name":              types.StringType,
		"timeouts":               timeouts.AttributesAll(context.Background()).GetType(),
	},
}

func flattenYandexMdbGreenplumClusterV2Datasource(ctx context.Context,
	yandexMdbGreenplumClusterV2Datasource *greenplum.Cluster,
	state yandexMdbGreenplumClusterV2DatasourceModel,
	to timeouts.Value,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2Datasource == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2DatasourceModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2DatasourceModelType.AttrTypes, yandexMdbGreenplumClusterV2DatasourceModel{
		CloudStorage:        flattenYandexMdbGreenplumClusterV2CloudStorage(ctx, yandexMdbGreenplumClusterV2Datasource.GetCloudStorage(), diags),
		ClusterConfig:       flattenYandexMdbGreenplumClusterV2ClusterConfig(ctx, yandexMdbGreenplumClusterV2Datasource.GetClusterConfig(), converter.ExpandObject(ctx, state.ClusterConfig, yandexMdbGreenplumClusterV2ClusterConfigModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigModel), diags),
		ID:                  types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetId()),
		Config:              flattenYandexMdbGreenplumClusterV2Config(ctx, yandexMdbGreenplumClusterV2Datasource.GetConfig(), converter.ExpandObject(ctx, state.Config, yandexMdbGreenplumClusterV2ConfigModel{}, diags).(yandexMdbGreenplumClusterV2ConfigModel), diags),
		CreatedAt:           types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetCreatedAt().AsTime().Format(time.RFC3339)),
		DeletionProtection:  types.BoolValue(yandexMdbGreenplumClusterV2Datasource.GetDeletionProtection()),
		Description:         types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetDescription()),
		Environment:         types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetEnvironment().String()),
		FolderId:            types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetFolderId()),
		HostGroupIds:        flattenYandexMdbGreenplumClusterV2HostGroupIds(ctx, yandexMdbGreenplumClusterV2Datasource.GetHostGroupIds(), state.HostGroupIds, diags),
		Labels:              flattenYandexMdbGreenplumClusterV2Labels(ctx, yandexMdbGreenplumClusterV2Datasource.GetLabels(), state.Labels, diags),
		Logging:             flattenYandexMdbGreenplumClusterV2Logging(ctx, yandexMdbGreenplumClusterV2Datasource.GetLogging(), diags),
		MaintenanceWindow:   flattenYandexMdbGreenplumClusterV2MaintenanceWindow(ctx, yandexMdbGreenplumClusterV2Datasource.GetMaintenanceWindow(), diags),
		MasterConfig:        flattenYandexMdbGreenplumClusterV2MasterConfig(ctx, yandexMdbGreenplumClusterV2Datasource.GetMasterConfig(), diags),
		MasterHostCount:     types.Int64Value(yandexMdbGreenplumClusterV2Datasource.GetMasterHostCount()),
		MasterHostGroupIds:  flattenYandexMdbGreenplumClusterV2HostGroupIds(ctx, yandexMdbGreenplumClusterV2Datasource.GetMasterHostGroupIds(), state.MasterHostGroupIds, diags),
		Monitoring:          flattenYandexMdbGreenplumClusterV2Monitoring(ctx, yandexMdbGreenplumClusterV2Datasource.GetMonitoring(), state.Monitoring, diags),
		Name:                types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetName()),
		NetworkId:           types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetNetworkId()),
		PlannedOperation:    flattenYandexMdbGreenplumClusterV2PlannedOperation(ctx, yandexMdbGreenplumClusterV2Datasource.GetPlannedOperation(), diags),
		SecurityGroupIds:    flattenYandexMdbGreenplumClusterV2SecurityGroupIds(ctx, yandexMdbGreenplumClusterV2Datasource.GetSecurityGroupIds(), state.SecurityGroupIds, diags),
		SegmentConfig:       flattenYandexMdbGreenplumClusterV2SegmentConfig(ctx, yandexMdbGreenplumClusterV2Datasource.GetSegmentConfig(), diags),
		SegmentHostCount:    types.Int64Value(yandexMdbGreenplumClusterV2Datasource.GetSegmentHostCount()),
		SegmentHostGroupIds: flattenYandexMdbGreenplumClusterV2HostGroupIds(ctx, yandexMdbGreenplumClusterV2Datasource.GetSegmentHostGroupIds(), state.SegmentHostGroupIds, diags),
		SegmentInHost:       types.Int64Value(yandexMdbGreenplumClusterV2Datasource.GetSegmentInHost()),
		ServiceAccountId:    types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetServiceAccountId()),
		UserName:            types.StringValue(yandexMdbGreenplumClusterV2Datasource.GetUserName()),
		Timeouts:            to,
	})
	diags.Append(diag...)
	return value
}
