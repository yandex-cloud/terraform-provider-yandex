package mdb_mysql_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func flattenAccess(ctx context.Context, pgAccess *mysql.Access, diags *diag.Diagnostics) types.Object {
	if pgAccess == nil {
		return types.ObjectNull(AccessAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, AccessAttrTypes, Access{
			DataLens:     types.BoolValue(pgAccess.DataLens),
			DataTransfer: types.BoolValue(pgAccess.DataTransfer),
			WebSql:       types.BoolValue(pgAccess.WebSql),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenPerformanceDiagnostics(ctx context.Context, pd *mysql.PerformanceDiagnostics, diags *diag.Diagnostics) types.Object {
	if pd == nil {
		return types.ObjectNull(PerformanceDiagnosticsAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, PerformanceDiagnosticsAttrTypes, PerformanceDiagnostics{
			Enabled:                    types.BoolValue(pd.Enabled),
			SessionsSamplingInterval:   types.Int64Value(pd.SessionsSamplingInterval),
			StatementsSamplingInterval: types.Int64Value(pd.StatementsSamplingInterval),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenDiskSizeAutoscaling(ctx context.Context, dsa *mysql.DiskSizeAutoscaling, diags *diag.Diagnostics) types.Object {
	if dsa == nil {
		return types.ObjectNull(DiskSizeAutoscalingAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DiskSizeAutoscalingAttrTypes, DiskSizeAutoscaling{
			DiskSizeLimit:           types.Int64Value(datasize.ToGigabytes(dsa.GetDiskSizeLimit())),
			PlannedUsageThreshold:   types.Int64Value(dsa.PlannedUsageThreshold),
			EmergencyUsageThreshold: types.Int64Value(dsa.EmergencyUsageThreshold),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenConfig(
	ctx context.Context,
	stateMSCfg mdbcommon.SettingsMapValue,
	c *mysql.ClusterConfig, diags *diag.Diagnostics,
) Config {
	if c == nil {
		diags.AddError("Failed to flatten config.", "Config of cluster can't be nil. It's error in provider")
		return Config{}
	}

	if stateMSCfg.IsNull() || stateMSCfg.IsUnknown() {
		stateMSCfg = flattenMySQLConfig(ctx, c.MysqlConfig, diags)
	}

	return Config{
		Version:                types.StringValue(c.Version),
		Resources:              mdbcommon.FlattenResources(ctx, c.Resources, diags),
		Access:                 flattenAccess(ctx, c.Access, diags),
		PerformanceDiagnostics: flattenPerformanceDiagnostics(ctx, c.PerformanceDiagnostics, diags),
		DiskSizeAutoscaling:    flattenDiskSizeAutoscaling(ctx, c.DiskSizeAutoscaling, diags),
		BackupRetainPeriodDays: mdbcommon.FlattenInt64Wrapper(ctx, c.BackupRetainPeriodDays, diags),
		BackupWindowStart:      mdbcommon.FlattenBackupWindowStart(ctx, c.BackupWindowStart, diags),
		MySQLConfig:            stateMSCfg,
	}
}

func flattenMySQLConfig(ctx context.Context, c mysql.ClusterConfig_MysqlConfig, diags *diag.Diagnostics) mdbcommon.SettingsMapValue {
	a := protobuf_adapter.NewProtobufMapDataAdapter()
	uc := mdbcommon.GetUserConfig(ctx, c, "mysql_config", diags)
	if diags.HasError() {
		return NewMsSettingsMapNull()
	}

	attrs := a.Extract(ctx, uc, diags)
	if diags.HasError() {
		return NewMsSettingsMapNull()
	}

	attrsPresent := make(map[string]attr.Value)
	for attr, val := range attrs {
		if ok := mdbcommon.IsAttrZeroValue(val, diags); !ok {
			attrsPresent[attr] = val
		}

		if diags.HasError() {
			diags.AddError("Flatten MySQL Config Erorr", fmt.Sprintf("Can't check zero attribute %s", attr))
		}
	}

	mv, d := NewMsSettingsMapValue(attrsPresent)
	diags.Append(d...)
	return mv
}
