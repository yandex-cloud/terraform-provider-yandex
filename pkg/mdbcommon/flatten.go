package mdbcommon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func FlattenResources[V any, T resourceModel[V]](ctx context.Context, r T, diags *diag.Diagnostics) types.Object {
	if r == nil {
		diags.AddError("Failed to flatten resources.", "Resources of cluster can't be nil. It's error in provider")
		return types.ObjectNull(ResourceType.AttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, ResourceType.AttrTypes, Resource{
		ResourcePresetId: types.StringValue(r.GetResourcePresetId()),
		DiskSize:         types.Int64Value(datasize.ToGigabytes(r.GetDiskSize())),
		DiskTypeId:       types.StringValue(r.GetDiskTypeId()),
	})

	diags.Append(d...)
	return obj
}

func FlattenBackupWindowStart(ctx context.Context, bws *timeofday.TimeOfDay, diags *diag.Diagnostics) types.Object {
	if bws == nil {
		return types.ObjectNull(BackupWindowType.AttrTypes)
	}

	bwsObj, d := types.ObjectValueFrom(ctx, BackupWindowType.AttrTypes, BackupWindow{
		Hours:   types.Int64Value(int64(bws.GetHours())),
		Minutes: types.Int64Value(int64(bws.GetMinutes())),
	})
	diags.Append(d...)
	return bwsObj
}

func FlattenBoolWrapper(ctx context.Context, wb *wrapperspb.BoolValue, diags *diag.Diagnostics) types.Bool {
	if wb == nil {
		return types.BoolNull()
	}
	return types.BoolValue(wb.GetValue())
}

func FlattenSetString(ctx context.Context, ss []string, diags *diag.Diagnostics) types.Set {
	obj, d := types.SetValueFrom(ctx, types.StringType, ss)
	diags.Append(d...)
	return obj
}

func FlattenMapString(ctx context.Context, ms map[string]string, diags *diag.Diagnostics) types.Map {
	obj, d := types.MapValueFrom(ctx, types.StringType, ms)
	diags.Append(d...)
	return obj
}

func FlattenMaintenanceWindow[
	V any,
	VW any, VA any,

	WD ~int32,

	W weeklyMaintenanceWindow[VW, WD],
	A anytimeMaintenanceWindow[VA],

	T maintenanceWindow[V, VW, VA, WD, W, A],
](ctx context.Context, mw T, diags *diag.Diagnostics) types.Object {

	if mw == nil {
		diags.AddError("Failed to flatten maintenance window.", "Unsupported nil maintenance window type.")
		return types.ObjectNull(MaintenanceWindowType.AttrTypes)
	}

	var maintenanceWindow MaintenanceWindow

	if ap := mw.GetAnytime(); ap != nil {
		maintenanceWindow.Type = types.StringValue("ANYTIME")
	} else if wp := mw.GetWeeklyMaintenanceWindow(); wp != nil {
		maintenanceWindow.Type = types.StringValue("WEEKLY")
		maintenanceWindow.Day = types.StringValue(
			weekdayNums[int32(wp.GetDay())],
		)
		maintenanceWindow.Hour = types.Int64Value(int64(wp.GetHour()))
	} else {
		diags.AddError("Failed to flatten maintenance window.", "Unsupported maintenance policy type.")
		return types.ObjectNull(MaintenanceWindowType.AttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, MaintenanceWindowType.AttrTypes, maintenanceWindow)
	diags.Append(d...)

	return obj
}

func FlattenInt64Wrapper(ctx context.Context, pgBrpd *wrapperspb.Int64Value, diags *diag.Diagnostics) types.Int64 {
	if pgBrpd == nil {
		return types.Int64Null()
	}
	return types.Int64Value(pgBrpd.GetValue())
}
