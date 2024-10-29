package cluster

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Cluster struct {
	Id          types.String `tfsdk:"id"`
	FolderId    types.String `tfsdk:"folder_id"`
	NetworkId   types.String `tfsdk:"network_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Environment types.String `tfsdk:"environment"`
	Labels      types.Map    `tfsdk:"labels"`
	Config      types.Object `tfsdk:"config"`
	HostSpecs   types.Map    `tfsdk:"hosts"`
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
	Version   types.String `tfsdk:"version"`
	Resources types.Object `tfsdk:"resources"`
}

var ConfigAttrTypes = map[string]attr.Type{
	"version":   types.StringType,
	"resources": types.ObjectType{AttrTypes: ResourcesAttrTypes},
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
