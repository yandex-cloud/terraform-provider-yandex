package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
