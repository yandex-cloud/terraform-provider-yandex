package yq_ydb_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ydbConnectionModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	DatabaseID       types.String `tfsdk:"database_id"`
}
