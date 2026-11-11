package mdb_sharded_postgresql_user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func flattenPermissions(userPerms []*spqr.Permission) types.Set {
	if userPerms == nil {
		return types.SetNull(permissionType)
	}
	perms := []attr.Value{}
	for _, p := range userPerms {
		perms = append(perms, types.ObjectValueMust(
			permissionType.AttrTypes,
			map[string]attr.Value{
				"database": types.StringValue(p.DatabaseName),
			}),
		)
	}
	return types.SetValueMust(permissionType, perms)
}

func flattenGrants(userGrants []string) types.Set {
	if userGrants == nil {
		return types.SetNull(types.StringType)
	}
	grants := []attr.Value{}
	for _, g := range userGrants {
		grants = append(grants, types.StringValue(g))
	}
	return types.SetValueMust(types.StringType, grants)
}

func flattenSettings(ctx context.Context, c any, diags *diag.Diagnostics) mdbcommon.SettingsMapValue {
	if c == nil {
		return mdbcommon.NewSettingsMapNull()
	}

	a := protobuf_adapter.NewProtobufMapDataAdapter()

	attrs := a.Extract(ctx, c, diags)
	if diags.HasError() {
		return mdbcommon.NewSettingsMapNull()
	}

	attrsPresent := make(map[string]attr.Value)
	for attr, val := range attrs {
		if val.IsNull() || val.IsUnknown() {
			continue
		}

		if valInt, ok := val.(types.Int64); ok {
			if valInt.ValueInt64() != 0 {
				attrsPresent[attr] = val
			}
			continue
		}

		if valStr, ok := val.(types.String); ok {
			if valStr.ValueString() != "" {
				attrsPresent[attr] = val
			}
			continue
		}

		if _, ok := val.(types.Bool); ok {
			attrsPresent[attr] = val
			continue
		}

		if _, ok := val.(types.List); ok {
			attrsPresent[attr] = val
			continue
		}

		if valFloat, ok := val.(types.Float64); ok {
			if valFloat.ValueFloat64() != 0 {
				attrsPresent[attr] = val
			}
			continue
		}

		if valNum, ok := val.(types.Number); ok {
			i, _ := valNum.ValueBigFloat().Int64()
			if !valNum.ValueBigFloat().IsInt() || i != 0 {
				attrsPresent[attr] = val
			}
			continue
		}

		if _, ok := val.(types.Tuple); ok {
			attrsPresent[attr] = val
			continue
		}

		diags.AddError("Flatten ShardedPostgresql Config Erorr", fmt.Sprintf("Attribute %s has a unknown handling value %v", attr, val.String()))

	}

	settings, d := mdbcommon.NewSettingsMapValue(attrsPresent, &SettingsAttributeInfoProvider{})
	diags.Append(d...)
	return settings
}
