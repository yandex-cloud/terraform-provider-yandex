package yq_monitoring_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type monitoringConnectionModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	CloudID          types.String `tfsdk:"cloud_id"`
	FolderID         types.String `tfsdk:"folder_id"`
}
