package mdb_sharded_postgresql_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func flattenMapString(ctx context.Context, ms map[string]string, diags *diag.Diagnostics) types.Map {
	obj, d := types.MapValueFrom(ctx, types.StringType, ms)
	diags.Append(d...)
	return obj
}

func flattenMaintenanceWindow(ctx context.Context, mw *spqr.MaintenanceWindow, diags *diag.Diagnostics) types.Object {
	var maintenanceWindow MaintenanceWindow
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *spqr.MaintenanceWindow_Anytime:
			maintenanceWindow.Type = types.StringValue("ANYTIME")
			// do nothing
		case *spqr.MaintenanceWindow_WeeklyMaintenanceWindow:
			maintenanceWindow.Type = types.StringValue("WEEKLY")
			maintenanceWindow.Day = types.StringValue(
				spqr.WeeklyMaintenanceWindow_WeekDay_name[int32(p.WeeklyMaintenanceWindow.GetDay())],
			)
			maintenanceWindow.Hour = types.Int64Value(p.WeeklyMaintenanceWindow.Hour)
		default:
			diags.AddError("Failed to flatten maintenance window.", "Unsupported MySQL maintenance policy type.")
			return types.ObjectNull(MaintenanceWindowAttrTypes)
		}
	} else {
		diags.AddError("Failed to flatten maintenance window.", "Unsupported nil MySQL maintenance window type.")
		return types.ObjectNull(MaintenanceWindowAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, MaintenanceWindowAttrTypes, maintenanceWindow)
	diags.Append(d...)

	return obj
}

func flattenConfig(ctx context.Context, cfgState Config, c *spqr.ClusterConfig, diags *diag.Diagnostics) types.Object {
	if c == nil {
		diags.AddError("Failed to flatten config.", "Config of cluster can't be nil. It's error in provider")
		return types.ObjectNull(ConfigAttrTypes)
	}

	cfg := &Config{
		Access:                 mdbcommon.FlattenAccess[Access](ctx, c.Access.ProtoReflect(), accessAttrTypes, diags),
		BackupRetainPeriodDays: flattenBackupRetainPeriodDays(ctx, c.BackupRetainPeriodDays, diags),
		BackupWindowStart:      flattenBackupWindowStart(ctx, c.BackupWindowStart, diags),
		SPQRConfig:             flattenSPQRConfig(ctx, cfgState, c.SpqrConfig, diags),
	}
	obj, d := types.ObjectValueFrom(ctx, ConfigAttrTypes, cfg)
	diags.Append(d...)
	return obj
}

func flattenBackupRetainPeriodDays(ctx context.Context, brpd *wrapperspb.Int64Value, diags *diag.Diagnostics) types.Int64 {
	if brpd == nil {
		return types.Int64Null()
	}
	return types.Int64Value(brpd.GetValue())
}

func flattenBackupWindowStart(ctx context.Context, bws *timeofday.TimeOfDay, diags *diag.Diagnostics) types.Object {
	if bws == nil {
		return types.ObjectNull(BackupWindowStartAttrTypes)
	}

	bwsObj, d := types.ObjectValueFrom(ctx, BackupWindowStartAttrTypes, BackupWindowStart{
		Hours:   types.Int64Value(int64(bws.GetHours())),
		Minutes: types.Int64Value(int64(bws.GetMinutes())),
	})
	diags.Append(d...)
	return bwsObj
}

func flattenSPQRConfig(ctx context.Context, cfgState Config, c *spqr.SPQRConfig, diags *diag.Diagnostics) ShardedPostgreSQLConfig {
	cfg := ShardedPostgreSQLConfig{}

	if c == nil {
		c = &spqr.SPQRConfig{}
	}

	if c.Router != nil {
		cfg.Router = &ComponentConfig{Config: NewSettingsMapEmpty()}
		cfg.Router.Resources = mdbcommon.FlattenResources(ctx, c.Router.Resources, diags)
		cfg.Router.Config = flattenComponentConfig(ctx, c.Router.Config, diags)
		c.Router = nil
	}

	if c.Coordinator != nil {
		cfg.Coordinator = &ComponentConfig{Config: NewSettingsMapEmpty()}
		cfg.Coordinator.Resources = mdbcommon.FlattenResources(ctx, c.Coordinator.Resources, diags)
		cfg.Coordinator.Config = flattenComponentConfig(ctx, c.Coordinator.Config, diags)
		c.Coordinator = nil
	}

	if c.Infra != nil {
		cfg.Infra = &InfraConfig{Router: NewSettingsMapEmpty(), Coordinator: NewSettingsMapEmpty()}
		cfg.Infra.Resources = mdbcommon.FlattenResources(ctx, c.Infra.Resources, diags)
		cfg.Infra.Router = flattenComponentConfig(ctx, c.Infra.Router, diags)
		cfg.Infra.Coordinator = flattenComponentConfig(ctx, c.Infra.Coordinator, diags)
		c.Infra = nil
	}

	if c.Balancer != nil {
		cfg.Balancer = flattenComponentConfig(ctx, c.Balancer, diags)
		c.Balancer = nil
	}

	cfgElements := flattenComponentConfig(ctx, c, diags).PrimitiveElements(ctx, diags)
	if v, ok := cfgState.SPQRConfig.Common.PrimitiveElements(ctx, diags)["console_password"]; ok {
		cfgElements["console_password"] = v
	}
	cfg.Common = mdbcommon.NewSettingsMapValueMust(cfgElements, attrProvider)

	return cfg
}

func flattenComponentConfig(ctx context.Context, c any, diags *diag.Diagnostics) mdbcommon.SettingsMapValue {
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
