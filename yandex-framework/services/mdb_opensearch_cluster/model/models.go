package model

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

type OpenSearch struct {
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
	ID                  types.String   `tfsdk:"id"`
	ClusterID           types.String   `tfsdk:"cluster_id"`
	FolderID            types.String   `tfsdk:"folder_id"`
	CreatedAt           types.String   `tfsdk:"created_at"`
	Name                types.String   `tfsdk:"name"`
	Description         types.String   `tfsdk:"description"`
	Labels              types.Map      `tfsdk:"labels"`
	Environment         types.String   `tfsdk:"environment"`
	Config              types.Object   `tfsdk:"config"`
	Hosts               types.List     `tfsdk:"hosts"`
	NetworkID           types.String   `tfsdk:"network_id"`
	Health              types.String   `tfsdk:"health"`
	Status              types.String   `tfsdk:"status"`
	SecurityGroupIDs    types.Set      `tfsdk:"security_group_ids"`
	ServiceAccountID    types.String   `tfsdk:"service_account_id"`
	DeletionProtection  types.Bool     `tfsdk:"deletion_protection"`
	MaintenanceWindow   types.Object   `tfsdk:"maintenance_window"`
	AuthSettings        types.Object   `tfsdk:"auth_settings"`
	DiskEncryptionKeyID types.String   `tfsdk:"disk_encryption_key_id"`
}

type Config struct {
	Version                types.String `tfsdk:"version"`
	AdminPassword          types.String `tfsdk:"admin_password"`
	AdminPasswordWo        types.String `tfsdk:"admin_password_wo"`
	AdminPasswordWoVersion types.Int64  `tfsdk:"admin_password_wo_version"`
	OpenSearch             types.Object `tfsdk:"opensearch"`
	Dashboards             types.Object `tfsdk:"dashboards"`
	Access                 types.Object `tfsdk:"access"`
	AuditLog               types.Object `tfsdk:"audit_log"`
}

var ConfigAttrTypes = map[string]attr.Type{
	"version":                   types.StringType,
	"admin_password":            types.StringType,
	"admin_password_wo":         types.StringType,
	"admin_password_wo_version": types.Int64Type,
	"opensearch":                types.ObjectType{AttrTypes: OpenSearchSubConfigAttrTypes},
	"dashboards":                types.ObjectType{AttrTypes: DashboardsSubConfigAttrTypes},
	"access":                    types.ObjectType{AttrTypes: accessAttrTypes},
	"audit_log":                 types.ObjectType{AttrTypes: AuditLogTypes},
}

type dataSourceConfig struct {
	Version       types.String `tfsdk:"version"`
	AdminPassword types.String `tfsdk:"admin_password"`
	OpenSearch    types.Object `tfsdk:"opensearch"`
	Dashboards    types.Object `tfsdk:"dashboards"`
	Access        types.Object `tfsdk:"access"`
	AuditLog      types.Object `tfsdk:"audit_log"`
}

var dataSourceConfigAttrTypes = map[string]attr.Type{
	"version":        types.StringType,
	"admin_password": types.StringType,
	"opensearch":     types.ObjectType{AttrTypes: OpenSearchSubConfigAttrTypes},
	"dashboards":     types.ObjectType{AttrTypes: DashboardsSubConfigAttrTypes},
	"access":         types.ObjectType{AttrTypes: accessAttrTypes},
	"audit_log":      types.ObjectType{AttrTypes: AuditLogTypes},
}

