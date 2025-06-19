package yq_object_storage_binding

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type objectStorageBindingModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ConnectionID  types.String `tfsdk:"connection_id"`
	Format        types.String `tfsdk:"format"`
	Compression   types.String `tfsdk:"compression"`
	PathPattern   types.String `tfsdk:"path_pattern"`
	FormatSetting types.Map    `tfsdk:"format_setting"`
	Projection    types.Map    `tfsdk:"projection"`
	PartitionedBy types.List   `tfsdk:"partitioned_by"`
	Column        types.List   `tfsdk:"column"`
}
