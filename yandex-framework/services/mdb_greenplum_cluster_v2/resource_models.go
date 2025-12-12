package mdb_greenplum_cluster_v2

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

const GB = 1024 * 1024 * 1024

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel) GetHours() types.Int64 {
	return m.Hours
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel) GetMinutes() types.Int64 {
	return m.Minutes
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel{
		Hours:   types.Int64Null(),
		Minutes: types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel {
	if target.Hours.IsUnknown() || target.Hours.IsNull() {
		target.Hours = types.Int64Null()
	}
	if target.Minutes.IsUnknown() || target.Minutes.IsNull() {
		target.Minutes = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct *greenplum.BackgroundActivityStartAt,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel{
		Hours:   types.Int64Value(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct.GetHours()),
		Minutes: types.Int64Value(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct.GetMinutes()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructState types.Object, diags *diag.Diagnostics) *greenplum.BackgroundActivityStartAt {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel, diags *diag.Diagnostics) *greenplum.BackgroundActivityStartAt {
	value := &greenplum.BackgroundActivityStartAt{}
	value.SetHours(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructState.Hours.ValueInt64())
	value.SetMinutes(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructState.Minutes.ValueInt64())
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2MonitoringStructModel struct {
	Description types.String `tfsdk:"description"`
	Link        types.String `tfsdk:"link"`
	Name        types.String `tfsdk:"name"`
}

func (m *yandexMdbGreenplumClusterV2MonitoringStructModel) GetDescription() types.String {
	return m.Description
}
func (m *yandexMdbGreenplumClusterV2MonitoringStructModel) GetLink() types.String {
	return m.Link
}
func (m *yandexMdbGreenplumClusterV2MonitoringStructModel) GetName() types.String {
	return m.Name
}

func NewYandexMdbGreenplumClusterV2MonitoringStructModel() yandexMdbGreenplumClusterV2MonitoringStructModel {
	return yandexMdbGreenplumClusterV2MonitoringStructModel{
		Description: types.StringNull(),
		Link:        types.StringNull(),
		Name:        types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2MonitoringStructModelFillUnknown(target yandexMdbGreenplumClusterV2MonitoringStructModel) yandexMdbGreenplumClusterV2MonitoringStructModel {
	if target.Description.IsUnknown() || target.Description.IsNull() {
		target.Description = types.StringNull()
	}
	if target.Link.IsUnknown() || target.Link.IsNull() {
		target.Link = types.StringNull()
	}
	if target.Name.IsUnknown() || target.Name.IsNull() {
		target.Name = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2MonitoringStructModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"description": types.StringType,
		"link":        types.StringType,
		"name":        types.StringType,
	},
}

func flattenYandexMdbGreenplumClusterV2MonitoringStruct(ctx context.Context,
	yandexMdbGreenplumClusterV2MonitoringStruct *greenplum.Monitoring,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2MonitoringStruct == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2MonitoringStructModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2MonitoringStructModelType.AttrTypes, yandexMdbGreenplumClusterV2MonitoringStructModel{
		Description: types.StringValue(yandexMdbGreenplumClusterV2MonitoringStruct.GetDescription()),
		Link:        types.StringValue(yandexMdbGreenplumClusterV2MonitoringStruct.GetLink()),
		Name:        types.StringValue(yandexMdbGreenplumClusterV2MonitoringStruct.GetName()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2MonitoringStruct(ctx context.Context, yandexMdbGreenplumClusterV2MonitoringStructState types.Object, diags *diag.Diagnostics) *greenplum.Monitoring {
	if yandexMdbGreenplumClusterV2MonitoringStructState.IsNull() || yandexMdbGreenplumClusterV2MonitoringStructState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2MonitoringStruct yandexMdbGreenplumClusterV2MonitoringStructModel
	diags.Append(yandexMdbGreenplumClusterV2MonitoringStructState.As(ctx, &yandexMdbGreenplumClusterV2MonitoringStruct, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2MonitoringStructModel(ctx, yandexMdbGreenplumClusterV2MonitoringStruct, diags)
}

func expandYandexMdbGreenplumClusterV2MonitoringStructModel(ctx context.Context, yandexMdbGreenplumClusterV2MonitoringStructState yandexMdbGreenplumClusterV2MonitoringStructModel, diags *diag.Diagnostics) *greenplum.Monitoring {
	value := &greenplum.Monitoring{}
	value.SetDescription(yandexMdbGreenplumClusterV2MonitoringStructState.Description.ValueString())
	value.SetLink(yandexMdbGreenplumClusterV2MonitoringStructState.Link.ValueString())
	value.SetName(yandexMdbGreenplumClusterV2MonitoringStructState.Name.ValueString())
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2Model struct {
	Restore             types.Object   `tfsdk:"restore"`
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
	UserPassword        types.String   `tfsdk:"user_password"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

func (m *yandexMdbGreenplumClusterV2Model) GetRestore() types.Object {
	return m.Restore
}
func (m *yandexMdbGreenplumClusterV2Model) GetCloudStorage() types.Object {
	return m.CloudStorage
}
func (m *yandexMdbGreenplumClusterV2Model) GetClusterConfig() types.Object {
	return m.ClusterConfig
}
func (m *yandexMdbGreenplumClusterV2Model) GetID() types.String {
	return m.ID
}
func (m *yandexMdbGreenplumClusterV2Model) GetConfig() types.Object {
	return m.Config
}
func (m *yandexMdbGreenplumClusterV2Model) GetCreatedAt() types.String {
	return m.CreatedAt
}
func (m *yandexMdbGreenplumClusterV2Model) GetDeletionProtection() types.Bool {
	return m.DeletionProtection
}
func (m *yandexMdbGreenplumClusterV2Model) GetDescription() types.String {
	return m.Description
}
func (m *yandexMdbGreenplumClusterV2Model) GetEnvironment() types.String {
	return m.Environment
}
func (m *yandexMdbGreenplumClusterV2Model) GetFolderId() types.String {
	return m.FolderId
}
func (m *yandexMdbGreenplumClusterV2Model) GetHostGroupIds() types.Set {
	return m.HostGroupIds
}
func (m *yandexMdbGreenplumClusterV2Model) GetLabels() types.Map {
	return m.Labels
}
func (m *yandexMdbGreenplumClusterV2Model) GetLogging() types.Object {
	return m.Logging
}
func (m *yandexMdbGreenplumClusterV2Model) GetMaintenanceWindow() types.Object {
	return m.MaintenanceWindow
}
func (m *yandexMdbGreenplumClusterV2Model) GetMasterConfig() types.Object {
	return m.MasterConfig
}
func (m *yandexMdbGreenplumClusterV2Model) GetMasterHostCount() types.Int64 {
	return m.MasterHostCount
}
func (m *yandexMdbGreenplumClusterV2Model) GetMasterHostGroupIds() types.Set {
	return m.MasterHostGroupIds
}
func (m *yandexMdbGreenplumClusterV2Model) GetMonitoring() types.Set {
	return m.Monitoring
}
func (m *yandexMdbGreenplumClusterV2Model) GetName() types.String {
	return m.Name
}
func (m *yandexMdbGreenplumClusterV2Model) GetNetworkId() types.String {
	return m.NetworkId
}
func (m *yandexMdbGreenplumClusterV2Model) GetPlannedOperation() types.Object {
	return m.PlannedOperation
}
func (m *yandexMdbGreenplumClusterV2Model) GetSecurityGroupIds() types.Set {
	return m.SecurityGroupIds
}
func (m *yandexMdbGreenplumClusterV2Model) GetSegmentConfig() types.Object {
	return m.SegmentConfig
}
func (m *yandexMdbGreenplumClusterV2Model) GetSegmentHostCount() types.Int64 {
	return m.SegmentHostCount
}
func (m *yandexMdbGreenplumClusterV2Model) GetSegmentHostGroupIds() types.Set {
	return m.SegmentHostGroupIds
}
func (m *yandexMdbGreenplumClusterV2Model) GetSegmentInHost() types.Int64 {
	return m.SegmentInHost
}
func (m *yandexMdbGreenplumClusterV2Model) GetServiceAccountId() types.String {
	return m.ServiceAccountId
}
func (m *yandexMdbGreenplumClusterV2Model) GetUserName() types.String {
	return m.UserName
}
func (m *yandexMdbGreenplumClusterV2Model) GetUserPassword() types.String {
	return m.UserPassword
}

func NewYandexMdbGreenplumClusterV2Model() yandexMdbGreenplumClusterV2Model {
	return yandexMdbGreenplumClusterV2Model{
		Restore:             types.ObjectNull(yandexMdbGreenplumClusterV2RestoreModelType.AttrTypes),
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
		UserPassword:        types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2ModelFillUnknown(target yandexMdbGreenplumClusterV2Model) yandexMdbGreenplumClusterV2Model {
	if target.Restore.IsUnknown() || target.Restore.IsNull() {
		target.Restore = types.ObjectNull(yandexMdbGreenplumClusterV2RestoreModelType.AttrTypes)
	}
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
	if target.UserPassword.IsUnknown() || target.UserPassword.IsNull() {
		target.UserPassword = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2ModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"restore":                yandexMdbGreenplumClusterV2RestoreModelType,
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
		"user_password":          types.StringType,
		"timeouts":               timeouts.AttributesAll(context.Background()).GetType(),
	},
}

func flattenYandexMdbGreenplumClusterV2(ctx context.Context,
	yandexMdbGreenplumClusterV2 *greenplum.Cluster,
	state yandexMdbGreenplumClusterV2Model,
	to timeouts.Value,
	diags *diag.Diagnostics) yandexMdbGreenplumClusterV2Model {
	if yandexMdbGreenplumClusterV2 == nil {
		return yandexMdbGreenplumClusterV2Model{}
	}
	return yandexMdbGreenplumClusterV2Model{
		Restore:             state.Restore,
		CloudStorage:        flattenYandexMdbGreenplumClusterV2CloudStorage(ctx, yandexMdbGreenplumClusterV2.GetCloudStorage(), diags),
		ClusterConfig:       flattenYandexMdbGreenplumClusterV2ClusterConfig(ctx, yandexMdbGreenplumClusterV2.GetClusterConfig(), converter.ExpandObject(ctx, state.ClusterConfig, yandexMdbGreenplumClusterV2ClusterConfigModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigModel), diags),
		ID:                  types.StringValue(yandexMdbGreenplumClusterV2.GetId()),
		Config:              flattenYandexMdbGreenplumClusterV2Config(ctx, yandexMdbGreenplumClusterV2.GetConfig(), converter.ExpandObject(ctx, state.Config, yandexMdbGreenplumClusterV2ConfigModel{}, diags).(yandexMdbGreenplumClusterV2ConfigModel), diags),
		CreatedAt:           types.StringValue(yandexMdbGreenplumClusterV2.GetCreatedAt().AsTime().Format(time.RFC3339)),
		DeletionProtection:  types.BoolValue(yandexMdbGreenplumClusterV2.GetDeletionProtection()),
		Description:         types.StringValue(yandexMdbGreenplumClusterV2.GetDescription()),
		Environment:         flattenEnum(yandexMdbGreenplumClusterV2.GetEnvironment()),
		FolderId:            types.StringValue(yandexMdbGreenplumClusterV2.GetFolderId()),
		HostGroupIds:        flattenYandexMdbGreenplumClusterV2HostGroupIds(ctx, yandexMdbGreenplumClusterV2.GetHostGroupIds(), state.HostGroupIds, diags),
		Labels:              flattenYandexMdbGreenplumClusterV2Labels(ctx, yandexMdbGreenplumClusterV2.GetLabels(), state.Labels, diags),
		Logging:             flattenYandexMdbGreenplumClusterV2Logging(ctx, yandexMdbGreenplumClusterV2.GetLogging(), diags),
		MaintenanceWindow:   flattenYandexMdbGreenplumClusterV2MaintenanceWindow(ctx, yandexMdbGreenplumClusterV2.GetMaintenanceWindow(), diags),
		MasterConfig:        flattenYandexMdbGreenplumClusterV2MasterConfig(ctx, yandexMdbGreenplumClusterV2.GetMasterConfig(), diags),
		MasterHostCount:     types.Int64Value(yandexMdbGreenplumClusterV2.GetMasterHostCount()),
		MasterHostGroupIds:  flattenYandexMdbGreenplumClusterV2HostGroupIds(ctx, yandexMdbGreenplumClusterV2.GetMasterHostGroupIds(), state.MasterHostGroupIds, diags),
		Monitoring:          flattenYandexMdbGreenplumClusterV2Monitoring(ctx, yandexMdbGreenplumClusterV2.GetMonitoring(), state.Monitoring, diags),
		Name:                types.StringValue(yandexMdbGreenplumClusterV2.GetName()),
		NetworkId:           types.StringValue(yandexMdbGreenplumClusterV2.GetNetworkId()),
		PlannedOperation:    flattenYandexMdbGreenplumClusterV2PlannedOperation(ctx, yandexMdbGreenplumClusterV2.GetPlannedOperation(), diags),
		SecurityGroupIds:    flattenYandexMdbGreenplumClusterV2SecurityGroupIds(ctx, yandexMdbGreenplumClusterV2.GetSecurityGroupIds(), state.SecurityGroupIds, diags),
		SegmentConfig:       flattenYandexMdbGreenplumClusterV2SegmentConfig(ctx, yandexMdbGreenplumClusterV2.GetSegmentConfig(), diags),
		SegmentHostCount:    types.Int64Value(yandexMdbGreenplumClusterV2.GetSegmentHostCount()),
		SegmentHostGroupIds: flattenYandexMdbGreenplumClusterV2HostGroupIds(ctx, yandexMdbGreenplumClusterV2.GetSegmentHostGroupIds(), state.SegmentHostGroupIds, diags),
		SegmentInHost:       types.Int64Value(yandexMdbGreenplumClusterV2.GetSegmentInHost()),
		ServiceAccountId:    types.StringValue(yandexMdbGreenplumClusterV2.GetServiceAccountId()),
		UserName:            types.StringValue(yandexMdbGreenplumClusterV2.GetUserName()),
		UserPassword:        state.UserPassword,
		Timeouts:            to,
	}
}

type yandexMdbGreenplumClusterV2RestoreModel struct {
	BackupId      types.String `tfsdk:"backup_id"`
	TimeInclusive types.Bool   `tfsdk:"time_inclusive"`
	Time          types.String `tfsdk:"time"`
}

func (m *yandexMdbGreenplumClusterV2RestoreModel) GetBackupId() types.String {
	return m.BackupId
}

func (m *yandexMdbGreenplumClusterV2RestoreModel) GetTimeInclusive() types.Bool {
	return m.TimeInclusive
}

func (m *yandexMdbGreenplumClusterV2RestoreModel) GetTime() types.String {
	return m.Time
}

func NewYandexMdbGreenplumClusterV2RestoreModel() yandexMdbGreenplumClusterV2RestoreModel {
	return yandexMdbGreenplumClusterV2RestoreModel{
		BackupId:      types.StringNull(),
		TimeInclusive: types.BoolNull(),
		Time:          types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2RestoreModelFillUnknown(target yandexMdbGreenplumClusterV2RestoreModel) yandexMdbGreenplumClusterV2RestoreModel {
	if target.BackupId.IsUnknown() || target.BackupId.IsNull() {
		target.BackupId = types.StringNull()
	}
	if target.TimeInclusive.IsUnknown() || target.TimeInclusive.IsNull() {
		target.TimeInclusive = types.BoolNull()
	}
	if target.Time.IsUnknown() || target.Time.IsNull() {
		target.Time = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2RestoreModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"backup_id":      types.StringType,
		"time_inclusive": types.BoolType,
		"time":           types.StringType,
	},
}

type yandexMdbGreenplumClusterV2CloudStorageModel struct {
	Enable types.Bool `tfsdk:"enable"`
}

func (m *yandexMdbGreenplumClusterV2CloudStorageModel) GetEnable() types.Bool {
	return m.Enable
}

func NewYandexMdbGreenplumClusterV2CloudStorageModel() yandexMdbGreenplumClusterV2CloudStorageModel {
	return yandexMdbGreenplumClusterV2CloudStorageModel{
		Enable: types.BoolNull(),
	}
}

func yandexMdbGreenplumClusterV2CloudStorageModelFillUnknown(target yandexMdbGreenplumClusterV2CloudStorageModel) yandexMdbGreenplumClusterV2CloudStorageModel {
	if target.Enable.IsUnknown() || target.Enable.IsNull() {
		target.Enable = types.BoolNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2CloudStorageModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"enable": types.BoolType,
	},
}

func flattenYandexMdbGreenplumClusterV2CloudStorage(ctx context.Context,
	yandexMdbGreenplumClusterV2CloudStorage *greenplum.CloudStorage,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2CloudStorage == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2CloudStorageModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2CloudStorageModelType.AttrTypes, yandexMdbGreenplumClusterV2CloudStorageModel{
		Enable: types.BoolValue(yandexMdbGreenplumClusterV2CloudStorage.GetEnable()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2CloudStorage(ctx context.Context, yandexMdbGreenplumClusterV2CloudStorageState types.Object, diags *diag.Diagnostics) *greenplum.CloudStorage {
	if yandexMdbGreenplumClusterV2CloudStorageState.IsNull() || yandexMdbGreenplumClusterV2CloudStorageState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2CloudStorage yandexMdbGreenplumClusterV2CloudStorageModel
	diags.Append(yandexMdbGreenplumClusterV2CloudStorageState.As(ctx, &yandexMdbGreenplumClusterV2CloudStorage, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2CloudStorageModel(ctx, yandexMdbGreenplumClusterV2CloudStorage, diags)
}

func expandYandexMdbGreenplumClusterV2CloudStorageModel(ctx context.Context, yandexMdbGreenplumClusterV2CloudStorageState yandexMdbGreenplumClusterV2CloudStorageModel, diags *diag.Diagnostics) *greenplum.CloudStorage {
	value := &greenplum.CloudStorage{}
	value.SetEnable(yandexMdbGreenplumClusterV2CloudStorageState.Enable.ValueBool())
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigModel struct {
	BackgroundActivities types.Object `tfsdk:"background_activities"`
	GreenplumConfig6     types.Object `tfsdk:"greenplum_config"`
	Pool                 types.Object `tfsdk:"pool"`
	PxfConfig            types.Object `tfsdk:"pxf_config"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigModel) GetBackgroundActivities() types.Object {
	return m.BackgroundActivities
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigModel) GetGreenplumConfig6() types.Object {
	return m.GreenplumConfig6
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigModel) GetPool() types.Object {
	return m.Pool
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigModel) GetPxfConfig() types.Object {
	return m.PxfConfig
}

func NewYandexMdbGreenplumClusterV2ClusterConfigModel() yandexMdbGreenplumClusterV2ClusterConfigModel {
	return yandexMdbGreenplumClusterV2ClusterConfigModel{
		BackgroundActivities: types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModelType.AttrTypes),
		GreenplumConfig6:     types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelType.AttrTypes),
		Pool:                 types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigPoolModelType.AttrTypes),
		PxfConfig:            types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModelType.AttrTypes),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigModel) yandexMdbGreenplumClusterV2ClusterConfigModel {
	if target.BackgroundActivities.IsUnknown() || target.BackgroundActivities.IsNull() {
		target.BackgroundActivities = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModelType.AttrTypes)
	}
	if target.GreenplumConfig6.IsUnknown() || target.GreenplumConfig6.IsNull() {
		target.GreenplumConfig6 = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelType.AttrTypes)
	}
	if target.Pool.IsUnknown() || target.Pool.IsNull() {
		target.Pool = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigPoolModelType.AttrTypes)
	}
	if target.PxfConfig.IsUnknown() || target.PxfConfig.IsNull() {
		target.PxfConfig = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModelType.AttrTypes)
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"background_activities": yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModelType,
		"greenplum_config":      yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelType,
		"pool":                  yandexMdbGreenplumClusterV2ClusterConfigPoolModelType,
		"pxf_config":            yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModelType,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfig(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfig *greenplum.ClusterConfigSet,
	state yandexMdbGreenplumClusterV2ClusterConfigModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfig == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigModel{
		BackgroundActivities: flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities(ctx, yandexMdbGreenplumClusterV2ClusterConfig.GetBackgroundActivities(), converter.ExpandObject(ctx, state.BackgroundActivities, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel), diags),
		GreenplumConfig6:     flattenYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6(ctx, yandexMdbGreenplumClusterV2ClusterConfig.GetGreenplumConfigSet_6(), diags),
		Pool:                 flattenYandexMdbGreenplumClusterV2ClusterConfigPool(ctx, yandexMdbGreenplumClusterV2ClusterConfig.GetPool(), converter.ExpandObject(ctx, state.Pool, yandexMdbGreenplumClusterV2ClusterConfigPoolModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigPoolModel), diags),
		PxfConfig:            flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfig.GetPxfConfig(), converter.ExpandObject(ctx, state.PxfConfig, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel), diags),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfig_create(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigState types.Object, diags *diag.Diagnostics) *greenplum.ConfigSpec {
	if yandexMdbGreenplumClusterV2ClusterConfigState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfig yandexMdbGreenplumClusterV2ClusterConfigModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfig, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigModel_create(ctx, yandexMdbGreenplumClusterV2ClusterConfig, diags)
}
func expandYandexMdbGreenplumClusterV2ClusterConfig_update(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigState types.Object, diags *diag.Diagnostics) *greenplum.ConfigSpec {
	if yandexMdbGreenplumClusterV2ClusterConfigState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfig yandexMdbGreenplumClusterV2ClusterConfigModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfig, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigModel_update(ctx, yandexMdbGreenplumClusterV2ClusterConfig, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigModel_create(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigState yandexMdbGreenplumClusterV2ClusterConfigModel, diags *diag.Diagnostics) *greenplum.ConfigSpec {
	value := &greenplum.ConfigSpec{}
	value.SetBackgroundActivities(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.BackgroundActivities, diags))
	if !(yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6.IsUnknown() || yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6.Equal(types.Object{})) {
		value.SetGreenplumConfig_6(expandYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6, diags))
	}
	value.SetPool(expandYandexMdbGreenplumClusterV2ClusterConfigPool_create(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.Pool, diags))
	value.SetPxfConfig(expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfig_create(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.PxfConfig, diags))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2ClusterConfigModel_update(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigState yandexMdbGreenplumClusterV2ClusterConfigModel, diags *diag.Diagnostics) *greenplum.ConfigSpec {
	value := &greenplum.ConfigSpec{}
	value.SetBackgroundActivities(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.BackgroundActivities, diags))
	if !(yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6.IsUnknown() || yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6.Equal(types.Object{})) {
		value.SetGreenplumConfig_6(expandYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6, diags))
	}
	value.SetPool(expandYandexMdbGreenplumClusterV2ClusterConfigPool_update(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.Pool, diags))
	value.SetPxfConfig(expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfig_update(ctx, yandexMdbGreenplumClusterV2ClusterConfigState.PxfConfig, diags))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel struct {
	AnalyzeAndVacuum   types.Object `tfsdk:"analyze_and_vacuum"`
	QueryKillerScripts types.Object `tfsdk:"query_killer_scripts"`
	TableSizes         types.Object `tfsdk:"table_sizes"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel) GetAnalyzeAndVacuum() types.Object {
	return m.AnalyzeAndVacuum
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel) GetQueryKillerScripts() types.Object {
	return m.QueryKillerScripts
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel) GetTableSizes() types.Object {
	return m.TableSizes
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel{
		AnalyzeAndVacuum:   types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModelType.AttrTypes),
		QueryKillerScripts: types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModelType.AttrTypes),
		TableSizes:         types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModelType.AttrTypes),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel {
	if target.AnalyzeAndVacuum.IsUnknown() || target.AnalyzeAndVacuum.IsNull() {
		target.AnalyzeAndVacuum = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModelType.AttrTypes)
	}
	if target.QueryKillerScripts.IsUnknown() || target.QueryKillerScripts.IsNull() {
		target.QueryKillerScripts = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModelType.AttrTypes)
	}
	if target.TableSizes.IsUnknown() || target.TableSizes.IsNull() {
		target.TableSizes = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModelType.AttrTypes)
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"analyze_and_vacuum":   yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModelType,
		"query_killer_scripts": yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModelType,
		"table_sizes":          yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModelType,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities *greenplum.BackgroundActivitiesConfig,
	state yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel{
		AnalyzeAndVacuum:   flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities.GetAnalyzeAndVacuum(), diags),
		QueryKillerScripts: flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities.GetQueryKillerScripts(), converter.ExpandObject(ctx, state.QueryKillerScripts, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel), diags),
		TableSizes:         flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities.GetTableSizes(), converter.ExpandObject(ctx, state.TableSizes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel), diags),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState types.Object, diags *diag.Diagnostics) *greenplum.BackgroundActivitiesConfig {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivities, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel, diags *diag.Diagnostics) *greenplum.BackgroundActivitiesConfig {
	value := &greenplum.BackgroundActivitiesConfig{}
	value.SetAnalyzeAndVacuum(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.AnalyzeAndVacuum, diags))
	value.SetQueryKillerScripts(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.QueryKillerScripts, diags))
	value.SetTableSizes(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.TableSizes, diags))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel struct {
	AnalyzeTimeout types.Int64  `tfsdk:"analyze_timeout"`
	Start          types.Object `tfsdk:"start"`
	VacuumTimeout  types.Int64  `tfsdk:"vacuum_timeout"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel) GetAnalyzeTimeout() types.Int64 {
	return m.AnalyzeTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel) GetStart() types.Object {
	return m.Start
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel) GetVacuumTimeout() types.Int64 {
	return m.VacuumTimeout
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel{
		AnalyzeTimeout: types.Int64Null(),
		Start:          types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModelType.AttrTypes),
		VacuumTimeout:  types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel {
	if target.AnalyzeTimeout.IsUnknown() || target.AnalyzeTimeout.IsNull() {
		target.AnalyzeTimeout = types.Int64Null()
	}
	if target.Start.IsUnknown() || target.Start.IsNull() {
		target.Start = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModelType.AttrTypes)
	}
	if target.VacuumTimeout.IsUnknown() || target.VacuumTimeout.IsNull() {
		target.VacuumTimeout = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"analyze_timeout": types.Int64Type,
		"start":           yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModelType,
		"vacuum_timeout":  types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum *greenplum.AnalyzeAndVacuum,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel{
		AnalyzeTimeout: flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum.GetAnalyzeTimeout()),
		Start:          flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum.GetStart(), diags),
		VacuumTimeout:  flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum.GetVacuumTimeout()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState types.Object, diags *diag.Diagnostics) *greenplum.AnalyzeAndVacuum {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuum, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel, diags *diag.Diagnostics) *greenplum.AnalyzeAndVacuum {
	value := &greenplum.AnalyzeAndVacuum{}
	value.SetAnalyzeTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.AnalyzeTimeout))
	value.SetStart(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.Start, diags))
	value.SetVacuumTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.VacuumTimeout))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel) GetHours() types.Int64 {
	return m.Hours
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel) GetMinutes() types.Int64 {
	return m.Minutes
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel{
		Hours:   types.Int64Null(),
		Minutes: types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel {
	if target.Hours.IsUnknown() || target.Hours.IsNull() {
		target.Hours = types.Int64Null()
	}
	if target.Minutes.IsUnknown() || target.Minutes.IsNull() {
		target.Minutes = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart *greenplum.BackgroundActivityStartAt,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel{
		Hours:   types.Int64Value(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart.GetHours()),
		Minutes: types.Int64Value(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart.GetMinutes()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState types.Object, diags *diag.Diagnostics) *greenplum.BackgroundActivityStartAt {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStart, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel, diags *diag.Diagnostics) *greenplum.BackgroundActivityStartAt {
	value := &greenplum.BackgroundActivityStartAt{}
	value.SetHours(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState.Hours.ValueInt64())
	value.SetMinutes(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState.Minutes.ValueInt64())
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel struct {
	Idle              types.Object `tfsdk:"idle"`
	IdleInTransaction types.Object `tfsdk:"idle_in_transaction"`
	LongRunning       types.Object `tfsdk:"long_running"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel) GetIdle() types.Object {
	return m.Idle
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel) GetIdleInTransaction() types.Object {
	return m.IdleInTransaction
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel) GetLongRunning() types.Object {
	return m.LongRunning
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel{
		Idle:              types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModelType.AttrTypes),
		IdleInTransaction: types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModelType.AttrTypes),
		LongRunning:       types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModelType.AttrTypes),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel {
	if target.Idle.IsUnknown() || target.Idle.IsNull() {
		target.Idle = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModelType.AttrTypes)
	}
	if target.IdleInTransaction.IsUnknown() || target.IdleInTransaction.IsNull() {
		target.IdleInTransaction = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModelType.AttrTypes)
	}
	if target.LongRunning.IsUnknown() || target.LongRunning.IsNull() {
		target.LongRunning = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModelType.AttrTypes)
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"idle":                yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModelType,
		"idle_in_transaction": yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModelType,
		"long_running":        yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModelType,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts *greenplum.QueryKillerScripts,
	state yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel{
		Idle:              flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts.GetIdle(), converter.ExpandObject(ctx, state.Idle, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel), diags),
		IdleInTransaction: flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts.GetIdleInTransaction(), converter.ExpandObject(ctx, state.IdleInTransaction, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel), diags),
		LongRunning:       flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts.GetLongRunning(), converter.ExpandObject(ctx, state.LongRunning, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel{}, diags).(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel), diags),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState types.Object, diags *diag.Diagnostics) *greenplum.QueryKillerScripts {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScripts, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel, diags *diag.Diagnostics) *greenplum.QueryKillerScripts {
	value := &greenplum.QueryKillerScripts{}
	value.SetIdle(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.Idle, diags))
	value.SetIdleInTransaction(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.IdleInTransaction, diags))
	value.SetLongRunning(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.LongRunning, diags))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel struct {
	Enable      types.Bool  `tfsdk:"enable"`
	IgnoreUsers types.Set   `tfsdk:"ignore_users"`
	MaxAge      types.Int64 `tfsdk:"max_age"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel) GetEnable() types.Bool {
	return m.Enable
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel) GetIgnoreUsers() types.Set {
	return m.IgnoreUsers
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel) GetMaxAge() types.Int64 {
	return m.MaxAge
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel{
		Enable:      types.BoolNull(),
		IgnoreUsers: types.SetNull(types.StringType),
		MaxAge:      types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel {
	if target.Enable.IsUnknown() || target.Enable.IsNull() {
		target.Enable = types.BoolNull()
	}
	if target.IgnoreUsers.IsUnknown() || target.IgnoreUsers.IsNull() {
		target.IgnoreUsers = types.SetNull(types.StringType)
	}
	if target.MaxAge.IsUnknown() || target.MaxAge.IsNull() {
		target.MaxAge = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"enable":       types.BoolType,
		"ignore_users": types.SetType{ElemType: types.StringType},
		"max_age":      types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle *greenplum.QueryKiller,
	state yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel{
		Enable:      flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle.GetEnable()),
		IgnoreUsers: flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsers(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle.GetIgnoreUsers(), state.IgnoreUsers, diags),
		MaxAge:      flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle.GetMaxAge()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState types.Object, diags *diag.Diagnostics) *greenplum.QueryKiller {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdle, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel, diags *diag.Diagnostics) *greenplum.QueryKiller {
	value := &greenplum.QueryKiller{}
	value.SetEnable(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.Enable))
	value.SetIgnoreUsers(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsers(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.IgnoreUsers, diags))
	value.SetMaxAge(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.MaxAge))
	if diags.HasError() {
		return nil
	}
	return value
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsers(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsers []string, listState types.Set, diags *diag.Diagnostics) types.Set {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsers == nil {
		if !listState.IsNull() && !listState.IsUnknown() && len(listState.Elements()) == 0 {
			return listState
		}
		return types.SetNull(types.StringType)
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersValues []attr.Value
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsers {
		val := types.StringValue(elem)
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersValues = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersValues, val)
	}

	value, diag := types.SetValue(types.StringType, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsers(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersState types.Set, diags *diag.Diagnostics) []string {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersState.Elements()) == 0 {
		return []string{}
	}
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersRes := make([]string, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersState.Elements()))
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersType := make([]types.String, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersState.Elements()))
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersType {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersRes = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersRes, elem.ValueString())
	}
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleIgnoreUsersRes
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel struct {
	Enable      types.Bool  `tfsdk:"enable"`
	IgnoreUsers types.Set   `tfsdk:"ignore_users"`
	MaxAge      types.Int64 `tfsdk:"max_age"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel) GetEnable() types.Bool {
	return m.Enable
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel) GetIgnoreUsers() types.Set {
	return m.IgnoreUsers
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel) GetMaxAge() types.Int64 {
	return m.MaxAge
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel{
		Enable:      types.BoolNull(),
		IgnoreUsers: types.SetNull(types.StringType),
		MaxAge:      types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel {
	if target.Enable.IsUnknown() || target.Enable.IsNull() {
		target.Enable = types.BoolNull()
	}
	if target.IgnoreUsers.IsUnknown() || target.IgnoreUsers.IsNull() {
		target.IgnoreUsers = types.SetNull(types.StringType)
	}
	if target.MaxAge.IsUnknown() || target.MaxAge.IsNull() {
		target.MaxAge = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"enable":       types.BoolType,
		"ignore_users": types.SetType{ElemType: types.StringType},
		"max_age":      types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction *greenplum.QueryKiller,
	state yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel{
		Enable:      flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction.GetEnable()),
		IgnoreUsers: flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsers(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction.GetIgnoreUsers(), state.IgnoreUsers, diags),
		MaxAge:      flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction.GetMaxAge()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState types.Object, diags *diag.Diagnostics) *greenplum.QueryKiller {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransaction, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel, diags *diag.Diagnostics) *greenplum.QueryKiller {
	value := &greenplum.QueryKiller{}
	value.SetEnable(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.Enable))
	value.SetIgnoreUsers(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsers(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.IgnoreUsers, diags))
	value.SetMaxAge(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.MaxAge))
	if diags.HasError() {
		return nil
	}
	return value
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsers(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsers []string, listState types.Set, diags *diag.Diagnostics) types.Set {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsers == nil {
		if !listState.IsNull() && !listState.IsUnknown() && len(listState.Elements()) == 0 {
			return listState
		}
		return types.SetNull(types.StringType)
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersValues []attr.Value
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsers {
		val := types.StringValue(elem)
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersValues = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersValues, val)
	}

	value, diag := types.SetValue(types.StringType, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsers(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersState types.Set, diags *diag.Diagnostics) []string {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersState.Elements()) == 0 {
		return []string{}
	}
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersRes := make([]string, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersState.Elements()))
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersType := make([]types.String, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersState.Elements()))
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersType {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersRes = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersRes, elem.ValueString())
	}
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionIgnoreUsersRes
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel struct {
	Enable      types.Bool  `tfsdk:"enable"`
	IgnoreUsers types.Set   `tfsdk:"ignore_users"`
	MaxAge      types.Int64 `tfsdk:"max_age"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel) GetEnable() types.Bool {
	return m.Enable
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel) GetIgnoreUsers() types.Set {
	return m.IgnoreUsers
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel) GetMaxAge() types.Int64 {
	return m.MaxAge
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel{
		Enable:      types.BoolNull(),
		IgnoreUsers: types.SetNull(types.StringType),
		MaxAge:      types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel {
	if target.Enable.IsUnknown() || target.Enable.IsNull() {
		target.Enable = types.BoolNull()
	}
	if target.IgnoreUsers.IsUnknown() || target.IgnoreUsers.IsNull() {
		target.IgnoreUsers = types.SetNull(types.StringType)
	}
	if target.MaxAge.IsUnknown() || target.MaxAge.IsNull() {
		target.MaxAge = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"enable":       types.BoolType,
		"ignore_users": types.SetType{ElemType: types.StringType},
		"max_age":      types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning *greenplum.QueryKiller,
	state yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel{
		Enable:      flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning.GetEnable()),
		IgnoreUsers: flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsers(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning.GetIgnoreUsers(), state.IgnoreUsers, diags),
		MaxAge:      flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning.GetMaxAge()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState types.Object, diags *diag.Diagnostics) *greenplum.QueryKiller {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunning, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel, diags *diag.Diagnostics) *greenplum.QueryKiller {
	value := &greenplum.QueryKiller{}
	value.SetEnable(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.Enable))
	value.SetIgnoreUsers(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsers(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.IgnoreUsers, diags))
	value.SetMaxAge(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.MaxAge))
	if diags.HasError() {
		return nil
	}
	return value
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsers(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsers []string, listState types.Set, diags *diag.Diagnostics) types.Set {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsers == nil {
		if !listState.IsNull() && !listState.IsUnknown() && len(listState.Elements()) == 0 {
			return listState
		}
		return types.SetNull(types.StringType)
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersValues []attr.Value
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsers {
		val := types.StringValue(elem)
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersValues = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersValues, val)
	}

	value, diag := types.SetValue(types.StringType, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsers(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersState types.Set, diags *diag.Diagnostics) []string {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersState.Elements()) == 0 {
		return []string{}
	}
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersRes := make([]string, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersState.Elements()))
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersType := make([]types.String, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersState.Elements()))
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersType {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersRes = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersRes, elem.ValueString())
	}
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningIgnoreUsersRes
}

type yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel struct {
	Starts types.Set `tfsdk:"starts"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel) GetStarts() types.Set {
	return m.Starts
}

func NewYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel() yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel {
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel{
		Starts: types.SetNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel) yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel {
	if target.Starts.IsUnknown() || target.Starts.IsNull() {
		target.Starts = types.SetNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType)
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"starts": types.SetType{ElemType: yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType},
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes *greenplum.TableSizes,
	state yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel{
		Starts: flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStarts(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes.GetStarts(), state.Starts, diags),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState types.Object, diags *diag.Diagnostics) *greenplum.TableSizes {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizes, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel, diags *diag.Diagnostics) *greenplum.TableSizes {
	value := &greenplum.TableSizes{}
	value.SetStarts(expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStarts(ctx, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState.Starts, diags))
	if diags.HasError() {
		return nil
	}
	return value
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStarts(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStarts []*greenplum.BackgroundActivityStartAt, listState types.Set, diags *diag.Diagnostics) types.Set {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStarts == nil {
		if !listState.IsNull() && !listState.IsUnknown() && len(listState.Elements()) == 0 {
			return listState
		}
		return types.SetNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType)
	}
	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsValues []attr.Value
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStarts {
		val := flattenYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStruct(ctx, elem, diags)
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsValues = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsValues, val)
	}

	value, diag := types.SetValue(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStarts(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsState types.Set, diags *diag.Diagnostics) []*greenplum.BackgroundActivityStartAt {
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsState.Elements()) == 0 {
		return []*greenplum.BackgroundActivityStartAt{}
	}
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsRes := make([]*greenplum.BackgroundActivityStartAt, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsState.Elements()))
	yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsType := make([]yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel, 0, len(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsState.Elements()))
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsType {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsRes = append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsRes, expandYandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModel(ctx, elem, diags))
	}
	return yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesStartsRes
}

type yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model struct {
	GpAddColumnInheritsTableSetting  types.Bool    `tfsdk:"gp_add_column_inherits_table_setting"`
	GpAutostatsMode                  types.String  `tfsdk:"gp_autostats_mode"`
	GpAutostatsOnChangeThreshold     types.Int64   `tfsdk:"gp_autostats_on_change_threshold"`
	GpCachedSegworkersThreshold      types.Int64   `tfsdk:"gp_cached_segworkers_threshold"`
	GpEnableGlobalDeadlockDetector   types.Bool    `tfsdk:"gp_enable_global_deadlock_detector"`
	GpEnableZstdMemoryAccounting     types.Bool    `tfsdk:"gp_enable_zstd_memory_accounting"`
	GpGlobalDeadlockDetectorPeriod   types.Int64   `tfsdk:"gp_global_deadlock_detector_period"`
	GpMaxPlanSize                    types.Int64   `tfsdk:"gp_max_plan_size"`
	GpMaxSlices                      types.Int64   `tfsdk:"gp_max_slices"`
	GpResourceGroupMemoryLimit       types.Float64 `tfsdk:"gp_resource_group_memory_limit"`
	GpVmemProtectSegworkerCacheLimit types.Int64   `tfsdk:"gp_vmem_protect_segworker_cache_limit"`
	GpWorkfileCompression            types.Bool    `tfsdk:"gp_workfile_compression"`
	GpWorkfileLimitFilesPerQuery     types.Int64   `tfsdk:"gp_workfile_limit_files_per_query"`
	GpWorkfileLimitPerQuery          types.Int64   `tfsdk:"gp_workfile_limit_per_query"`
	GpWorkfileLimitPerSegment        types.Int64   `tfsdk:"gp_workfile_limit_per_segment"`
	IdleInTransactionSessionTimeout  types.Int64   `tfsdk:"idle_in_transaction_session_timeout"`
	LockTimeout                      types.Int64   `tfsdk:"lock_timeout"`
	LogStatement                     types.String  `tfsdk:"log_statement"`
	MaxConnections                   types.Int64   `tfsdk:"max_connections"`
	MaxPreparedTransactions          types.Int64   `tfsdk:"max_prepared_transactions"`
	MaxSlotWalKeepSize               types.Int64   `tfsdk:"max_slot_wal_keep_size"`
	MaxStatementMem                  types.Int64   `tfsdk:"max_statement_mem"`
	RunawayDetectorActivationPercent types.Int64   `tfsdk:"runaway_detector_activation_percent"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpAddColumnInheritsTableSetting() types.Bool {
	return m.GpAddColumnInheritsTableSetting
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpAutostatsMode() types.String {
	return m.GpAutostatsMode
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpAutostatsOnChangeThreshold() types.Int64 {
	return m.GpAutostatsOnChangeThreshold
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpCachedSegworkersThreshold() types.Int64 {
	return m.GpCachedSegworkersThreshold
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpEnableGlobalDeadlockDetector() types.Bool {
	return m.GpEnableGlobalDeadlockDetector
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpEnableZstdMemoryAccounting() types.Bool {
	return m.GpEnableZstdMemoryAccounting
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpGlobalDeadlockDetectorPeriod() types.Int64 {
	return m.GpGlobalDeadlockDetectorPeriod
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpMaxPlanSize() types.Int64 {
	return m.GpMaxPlanSize
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpMaxSlices() types.Int64 {
	return m.GpMaxSlices
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpResourceGroupMemoryLimit() types.Float64 {
	return m.GpResourceGroupMemoryLimit
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpVmemProtectSegworkerCacheLimit() types.Int64 {
	return m.GpVmemProtectSegworkerCacheLimit
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpWorkfileCompression() types.Bool {
	return m.GpWorkfileCompression
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpWorkfileLimitFilesPerQuery() types.Int64 {
	return m.GpWorkfileLimitFilesPerQuery
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpWorkfileLimitPerQuery() types.Int64 {
	return m.GpWorkfileLimitPerQuery
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetGpWorkfileLimitPerSegment() types.Int64 {
	return m.GpWorkfileLimitPerSegment
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetIdleInTransactionSessionTimeout() types.Int64 {
	return m.IdleInTransactionSessionTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetLockTimeout() types.Int64 {
	return m.LockTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetLogStatement() types.String {
	return m.LogStatement
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetMaxConnections() types.Int64 {
	return m.MaxConnections
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetMaxPreparedTransactions() types.Int64 {
	return m.MaxPreparedTransactions
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetMaxSlotWalKeepSize() types.Int64 {
	return m.MaxSlotWalKeepSize
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetMaxStatementMem() types.Int64 {
	return m.MaxStatementMem
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) GetRunawayDetectorActivationPercent() types.Int64 {
	return m.RunawayDetectorActivationPercent
}

func NewYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model() yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model {
	return yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model{
		GpAddColumnInheritsTableSetting:  types.BoolNull(),
		GpAutostatsMode:                  types.StringNull(),
		GpAutostatsOnChangeThreshold:     types.Int64Null(),
		GpCachedSegworkersThreshold:      types.Int64Null(),
		GpEnableGlobalDeadlockDetector:   types.BoolNull(),
		GpEnableZstdMemoryAccounting:     types.BoolNull(),
		GpGlobalDeadlockDetectorPeriod:   types.Int64Null(),
		GpMaxPlanSize:                    types.Int64Null(),
		GpMaxSlices:                      types.Int64Null(),
		GpResourceGroupMemoryLimit:       types.Float64Null(),
		GpVmemProtectSegworkerCacheLimit: types.Int64Null(),
		GpWorkfileCompression:            types.BoolNull(),
		GpWorkfileLimitFilesPerQuery:     types.Int64Null(),
		GpWorkfileLimitPerQuery:          types.Int64Null(),
		GpWorkfileLimitPerSegment:        types.Int64Null(),
		IdleInTransactionSessionTimeout:  types.Int64Null(),
		LockTimeout:                      types.Int64Null(),
		LogStatement:                     types.StringNull(),
		MaxConnections:                   types.Int64Null(),
		MaxPreparedTransactions:          types.Int64Null(),
		MaxSlotWalKeepSize:               types.Int64Null(),
		MaxStatementMem:                  types.Int64Null(),
		RunawayDetectorActivationPercent: types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model) yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model {
	if target.GpAddColumnInheritsTableSetting.IsUnknown() || target.GpAddColumnInheritsTableSetting.IsNull() {
		target.GpAddColumnInheritsTableSetting = types.BoolNull()
	}
	if target.GpAutostatsMode.IsUnknown() || target.GpAutostatsMode.IsNull() {
		target.GpAutostatsMode = types.StringNull()
	}
	if target.GpAutostatsOnChangeThreshold.IsUnknown() || target.GpAutostatsOnChangeThreshold.IsNull() {
		target.GpAutostatsOnChangeThreshold = types.Int64Null()
	}
	if target.GpCachedSegworkersThreshold.IsUnknown() || target.GpCachedSegworkersThreshold.IsNull() {
		target.GpCachedSegworkersThreshold = types.Int64Null()
	}
	if target.GpEnableGlobalDeadlockDetector.IsUnknown() || target.GpEnableGlobalDeadlockDetector.IsNull() {
		target.GpEnableGlobalDeadlockDetector = types.BoolNull()
	}
	if target.GpEnableZstdMemoryAccounting.IsUnknown() || target.GpEnableZstdMemoryAccounting.IsNull() {
		target.GpEnableZstdMemoryAccounting = types.BoolNull()
	}
	if target.GpGlobalDeadlockDetectorPeriod.IsUnknown() || target.GpGlobalDeadlockDetectorPeriod.IsNull() {
		target.GpGlobalDeadlockDetectorPeriod = types.Int64Null()
	}
	if target.GpMaxPlanSize.IsUnknown() || target.GpMaxPlanSize.IsNull() {
		target.GpMaxPlanSize = types.Int64Null()
	}
	if target.GpMaxSlices.IsUnknown() || target.GpMaxSlices.IsNull() {
		target.GpMaxSlices = types.Int64Null()
	}
	if target.GpResourceGroupMemoryLimit.IsUnknown() || target.GpResourceGroupMemoryLimit.IsNull() {
		target.GpResourceGroupMemoryLimit = types.Float64Null()
	}
	if target.GpVmemProtectSegworkerCacheLimit.IsUnknown() || target.GpVmemProtectSegworkerCacheLimit.IsNull() {
		target.GpVmemProtectSegworkerCacheLimit = types.Int64Null()
	}
	if target.GpWorkfileCompression.IsUnknown() || target.GpWorkfileCompression.IsNull() {
		target.GpWorkfileCompression = types.BoolNull()
	}
	if target.GpWorkfileLimitFilesPerQuery.IsUnknown() || target.GpWorkfileLimitFilesPerQuery.IsNull() {
		target.GpWorkfileLimitFilesPerQuery = types.Int64Null()
	}
	if target.GpWorkfileLimitPerQuery.IsUnknown() || target.GpWorkfileLimitPerQuery.IsNull() {
		target.GpWorkfileLimitPerQuery = types.Int64Null()
	}
	if target.GpWorkfileLimitPerSegment.IsUnknown() || target.GpWorkfileLimitPerSegment.IsNull() {
		target.GpWorkfileLimitPerSegment = types.Int64Null()
	}
	if target.IdleInTransactionSessionTimeout.IsUnknown() || target.IdleInTransactionSessionTimeout.IsNull() {
		target.IdleInTransactionSessionTimeout = types.Int64Null()
	}
	if target.LockTimeout.IsUnknown() || target.LockTimeout.IsNull() {
		target.LockTimeout = types.Int64Null()
	}
	if target.LogStatement.IsUnknown() || target.LogStatement.IsNull() {
		target.LogStatement = types.StringNull()
	}
	if target.MaxConnections.IsUnknown() || target.MaxConnections.IsNull() {
		target.MaxConnections = types.Int64Null()
	}
	if target.MaxPreparedTransactions.IsUnknown() || target.MaxPreparedTransactions.IsNull() {
		target.MaxPreparedTransactions = types.Int64Null()
	}
	if target.MaxSlotWalKeepSize.IsUnknown() || target.MaxSlotWalKeepSize.IsNull() {
		target.MaxSlotWalKeepSize = types.Int64Null()
	}
	if target.MaxStatementMem.IsUnknown() || target.MaxStatementMem.IsNull() {
		target.MaxStatementMem = types.Int64Null()
	}
	if target.RunawayDetectorActivationPercent.IsUnknown() || target.RunawayDetectorActivationPercent.IsNull() {
		target.RunawayDetectorActivationPercent = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"gp_add_column_inherits_table_setting":  types.BoolType,
		"gp_autostats_mode":                     types.StringType,
		"gp_autostats_on_change_threshold":      types.Int64Type,
		"gp_cached_segworkers_threshold":        types.Int64Type,
		"gp_enable_global_deadlock_detector":    types.BoolType,
		"gp_enable_zstd_memory_accounting":      types.BoolType,
		"gp_global_deadlock_detector_period":    types.Int64Type,
		"gp_max_plan_size":                      types.Int64Type,
		"gp_max_slices":                         types.Int64Type,
		"gp_resource_group_memory_limit":        types.Float64Type,
		"gp_vmem_protect_segworker_cache_limit": types.Int64Type,
		"gp_workfile_compression":               types.BoolType,
		"gp_workfile_limit_files_per_query":     types.Int64Type,
		"gp_workfile_limit_per_query":           types.Int64Type,
		"gp_workfile_limit_per_segment":         types.Int64Type,
		"idle_in_transaction_session_timeout":   types.Int64Type,
		"lock_timeout":                          types.Int64Type,
		"log_statement":                         types.StringType,
		"max_connections":                       types.Int64Type,
		"max_prepared_transactions":             types.Int64Type,
		"max_slot_wal_keep_size":                types.Int64Type,
		"max_statement_mem":                     types.Int64Type,
		"runaway_detector_activation_percent":   types.Int64Type,
	},
}

func expandYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State types.Object, diags *diag.Diagnostics) *greenplum.GreenplumConfig6 {
	if yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6 yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model(ctx, yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model, diags *diag.Diagnostics) *greenplum.GreenplumConfig6 {
	value := &greenplum.GreenplumConfig6{}
	value.SetGpAddColumnInheritsTableSetting(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpAddColumnInheritsTableSetting))
	value.SetGpAutostatsMode(greenplum.GPAutostatsMode(greenplum.GPAutostatsMode_value[yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpAutostatsMode.ValueString()]))
	value.SetGpAutostatsOnChangeThreshold(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpAutostatsOnChangeThreshold))
	value.SetGpCachedSegworkersThreshold(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpCachedSegworkersThreshold))
	value.SetGpEnableGlobalDeadlockDetector(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpEnableGlobalDeadlockDetector))
	value.SetGpEnableZstdMemoryAccounting(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpEnableZstdMemoryAccounting))
	value.SetGpGlobalDeadlockDetectorPeriod(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpGlobalDeadlockDetectorPeriod))
	value.SetGpMaxPlanSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpMaxPlanSize))
	value.SetGpMaxSlices(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpMaxSlices))
	value.SetGpResourceGroupMemoryLimit(expandFloat64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpResourceGroupMemoryLimit))
	value.SetGpVmemProtectSegworkerCacheLimit(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpVmemProtectSegworkerCacheLimit))
	value.SetGpWorkfileCompression(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileCompression))
	value.SetGpWorkfileLimitFilesPerQuery(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileLimitFilesPerQuery))
	value.SetGpWorkfileLimitPerQuery(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileLimitPerQuery))
	value.SetGpWorkfileLimitPerSegment(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileLimitPerSegment))
	value.SetIdleInTransactionSessionTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.IdleInTransactionSessionTimeout))
	value.SetLockTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.LockTimeout))
	value.SetLogStatement(greenplum.LogStatement(greenplum.LogStatement_value[yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.LogStatement.ValueString()]))
	value.SetMaxConnections(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxConnections))
	value.SetMaxPreparedTransactions(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxPreparedTransactions))
	value.SetMaxSlotWalKeepSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxSlotWalKeepSize))
	value.SetMaxStatementMem(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxStatementMem))
	value.SetRunawayDetectorActivationPercent(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.RunawayDetectorActivationPercent))
	if diags.HasError() {
		return nil
	}
	return value
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfigSet6 *greenplum.GreenplumConfigSet6,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfigSet6 == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelType.AttrTypes)
	}
	yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6 := yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfigSet6.UserConfig
	if yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6 == nil {
		value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelType.AttrTypes, NewYandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model())
		diags.Append(diag...)
		return value
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6ModelType.AttrTypes,
		yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model{
			GpAddColumnInheritsTableSetting:  flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpAddColumnInheritsTableSetting()),
			GpAutostatsMode:                  flattenEnum(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpAutostatsMode()),
			GpAutostatsOnChangeThreshold:     flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpAutostatsOnChangeThreshold()),
			GpCachedSegworkersThreshold:      flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpCachedSegworkersThreshold()),
			GpEnableGlobalDeadlockDetector:   flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpEnableGlobalDeadlockDetector()),
			GpEnableZstdMemoryAccounting:     flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpEnableZstdMemoryAccounting()),
			GpGlobalDeadlockDetectorPeriod:   flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpGlobalDeadlockDetectorPeriod()),
			GpMaxPlanSize:                    flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpMaxPlanSize()),
			GpMaxSlices:                      flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpMaxSlices()),
			GpResourceGroupMemoryLimit:       flattenFloat64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpResourceGroupMemoryLimit()),
			GpVmemProtectSegworkerCacheLimit: flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpVmemProtectSegworkerCacheLimit()),
			GpWorkfileCompression:            flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpWorkfileCompression()),
			GpWorkfileLimitFilesPerQuery:     flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpWorkfileLimitFilesPerQuery()),
			GpWorkfileLimitPerQuery:          flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpWorkfileLimitPerQuery()),
			GpWorkfileLimitPerSegment:        flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetGpWorkfileLimitPerSegment()),
			IdleInTransactionSessionTimeout:  flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetIdleInTransactionSessionTimeout()),
			LockTimeout:                      flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetLockTimeout()),
			LogStatement:                     flattenEnum(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetLogStatement()),
			MaxConnections:                   flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetMaxConnections()),
			MaxPreparedTransactions:          flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetMaxPreparedTransactions()),
			MaxSlotWalKeepSize:               flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetMaxSlotWalKeepSize()),
			MaxStatementMem:                  flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetMaxStatementMem()),
			RunawayDetectorActivationPercent: flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6.GetRunawayDetectorActivationPercent()),
		})
	diags.Append(diag...)
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigPoolModel struct {
	ClientIdleTimeout        types.Int64  `tfsdk:"client_idle_timeout"`
	IdleInTransactionTimeout types.Int64  `tfsdk:"idle_in_transaction_timeout"`
	Mode                     types.String `tfsdk:"mode"`
	Size                     types.Int64  `tfsdk:"size"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolModel) GetClientIdleTimeout() types.Int64 {
	return m.ClientIdleTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolModel) GetIdleInTransactionTimeout() types.Int64 {
	return m.IdleInTransactionTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolModel) GetMode() types.String {
	return m.Mode
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolModel) GetSize() types.Int64 {
	return m.Size
}

func NewYandexMdbGreenplumClusterV2ClusterConfigPoolModel() yandexMdbGreenplumClusterV2ClusterConfigPoolModel {
	return yandexMdbGreenplumClusterV2ClusterConfigPoolModel{
		ClientIdleTimeout:        types.Int64Null(),
		IdleInTransactionTimeout: types.Int64Null(),
		Mode:                     types.StringNull(),
		Size:                     types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigPoolModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigPoolModel) yandexMdbGreenplumClusterV2ClusterConfigPoolModel {
	if target.ClientIdleTimeout.IsUnknown() || target.ClientIdleTimeout.IsNull() {
		target.ClientIdleTimeout = types.Int64Null()
	}
	if target.IdleInTransactionTimeout.IsUnknown() || target.IdleInTransactionTimeout.IsNull() {
		target.IdleInTransactionTimeout = types.Int64Null()
	}
	if target.Mode.IsUnknown() || target.Mode.IsNull() {
		target.Mode = types.StringNull()
	}
	if target.Size.IsUnknown() || target.Size.IsNull() {
		target.Size = types.Int64Null()
	}
	if target.ClientIdleTimeout.IsUnknown() || target.ClientIdleTimeout.IsNull() {
		target.ClientIdleTimeout = types.Int64Null()
	}
	if target.IdleInTransactionTimeout.IsUnknown() || target.IdleInTransactionTimeout.IsNull() {
		target.IdleInTransactionTimeout = types.Int64Null()
	}
	if target.Mode.IsUnknown() || target.Mode.IsNull() {
		target.Mode = types.StringNull()
	}
	if target.Size.IsUnknown() || target.Size.IsNull() {
		target.Size = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigPoolModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"client_idle_timeout":         types.Int64Type,
		"idle_in_transaction_timeout": types.Int64Type,
		"mode":                        types.StringType,
		"size":                        types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigPool(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigPool *greenplum.ConnectionPoolerConfigSet,
	state yandexMdbGreenplumClusterV2ClusterConfigPoolModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigPool == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigPoolModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigPoolModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigPoolModel{
		ClientIdleTimeout:        flattenYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPool.GetUserConfig(), diags).ClientIdleTimeout,
		IdleInTransactionTimeout: flattenYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPool.GetUserConfig(), diags).IdleInTransactionTimeout,
		Mode:                     flattenYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPool.GetUserConfig(), diags).Mode,
		Size:                     flattenYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPool.GetUserConfig(), diags).Size,
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigPool(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPoolState types.Object, diags *diag.Diagnostics) *greenplum.ConnectionPoolerConfigSet {
	if yandexMdbGreenplumClusterV2ClusterConfigPoolState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigPoolState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigPool yandexMdbGreenplumClusterV2ClusterConfigPoolModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigPoolState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPool, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigPoolModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigPool, diags)
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPool_create(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPoolState types.Object, diags *diag.Diagnostics) *greenplum.ConnectionPoolerConfig {
	if yandexMdbGreenplumClusterV2ClusterConfigPoolState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigPoolState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigPool yandexMdbGreenplumClusterV2ClusterConfigPoolModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigPoolState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPool, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigPoolModel_create(ctx, yandexMdbGreenplumClusterV2ClusterConfigPool, diags)
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPool_update(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPoolState types.Object, diags *diag.Diagnostics) *greenplum.ConnectionPoolerConfig {
	if yandexMdbGreenplumClusterV2ClusterConfigPoolState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigPoolState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigPool yandexMdbGreenplumClusterV2ClusterConfigPoolModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigPoolState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPool, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigPoolModel_update(ctx, yandexMdbGreenplumClusterV2ClusterConfigPool, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigPoolModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPoolState yandexMdbGreenplumClusterV2ClusterConfigPoolModel, diags *diag.Diagnostics) *greenplum.ConnectionPoolerConfigSet {
	value := &greenplum.ConnectionPoolerConfigSet{}
	value.SetUserConfig(expandYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel{
		ClientIdleTimeout:        yandexMdbGreenplumClusterV2ClusterConfigPoolState.ClientIdleTimeout,
		IdleInTransactionTimeout: yandexMdbGreenplumClusterV2ClusterConfigPoolState.IdleInTransactionTimeout,
		Mode:                     yandexMdbGreenplumClusterV2ClusterConfigPoolState.Mode,
		Size:                     yandexMdbGreenplumClusterV2ClusterConfigPoolState.Size,
	}, diags))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPoolModel_create(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPoolState yandexMdbGreenplumClusterV2ClusterConfigPoolModel, diags *diag.Diagnostics) *greenplum.ConnectionPoolerConfig {
	value := &greenplum.ConnectionPoolerConfig{}
	value.SetClientIdleTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolState.ClientIdleTimeout))
	value.SetIdleInTransactionTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolState.IdleInTransactionTimeout))
	value.SetMode(greenplum.ConnectionPoolerConfig_PoolMode(greenplum.ConnectionPoolerConfig_PoolMode_value[yandexMdbGreenplumClusterV2ClusterConfigPoolState.Mode.ValueString()]))
	value.SetSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolState.Size))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPoolModel_update(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPoolState yandexMdbGreenplumClusterV2ClusterConfigPoolModel, diags *diag.Diagnostics) *greenplum.ConnectionPoolerConfig {
	value := &greenplum.ConnectionPoolerConfig{}
	value.SetClientIdleTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolState.ClientIdleTimeout))
	value.SetIdleInTransactionTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolState.IdleInTransactionTimeout))
	value.SetMode(greenplum.ConnectionPoolerConfig_PoolMode(greenplum.ConnectionPoolerConfig_PoolMode_value[yandexMdbGreenplumClusterV2ClusterConfigPoolState.Mode.ValueString()]))
	value.SetSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolState.Size))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel struct {
	ClientIdleTimeout        types.Int64  `tfsdk:"client_idle_timeout"`
	IdleInTransactionTimeout types.Int64  `tfsdk:"idle_in_transaction_timeout"`
	Mode                     types.String `tfsdk:"mode"`
	Size                     types.Int64  `tfsdk:"size"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel) GetClientIdleTimeout() types.Int64 {
	return m.ClientIdleTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel) GetIdleInTransactionTimeout() types.Int64 {
	return m.IdleInTransactionTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel) GetMode() types.String {
	return m.Mode
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel) GetSize() types.Int64 {
	return m.Size
}

func NewYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel() yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel {
	return yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel{
		ClientIdleTimeout:        types.Int64Null(),
		IdleInTransactionTimeout: types.Int64Null(),
		Mode:                     types.StringNull(),
		Size:                     types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel) yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel {
	if target.ClientIdleTimeout.IsUnknown() || target.ClientIdleTimeout.IsNull() {
		target.ClientIdleTimeout = types.Int64Null()
	}
	if target.IdleInTransactionTimeout.IsUnknown() || target.IdleInTransactionTimeout.IsNull() {
		target.IdleInTransactionTimeout = types.Int64Null()
	}
	if target.Mode.IsUnknown() || target.Mode.IsNull() {
		target.Mode = types.StringNull()
	}
	if target.Size.IsUnknown() || target.Size.IsNull() {
		target.Size = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"client_idle_timeout":         types.Int64Type,
		"idle_in_transaction_timeout": types.Int64Type,
		"mode":                        types.StringType,
		"size":                        types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig *greenplum.ConnectionPoolerConfig,
	diags *diag.Diagnostics) yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel {
	if yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig == nil {
		return NewYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel()
	}
	return yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel{
		ClientIdleTimeout:        flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig.GetClientIdleTimeout()),
		IdleInTransactionTimeout: flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig.GetIdleInTransactionTimeout()),
		Mode:                     flattenEnum(yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig.GetMode()),
		Size:                     flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfig.GetSize()),
	}
}

func expandYandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigState yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigModel, diags *diag.Diagnostics) *greenplum.ConnectionPoolerConfig {
	value := &greenplum.ConnectionPoolerConfig{}
	value.SetClientIdleTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigState.ClientIdleTimeout))
	value.SetIdleInTransactionTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigState.IdleInTransactionTimeout))
	value.SetMode(greenplum.ConnectionPoolerConfig_PoolMode(greenplum.ConnectionPoolerConfig_PoolMode_value[yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigState.Mode.ValueString()]))
	value.SetSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPoolUserConfigState.Size))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel struct {
	ConnectionTimeout          types.Int64 `tfsdk:"connection_timeout"`
	MaxThreads                 types.Int64 `tfsdk:"max_threads"`
	PoolAllowCoreThreadTimeout types.Bool  `tfsdk:"pool_allow_core_thread_timeout"`
	PoolCoreSize               types.Int64 `tfsdk:"pool_core_size"`
	PoolMaxSize                types.Int64 `tfsdk:"pool_max_size"`
	PoolQueueCapacity          types.Int64 `tfsdk:"pool_queue_capacity"`
	UploadTimeout              types.Int64 `tfsdk:"upload_timeout"`
	Xms                        types.Int64 `tfsdk:"xms"`
	Xmx                        types.Int64 `tfsdk:"xmx"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetConnectionTimeout() types.Int64 {
	return m.ConnectionTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetMaxThreads() types.Int64 {
	return m.MaxThreads
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetPoolAllowCoreThreadTimeout() types.Bool {
	return m.PoolAllowCoreThreadTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetPoolCoreSize() types.Int64 {
	return m.PoolCoreSize
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetPoolMaxSize() types.Int64 {
	return m.PoolMaxSize
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetPoolQueueCapacity() types.Int64 {
	return m.PoolQueueCapacity
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetUploadTimeout() types.Int64 {
	return m.UploadTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetXms() types.Int64 {
	return m.Xms
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) GetXmx() types.Int64 {
	return m.Xmx
}

func NewYandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel() yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel {
	return yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel{
		ConnectionTimeout:          types.Int64Null(),
		MaxThreads:                 types.Int64Null(),
		PoolAllowCoreThreadTimeout: types.BoolNull(),
		PoolCoreSize:               types.Int64Null(),
		PoolMaxSize:                types.Int64Null(),
		PoolQueueCapacity:          types.Int64Null(),
		UploadTimeout:              types.Int64Null(),
		Xms:                        types.Int64Null(),
		Xmx:                        types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel) yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel {
	if target.ConnectionTimeout.IsUnknown() || target.ConnectionTimeout.IsNull() {
		target.ConnectionTimeout = types.Int64Null()
	}
	if target.MaxThreads.IsUnknown() || target.MaxThreads.IsNull() {
		target.MaxThreads = types.Int64Null()
	}
	if target.PoolAllowCoreThreadTimeout.IsUnknown() || target.PoolAllowCoreThreadTimeout.IsNull() {
		target.PoolAllowCoreThreadTimeout = types.BoolNull()
	}
	if target.PoolCoreSize.IsUnknown() || target.PoolCoreSize.IsNull() {
		target.PoolCoreSize = types.Int64Null()
	}
	if target.PoolMaxSize.IsUnknown() || target.PoolMaxSize.IsNull() {
		target.PoolMaxSize = types.Int64Null()
	}
	if target.PoolQueueCapacity.IsUnknown() || target.PoolQueueCapacity.IsNull() {
		target.PoolQueueCapacity = types.Int64Null()
	}
	if target.UploadTimeout.IsUnknown() || target.UploadTimeout.IsNull() {
		target.UploadTimeout = types.Int64Null()
	}
	if target.ConnectionTimeout.IsUnknown() || target.ConnectionTimeout.IsNull() {
		target.ConnectionTimeout = types.Int64Null()
	}
	if target.MaxThreads.IsUnknown() || target.MaxThreads.IsNull() {
		target.MaxThreads = types.Int64Null()
	}
	if target.PoolAllowCoreThreadTimeout.IsUnknown() || target.PoolAllowCoreThreadTimeout.IsNull() {
		target.PoolAllowCoreThreadTimeout = types.BoolNull()
	}
	if target.PoolCoreSize.IsUnknown() || target.PoolCoreSize.IsNull() {
		target.PoolCoreSize = types.Int64Null()
	}
	if target.PoolMaxSize.IsUnknown() || target.PoolMaxSize.IsNull() {
		target.PoolMaxSize = types.Int64Null()
	}
	if target.PoolQueueCapacity.IsUnknown() || target.PoolQueueCapacity.IsNull() {
		target.PoolQueueCapacity = types.Int64Null()
	}
	if target.UploadTimeout.IsUnknown() || target.UploadTimeout.IsNull() {
		target.UploadTimeout = types.Int64Null()
	}
	if target.Xms.IsUnknown() || target.Xms.IsNull() {
		target.Xms = types.Int64Null()
	}
	if target.Xmx.IsUnknown() || target.Xmx.IsNull() {
		target.Xmx = types.Int64Null()
	}
	if target.Xms.IsUnknown() || target.Xms.IsNull() {
		target.Xms = types.Int64Null()
	}
	if target.Xmx.IsUnknown() || target.Xmx.IsNull() {
		target.Xmx = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"connection_timeout":             types.Int64Type,
		"max_threads":                    types.Int64Type,
		"pool_allow_core_thread_timeout": types.BoolType,
		"pool_core_size":                 types.Int64Type,
		"pool_max_size":                  types.Int64Type,
		"pool_queue_capacity":            types.Int64Type,
		"upload_timeout":                 types.Int64Type,
		"xms":                            types.Int64Type,
		"xmx":                            types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfig(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigPxfConfig *greenplum.PXFConfigSet,
	state yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ClusterConfigPxfConfig == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModelType.AttrTypes, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel{
		ConnectionTimeout:          flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).ConnectionTimeout,
		MaxThreads:                 flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).MaxThreads,
		PoolAllowCoreThreadTimeout: flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).PoolAllowCoreThreadTimeout,
		PoolCoreSize:               flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).PoolCoreSize,
		PoolMaxSize:                flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).PoolMaxSize,
		PoolQueueCapacity:          flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).PoolQueueCapacity,
		UploadTimeout:              flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).UploadTimeout,
		Xms:                        flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).Xms,
		Xmx:                        flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig.GetUserConfig(), diags).Xmx,
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfig(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState types.Object, diags *diag.Diagnostics) *greenplum.PXFConfigSet {
	if yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigPxfConfig yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPxfConfig, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig, diags)
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfig_create(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState types.Object, diags *diag.Diagnostics) *greenplum.PXFConfig {
	if yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigPxfConfig yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPxfConfig, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel_create(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig, diags)
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfig_update(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState types.Object, diags *diag.Diagnostics) *greenplum.PXFConfig {
	if yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.IsNull() || yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ClusterConfigPxfConfig yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel
	diags.Append(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPxfConfig, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel_update(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfig, diags)
}

func expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel, diags *diag.Diagnostics) *greenplum.PXFConfigSet {
	value := &greenplum.PXFConfigSet{}
	value.SetUserConfig(expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel(ctx, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel{
		ConnectionTimeout:          yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.ConnectionTimeout,
		MaxThreads:                 yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.MaxThreads,
		PoolAllowCoreThreadTimeout: yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolAllowCoreThreadTimeout,
		PoolCoreSize:               yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolCoreSize,
		PoolMaxSize:                yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolMaxSize,
		PoolQueueCapacity:          yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolQueueCapacity,
		UploadTimeout:              yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.UploadTimeout,
		Xms:                        yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xms,
		Xmx:                        yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xmx,
	}, diags))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel_create(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel, diags *diag.Diagnostics) *greenplum.PXFConfig {
	value := &greenplum.PXFConfig{}
	value.SetConnectionTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.ConnectionTimeout))
	value.SetMaxThreads(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.MaxThreads))
	value.SetPoolAllowCoreThreadTimeout(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolAllowCoreThreadTimeout))
	value.SetPoolCoreSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolCoreSize))
	value.SetPoolMaxSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolMaxSize))
	value.SetPoolQueueCapacity(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolQueueCapacity))
	value.SetUploadTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.UploadTimeout))
	value.SetXms(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xms))
	value.SetXmx(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xmx))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel_update(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel, diags *diag.Diagnostics) *greenplum.PXFConfig {
	value := &greenplum.PXFConfig{}
	value.SetConnectionTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.ConnectionTimeout))
	value.SetMaxThreads(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.MaxThreads))
	value.SetPoolAllowCoreThreadTimeout(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolAllowCoreThreadTimeout))
	value.SetPoolCoreSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolCoreSize))
	value.SetPoolMaxSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolMaxSize))
	value.SetPoolQueueCapacity(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolQueueCapacity))
	value.SetUploadTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.UploadTimeout))
	value.SetXms(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xms))
	value.SetXmx(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xmx))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel struct {
	ConnectionTimeout          types.Int64 `tfsdk:"connection_timeout"`
	MaxThreads                 types.Int64 `tfsdk:"max_threads"`
	PoolAllowCoreThreadTimeout types.Bool  `tfsdk:"pool_allow_core_thread_timeout"`
	PoolCoreSize               types.Int64 `tfsdk:"pool_core_size"`
	PoolMaxSize                types.Int64 `tfsdk:"pool_max_size"`
	PoolQueueCapacity          types.Int64 `tfsdk:"pool_queue_capacity"`
	UploadTimeout              types.Int64 `tfsdk:"upload_timeout"`
	Xms                        types.Int64 `tfsdk:"xms"`
	Xmx                        types.Int64 `tfsdk:"xmx"`
}

func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetConnectionTimeout() types.Int64 {
	return m.ConnectionTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetMaxThreads() types.Int64 {
	return m.MaxThreads
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetPoolAllowCoreThreadTimeout() types.Bool {
	return m.PoolAllowCoreThreadTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetPoolCoreSize() types.Int64 {
	return m.PoolCoreSize
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetPoolMaxSize() types.Int64 {
	return m.PoolMaxSize
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetPoolQueueCapacity() types.Int64 {
	return m.PoolQueueCapacity
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetUploadTimeout() types.Int64 {
	return m.UploadTimeout
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetXms() types.Int64 {
	return m.Xms
}
func (m *yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) GetXmx() types.Int64 {
	return m.Xmx
}

func NewYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel() yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel {
	return yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel{
		ConnectionTimeout:          types.Int64Null(),
		MaxThreads:                 types.Int64Null(),
		PoolAllowCoreThreadTimeout: types.BoolNull(),
		PoolCoreSize:               types.Int64Null(),
		PoolMaxSize:                types.Int64Null(),
		PoolQueueCapacity:          types.Int64Null(),
		UploadTimeout:              types.Int64Null(),
		Xms:                        types.Int64Null(),
		Xmx:                        types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModelFillUnknown(target yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel) yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel {
	if target.ConnectionTimeout.IsUnknown() || target.ConnectionTimeout.IsNull() {
		target.ConnectionTimeout = types.Int64Null()
	}
	if target.MaxThreads.IsUnknown() || target.MaxThreads.IsNull() {
		target.MaxThreads = types.Int64Null()
	}
	if target.PoolAllowCoreThreadTimeout.IsUnknown() || target.PoolAllowCoreThreadTimeout.IsNull() {
		target.PoolAllowCoreThreadTimeout = types.BoolNull()
	}
	if target.PoolCoreSize.IsUnknown() || target.PoolCoreSize.IsNull() {
		target.PoolCoreSize = types.Int64Null()
	}
	if target.PoolMaxSize.IsUnknown() || target.PoolMaxSize.IsNull() {
		target.PoolMaxSize = types.Int64Null()
	}
	if target.PoolQueueCapacity.IsUnknown() || target.PoolQueueCapacity.IsNull() {
		target.PoolQueueCapacity = types.Int64Null()
	}
	if target.UploadTimeout.IsUnknown() || target.UploadTimeout.IsNull() {
		target.UploadTimeout = types.Int64Null()
	}
	if target.Xms.IsUnknown() || target.Xms.IsNull() {
		target.Xms = types.Int64Null()
	}
	if target.Xmx.IsUnknown() || target.Xmx.IsNull() {
		target.Xmx = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"connection_timeout":             types.Int64Type,
		"max_threads":                    types.Int64Type,
		"pool_allow_core_thread_timeout": types.BoolType,
		"pool_core_size":                 types.Int64Type,
		"pool_max_size":                  types.Int64Type,
		"pool_queue_capacity":            types.Int64Type,
		"upload_timeout":                 types.Int64Type,
		"xms":                            types.Int64Type,
		"xmx":                            types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig(ctx context.Context,
	yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig *greenplum.PXFConfig,
	diags *diag.Diagnostics) yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel {
	if yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig == nil {
		return NewYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel()
	}
	return yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel{
		ConnectionTimeout:          flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetConnectionTimeout()),
		MaxThreads:                 flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetMaxThreads()),
		PoolAllowCoreThreadTimeout: flattenBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetPoolAllowCoreThreadTimeout()),
		PoolCoreSize:               flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetPoolCoreSize()),
		PoolMaxSize:                flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetPoolMaxSize()),
		PoolQueueCapacity:          flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetPoolQueueCapacity()),
		UploadTimeout:              flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetUploadTimeout()),
		Xms:                        flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetXms()),
		Xmx:                        flattenInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfig.GetXmx()),
	}
}

func expandYandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel(ctx context.Context, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigModel, diags *diag.Diagnostics) *greenplum.PXFConfig {
	value := &greenplum.PXFConfig{}
	value.SetConnectionTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.ConnectionTimeout))
	value.SetMaxThreads(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.MaxThreads))
	value.SetPoolAllowCoreThreadTimeout(expandBoolWrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.PoolAllowCoreThreadTimeout))
	value.SetPoolCoreSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.PoolCoreSize))
	value.SetPoolMaxSize(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.PoolMaxSize))
	value.SetPoolQueueCapacity(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.PoolQueueCapacity))
	value.SetUploadTimeout(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.UploadTimeout))
	value.SetXms(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.Xms))
	value.SetXmx(expandInt64Wrapper(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigUserConfigState.Xmx))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ConfigModel struct {
	Access                 types.Object `tfsdk:"access"`
	AssignPublicIp         types.Bool   `tfsdk:"assign_public_ip"`
	BackupRetainPeriodDays types.Int64  `tfsdk:"backup_retain_period_days"`
	BackupWindowStart      types.String `tfsdk:"backup_window_start"`
	SubnetId               types.String `tfsdk:"subnet_id"`
	Version                types.String `tfsdk:"version"`
	ZoneId                 types.String `tfsdk:"zone_id"`
}

func (m *yandexMdbGreenplumClusterV2ConfigModel) GetAccess() types.Object {
	return m.Access
}
func (m *yandexMdbGreenplumClusterV2ConfigModel) GetAssignPublicIp() types.Bool {
	return m.AssignPublicIp
}
func (m *yandexMdbGreenplumClusterV2ConfigModel) GetBackupRetainPeriodDays() types.Int64 {
	return m.BackupRetainPeriodDays
}
func (m *yandexMdbGreenplumClusterV2ConfigModel) GetBackupWindowStart() types.String {
	return m.BackupWindowStart
}
func (m *yandexMdbGreenplumClusterV2ConfigModel) GetSubnetId() types.String {
	return m.SubnetId
}
func (m *yandexMdbGreenplumClusterV2ConfigModel) GetVersion() types.String {
	return m.Version
}
func (m *yandexMdbGreenplumClusterV2ConfigModel) GetZoneId() types.String {
	return m.ZoneId
}

func NewYandexMdbGreenplumClusterV2ConfigModel() yandexMdbGreenplumClusterV2ConfigModel {
	return yandexMdbGreenplumClusterV2ConfigModel{
		Access:                 types.ObjectNull(yandexMdbGreenplumClusterV2ConfigAccessModelType.AttrTypes),
		AssignPublicIp:         types.BoolNull(),
		BackupRetainPeriodDays: types.Int64Null(),
		BackupWindowStart:      types.StringNull(),
		SubnetId:               types.StringNull(),
		Version:                types.StringNull(),
		ZoneId:                 types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2ConfigModelFillUnknown(target yandexMdbGreenplumClusterV2ConfigModel) yandexMdbGreenplumClusterV2ConfigModel {
	if target.Access.IsUnknown() || target.Access.IsNull() {
		target.Access = types.ObjectNull(yandexMdbGreenplumClusterV2ConfigAccessModelType.AttrTypes)
	}
	if target.AssignPublicIp.IsUnknown() || target.AssignPublicIp.IsNull() {
		target.AssignPublicIp = types.BoolNull()
	}
	if target.BackupRetainPeriodDays.IsUnknown() || target.BackupRetainPeriodDays.IsNull() {
		target.BackupRetainPeriodDays = types.Int64Null()
	}
	if target.BackupWindowStart.IsUnknown() || target.BackupWindowStart.IsNull() {
		target.BackupWindowStart = types.StringNull()
	}
	if target.SubnetId.IsUnknown() || target.SubnetId.IsNull() {
		target.SubnetId = types.StringNull()
	}
	if target.Version.IsUnknown() || target.Version.IsNull() {
		target.Version = types.StringNull()
	}
	if target.ZoneId.IsUnknown() || target.ZoneId.IsNull() {
		target.ZoneId = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2ConfigModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"access":                    yandexMdbGreenplumClusterV2ConfigAccessModelType,
		"assign_public_ip":          types.BoolType,
		"backup_retain_period_days": types.Int64Type,
		"backup_window_start":       types.StringType,
		"subnet_id":                 types.StringType,
		"version":                   types.StringType,
		"zone_id":                   types.StringType,
	},
}

func flattenYandexMdbGreenplumClusterV2Config(ctx context.Context,
	yandexMdbGreenplumClusterV2Config *greenplum.GreenplumConfig,
	state yandexMdbGreenplumClusterV2ConfigModel,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2Config == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ConfigModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ConfigModelType.AttrTypes, yandexMdbGreenplumClusterV2ConfigModel{
		Access:                 flattenYandexMdbGreenplumClusterV2ConfigAccess(ctx, yandexMdbGreenplumClusterV2Config.GetAccess(), diags),
		AssignPublicIp:         types.BoolValue(yandexMdbGreenplumClusterV2Config.GetAssignPublicIp()),
		BackupRetainPeriodDays: flattenInt64Wrapper(yandexMdbGreenplumClusterV2Config.GetBackupRetainPeriodDays()),
		BackupWindowStart:      types.StringValue(converter.GetTimeOfDay(yandexMdbGreenplumClusterV2Config.GetBackupWindowStart(), state.BackupWindowStart.ValueString(), diags)),
		SubnetId:               types.StringValue(yandexMdbGreenplumClusterV2Config.GetSubnetId()),
		Version:                types.StringValue(yandexMdbGreenplumClusterV2Config.GetVersion()),
		ZoneId:                 types.StringValue(yandexMdbGreenplumClusterV2Config.GetZoneId()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2Config(ctx context.Context, yandexMdbGreenplumClusterV2ConfigState types.Object, diags *diag.Diagnostics) *greenplum.GreenplumConfig {
	if yandexMdbGreenplumClusterV2ConfigState.IsNull() || yandexMdbGreenplumClusterV2ConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2Config yandexMdbGreenplumClusterV2ConfigModel
	diags.Append(yandexMdbGreenplumClusterV2ConfigState.As(ctx, &yandexMdbGreenplumClusterV2Config, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ConfigModel(ctx, yandexMdbGreenplumClusterV2Config, diags)
}

func expandYandexMdbGreenplumClusterV2ConfigModel(ctx context.Context, yandexMdbGreenplumClusterV2ConfigState yandexMdbGreenplumClusterV2ConfigModel, diags *diag.Diagnostics) *greenplum.GreenplumConfig {
	value := &greenplum.GreenplumConfig{}
	value.SetAccess(expandYandexMdbGreenplumClusterV2ConfigAccess(ctx, yandexMdbGreenplumClusterV2ConfigState.Access, diags))
	value.SetAssignPublicIp(yandexMdbGreenplumClusterV2ConfigState.AssignPublicIp.ValueBool())
	value.SetBackupRetainPeriodDays(expandInt64Wrapper(yandexMdbGreenplumClusterV2ConfigState.BackupRetainPeriodDays))
	value.SetBackupWindowStart(converter.ParseTimeOfDay(yandexMdbGreenplumClusterV2ConfigState.BackupWindowStart.ValueString(), diags))
	value.SetSubnetId(yandexMdbGreenplumClusterV2ConfigState.SubnetId.ValueString())
	value.SetVersion(yandexMdbGreenplumClusterV2ConfigState.Version.ValueString())
	value.SetZoneId(yandexMdbGreenplumClusterV2ConfigState.ZoneId.ValueString())
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2ConfigAccessModel struct {
	DataLens     types.Bool `tfsdk:"data_lens"`
	DataTransfer types.Bool `tfsdk:"data_transfer"`
	WebSql       types.Bool `tfsdk:"web_sql"`
	YandexQuery  types.Bool `tfsdk:"yandex_query"`
}

func (m *yandexMdbGreenplumClusterV2ConfigAccessModel) GetDataLens() types.Bool {
	return m.DataLens
}
func (m *yandexMdbGreenplumClusterV2ConfigAccessModel) GetDataTransfer() types.Bool {
	return m.DataTransfer
}
func (m *yandexMdbGreenplumClusterV2ConfigAccessModel) GetWebSql() types.Bool {
	return m.WebSql
}
func (m *yandexMdbGreenplumClusterV2ConfigAccessModel) GetYandexQuery() types.Bool {
	return m.YandexQuery
}

func NewYandexMdbGreenplumClusterV2ConfigAccessModel() yandexMdbGreenplumClusterV2ConfigAccessModel {
	return yandexMdbGreenplumClusterV2ConfigAccessModel{
		DataLens:     types.BoolNull(),
		DataTransfer: types.BoolNull(),
		WebSql:       types.BoolNull(),
		YandexQuery:  types.BoolNull(),
	}
}

func yandexMdbGreenplumClusterV2ConfigAccessModelFillUnknown(target yandexMdbGreenplumClusterV2ConfigAccessModel) yandexMdbGreenplumClusterV2ConfigAccessModel {
	if target.DataLens.IsUnknown() || target.DataLens.IsNull() {
		target.DataLens = types.BoolNull()
	}
	if target.DataTransfer.IsUnknown() || target.DataTransfer.IsNull() {
		target.DataTransfer = types.BoolNull()
	}
	if target.WebSql.IsUnknown() || target.WebSql.IsNull() {
		target.WebSql = types.BoolNull()
	}
	if target.YandexQuery.IsUnknown() || target.YandexQuery.IsNull() {
		target.YandexQuery = types.BoolNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2ConfigAccessModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"data_lens":     types.BoolType,
		"data_transfer": types.BoolType,
		"web_sql":       types.BoolType,
		"yandex_query":  types.BoolType,
	},
}

func flattenYandexMdbGreenplumClusterV2ConfigAccess(ctx context.Context,
	yandexMdbGreenplumClusterV2ConfigAccess *greenplum.Access,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2ConfigAccess == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2ConfigAccessModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ConfigAccessModelType.AttrTypes, yandexMdbGreenplumClusterV2ConfigAccessModel{
		DataLens:     types.BoolValue(yandexMdbGreenplumClusterV2ConfigAccess.GetDataLens()),
		DataTransfer: types.BoolValue(yandexMdbGreenplumClusterV2ConfigAccess.GetDataTransfer()),
		WebSql:       types.BoolValue(yandexMdbGreenplumClusterV2ConfigAccess.GetWebSql()),
		YandexQuery:  types.BoolValue(yandexMdbGreenplumClusterV2ConfigAccess.GetYandexQuery()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2ConfigAccess(ctx context.Context, yandexMdbGreenplumClusterV2ConfigAccessState types.Object, diags *diag.Diagnostics) *greenplum.Access {
	if yandexMdbGreenplumClusterV2ConfigAccessState.IsNull() || yandexMdbGreenplumClusterV2ConfigAccessState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2ConfigAccess yandexMdbGreenplumClusterV2ConfigAccessModel
	diags.Append(yandexMdbGreenplumClusterV2ConfigAccessState.As(ctx, &yandexMdbGreenplumClusterV2ConfigAccess, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2ConfigAccessModel(ctx, yandexMdbGreenplumClusterV2ConfigAccess, diags)
}

func expandYandexMdbGreenplumClusterV2ConfigAccessModel(ctx context.Context, yandexMdbGreenplumClusterV2ConfigAccessState yandexMdbGreenplumClusterV2ConfigAccessModel, diags *diag.Diagnostics) *greenplum.Access {
	value := &greenplum.Access{}
	value.SetDataLens(yandexMdbGreenplumClusterV2ConfigAccessState.DataLens.ValueBool())
	value.SetDataTransfer(yandexMdbGreenplumClusterV2ConfigAccessState.DataTransfer.ValueBool())
	value.SetWebSql(yandexMdbGreenplumClusterV2ConfigAccessState.WebSql.ValueBool())
	value.SetYandexQuery(yandexMdbGreenplumClusterV2ConfigAccessState.YandexQuery.ValueBool())
	if diags.HasError() {
		return nil
	}
	return value
}

func flattenYandexMdbGreenplumClusterV2HostGroupIds(ctx context.Context, yandexMdbGreenplumClusterV2HostGroupIds []string, setState types.Set, diags *diag.Diagnostics) types.Set {
	if yandexMdbGreenplumClusterV2HostGroupIds == nil {
		if !setState.IsNull() && !setState.IsUnknown() && len(setState.Elements()) == 0 {
			return setState
		}
		return types.SetNull(types.StringType)
	}
	var yandexMdbGreenplumClusterV2HostGroupIdsValues []attr.Value
	for _, elem := range yandexMdbGreenplumClusterV2HostGroupIds {
		val := types.StringValue(elem)
		yandexMdbGreenplumClusterV2HostGroupIdsValues = append(yandexMdbGreenplumClusterV2HostGroupIdsValues, val)
	}

	value, diag := types.SetValue(types.StringType, yandexMdbGreenplumClusterV2HostGroupIdsValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2HostGroupIds(ctx context.Context, yandexMdbGreenplumClusterV2HostGroupIdsState types.Set, diags *diag.Diagnostics) []string {
	if yandexMdbGreenplumClusterV2HostGroupIdsState.IsNull() || yandexMdbGreenplumClusterV2HostGroupIdsState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2HostGroupIdsState.Elements()) == 0 {
		return []string{}
	}
	yandexMdbGreenplumClusterV2HostGroupIdsRes := make([]string, 0, len(yandexMdbGreenplumClusterV2HostGroupIdsState.Elements()))
	yandexMdbGreenplumClusterV2HostGroupIdsType := make([]types.String, 0, len(yandexMdbGreenplumClusterV2HostGroupIdsState.Elements()))
	diags.Append(yandexMdbGreenplumClusterV2HostGroupIdsState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2HostGroupIdsType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range yandexMdbGreenplumClusterV2HostGroupIdsType {
		yandexMdbGreenplumClusterV2HostGroupIdsRes = append(yandexMdbGreenplumClusterV2HostGroupIdsRes, elem.ValueString())
	}
	return yandexMdbGreenplumClusterV2HostGroupIdsRes
}

func flattenYandexMdbGreenplumClusterV2Labels(ctx context.Context, yandexMdbGreenplumClusterV2Labels map[string]string, listState types.Map, diags *diag.Diagnostics) types.Map {
	if yandexMdbGreenplumClusterV2Labels == nil {
		if !listState.IsNull() && !listState.IsUnknown() && len(listState.Elements()) == 0 {
			return listState
		}
		return types.MapNull(types.StringType)
	}
	yandexMdbGreenplumClusterV2LabelsValues := make(map[string]attr.Value)
	for k, elem := range yandexMdbGreenplumClusterV2Labels {
		val := types.StringValue(elem)
		yandexMdbGreenplumClusterV2LabelsValues[k] = val
	}

	value, diag := types.MapValue(types.StringType, yandexMdbGreenplumClusterV2LabelsValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2Labels(ctx context.Context, yandexMdbGreenplumClusterV2LabelsState types.Map, diags *diag.Diagnostics) map[string]string {
	if yandexMdbGreenplumClusterV2LabelsState.IsNull() || yandexMdbGreenplumClusterV2LabelsState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2LabelsState.Elements()) == 0 {
		return map[string]string{}
	}
	yandexMdbGreenplumClusterV2LabelsRes := make(map[string]string)
	yandexMdbGreenplumClusterV2LabelsType := make(map[string]types.String)
	diags.Append(yandexMdbGreenplumClusterV2LabelsState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2LabelsType, false)...)
	if diags.HasError() {
		return nil
	}
	for k, elem := range yandexMdbGreenplumClusterV2LabelsType {
		yandexMdbGreenplumClusterV2LabelsRes[k] = elem.ValueString()
	}
	return yandexMdbGreenplumClusterV2LabelsRes
}

type yandexMdbGreenplumClusterV2LoggingModel struct {
	CommandCenterEnabled types.Bool   `tfsdk:"command_center_enabled"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	FolderId             types.String `tfsdk:"folder_id"`
	GreenplumEnabled     types.Bool   `tfsdk:"greenplum_enabled"`
	LogGroupId           types.String `tfsdk:"log_group_id"`
	PoolerEnabled        types.Bool   `tfsdk:"pooler_enabled"`
}

