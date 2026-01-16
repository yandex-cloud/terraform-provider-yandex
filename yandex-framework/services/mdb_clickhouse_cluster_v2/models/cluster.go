package models

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Cluster struct {
	Id                  types.String `tfsdk:"id"`
	FolderId            types.String `tfsdk:"folder_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Labels              types.Map    `tfsdk:"labels"`
	Environment         types.String `tfsdk:"environment"`
	NetworkId           types.String `tfsdk:"network_id"`
	ServiceAccountId    types.String `tfsdk:"service_account_id"`
	MaintenanceWindow   types.Object `tfsdk:"maintenance_window"`
	SecurityGroupIds    types.Set    `tfsdk:"security_group_ids"`
	DeletionProtection  types.Bool   `tfsdk:"deletion_protection"`
	DiskEncryptionKeyId types.String `tfsdk:"disk_encryption_key_id"`

	Version                types.String `tfsdk:"version"`
	ClickHouse             types.Object `tfsdk:"clickhouse"`
	ZooKeeper              types.Object `tfsdk:"zookeeper"`
	BackupWindowStart      types.Object `tfsdk:"backup_window_start"`
	Access                 types.Object `tfsdk:"access"`
	CloudStorage           types.Object `tfsdk:"cloud_storage"`
	SqlDatabaseManagement  types.Bool   `tfsdk:"sql_database_management"`
	SqlUserManagement      types.Bool   `tfsdk:"sql_user_management"`
	AdminPassword          types.String `tfsdk:"admin_password"`
	EmbeddedKeeper         types.Bool   `tfsdk:"embedded_keeper"`
	BackupRetainPeriodDays types.Int64  `tfsdk:"backup_retain_period_days"`

	FormatSchema types.Set  `tfsdk:"format_schema"`
	MLModel      types.Set  `tfsdk:"ml_model"`
	Shards       types.Map  `tfsdk:"shards"`
	ShardGroup   types.List `tfsdk:"shard_group"`

	HostSpecs            types.Map      `tfsdk:"hosts"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
	CopySchemaOnNewHosts types.Bool     `tfsdk:"copy_schema_on_new_hosts"`
}

var ClusterAttrTypes = map[string]attr.Type{
	"id":                        types.StringType,
	"folder_id":                 types.StringType,
	"created_at":                types.StringType,
	"name":                      types.StringType,
	"description":               types.StringType,
	"labels":                    types.MapType{ElemType: types.StringType},
	"environment":               types.StringType,
	"network_id":                types.StringType,
	"service_account_id":        types.StringType,
	"maintenance_window":        types.ObjectType{AttrTypes: MaintenanceWindowAttrTypes},
	"security_group_ids":        types.SetType{ElemType: types.StringType},
	"deletion_protection":       types.BoolType,
	"disk_encryption_key_id":    types.StringType,
	"version":                   types.StringType,
	"clickhouse":                types.ObjectType{AttrTypes: ClickhouseAttrTypes},
	"zookeeper":                 types.ObjectType{AttrTypes: ZookeeperAttrTypes},
	"backup_window_start":       types.ObjectType{AttrTypes: BackupWindowStartAttrTypes},
	"access":                    types.ObjectType{AttrTypes: AccessAttrTypes},
	"cloud_storage":             types.ObjectType{AttrTypes: CloudStorageAttrTypes},
	"sql_database_management":   types.BoolType,
	"sql_user_management":       types.BoolType,
	"admin_password":            types.StringType,
	"embedded_keeper":           types.BoolType,
	"backup_retain_period_days": types.Int64Type,

	"format_schema": types.SetType{ElemType: types.ObjectType{AttrTypes: FormatSchemaAttrTypes}},
	"ml_model":      types.SetType{ElemType: types.ObjectType{AttrTypes: MLModelAttrTypes}},
	"shards":        types.MapType{ElemType: types.ObjectType{AttrTypes: ShardAttrTypes}},
	"shard_group":   types.ListType{ElemType: types.ObjectType{AttrTypes: ShardGroupAttrTypes}},

	"hosts":                    types.MapType{ElemType: types.StringType},
	"timeouts":                 timeouts.Type{},
	"copy_schema_on_new_hosts": types.BoolType,
}
