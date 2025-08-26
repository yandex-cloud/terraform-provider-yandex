package mdb_sharded_postgresql_cluster

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

const (
	ConsolePasswordStubOnImport = "<real value unknown because resource was imported>"
)

type Cluster struct {
	Id                 types.String   `tfsdk:"id"`
	FolderId           types.String   `tfsdk:"folder_id"`
	NetworkId          types.String   `tfsdk:"network_id"`
	Name               types.String   `tfsdk:"name"`
	Description        types.String   `tfsdk:"description"`
	Environment        types.String   `tfsdk:"environment"`
	Labels             types.Map      `tfsdk:"labels"`
	Config             types.Object   `tfsdk:"config"`
	HostSpecs          types.Map      `tfsdk:"hosts"`
	MaintenanceWindow  types.Object   `tfsdk:"maintenance_window"`
	DeletionProtection types.Bool     `tfsdk:"deletion_protection"`
	SecurityGroupIds   types.Set      `tfsdk:"security_group_ids"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

type Host struct {
	Zone           types.String `tfsdk:"zone"`
	SubnetId       types.String `tfsdk:"subnet_id"`
	AssignPublicIp types.Bool   `tfsdk:"assign_public_ip"`
	FQDN           types.String `tfsdk:"fqdn"`
	Type           types.String `tfsdk:"type"`
}

type Config struct {
	Access                 types.Object            `tfsdk:"access"`
	BackupRetainPeriodDays types.Int64             `tfsdk:"backup_retain_period_days"`
	BackupWindowStart      types.Object            `tfsdk:"backup_window_start"`
	SPQRConfig             ShardedPostgreSQLConfig `tfsdk:"sharded_postgresql_config"`
}

type ShardedPostgreSQLConfig struct {
	Common      mdbcommon.SettingsMapValue `tfsdk:"common"`
	Router      *ComponentConfig           `tfsdk:"router"`
	Coordinator *ComponentConfig           `tfsdk:"coordinator"`
	Infra       *InfraConfig               `tfsdk:"infra"`
	Balancer    mdbcommon.SettingsMapValue `tfsdk:"balancer"`
}

type ComponentConfig struct {
	Resources types.Object               `tfsdk:"resources"`
	Config    mdbcommon.SettingsMapValue `tfsdk:"config"`
}

type InfraConfig struct {
	Router      mdbcommon.SettingsMapValue `tfsdk:"router"`
	Coordinator mdbcommon.SettingsMapValue `tfsdk:"coordinator"`
	Resources   types.Object               `tfsdk:"resources"`
}

type Access struct {
	DataLens     types.Bool `tfsdk:"data_lens"`
	WebSql       types.Bool `tfsdk:"web_sql"`
	DataTransfer types.Bool `tfsdk:"data_transfer"`
	Serverless   types.Bool `tfsdk:"serverless"`
}

type MaintenanceWindow struct {
	Type types.String `tfsdk:"type"`
	Day  types.String `tfsdk:"day"`
	Hour types.Int64  `tfsdk:"hour"`
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

var MaintenanceWindowAttrTypes = map[string]attr.Type{
	"type": types.StringType,
	"day":  types.StringType,
	"hour": types.Int64Type,
}

type BackupWindowStart struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

var BackupWindowStartAttrTypes = map[string]attr.Type{
	"hours":   types.Int64Type,
	"minutes": types.Int64Type,
}

var ConfigAttrTypes = map[string]attr.Type{
	"access":                    types.ObjectType{AttrTypes: mdbcommon.AccessAttrTypes},
	"backup_retain_period_days": types.Int64Type,
	"backup_window_start":       types.ObjectType{AttrTypes: BackupWindowStartAttrTypes},
	"sharded_postgresql_config": types.ObjectType{AttrTypes: ShardedPostgreSQLConfigAttrTypes},
}

var ShardedPostgreSQLConfigAttrTypes = map[string]attr.Type{
	"common":      mdbcommon.NewSettingsMapType(attrProvider),
	"router":      types.ObjectType{AttrTypes: ComponentsAttrTypes},
	"coordinator": types.ObjectType{AttrTypes: ComponentsAttrTypes},
	"infra":       types.ObjectType{AttrTypes: InfraAttrTypes},
	"balancer":    mdbcommon.NewSettingsMapType(attrProvider),
}

var ComponentsAttrTypes = map[string]attr.Type{
	"config":    mdbcommon.NewSettingsMapType(attrProvider),
	"resources": types.ObjectType{AttrTypes: ResourcesAttrTypes},
}

var InfraAttrTypes = map[string]attr.Type{
	"resources":   types.ObjectType{AttrTypes: ResourcesAttrTypes},
	"router":      mdbcommon.NewSettingsMapType(attrProvider),
	"coordinator": mdbcommon.NewSettingsMapType(attrProvider),
}

var hostType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"zone":             types.StringType,
		"subnet_id":        types.StringType,
		"assign_public_ip": types.BoolType,
		"fqdn":             types.StringType,
		"type":             types.StringType,
	},
}
