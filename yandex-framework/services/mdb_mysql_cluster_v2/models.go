package mdb_mysql_cluster_v2

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type Cluster struct {
	Id                     types.String               `tfsdk:"id"`
	FolderId               types.String               `tfsdk:"folder_id"`
	NetworkId              types.String               `tfsdk:"network_id"`
	Name                   types.String               `tfsdk:"name"`
	Description            types.String               `tfsdk:"description"`
	Environment            types.String               `tfsdk:"environment"`
	Labels                 types.Map                  `tfsdk:"labels"`
	HostSpecs              types.Map                  `tfsdk:"hosts"`
	MaintenanceWindow      types.Object               `tfsdk:"maintenance_window"`
	DeletionProtection     types.Bool                 `tfsdk:"deletion_protection"`
	SecurityGroupIds       types.Set                  `tfsdk:"security_group_ids"`
	Version                types.String               `tfsdk:"version"`
	Resources              types.Object               `tfsdk:"resources"`
	Access                 types.Object               `tfsdk:"access"`
	PerformanceDiagnostics types.Object               `tfsdk:"performance_diagnostics"`
	BackupRetainPeriodDays types.Int64                `tfsdk:"backup_retain_period_days"`
	BackupWindowStart      types.Object               `tfsdk:"backup_window_start"`
	MySQLConfig            mdbcommon.SettingsMapValue `tfsdk:"mysql_config"`
	DiskEncryptionKeyId    types.String               `tfsdk:"disk_encryption_key_id"`
	Timeouts               timeouts.Value             `tfsdk:"timeouts"`
}

type Host struct {
	Zone              types.String `tfsdk:"zone"`
	SubnetId          types.String `tfsdk:"subnet_id"`
	AssignPublicIp    types.Bool   `tfsdk:"assign_public_ip"`
	FQDN              types.String `tfsdk:"fqdn"`
	ReplicationSource types.String `tfsdk:"replication_source"`
}

var hostType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"zone":               types.StringType,
		"subnet_id":          types.StringType,
		"assign_public_ip":   types.BoolType,
		"fqdn":               types.StringType,
		"replication_source": types.StringType,
	},
}

type Config struct {
	Version                types.String               `tfsdk:"version"`
	Resources              types.Object               `tfsdk:"resources"`
	Access                 types.Object               `tfsdk:"access"`
	PerformanceDiagnostics types.Object               `tfsdk:"performance_diagnostics"`
	BackupRetainPeriodDays types.Int64                `tfsdk:"backup_retain_period_days"`
	BackupWindowStart      types.Object               `tfsdk:"backup_window_start"`
	MySQLConfig            mdbcommon.SettingsMapValue `tfsdk:"mysql_config"`
}

type MaintenanceWindow struct {
	Type types.String `tfsdk:"type"`
	Day  types.String `tfsdk:"day"`
	Hour types.Int64  `tfsdk:"hour"`
}

var MaintenanceWindowAttrTypes = map[string]attr.Type{
	"type": types.StringType,
	"day":  types.StringType,
	"hour": types.Int64Type,
}

var ConfigAttrTypes = map[string]attr.Type{
	"version":                   types.StringType,
	"resources":                 types.ObjectType{AttrTypes: ResourcesAttrTypes},
	"access":                    types.ObjectType{AttrTypes: AccessAttrTypes},
	"performance_diagnostics":   types.ObjectType{AttrTypes: PerformanceDiagnosticsAttrTypes},
	"backup_retain_period_days": types.Int64Type,
	"backup_window_start":       types.ObjectType{AttrTypes: BackupWindowStartAttrTypes},
	"mysql_config":              mdbcommon.NewSettingsMapType(msAttrProvider),
}

type Access struct {
	DataLens     types.Bool `tfsdk:"data_lens"`
	WebSql       types.Bool `tfsdk:"web_sql"`
	DataTransfer types.Bool `tfsdk:"data_transfer"`
}

var AccessAttrTypes = map[string]attr.Type{
	"data_lens":     types.BoolType,
	"web_sql":       types.BoolType,
	"data_transfer": types.BoolType,
}

type PerformanceDiagnostics struct {
	Enabled                    types.Bool  `tfsdk:"enabled"`
	SessionsSamplingInterval   types.Int64 `tfsdk:"sessions_sampling_interval"`
	StatementsSamplingInterval types.Int64 `tfsdk:"statements_sampling_interval"`
}

var PerformanceDiagnosticsAttrTypes = map[string]attr.Type{
	"enabled":                      types.BoolType,
	"sessions_sampling_interval":   types.Int64Type,
	"statements_sampling_interval": types.Int64Type,
}

type Resources struct {
	ResourcePresetID types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskTypeID       types.String `tfsdk:"disk_type_id"`
}

var ResourcesAttrTypes = map[string]attr.Type{
	"resource_preset_id": types.StringType,
	"disk_size":          types.Int64Type,
	"disk_type_id":       types.StringType,
}

type BackupWindowStart struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

var BackupWindowStartAttrTypes = map[string]attr.Type{
	"hours":   types.Int64Type,
	"minutes": types.Int64Type,
}
