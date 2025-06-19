package yq_yds_binding

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ydsBindingModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ConnectionID  types.String `tfsdk:"connection_id"`
	Stream        types.String `tfsdk:"stream"`
	Format        types.String `tfsdk:"format"`
	Compression   types.String `tfsdk:"compression"`
	FormatSetting types.Map    `tfsdk:"format_setting"`
	Column        types.List   `tfsdk:"column"`
}
