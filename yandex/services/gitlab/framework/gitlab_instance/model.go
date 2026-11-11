package gitlab_instance

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InstanceModel struct {
	AdminEmail                types.String   `tfsdk:"admin_email"`
	AdminLogin                types.String   `tfsdk:"admin_login"`
	ApprovalRulesId           types.String   `tfsdk:"approval_rules_id"`
	ApprovalRulesToken        types.String   `tfsdk:"approval_rules_token"`
	BackupRetainPeriodDays    types.Int64    `tfsdk:"backup_retain_period_days"`
	CreatedAt                 types.String   `tfsdk:"created_at"`
	DeletionProtection        types.Bool     `tfsdk:"deletion_protection"`
	Description               types.String   `tfsdk:"description"`
	DiskSize                  types.Int64    `tfsdk:"disk_size"`
	Domain                    types.String   `tfsdk:"domain"`
	FolderId                  types.String   `tfsdk:"folder_id"`
	GitlabVersion             types.String   `tfsdk:"gitlab_version"`
	Id                        types.String   `tfsdk:"id"`
	Labels                    types.Map      `tfsdk:"labels"`
	MaintenanceDeleteUntagged types.Bool     `tfsdk:"maintenance_delete_untagged"`
	Name                      types.String   `tfsdk:"name"`
	ResourcePresetId          types.String   `tfsdk:"resource_preset_id"`
	Status                    types.String   `tfsdk:"status"`
	SubnetId                  types.String   `tfsdk:"subnet_id"`
	UpdatedAt                 types.String   `tfsdk:"updated_at"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
}
