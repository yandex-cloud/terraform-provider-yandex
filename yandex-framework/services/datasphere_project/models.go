package datasphere_project

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type projectDataModel struct {
	Id          types.String   `tfsdk:"id"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Labels      types.Map      `tfsdk:"labels"`
	CreatedBy   types.String   `tfsdk:"created_by"`
	Settings    types.Object   `tfsdk:"settings"`
	Limits      types.Object   `tfsdk:"limits"`
	CommunityId types.String   `tfsdk:"community_id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

type limitsObjectModel struct {
	// The number of units that can be spent per hour.
	MaxUnitsPerHour types.Int64 `tfsdk:"max_units_per_hour"`
	// The number of units that can be spent on the one execution.
	MaxUnitsPerExecution types.Int64 `tfsdk:"max_units_per_execution"`
	// The number of units available to the project.
	Balance types.Int64 `tfsdk:"balance"`
}

func (m *limitsObjectModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"max_units_per_hour":      types.Int64Type,
		"max_units_per_execution": types.Int64Type,
		"balance":                 types.Int64Type,
	}
}

type settingsObjectModel struct {
	ServiceAccountId     types.String `tfsdk:"service_account_id"`
	SubnetId             types.String `tfsdk:"subnet_id"`
	DataProcClusterId    types.String `tfsdk:"data_proc_cluster_id"`
	SecurityGroupIds     types.Set    `tfsdk:"security_group_ids"`
	DefaultFolderId      types.String `tfsdk:"default_folder_id"`
	StaleExecTimeoutMode types.String `tfsdk:"stale_exec_timeout_mode"`
}

func (m *settingsObjectModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"service_account_id":      types.StringType,
		"subnet_id":               types.StringType,
		"data_proc_cluster_id":    types.StringType,
		"security_group_ids":      types.SetType{ElemType: types.StringType},
		"default_folder_id":       types.StringType,
		"stale_exec_timeout_mode": types.StringType,
	}
}
