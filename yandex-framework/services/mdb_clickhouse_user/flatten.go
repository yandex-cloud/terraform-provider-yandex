package mdb_clickhouse_user

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon"
)

func flattenPermissions(ctx context.Context, permissions []*clickhouse.Permission, diags *diag.Diagnostics) types.Set {
	if permissions == nil {
		return types.SetNull(permissionType)
	}

	var permissionValues []attr.Value

	for _, permission := range permissions {
		permissionValue, diag := types.ObjectValue(permissionType.AttrTypes, map[string]attr.Value{
			"database_name": types.StringValue(permission.DatabaseName),
		})

		permissionValues = append(permissionValues, permissionValue)
		diags.Append(diag...)
	}

	value, diag := types.SetValue(permissionType, permissionValues)
	diags.Append(diag...)

	return value
}

func flattenQuotas(ctx context.Context, quotas []*clickhouse.UserQuota, diags *diag.Diagnostics) types.Set {
	if quotas == nil {
		return types.SetNull(quotaType)
	}

	var quotaValues []attr.Value

	for _, quota := range quotas {
		quotaValue, diag := types.ObjectValue(quotaType.AttrTypes, map[string]attr.Value{
			"interval_duration": chcommon.Int64FromWrapper(quota.IntervalDuration),
			"queries":           chcommon.Int64FromWrapper(quota.Queries),
			"errors":            chcommon.Int64FromWrapper(quota.Errors),
			"result_rows":       chcommon.Int64FromWrapper(quota.ResultRows),
			"read_rows":         chcommon.Int64FromWrapper(quota.ReadRows),
			"execution_time":    chcommon.Int64FromWrapper(quota.ExecutionTime),
		})

		quotaValues = append(quotaValues, quotaValue)
		diags.Append(diag...)
	}

	value, diag := types.SetValue(quotaType, quotaValues)
	diags.Append(diag...)

	return value
}

func flattenConnectionManager(ctx context.Context, connectionManager *clickhouse.ConnectionManager, diags *diag.Diagnostics) types.Object {
	if connectionManager == nil {
		return types.ObjectNull(connectionManagerType)
	}

	obj, d := types.ObjectValueFrom(
		ctx, connectionManagerType, ConnectionManager{
			ConnectionId: chcommon.NullableString(connectionManager.ConnectionId),
		},
	)

	log.Printf("[TRACE] mdb_clickhouse_user: flatten connection_manager to state: %+v\n", obj)
	diags.Append(d...)
	return obj
}
