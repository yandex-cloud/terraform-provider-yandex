package cloudregistry_ip_permission

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
)

type yandexCloudregistryIPPermissionModel struct {
	ID         types.String   `tfsdk:"id"`
	RegistryId types.String   `tfsdk:"registry_id"`
	Push       types.Set      `tfsdk:"push"`
	Pull       types.Set      `tfsdk:"pull"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

var yandexCloudregistryIPPermissionModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":          types.StringType,
		"registry_id": types.StringType,
		"push":        types.SetType{ElemType: types.StringType},
		"pull":        types.SetType{ElemType: types.StringType},
		"timeouts":    timeouts.AttributesAll(context.Background()).GetType(),
	},
}

func flattenYandexCloudregistryIPPermission(ctx context.Context,
	permissions []*cloudregistry.IpPermission,
	state yandexCloudregistryIPPermissionModel,
	registryID types.String,
	to timeouts.Value,
	diags *diag.Diagnostics) types.Object {

	if permissions == nil {
		return types.ObjectNull(yandexCloudregistryIPPermissionModelType.AttrTypes)
	}
	var pushIPs, pullIPs []string
	for _, perm := range permissions {
		switch perm.Action {
		case cloudregistry.IpPermission_PUSH:
			pushIPs = append(pushIPs, perm.Ip)
		case cloudregistry.IpPermission_PULL:
			pullIPs = append(pullIPs, perm.Ip)
		}
	}

	pushSet, diagSet := types.SetValueFrom(ctx, types.StringType, pushIPs)
	diags.Append(diagSet...)

	pullSet, diagSet := types.SetValueFrom(ctx, types.StringType, pullIPs)
	diags.Append(diagSet...)

	value, diagObj := types.ObjectValueFrom(ctx, yandexCloudregistryIPPermissionModelType.AttrTypes, yandexCloudregistryIPPermissionModel{
		ID:         registryID,
		RegistryId: registryID,
		Push:       pushSet,
		Pull:       pullSet,
		Timeouts:   to,
	})
	diags.Append(diagObj...)
	return value
}

func expandYandexCloudregistryIPPermission(ctx context.Context, state types.Object, diags *diag.Diagnostics) *cloudregistry.IpPermission {
	if state.IsNull() || state.IsUnknown() {
		return nil
	}
	var model yandexCloudregistryIPPermissionModel
	diags.Append(state.As(ctx, &model, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	return expandYandexCloudregistryIPPermissionModel(ctx, model, diags)
}

func expandYandexCloudregistryIPPermissionModel(ctx context.Context, model yandexCloudregistryIPPermissionModel, diags *diag.Diagnostics) *cloudregistry.IpPermission {
	permission := &cloudregistry.IpPermission{}

	if !model.Push.IsNull() && !model.Push.IsUnknown() {
		var push []string
		diags.Append(model.Push.ElementsAs(ctx, &push, false)...)
		permission.Ip = push[0]
		permission.Action = cloudregistry.IpPermission_PUSH
	}

	if !model.Pull.IsNull() && !model.Pull.IsUnknown() {
		var pull []string
		diags.Append(model.Pull.ElementsAs(ctx, &pull, false)...)
		permission.Ip = pull[0]
		permission.Action = cloudregistry.IpPermission_PULL
	}

	return permission
}