func (m *yandexMdbGreenplumClusterV2LoggingModel) GetCommandCenterEnabled() types.Bool {
	return m.CommandCenterEnabled
}
func (m *yandexMdbGreenplumClusterV2LoggingModel) GetEnabled() types.Bool {
	return m.Enabled
}
func (m *yandexMdbGreenplumClusterV2LoggingModel) GetFolderId() types.String {
	return m.FolderId
}
func (m *yandexMdbGreenplumClusterV2LoggingModel) GetGreenplumEnabled() types.Bool {
	return m.GreenplumEnabled
}
func (m *yandexMdbGreenplumClusterV2LoggingModel) GetLogGroupId() types.String {
	return m.LogGroupId
}
func (m *yandexMdbGreenplumClusterV2LoggingModel) GetPoolerEnabled() types.Bool {
	return m.PoolerEnabled
}

func NewYandexMdbGreenplumClusterV2LoggingModel() yandexMdbGreenplumClusterV2LoggingModel {
	return yandexMdbGreenplumClusterV2LoggingModel{
		CommandCenterEnabled: types.BoolNull(),
		Enabled:              types.BoolNull(),
		FolderId:             types.StringNull(),
		GreenplumEnabled:     types.BoolNull(),
		LogGroupId:           types.StringNull(),
		PoolerEnabled:        types.BoolNull(),
	}
}

