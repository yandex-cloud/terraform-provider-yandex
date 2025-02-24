package mdb_postgresql_cluster_beta

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func flattenAccess(ctx context.Context, pgAccess *postgresql.Access, diags *diag.Diagnostics) types.Object {
	if pgAccess == nil {
		return types.ObjectNull(AccessAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, AccessAttrTypes, Access{
			DataLens:     types.BoolValue(pgAccess.DataLens),
			DataTransfer: types.BoolValue(pgAccess.DataTransfer),
			Serverless:   types.BoolValue(pgAccess.Serverless),
			WebSql:       types.BoolValue(pgAccess.WebSql),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenMaintenanceWindow(ctx context.Context, mw *postgresql.MaintenanceWindow, diags *diag.Diagnostics) types.Object {

	var maintenanceWindow MaintenanceWindow
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *postgresql.MaintenanceWindow_Anytime:
			maintenanceWindow.Type = types.StringValue("ANYTIME")
			// do nothing
		case *postgresql.MaintenanceWindow_WeeklyMaintenanceWindow:
			maintenanceWindow.Type = types.StringValue("WEEKLY")
			maintenanceWindow.Day = types.StringValue(
				postgresql.WeeklyMaintenanceWindow_WeekDay_name[int32(p.WeeklyMaintenanceWindow.GetDay())],
			)
			maintenanceWindow.Hour = types.Int64Value(p.WeeklyMaintenanceWindow.Hour)
		default:
			diags.AddError("Failed to flatten maintenance window.", "Unsupported PostgreSQL maintenance policy type.")
			return types.ObjectNull(MaintenanceWindowAttrTypes)
		}
	} else {
		diags.AddError("Failed to flatten maintenance window.", "Unsupported nil PostgreSQL maintenance window type.")
		return types.ObjectNull(MaintenanceWindowAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, MaintenanceWindowAttrTypes, maintenanceWindow)
	diags.Append(d...)

	return obj
}

func flattenPerformanceDiagnostics(ctx context.Context, pd *postgresql.PerformanceDiagnostics, diags *diag.Diagnostics) types.Object {
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

func flattenBackupRetainPeriodDays(ctx context.Context, pgBrpd *wrapperspb.Int64Value, diags *diag.Diagnostics) types.Int64 {
	if pgBrpd == nil {
		return types.Int64Null()
	}
	return types.Int64Value(pgBrpd.GetValue())
}

func flattenBackupWindowStart(ctx context.Context, pgBws *timeofday.TimeOfDay, diags *diag.Diagnostics) types.Object {
	if pgBws == nil {
		return types.ObjectNull(BackupWindowStartAttrTypes)
	}

	bwsObj, d := types.ObjectValueFrom(ctx, BackupWindowStartAttrTypes, BackupWindowStart{
		Hours:   types.Int64Value(int64(pgBws.GetHours())),
		Minutes: types.Int64Value(int64(pgBws.GetMinutes())),
	})
	diags.Append(d...)
	return bwsObj
}

func flattenMapString(ctx context.Context, ms map[string]string, diags *diag.Diagnostics) types.Map {
	obj, d := types.MapValueFrom(ctx, types.StringType, ms)
	diags.Append(d...)
	return obj
}

func flattenSetString(ctx context.Context, ss []string, diags *diag.Diagnostics) types.Set {
	if ss == nil {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}

	obj, d := types.SetValueFrom(ctx, types.StringType, ss)
	diags.Append(d...)
	return obj
}

func flattenBoolWrapper(ctx context.Context, wb *wrapperspb.BoolValue, diags *diag.Diagnostics) types.Bool {
	if wb == nil {
		return types.BoolNull()
	}
	return types.BoolValue(wb.GetValue())
}

func flattenResources(ctx context.Context, r *postgresql.Resources, diags *diag.Diagnostics) types.Object {
	if r == nil {
		diags.AddError("Failed to flatten resources.", "Resources of cluster can't be nil. It's error in provider")
		return types.ObjectNull(ResourcesAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, ResourcesAttrTypes, Resources{
		ResourcePresetID: types.StringValue(r.ResourcePresetId),
		DiskSize:         types.Int64Value(datasize.ToGigabytes(r.DiskSize)),
		DiskTypeID:       types.StringValue(r.DiskTypeId),
	})

	diags.Append(d...)
	return obj
}

func flattenConfig(ctx context.Context, statePGCfg PgSettingsMapValue, c *postgresql.ClusterConfig, diags *diag.Diagnostics) types.Object {
	if c == nil {
		diags.AddError("Failed to flatten config.", "Config of cluster can't be nil. It's error in provider")
		return types.ObjectNull(ConfigAttrTypes)
	}

	if statePGCfg.IsNull() || statePGCfg.IsUnknown() {
		statePGCfg = flattenPostgresqlConfig(ctx, c.PostgresqlConfig, diags)
	}

	obj, d := types.ObjectValueFrom(ctx, ConfigAttrTypes, Config{
		Version:                types.StringValue(c.Version),
		Resources:              flattenResources(ctx, c.Resources, diags),
		Autofailover:           flattenBoolWrapper(ctx, c.GetAutofailover(), diags),
		Access:                 flattenAccess(ctx, c.Access, diags),
		PerformanceDiagnostics: flattenPerformanceDiagnostics(ctx, c.PerformanceDiagnostics, diags),
		BackupRetainPeriodDays: flattenBackupRetainPeriodDays(ctx, c.BackupRetainPeriodDays, diags),
		BackupWindowStart:      flattenBackupWindowStart(ctx, c.BackupWindowStart, diags),
		PostgtgreSQLConfig:     statePGCfg,
	})
	diags.Append(d...)
	return obj
}

func getUserConfig(ctx context.Context, c postgresql.ClusterConfig_PostgresqlConfig, diags *diag.Diagnostics) interface{} {

	if c == nil {
		return nil
	}

	rc := reflect.ValueOf(c)

	if rc.Kind() == reflect.Ptr {
		if rc.IsNil() {
			diags.AddError(
				"Failed to flatten postgresql_config.",
				fmt.Sprintf("Can't scan type %T for extract attributes. It's error in provider", c),
			)
			return nil
		}

		rc = rc.Elem()
	}

	if rc.Kind() != reflect.Struct {
		diags.AddError(
			"Failed to flatten postgresql_config.",
			fmt.Sprintf("Can't scan type %T for extract attributes. It's error in provider", c),
		)
		return nil
	}

	rcType := rc.Type()
	var pgConf reflect.Value
	for i := 0; i < rcType.NumField(); i++ {
		field := rcType.Field(i)
		t, ok := protobuf_adapter.FindTag(field, "protobuf", "name")
		if !ok {
			continue
		}

		if !strings.Contains(t, "postgresql_config") {
			continue
		}

		pgConf = rc.Field(i)
	}
	if !pgConf.IsValid() {
		diags.AddError(
			"Failed to flatten postgresql_config.",
			fmt.Sprintf(
				`
				Can't find postgresql config in source struct type %T
				It's error in provider.
				`, c,
			),
		)
		return nil
	}

	if pgConf.Kind() == reflect.Ptr {
		pgConf = pgConf.Elem()
	}
	if pgConf.Kind() != reflect.Struct {
		diags.AddError(
			"Failed to flatten postgresql_config.",
			fmt.Sprintf(
				`
				Can't scan type %T for extract attributes: postgresql_config must be a struct. 
				It's error in provider.
				`, c,
			),
		)
		return nil
	}

	pgConfType := pgConf.Type()
	var uConf interface{}
	for i := 0; i < pgConfType.NumField(); i++ {
		field := pgConfType.Field(i)
		t, ok := protobuf_adapter.FindTag(field, "protobuf", "name")
		if !ok {
			continue
		}

		if t != "user_config" {
			continue
		}

		uConf = pgConf.Field(i).Interface()
	}

	if uConf == nil {
		diags.AddError(
			"Failed to flatten postgresql_config.",
			fmt.Sprintf(
				`
				Can't find user config in source struct type %T
				It's error in provider.
				`, c,
			),
		)
	}

	return uConf
}

func flattenPostgresqlConfig(ctx context.Context, c postgresql.ClusterConfig_PostgresqlConfig, diags *diag.Diagnostics) PgSettingsMapValue {

	a := protobuf_adapter.NewProtobufMapDataAdapter()
	uc := getUserConfig(ctx, c, diags)
	if diags.HasError() {
		return NewPgSettingsMapNull()
	}

	attrs := a.Extract(ctx, uc, diags)
	if diags.HasError() {
		return NewPgSettingsMapNull()
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

		diags.AddError("Flatten Postgresql Config Erorr", fmt.Sprintf("Attribute %s has a unknown handling value %v", attr, val.String()))

	}

	mv, d := NewPgSettingsMapValue(attrsPresent)
	diags.Append(d...)
	return mv
}