func ClusterToState(ctx context.Context, cluster *opensearch.Cluster, state *OpenSearch) diag.Diagnostics {
	state.FolderID = types.StringValue(cluster.GetFolderId())
	state.CreatedAt = types.StringValue(timestamp.Get(cluster.GetCreatedAt()))
	state.Name = types.StringValue(cluster.GetName())

	newDescription := types.StringValue(cluster.GetDescription())
	if !stringsAreEqual(state.Description, newDescription) {
		state.Description = newDescription
	}

	labels, diags := types.MapValueFrom(ctx, types.StringType, cluster.Labels)
	if diags.HasError() {
		return diags
	}
	if !mapsAreEqual(state.Labels, labels) {
		state.Labels = labels
	}

	state.Environment = types.StringValue(cluster.GetEnvironment().String())

	state.Config, diags = configToState(ctx, cluster.Config, state)
	if diags.HasError() {
		return diags
	}

	state.NetworkID = types.StringValue(cluster.GetNetworkId())
	state.Health = types.StringValue(cluster.GetHealth().String())
	state.Status = types.StringValue(cluster.GetStatus().String())

	securityGroupIDs := mdbcommon.FlattenSetString(ctx, cluster.SecurityGroupIds, &diags)
	if diags.HasError() {
		return diags
	}

	if !setsAreEqual(state.SecurityGroupIDs, securityGroupIDs) {
		state.SecurityGroupIDs = securityGroupIDs
	}

	newServiceAccountId := types.StringValue(cluster.GetServiceAccountId())
	if !stringsAreEqual(state.ServiceAccountID, newServiceAccountId) {
		state.ServiceAccountID = newServiceAccountId
	}

	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())
	state.MaintenanceWindow, diags = maintenanceWindowToObject(ctx, cluster.MaintenanceWindow)
	state.DiskEncryptionKeyID = mdbcommon.FlattenStringWrapper(ctx, cluster.GetDiskEncryptionKeyId(), &diags)
	return diags
}

func configToState(ctx context.Context, cfg *opensearch.ClusterConfig, state *OpenSearch) (types.Object, diag.Diagnostics) {
	stateCfg, diags := ParseConfig(ctx, state)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	adminPassword := types.StringNull()
	adminPasswordWoVersion := types.Int64Null()
	if !(stateCfg == nil || stateCfg.AdminPassword.IsNull() || stateCfg.AdminPassword.IsUnknown()) {
		adminPassword, diags = stateCfg.AdminPassword.ToStringValue(ctx)
		if diags.HasError() {
			return types.ObjectUnknown(ConfigAttrTypes), diags
		}
	}
	if !(stateCfg == nil || stateCfg.AdminPasswordWoVersion.IsNull() || stateCfg.AdminPasswordWoVersion.IsUnknown()) {
		adminPasswordWoVersion = stateCfg.AdminPasswordWoVersion
	}

	//It is required to have a config.opensearch block, so we can skip checking it
	stateOpenSearch, diags := ParseOpenSearchSubConfig(ctx, stateCfg)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	opensearchSubConfig, diags := openSearchSubConfigToObject(ctx, cfg.Opensearch, stateOpenSearch)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	stateDashboards, diags := ParseDashboardSubConfig(ctx, stateCfg)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	dashboardSubConfig, diags := dashboardSubConfigToObject(ctx, cfg.Dashboards, stateDashboards)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	access, diags := accessToObject(ctx, cfg.Access)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	auditLog, diags := auditLogToObject(ctx, cfg.AuditLog)
	if diags.HasError() {
		return types.ObjectUnknown(AuditLogTypes), diags
	}

	if _, isResourceConfig := state.Config.AttributeTypes(ctx)["admin_password_wo"]; isResourceConfig {
		return types.ObjectValueFrom(ctx, ConfigAttrTypes, Config{
			Version:                types.StringValue(cfg.GetVersion()),
			AdminPassword:          adminPassword,
			AdminPasswordWo:        types.StringNull(),
			AdminPasswordWoVersion: adminPasswordWoVersion,
			OpenSearch:             opensearchSubConfig,
			Dashboards:             dashboardSubConfig,
			Access:                 access,
			AuditLog:               auditLog,
		})
	}

	return types.ObjectValueFrom(ctx, dataSourceConfigAttrTypes, dataSourceConfig{
		Version:       types.StringValue(cfg.GetVersion()),
		AdminPassword: adminPassword,
		OpenSearch:    opensearchSubConfig,
		Dashboards:    dashboardSubConfig,
		Access:        access,
		AuditLog:      auditLog,
	})
}

func rolesToSet(roles []opensearch.OpenSearch_GroupRole) (types.Set, diag.Diagnostics) {
	if roles == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	res := make([]attr.Value, 0, len(roles))
	for _, v := range roles {
		res = append(res, types.StringValue(v.String()))
	}

	return types.SetValue(types.StringType, res)
}

