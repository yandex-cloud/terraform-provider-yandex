package mdb_clickhouse_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon"
)

func expandPermissionsFromState(ctx context.Context, permissionsSet types.Set, diags *diag.Diagnostics) []*clickhouse.Permission {
	if permissionsSet.IsNull() || permissionsSet.IsUnknown() {
		return nil
	}

	permissionsRes := make([]*clickhouse.Permission, 0, len(permissionsSet.Elements()))
	permissionsType := make([]Permission, 0, len(permissionsSet.Elements()))
	diag := permissionsSet.ElementsAs(ctx, &permissionsType, false)
	diags.Append(diag...)
	if diag.HasError() {
		return nil
	}

	for _, permission := range permissionsType {
		permissionsRes = append(permissionsRes, &clickhouse.Permission{
			DatabaseName: permission.DatabaseName.ValueString(),
		})
	}

	return permissionsRes
}

func expandQuotasFromState(ctx context.Context, quotasState types.Set, diags *diag.Diagnostics) []*clickhouse.UserQuota {
	if quotasState.IsNull() || quotasState.IsUnknown() {
		return nil
	}

	quotasRes := make([]*clickhouse.UserQuota, 0, len(quotasState.Elements()))
	quotasTypes := make([]Quota, 0, len(quotasState.Elements()))
	diag := quotasState.ElementsAs(ctx, &quotasTypes, false)
	diags.Append(diag...)
	if diag.HasError() {
		return nil
	}

	for _, quota := range quotasTypes {
		quotasRes = append(quotasRes, &clickhouse.UserQuota{
			IntervalDuration: chcommon.WrapInt64(quota.IntervalDuration),
			Queries:          chcommon.WrapInt64(quota.Queries),
			Errors:           chcommon.WrapInt64(quota.Errors),
			ResultRows:       chcommon.WrapInt64(quota.ResultRows),
			ReadRows:         chcommon.WrapInt64(quota.ReadRows),
			ExecutionTime:    chcommon.WrapInt64(quota.ExecutionTime),
		})
	}

	return quotasRes
}
