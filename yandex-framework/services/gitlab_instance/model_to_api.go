package gitlab_instance

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func BuildCreateInstanceRequest(ctx context.Context, instanceModel *InstanceModel, providerConfig *config.State) (*gitlab.CreateInstanceRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	folderID, d := validate.FolderID(instanceModel.FolderId, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	common, _, dd := buildBaseInstanceProperties(ctx, instanceModel, nil)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	if instanceModel.Domain.IsNull() || instanceModel.Domain.IsUnknown() || instanceModel.Domain.String() == "" {
		diags.Append(diag.NewErrorDiagnostic("Domain is required", "Domain is required and not provided."))
		return nil, diags
	}

	domain := instanceModel.Domain.ValueString()
	if !domainRegex.MatchString(domain) {
		diags.Append(diag.NewErrorDiagnostic("Provided invalid domain for gitlab instance", fmt.Sprintf("Domain: %s", instanceModel.Domain.ValueString())))
		return nil, diags
	}

	instanceCreateRequest := &gitlab.CreateInstanceRequest{
		FolderId:                  folderID,
		Name:                      common.Name,
		Description:               common.Description,
		Labels:                    common.Labels,
		BackupRetainPeriodDays:    common.BackupRetainPeriodDays,
		ResourcePresetId:          common.ResourcePresetId,
		DiskSize:                  common.DiskSize,
		MaintenanceDeleteUntagged: common.MaintenanceDeleteUntagged,
		DeletionProtection:        common.DeletionProtection,
		ApprovalRulesId:           common.ApprovalRulesId,
		AdminLogin:                instanceModel.AdminLogin.ValueString(),
		AdminEmail:                instanceModel.AdminEmail.ValueString(),
		DomainPrefix:              strings.Split(domain, ".")[0],
		SubnetId:                  instanceModel.SubnetId.ValueString(),
	}

	return instanceCreateRequest, diags
}

type BaseInstanceProperties struct {
	Name                      string
	Description               string
	Labels                    map[string]string
	BackupRetainPeriodDays    int64
	ResourcePresetId          string
	DiskSize                  int64
	MaintenanceDeleteUntagged bool
	DeletionProtection        bool
	ApprovalRulesId           string
}

func buildBaseInstanceProperties(ctx context.Context, plan, state *InstanceModel) (*BaseInstanceProperties, []string, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	updateMaskPaths := make([]string, 0)

	if state != nil {
		if !plan.Name.Equal(state.Name) {
			updateMaskPaths = append(updateMaskPaths, "name")
		}
		if !stringsAreEqual(plan.Description, state.Description) {
			updateMaskPaths = append(updateMaskPaths, "description")
		}
		if !plan.BackupRetainPeriodDays.Equal(state.BackupRetainPeriodDays) {
			updateMaskPaths = append(updateMaskPaths, "backup_retain_period_days")
		}
		if !plan.ResourcePresetId.Equal(state.ResourcePresetId) {
			updateMaskPaths = append(updateMaskPaths, "resource_preset_id")
		}
		if !plan.DiskSize.Equal(state.DiskSize) {
			updateMaskPaths = append(updateMaskPaths, "disk_size")
		}
		if !plan.MaintenanceDeleteUntagged.Equal(state.MaintenanceDeleteUntagged) {
			updateMaskPaths = append(updateMaskPaths, "maintenance_delete_untagged")
		}
		if !plan.DeletionProtection.Equal(state.DeletionProtection) {
			updateMaskPaths = append(updateMaskPaths, "deletion_protection")
		}
		if !stringsAreEqual(plan.ApprovalRulesId, state.ApprovalRulesId) {
			updateMaskPaths = append(updateMaskPaths, "approval_rules_id")
		}
	}

	labels := make(map[string]string, len(plan.Labels.Elements()))
	diags.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)

	if diags.HasError() {
		return nil, nil, diags
	}

	if state != nil && !mapsAreEqual(plan.Labels, state.Labels) {
		updateMaskPaths = append(updateMaskPaths, "labels")
	}

	return &BaseInstanceProperties{
		Name:                      plan.Name.ValueString(),
		Description:               plan.Description.ValueString(),
		Labels:                    labels,
		BackupRetainPeriodDays:    plan.BackupRetainPeriodDays.ValueInt64(),
		ResourcePresetId:          plan.ResourcePresetId.ValueString(),
		DiskSize:                  datasize.ToBytes(plan.DiskSize.ValueInt64()),
		MaintenanceDeleteUntagged: plan.MaintenanceDeleteUntagged.ValueBool(),
		DeletionProtection:        plan.DeletionProtection.ValueBool(),
		ApprovalRulesId:           plan.ApprovalRulesId.ValueString(),
	}, updateMaskPaths, diags
}
