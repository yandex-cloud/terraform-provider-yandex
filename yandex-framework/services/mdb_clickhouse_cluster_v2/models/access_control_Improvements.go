package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type AccessControlImprovements struct {
	SelectFromSystemDbRequiresGrant          types.Bool `tfsdk:"select_from_system_db_requires_grant"`
	SelectFromInformationSchemaRequiresGrant types.Bool `tfsdk:"select_from_information_schema_requires_grant"`
}

var AccessControlImprovementsAttrTypes = map[string]attr.Type{
	"select_from_system_db_requires_grant":          types.BoolType,
	"select_from_information_schema_requires_grant": types.BoolType,
}

func flattenAccessControlImprovements(ctx context.Context, aci *clickhouseConfig.ClickhouseConfig_AccessControlImprovements, diags *diag.Diagnostics) types.Object {
	if aci == nil {
		return types.ObjectNull(AccessControlImprovementsAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, AccessControlImprovementsAttrTypes, AccessControlImprovements{
			SelectFromSystemDbRequiresGrant:          mdbcommon.FlattenBoolWrapper(ctx, aci.SelectFromSystemDbRequiresGrant, diags),
			SelectFromInformationSchemaRequiresGrant: mdbcommon.FlattenBoolWrapper(ctx, aci.SelectFromInformationSchemaRequiresGrant, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func expandAccessControlImprovements(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_AccessControlImprovements {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var aci AccessControlImprovements
	diags.Append(obj.As(ctx, &aci, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_AccessControlImprovements{
		SelectFromSystemDbRequiresGrant:          mdbcommon.ExpandBoolWrapper(ctx, aci.SelectFromSystemDbRequiresGrant, diags),
		SelectFromInformationSchemaRequiresGrant: mdbcommon.ExpandBoolWrapper(ctx, aci.SelectFromInformationSchemaRequiresGrant, diags),
	}
}
