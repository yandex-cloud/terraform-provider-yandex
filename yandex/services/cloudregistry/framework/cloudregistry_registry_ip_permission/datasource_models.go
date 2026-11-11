package cloudregistry_registry_ip_permission

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
)

type yandexCloudregistryIPPermissionDataSourceModel struct {
	ID           types.String   `tfsdk:"id"`
	RegistryName types.String   `tfsdk:"registry_name"`
	RegistryId   types.String   `tfsdk:"registry_id"`
	Push         types.Set      `tfsdk:"push"`
	Pull         types.Set      `tfsdk:"pull"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

var yandexCloudregistryIPPermissionDataSourceModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":            types.StringType,
		"registry_name": types.StringType,
		"registry_id":   types.StringType,
		"push":          types.SetType{ElemType: types.StringType},
		"pull":          types.SetType{ElemType: types.StringType},
		"timeouts":      timeouts.AttributesAll(context.Background()).GetType(),
	},
}

func flattenYandexCloudregistryIPPermissionDataSource(ctx context.Context,
	permissions []*cloudregistry.IpPermission,
	state yandexCloudregistryIPPermissionDataSourceModel,
	registryName types.String,
	registryID types.String,
	to timeouts.Value,
	diags *diag.Diagnostics) types.Object {

	if permissions == nil {
		return types.ObjectNull(yandexCloudregistryIPPermissionDataSourceModelType.AttrTypes)
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

	value, diagObj := types.ObjectValueFrom(ctx, yandexCloudregistryIPPermissionDataSourceModelType.AttrTypes, yandexCloudregistryIPPermissionDataSourceModel{
		ID:           registryID,
		RegistryName: registryName,
		RegistryId:   registryID,
		Push:         pushSet,
		Pull:         pullSet,
		Timeouts:     to,
	})
	diags.Append(diagObj...)
	return value
}