func yandexMdbGreenplumClusterV2LoggingModelFillUnknown(target yandexMdbGreenplumClusterV2LoggingModel) yandexMdbGreenplumClusterV2LoggingModel {
	if target.CommandCenterEnabled.IsUnknown() || target.CommandCenterEnabled.IsNull() {
		target.CommandCenterEnabled = types.BoolNull()
	}
	if target.Enabled.IsUnknown() || target.Enabled.IsNull() {
		target.Enabled = types.BoolNull()
	}
	if target.FolderId.IsUnknown() || target.FolderId.IsNull() {
		target.FolderId = types.StringNull()
	}
	if target.GreenplumEnabled.IsUnknown() || target.GreenplumEnabled.IsNull() {
		target.GreenplumEnabled = types.BoolNull()
	}
	if target.LogGroupId.IsUnknown() || target.LogGroupId.IsNull() {
		target.LogGroupId = types.StringNull()
	}
	if target.PoolerEnabled.IsUnknown() || target.PoolerEnabled.IsNull() {
		target.PoolerEnabled = types.BoolNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2LoggingModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"command_center_enabled": types.BoolType,
		"enabled":                types.BoolType,
		"folder_id":              types.StringType,
		"greenplum_enabled":      types.BoolType,
		"log_group_id":           types.StringType,
		"pooler_enabled":         types.BoolType,
	},
}

