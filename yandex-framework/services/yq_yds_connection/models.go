package yq_yds_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ydsConnectionModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	DatabaseID       types.String `tfsdk:"database_id"`
	SharedReading    types.Bool   `tfsdk:"shared_reading"`
}
