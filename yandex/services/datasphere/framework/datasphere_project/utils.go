package datasphere_project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Convert from the API data model to the Terraform data model
// and refresh any attribute values.
func convertToTerraformModel(ctx context.Context, terraformModel *projectDataModel, grpcModel *datasphere.Project, diag *diag.Diagnostics, balance *wrapperspb.Int64Value) {
	terraformModel.Name = types.StringValue(grpcModel.Name)
	terraformModel.CreatedAt = types.StringValue(timestamp.Get(grpcModel.CreatedAt))
	terraformModel.Description = types.StringValue(grpcModel.Description)
	terraformModel.CreatedBy = types.StringValue(grpcModel.CreatedById)
	terraformModel.CommunityId = types.StringValue(grpcModel.CommunityId)

	labels, diags := types.MapValueFrom(ctx, types.StringType, grpcModel.Labels)
	terraformModel.Labels = labels
	diag.Append(diags...)

	if grpcModel.Settings != nil {
		var settings settingsObjectModel

		settings.ServiceAccountId = types.StringValue(grpcModel.Settings.ServiceAccountId)
		settings.SubnetId = types.StringValue(grpcModel.Settings.SubnetId)
		settings.DataProcClusterId = types.StringValue(grpcModel.Settings.DataProcClusterId)

		if grpcModel.Settings.SecurityGroupIds != nil && len(grpcModel.Settings.SecurityGroupIds) > 0 {
			securityGroups, diags := types.SetValueFrom(ctx, types.StringType, grpcModel.Settings.SecurityGroupIds)
			diag.Append(diags...)
			settings.SecurityGroupIds = securityGroups
		} else {
			settings.SecurityGroupIds = types.SetNull(types.StringType)
		}
		settings.DefaultFolderId = types.StringValue(grpcModel.Settings.DefaultFolderId)
		settings.StaleExecTimeoutMode = types.StringValue(grpcModel.Settings.StaleExecTimeoutMode.String())
		settingsObject, diags := types.ObjectValueFrom(ctx, settings.attributeTypes(), settings)
		diag.Append(diags...)
		terraformModel.Settings = settingsObject
	}
	if grpcModel.Limits != nil || balance != nil {
		var limits limitsObjectModel
		if grpcModel.Limits.MaxUnitsPerHour != nil {
			limits.MaxUnitsPerHour = types.Int64Value(grpcModel.Limits.MaxUnitsPerHour.Value)
		}
		if grpcModel.Limits.MaxUnitsPerExecution != nil {
			limits.MaxUnitsPerExecution = types.Int64Value(grpcModel.Limits.MaxUnitsPerExecution.Value)
		}
		if balance != nil {
			limits.Balance = types.Int64Value(balance.Value)
		}
		limitsObject, diags := types.ObjectValueFrom(ctx, limits.attributeTypes(), limits)
		diag.Append(diags...)
		terraformModel.Limits = limitsObject
	}
}