func flattenYandexMdbGreenplumClusterV2Logging(ctx context.Context,
	yandexMdbGreenplumClusterV2Logging *greenplum.LoggingConfig,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2Logging == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2LoggingModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2LoggingModelType.AttrTypes, yandexMdbGreenplumClusterV2LoggingModel{
		CommandCenterEnabled: types.BoolValue(yandexMdbGreenplumClusterV2Logging.GetCommandCenterEnabled()),
		Enabled:              types.BoolValue(yandexMdbGreenplumClusterV2Logging.GetEnabled()),
		FolderId:             types.StringValue(yandexMdbGreenplumClusterV2Logging.GetFolderId()),
		GreenplumEnabled:     types.BoolValue(yandexMdbGreenplumClusterV2Logging.GetGreenplumEnabled()),
		LogGroupId:           types.StringValue(yandexMdbGreenplumClusterV2Logging.GetLogGroupId()),
		PoolerEnabled:        types.BoolValue(yandexMdbGreenplumClusterV2Logging.GetPoolerEnabled()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2Logging(ctx context.Context, yandexMdbGreenplumClusterV2LoggingState types.Object, diags *diag.Diagnostics) *greenplum.LoggingConfig {
	if yandexMdbGreenplumClusterV2LoggingState.IsNull() || yandexMdbGreenplumClusterV2LoggingState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2Logging yandexMdbGreenplumClusterV2LoggingModel
	diags.Append(yandexMdbGreenplumClusterV2LoggingState.As(ctx, &yandexMdbGreenplumClusterV2Logging, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2LoggingModel(ctx, yandexMdbGreenplumClusterV2Logging, diags)
}

func expandYandexMdbGreenplumClusterV2LoggingModel(ctx context.Context, yandexMdbGreenplumClusterV2LoggingState yandexMdbGreenplumClusterV2LoggingModel, diags *diag.Diagnostics) *greenplum.LoggingConfig {
	value := &greenplum.LoggingConfig{}
	value.SetCommandCenterEnabled(yandexMdbGreenplumClusterV2LoggingState.CommandCenterEnabled.ValueBool())
	value.SetEnabled(yandexMdbGreenplumClusterV2LoggingState.Enabled.ValueBool())
	if !(yandexMdbGreenplumClusterV2LoggingState.FolderId.IsNull() || yandexMdbGreenplumClusterV2LoggingState.FolderId.IsUnknown() || yandexMdbGreenplumClusterV2LoggingState.FolderId.Equal(types.StringValue(""))) {
		value.SetFolderId(yandexMdbGreenplumClusterV2LoggingState.FolderId.ValueString())
	}
	value.SetGreenplumEnabled(yandexMdbGreenplumClusterV2LoggingState.GreenplumEnabled.ValueBool())
	if !(yandexMdbGreenplumClusterV2LoggingState.LogGroupId.IsNull() || yandexMdbGreenplumClusterV2LoggingState.LogGroupId.IsUnknown() || yandexMdbGreenplumClusterV2LoggingState.LogGroupId.Equal(types.StringValue(""))) {
		value.SetLogGroupId(yandexMdbGreenplumClusterV2LoggingState.LogGroupId.ValueString())
	}
	value.SetPoolerEnabled(yandexMdbGreenplumClusterV2LoggingState.PoolerEnabled.ValueBool())
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2MaintenanceWindowModel struct {
	Anytime                 types.Object `tfsdk:"anytime"`
	WeeklyMaintenanceWindow types.Object `tfsdk:"weekly_maintenance_window"`
}

func (m *yandexMdbGreenplumClusterV2MaintenanceWindowModel) GetAnytime() types.Object {
	return m.Anytime
}
func (m *yandexMdbGreenplumClusterV2MaintenanceWindowModel) GetWeeklyMaintenanceWindow() types.Object {
	return m.WeeklyMaintenanceWindow
}

func NewYandexMdbGreenplumClusterV2MaintenanceWindowModel() yandexMdbGreenplumClusterV2MaintenanceWindowModel {
	return yandexMdbGreenplumClusterV2MaintenanceWindowModel{
		Anytime:                 types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModelType.AttrTypes),
		WeeklyMaintenanceWindow: types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModelType.AttrTypes),
	}
}

func yandexMdbGreenplumClusterV2MaintenanceWindowModelFillUnknown(target yandexMdbGreenplumClusterV2MaintenanceWindowModel) yandexMdbGreenplumClusterV2MaintenanceWindowModel {
	if target.Anytime.IsUnknown() || target.Anytime.IsNull() {
		target.Anytime = types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModelType.AttrTypes)
	}
	if target.WeeklyMaintenanceWindow.IsUnknown() || target.WeeklyMaintenanceWindow.IsNull() {
		target.WeeklyMaintenanceWindow = types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModelType.AttrTypes)
	}
	return target
}

var yandexMdbGreenplumClusterV2MaintenanceWindowModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"anytime":                   yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModelType,
		"weekly_maintenance_window": yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModelType,
	},
}