func nullableStringSliceToSet(ctx context.Context, s []string) (types.Set, diag.Diagnostics) {
	if s == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	return types.SetValueFrom(ctx, types.StringType, s)
}

func nullableStringSliceToList(ctx context.Context, s []string) (types.List, diag.Diagnostics) {
	if s == nil {
		return types.ListNull(types.StringType), diag.Diagnostics{}
	}

	return types.ListValueFrom(ctx, types.StringType, s)
}

func setsAreEqual(set1, set2 types.Set) bool {
	if set1.IsUnknown() || set2.IsUnknown() {
		return false
	}

	// if one of sets is null and the other is empty then we assume that they are equal
	if len(set1.Elements()) == 0 && len(set2.Elements()) == 0 {
		return true
	}

	if !set1.IsNull() && set1.Equal(set2) {
		return true
	}

	return false
}

func mapsAreEqual(map1, map2 types.Map) bool {
	if map1.IsUnknown() || map2.IsUnknown() {
		return false
	}

	// if one of map is null and the other is empty then we assume that they are equal
	if len(map1.Elements()) == 0 && len(map2.Elements()) == 0 {
		return true
	}

	if !map1.IsNull() && map1.Equal(map2) {
		return true
	}

	return false
}

func stringsAreEqual(str1, str2 types.String) bool {
	if str1.IsUnknown() || str2.IsUnknown() {
		return false
	}

	// if one of strings is null and the other is empty then we assume that they are equal
	if str1.ValueString() == "" && str2.ValueString() == "" {
		return true
	}

	if !str1.IsNull() && str1.Equal(str2) {
		return true
	}

	return false
}

func ParseConfig(ctx context.Context, state *OpenSearch) (*Config, diag.Diagnostics) {
	if state.Config.IsNull() || state.Config.IsUnknown() {
		return &Config{
			Version:                types.StringNull(),
			AdminPassword:          types.StringNull(),
			AdminPasswordWo:        types.StringNull(),
			AdminPasswordWoVersion: types.Int64Null(),
			OpenSearch:             types.ObjectNull(OpenSearchSubConfigAttrTypes),
			Dashboards:             types.ObjectNull(DashboardsSubConfigAttrTypes),
			Access:                 types.ObjectNull(accessAttrTypes),
			AuditLog:               types.ObjectNull(AuditLogTypes),
		}, diag.Diagnostics{}
	}

	if _, isResourceConfig := state.Config.AttributeTypes(ctx)["admin_password_wo"]; !isResourceConfig {
		dataSourceCfg := &dataSourceConfig{}
		diags := state.Config.As(ctx, dataSourceCfg, datasize.DefaultOpts)
		if diags.HasError() {
			return nil, diags
		}

		return &Config{
			Version:                dataSourceCfg.Version,
			AdminPassword:          dataSourceCfg.AdminPassword,
			AdminPasswordWo:        types.StringNull(),
			AdminPasswordWoVersion: types.Int64Null(),
			OpenSearch:             dataSourceCfg.OpenSearch,
			Dashboards:             dataSourceCfg.Dashboards,
			Access:                 dataSourceCfg.Access,
			AuditLog:               dataSourceCfg.AuditLog,
		}, diag.Diagnostics{}
	}

	planConfig := &Config{}
	diags := state.Config.As(ctx, &planConfig, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return planConfig, diag.Diagnostics{}
}

func ParseGenerics[T any, V any](ctx context.Context, plan, state T, parse func(context.Context, T) (V, diag.Diagnostics)) (V, V, diag.Diagnostics) {
	planConfig, diags := parse(ctx, plan)
	if diags.HasError() {
		//NOTE: can't create an empty value result, so just dublicate planConfig
		return planConfig, planConfig, diags
	}

	stateConfig, diags := parse(ctx, state)
	if diags.HasError() {
		return planConfig, stateConfig, diags
	}

	return planConfig, stateConfig, diag.Diagnostics{}
}
