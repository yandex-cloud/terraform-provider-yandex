package cluster

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Cluster struct {
	Id                 types.String `tfsdk:"id"`
	FolderId           types.String `tfsdk:"folder_id"`
	NetworkId          types.String `tfsdk:"network_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Environment        types.String `tfsdk:"environment"`
	Labels             types.Map    `tfsdk:"labels"`
	Config             types.Object `tfsdk:"config"`
	HostSpecs          types.Map    `tfsdk:"hosts"`
	MaintenanceWindow  types.Object `tfsdk:"maintenance_window"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
	SecurityGroupIds   types.Set    `tfsdk:"security_group_ids"`
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
	Version      types.String `tfsdk:"version"`
	Resources    types.Object `tfsdk:"resources"`
	Autofailover types.Bool   `tfsdk:"autofailover"`
	Access       types.Object `tfsdk:"access"`
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
	"version":      types.StringType,
	"resources":    types.ObjectType{AttrTypes: ResourcesAttrTypes},
	"autofailover": types.BoolType,
	"access":       types.ObjectType{AttrTypes: AccessAttrTypes},
}

type Access struct {
	DataLens     types.Bool `tfsdk:"data_lens"`
	WebSql       types.Bool `tfsdk:"web_sql"`
	Serverless   types.Bool `tfsdk:"serverless"`
	DataTransfer types.Bool `tfsdk:"data_transfer"`
}

var AccessAttrTypes = map[string]attr.Type{
	"data_lens":     types.BoolType,
	"web_sql":       types.BoolType,
	"serverless":    types.BoolType,
	"data_transfer": types.BoolType,
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