func flattenYandexMdbGreenplumClusterV2MaintenanceWindow(ctx context.Context,
	yandexMdbGreenplumClusterV2MaintenanceWindow *greenplum.MaintenanceWindow,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2MaintenanceWindow == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2MaintenanceWindowModelType.AttrTypes, yandexMdbGreenplumClusterV2MaintenanceWindowModel{
		Anytime:                 flattenYandexMdbGreenplumClusterV2MaintenanceWindowAnytime(ctx, yandexMdbGreenplumClusterV2MaintenanceWindow.GetAnytime(), diags),
		WeeklyMaintenanceWindow: flattenYandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow(ctx, yandexMdbGreenplumClusterV2MaintenanceWindow.GetWeeklyMaintenanceWindow(), diags),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2MaintenanceWindow(ctx context.Context, yandexMdbGreenplumClusterV2MaintenanceWindowState types.Object, diags *diag.Diagnostics) *greenplum.MaintenanceWindow {
	if yandexMdbGreenplumClusterV2MaintenanceWindowState.IsNull() || yandexMdbGreenplumClusterV2MaintenanceWindowState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2MaintenanceWindow yandexMdbGreenplumClusterV2MaintenanceWindowModel
	diags.Append(yandexMdbGreenplumClusterV2MaintenanceWindowState.As(ctx, &yandexMdbGreenplumClusterV2MaintenanceWindow, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2MaintenanceWindowModel(ctx, yandexMdbGreenplumClusterV2MaintenanceWindow, diags)
}

func expandYandexMdbGreenplumClusterV2MaintenanceWindowModel(ctx context.Context, yandexMdbGreenplumClusterV2MaintenanceWindowState yandexMdbGreenplumClusterV2MaintenanceWindowModel, diags *diag.Diagnostics) *greenplum.MaintenanceWindow {
	value := &greenplum.MaintenanceWindow{}
	if !(yandexMdbGreenplumClusterV2MaintenanceWindowState.Anytime.IsNull() || yandexMdbGreenplumClusterV2MaintenanceWindowState.Anytime.IsUnknown() || yandexMdbGreenplumClusterV2MaintenanceWindowState.Anytime.Equal(types.Object{})) {
		value.SetAnytime(expandYandexMdbGreenplumClusterV2MaintenanceWindowAnytime(ctx, yandexMdbGreenplumClusterV2MaintenanceWindowState.Anytime, diags))
	}
	if !(yandexMdbGreenplumClusterV2MaintenanceWindowState.WeeklyMaintenanceWindow.IsNull() || yandexMdbGreenplumClusterV2MaintenanceWindowState.WeeklyMaintenanceWindow.IsUnknown() || yandexMdbGreenplumClusterV2MaintenanceWindowState.WeeklyMaintenanceWindow.Equal(types.Object{})) {
		value.SetWeeklyMaintenanceWindow(expandYandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow(ctx, yandexMdbGreenplumClusterV2MaintenanceWindowState.WeeklyMaintenanceWindow, diags))
	}
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel struct {
}

func NewYandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel() yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel {
	return yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel{}
}

func yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModelFillUnknown(target yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel) yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel {
	return target
}

var yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{},
}

func flattenYandexMdbGreenplumClusterV2MaintenanceWindowAnytime(ctx context.Context,
	yandexMdbGreenplumClusterV2MaintenanceWindowAnytime *greenplum.AnytimeMaintenanceWindow,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2MaintenanceWindowAnytime == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModelType.AttrTypes, yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel{})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2MaintenanceWindowAnytime(ctx context.Context, yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeState types.Object, diags *diag.Diagnostics) *greenplum.AnytimeMaintenanceWindow {
	if yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeState.IsNull() || yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2MaintenanceWindowAnytime yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel
	diags.Append(yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeState.As(ctx, &yandexMdbGreenplumClusterV2MaintenanceWindowAnytime, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel(ctx, yandexMdbGreenplumClusterV2MaintenanceWindowAnytime, diags)
}

func expandYandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel(ctx context.Context, yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeState yandexMdbGreenplumClusterV2MaintenanceWindowAnytimeModel, diags *diag.Diagnostics) *greenplum.AnytimeMaintenanceWindow {
	value := &greenplum.AnytimeMaintenanceWindow{}
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel struct {
	Day  types.String `tfsdk:"day"`
	Hour types.Int64  `tfsdk:"hour"`
}

func (m *yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel) GetDay() types.String {
	return m.Day
}
func (m *yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel) GetHour() types.Int64 {
	return m.Hour
}

func NewYandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel() yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel {
	return yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel{
		Day:  types.StringNull(),
		Hour: types.Int64Null(),
	}
}

func yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModelFillUnknown(target yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel) yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel {
	if target.Day.IsUnknown() || target.Day.IsNull() {
		target.Day = types.StringNull()
	}
	if target.Hour.IsUnknown() || target.Hour.IsNull() {
		target.Hour = types.Int64Null()
	}
	return target
}

var yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"day":  types.StringType,
		"hour": types.Int64Type,
	},
}

func flattenYandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow(ctx context.Context,
	yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow *greenplum.WeeklyMaintenanceWindow,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModelType.AttrTypes, yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel{
		Day:  flattenEnum(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow.GetDay()),
		Hour: types.Int64Value(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow.GetHour()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow(ctx context.Context, yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState types.Object, diags *diag.Diagnostics) *greenplum.WeeklyMaintenanceWindow {
	if yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState.IsNull() || yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel
	diags.Append(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState.As(ctx, &yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel(ctx, yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindow, diags)
}

func expandYandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel(ctx context.Context, yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel, diags *diag.Diagnostics) *greenplum.WeeklyMaintenanceWindow {
	value := &greenplum.WeeklyMaintenanceWindow{}
	value.SetDay(greenplum.WeeklyMaintenanceWindow_WeekDay(greenplum.WeeklyMaintenanceWindow_WeekDay_value[yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState.Day.ValueString()]))
	value.SetHour(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState.Hour.ValueInt64())
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2MasterConfigModel struct {
	Resources types.Object `tfsdk:"resources"`
}

func (m *yandexMdbGreenplumClusterV2MasterConfigModel) GetResources() types.Object {
	return m.Resources
}

func NewYandexMdbGreenplumClusterV2MasterConfigModel() yandexMdbGreenplumClusterV2MasterConfigModel {
	return yandexMdbGreenplumClusterV2MasterConfigModel{
		Resources: types.ObjectNull(yandexMdbGreenplumClusterV2MasterConfigResourcesModelType.AttrTypes),
	}
}

func yandexMdbGreenplumClusterV2MasterConfigModelFillUnknown(target yandexMdbGreenplumClusterV2MasterConfigModel) yandexMdbGreenplumClusterV2MasterConfigModel {
	if target.Resources.IsUnknown() || target.Resources.IsNull() {
		target.Resources = types.ObjectNull(yandexMdbGreenplumClusterV2MasterConfigResourcesModelType.AttrTypes)
	}
	return target
}

var yandexMdbGreenplumClusterV2MasterConfigModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"resources": yandexMdbGreenplumClusterV2MasterConfigResourcesModelType,
	},
}

func flattenYandexMdbGreenplumClusterV2MasterConfig(ctx context.Context,
	yandexMdbGreenplumClusterV2MasterConfig *greenplum.MasterSubclusterConfig,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2MasterConfig == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2MasterConfigModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2MasterConfigModelType.AttrTypes, yandexMdbGreenplumClusterV2MasterConfigModel{
		Resources: mdbcommon.FlattenResources(ctx, yandexMdbGreenplumClusterV2MasterConfig.GetResources(), diags),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2MasterConfig(ctx context.Context, yandexMdbGreenplumClusterV2MasterConfigState types.Object, diags *diag.Diagnostics) *greenplum.MasterSubclusterConfigSpec {
	if yandexMdbGreenplumClusterV2MasterConfigState.IsNull() || yandexMdbGreenplumClusterV2MasterConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2MasterConfig yandexMdbGreenplumClusterV2MasterConfigModel
	diags.Append(yandexMdbGreenplumClusterV2MasterConfigState.As(ctx, &yandexMdbGreenplumClusterV2MasterConfig, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2MasterConfigModel_create(ctx, yandexMdbGreenplumClusterV2MasterConfig, diags)
}

func expandYandexMdbGreenplumClusterV2MasterConfigModel(ctx context.Context, yandexMdbGreenplumClusterV2MasterConfigState yandexMdbGreenplumClusterV2MasterConfigModel, diags *diag.Diagnostics) *greenplum.MasterSubclusterConfig {
	value := &greenplum.MasterSubclusterConfig{}
	value.SetResources(mdbcommon.ExpandResources[greenplum.Resources](ctx, yandexMdbGreenplumClusterV2MasterConfigState.Resources, diags))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2MasterConfigModel_create(ctx context.Context, yandexMdbGreenplumClusterV2MasterConfigState yandexMdbGreenplumClusterV2MasterConfigModel, diags *diag.Diagnostics) *greenplum.MasterSubclusterConfigSpec {
	value := &greenplum.MasterSubclusterConfigSpec{}
	value.SetResources(mdbcommon.ExpandResources[greenplum.Resources](ctx, yandexMdbGreenplumClusterV2MasterConfigState.Resources, diags))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2MasterConfigModel_update(ctx context.Context, yandexMdbGreenplumClusterV2MasterConfigState yandexMdbGreenplumClusterV2MasterConfigModel, diags *diag.Diagnostics) *greenplum.MasterSubclusterConfigSpec {
	value := &greenplum.MasterSubclusterConfigSpec{}
	value.SetResources(mdbcommon.ExpandResources[greenplum.Resources](ctx, yandexMdbGreenplumClusterV2MasterConfigState.Resources, diags))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2MasterConfigResourcesModel struct {
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskTypeId       types.String `tfsdk:"disk_type_id"`
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
}

func (m *yandexMdbGreenplumClusterV2MasterConfigResourcesModel) GetDiskSize() types.Int64 {
	return m.DiskSize
}
func (m *yandexMdbGreenplumClusterV2MasterConfigResourcesModel) GetDiskTypeId() types.String {
	return m.DiskTypeId
}
func (m *yandexMdbGreenplumClusterV2MasterConfigResourcesModel) GetResourcePresetId() types.String {
	return m.ResourcePresetId
}

func NewYandexMdbGreenplumClusterV2MasterConfigResourcesModel() yandexMdbGreenplumClusterV2MasterConfigResourcesModel {
	return yandexMdbGreenplumClusterV2MasterConfigResourcesModel{
		DiskSize:         types.Int64Null(),
		DiskTypeId:       types.StringNull(),
		ResourcePresetId: types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2MasterConfigResourcesModelFillUnknown(target yandexMdbGreenplumClusterV2MasterConfigResourcesModel) yandexMdbGreenplumClusterV2MasterConfigResourcesModel {
	if target.DiskSize.IsUnknown() || target.DiskSize.IsNull() {
		target.DiskSize = types.Int64Null()
	}
	if target.DiskTypeId.IsUnknown() || target.DiskTypeId.IsNull() {
		target.DiskTypeId = types.StringNull()
	}
	if target.ResourcePresetId.IsUnknown() || target.ResourcePresetId.IsNull() {
		target.ResourcePresetId = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2MasterConfigResourcesModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"disk_size":          types.Int64Type,
		"disk_type_id":       types.StringType,
		"resource_preset_id": types.StringType,
	},
}

func flattenYandexMdbGreenplumClusterV2Monitoring(ctx context.Context, yandexMdbGreenplumClusterV2Monitoring []*greenplum.Monitoring, listState types.Set, diags *diag.Diagnostics) types.Set {
	if yandexMdbGreenplumClusterV2Monitoring == nil {
		if !listState.IsNull() && !listState.IsUnknown() && len(listState.Elements()) == 0 {
			return listState
		}
		return types.SetNull(yandexMdbGreenplumClusterV2MonitoringStructModelType)
	}
	var yandexMdbGreenplumClusterV2MonitoringValues []attr.Value
	for _, elem := range yandexMdbGreenplumClusterV2Monitoring {
		val := flattenYandexMdbGreenplumClusterV2MonitoringStruct(ctx, elem, diags)
		yandexMdbGreenplumClusterV2MonitoringValues = append(yandexMdbGreenplumClusterV2MonitoringValues, val)
	}

	value, diag := types.SetValue(yandexMdbGreenplumClusterV2MonitoringStructModelType, yandexMdbGreenplumClusterV2MonitoringValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2Monitoring(ctx context.Context, yandexMdbGreenplumClusterV2MonitoringState types.Set, diags *diag.Diagnostics) []*greenplum.Monitoring {
	if yandexMdbGreenplumClusterV2MonitoringState.IsNull() || yandexMdbGreenplumClusterV2MonitoringState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2MonitoringState.Elements()) == 0 {
		return []*greenplum.Monitoring{}
	}
	yandexMdbGreenplumClusterV2MonitoringRes := make([]*greenplum.Monitoring, 0, len(yandexMdbGreenplumClusterV2MonitoringState.Elements()))
	yandexMdbGreenplumClusterV2MonitoringType := make([]yandexMdbGreenplumClusterV2MonitoringStructModel, 0, len(yandexMdbGreenplumClusterV2MonitoringState.Elements()))
	diags.Append(yandexMdbGreenplumClusterV2MonitoringState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2MonitoringType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range yandexMdbGreenplumClusterV2MonitoringType {
		yandexMdbGreenplumClusterV2MonitoringRes = append(yandexMdbGreenplumClusterV2MonitoringRes, expandYandexMdbGreenplumClusterV2MonitoringStructModel(ctx, elem, diags))
	}
	return yandexMdbGreenplumClusterV2MonitoringRes
}

type yandexMdbGreenplumClusterV2PlannedOperationModel struct {
	DelayedUntil types.String `tfsdk:"delayed_until"`
	Info         types.String `tfsdk:"info"`
}

func (m *yandexMdbGreenplumClusterV2PlannedOperationModel) GetDelayedUntil() types.String {
	return m.DelayedUntil
}
func (m *yandexMdbGreenplumClusterV2PlannedOperationModel) GetInfo() types.String {
	return m.Info
}

func NewYandexMdbGreenplumClusterV2PlannedOperationModel() yandexMdbGreenplumClusterV2PlannedOperationModel {
	return yandexMdbGreenplumClusterV2PlannedOperationModel{
		DelayedUntil: types.StringNull(),
		Info:         types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2PlannedOperationModelFillUnknown(target yandexMdbGreenplumClusterV2PlannedOperationModel) yandexMdbGreenplumClusterV2PlannedOperationModel {
	if target.DelayedUntil.IsUnknown() || target.DelayedUntil.IsNull() {
		target.DelayedUntil = types.StringNull()
	}
	if target.Info.IsUnknown() || target.Info.IsNull() {
		target.Info = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2PlannedOperationModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"delayed_until": types.StringType,
		"info":          types.StringType,
	},
}

func flattenYandexMdbGreenplumClusterV2PlannedOperation(ctx context.Context,
	yandexMdbGreenplumClusterV2PlannedOperation *greenplum.MaintenanceOperation,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2PlannedOperation == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2PlannedOperationModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2PlannedOperationModelType.AttrTypes, yandexMdbGreenplumClusterV2PlannedOperationModel{
		DelayedUntil: types.StringValue(yandexMdbGreenplumClusterV2PlannedOperation.GetDelayedUntil().AsTime().Format(time.RFC3339)),
		Info:         types.StringValue(yandexMdbGreenplumClusterV2PlannedOperation.GetInfo()),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2PlannedOperation(ctx context.Context, yandexMdbGreenplumClusterV2PlannedOperationState types.Object, diags *diag.Diagnostics) *greenplum.MaintenanceOperation {
	if yandexMdbGreenplumClusterV2PlannedOperationState.IsNull() || yandexMdbGreenplumClusterV2PlannedOperationState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2PlannedOperation yandexMdbGreenplumClusterV2PlannedOperationModel
	diags.Append(yandexMdbGreenplumClusterV2PlannedOperationState.As(ctx, &yandexMdbGreenplumClusterV2PlannedOperation, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2PlannedOperationModel(ctx, yandexMdbGreenplumClusterV2PlannedOperation, diags)
}

func expandYandexMdbGreenplumClusterV2PlannedOperationModel(ctx context.Context, yandexMdbGreenplumClusterV2PlannedOperationState yandexMdbGreenplumClusterV2PlannedOperationModel, diags *diag.Diagnostics) *greenplum.MaintenanceOperation {
	value := &greenplum.MaintenanceOperation{}
	value.SetDelayedUntil(converter.ParseTimestamp(yandexMdbGreenplumClusterV2PlannedOperationState.DelayedUntil.ValueString(), diags))
	value.SetInfo(yandexMdbGreenplumClusterV2PlannedOperationState.Info.ValueString())
	if diags.HasError() {
		return nil
	}
	return value
}

func flattenYandexMdbGreenplumClusterV2SecurityGroupIds(ctx context.Context, yandexMdbGreenplumClusterV2SecurityGroupIds []string, listState types.Set, diags *diag.Diagnostics) types.Set {
	if yandexMdbGreenplumClusterV2SecurityGroupIds == nil {
		if !listState.IsNull() && !listState.IsUnknown() && len(listState.Elements()) == 0 {
			return listState
		}
		return types.SetNull(types.StringType)
	}
	var yandexMdbGreenplumClusterV2SecurityGroupIdsValues []attr.Value
	for _, elem := range yandexMdbGreenplumClusterV2SecurityGroupIds {
		val := types.StringValue(elem)
		yandexMdbGreenplumClusterV2SecurityGroupIdsValues = append(yandexMdbGreenplumClusterV2SecurityGroupIdsValues, val)
	}

	value, diag := types.SetValue(types.StringType, yandexMdbGreenplumClusterV2SecurityGroupIdsValues)
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2SecurityGroupIds(ctx context.Context, yandexMdbGreenplumClusterV2SecurityGroupIdsState types.Set, diags *diag.Diagnostics) []string {
	if yandexMdbGreenplumClusterV2SecurityGroupIdsState.IsNull() || yandexMdbGreenplumClusterV2SecurityGroupIdsState.IsUnknown() {
		return nil
	}
	if len(yandexMdbGreenplumClusterV2SecurityGroupIdsState.Elements()) == 0 {
		return []string{}
	}
	yandexMdbGreenplumClusterV2SecurityGroupIdsRes := make([]string, 0, len(yandexMdbGreenplumClusterV2SecurityGroupIdsState.Elements()))
	yandexMdbGreenplumClusterV2SecurityGroupIdsType := make([]types.String, 0, len(yandexMdbGreenplumClusterV2SecurityGroupIdsState.Elements()))
	diags.Append(yandexMdbGreenplumClusterV2SecurityGroupIdsState.ElementsAs(ctx, &yandexMdbGreenplumClusterV2SecurityGroupIdsType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range yandexMdbGreenplumClusterV2SecurityGroupIdsType {
		yandexMdbGreenplumClusterV2SecurityGroupIdsRes = append(yandexMdbGreenplumClusterV2SecurityGroupIdsRes, elem.ValueString())
	}
	return yandexMdbGreenplumClusterV2SecurityGroupIdsRes
}

type yandexMdbGreenplumClusterV2SegmentConfigModel struct {
	Resources types.Object `tfsdk:"resources"`
}

func (m *yandexMdbGreenplumClusterV2SegmentConfigModel) GetResources() types.Object {
	return m.Resources
}

func NewYandexMdbGreenplumClusterV2SegmentConfigModel() yandexMdbGreenplumClusterV2SegmentConfigModel {
	return yandexMdbGreenplumClusterV2SegmentConfigModel{
		Resources: types.ObjectNull(yandexMdbGreenplumClusterV2SegmentConfigResourcesModelType.AttrTypes),
	}
}

func yandexMdbGreenplumClusterV2SegmentConfigModelFillUnknown(target yandexMdbGreenplumClusterV2SegmentConfigModel) yandexMdbGreenplumClusterV2SegmentConfigModel {
	if target.Resources.IsUnknown() || target.Resources.IsNull() {
		target.Resources = types.ObjectNull(yandexMdbGreenplumClusterV2SegmentConfigResourcesModelType.AttrTypes)
	}
	return target
}

var yandexMdbGreenplumClusterV2SegmentConfigModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"resources": yandexMdbGreenplumClusterV2SegmentConfigResourcesModelType,
	},
}

func flattenYandexMdbGreenplumClusterV2SegmentConfig(ctx context.Context,
	yandexMdbGreenplumClusterV2SegmentConfig *greenplum.SegmentSubclusterConfig,
	diags *diag.Diagnostics) types.Object {
	if yandexMdbGreenplumClusterV2SegmentConfig == nil {
		return types.ObjectNull(yandexMdbGreenplumClusterV2SegmentConfigModelType.AttrTypes)
	}
	value, diag := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2SegmentConfigModelType.AttrTypes, yandexMdbGreenplumClusterV2SegmentConfigModel{
		Resources: mdbcommon.FlattenResources(ctx, yandexMdbGreenplumClusterV2SegmentConfig.GetResources(), diags),
	})
	diags.Append(diag...)
	return value
}

func expandYandexMdbGreenplumClusterV2SegmentConfig(ctx context.Context, yandexMdbGreenplumClusterV2SegmentConfigState types.Object, diags *diag.Diagnostics) *greenplum.SegmentSubclusterConfigSpec {
	if yandexMdbGreenplumClusterV2SegmentConfigState.IsNull() || yandexMdbGreenplumClusterV2SegmentConfigState.IsUnknown() {
		return nil
	}
	var yandexMdbGreenplumClusterV2SegmentConfig yandexMdbGreenplumClusterV2SegmentConfigModel
	diags.Append(yandexMdbGreenplumClusterV2SegmentConfigState.As(ctx, &yandexMdbGreenplumClusterV2SegmentConfig, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexMdbGreenplumClusterV2SegmentConfigModel_create(ctx, yandexMdbGreenplumClusterV2SegmentConfig, diags)
}

func expandYandexMdbGreenplumClusterV2SegmentConfigModel(ctx context.Context, yandexMdbGreenplumClusterV2SegmentConfigState yandexMdbGreenplumClusterV2SegmentConfigModel, diags *diag.Diagnostics) *greenplum.SegmentSubclusterConfig {
	value := &greenplum.SegmentSubclusterConfig{}
	value.SetResources(mdbcommon.ExpandResources[greenplum.Resources](ctx, yandexMdbGreenplumClusterV2SegmentConfigState.Resources, diags))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2SegmentConfigModel_create(ctx context.Context, yandexMdbGreenplumClusterV2SegmentConfigState yandexMdbGreenplumClusterV2SegmentConfigModel, diags *diag.Diagnostics) *greenplum.SegmentSubclusterConfigSpec {
	value := &greenplum.SegmentSubclusterConfigSpec{}
	value.SetResources(mdbcommon.ExpandResources[greenplum.Resources](ctx, yandexMdbGreenplumClusterV2SegmentConfigState.Resources, diags))
	if diags.HasError() {
		return nil
	}
	return value
}
func expandYandexMdbGreenplumClusterV2SegmentConfigModel_update(ctx context.Context, yandexMdbGreenplumClusterV2SegmentConfigState yandexMdbGreenplumClusterV2SegmentConfigModel, diags *diag.Diagnostics) *greenplum.SegmentSubclusterConfigSpec {
	value := &greenplum.SegmentSubclusterConfigSpec{}
	value.SetResources(mdbcommon.ExpandResources[greenplum.Resources](ctx, yandexMdbGreenplumClusterV2SegmentConfigState.Resources, diags))
	if diags.HasError() {
		return nil
	}
	return value
}

type yandexMdbGreenplumClusterV2SegmentConfigResourcesModel struct {
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskTypeId       types.String `tfsdk:"disk_type_id"`
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
}

func (m *yandexMdbGreenplumClusterV2SegmentConfigResourcesModel) GetDiskSize() types.Int64 {
	return m.DiskSize
}
func (m *yandexMdbGreenplumClusterV2SegmentConfigResourcesModel) GetDiskTypeId() types.String {
	return m.DiskTypeId
}
func (m *yandexMdbGreenplumClusterV2SegmentConfigResourcesModel) GetResourcePresetId() types.String {
	return m.ResourcePresetId
}

func NewYandexMdbGreenplumClusterV2SegmentConfigResourcesModel() yandexMdbGreenplumClusterV2SegmentConfigResourcesModel {
	return yandexMdbGreenplumClusterV2SegmentConfigResourcesModel{
		DiskSize:         types.Int64Null(),
		DiskTypeId:       types.StringNull(),
		ResourcePresetId: types.StringNull(),
	}
}

func yandexMdbGreenplumClusterV2SegmentConfigResourcesModelFillUnknown(target yandexMdbGreenplumClusterV2SegmentConfigResourcesModel) yandexMdbGreenplumClusterV2SegmentConfigResourcesModel {
	if target.DiskSize.IsUnknown() || target.DiskSize.IsNull() {
		target.DiskSize = types.Int64Null()
	}
	if target.DiskTypeId.IsUnknown() || target.DiskTypeId.IsNull() {
		target.DiskTypeId = types.StringNull()
	}
	if target.ResourcePresetId.IsUnknown() || target.ResourcePresetId.IsNull() {
		target.ResourcePresetId = types.StringNull()
	}
	return target
}

var yandexMdbGreenplumClusterV2SegmentConfigResourcesModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"disk_size":          types.Int64Type,
		"disk_type_id":       types.StringType,
		"resource_preset_id": types.StringType,
	},
}

type Restore struct {
	BackupId    types.String `tfsdk:"backup_id"`
	Time        types.String `tfsdk:"time"`
	RestoreOnly types.Set    `tfsdk:"restore_only"`
}

func expandYandexMdbGreenplumClusterV2RestoreOnly(ctx context.Context, restoreOnly types.Set, diags *diag.Diagnostics) []string {
	if restoreOnly.IsNull() || restoreOnly.IsUnknown() || len(restoreOnly.Elements()) == 0 {
		return nil
	}
	restoreOnlyRes := make([]string, 0, len(restoreOnly.Elements()))
	restoreOnlyType := make([]types.String, 0, len(restoreOnly.Elements()))
	diags.Append(restoreOnly.ElementsAs(ctx, &restoreOnlyType, false)...)
	if diags.HasError() {
		return nil
	}
	for _, elem := range restoreOnlyType {
		restoreOnlyRes = append(restoreOnlyRes, elem.ValueString())
	}
	return restoreOnlyRes
}
