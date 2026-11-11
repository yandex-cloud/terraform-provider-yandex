package yq_object_storage_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type objectStorageConnectionModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	Bucket           types.String `tfsdk:"bucket"`
}
