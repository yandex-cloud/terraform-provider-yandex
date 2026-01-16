package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BackupWindowStart struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

var BackupWindowStartAttrTypes = map[string]attr.Type{
	"hours":   types.Int64Type,
	"minutes": types.Int64Type,
}
