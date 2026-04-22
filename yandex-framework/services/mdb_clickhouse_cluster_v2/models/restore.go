package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Restore struct {
	BackupId        types.String `tfsdk:"backup_id"`
	IncludePatterns []string     `tfsdk:"include_patterns"`
	ExcludePatterns []string     `tfsdk:"exclude_patterns"`
}

var RestoreAttrTypes = map[string]attr.Type{
	"backup_id":        types.StringType,
	"include_patterns": types.ListType{ElemType: types.StringType},
	"exclude_patterns": types.ListType{ElemType: types.StringType},
}
