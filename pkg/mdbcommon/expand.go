package mdbcommon

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func ExpandBackupWindow(ctx context.Context, bw types.Object, diags *diag.Diagnostics) *timeofday.TimeOfDay {
	backupWindow := &BackupWindow{}
	diags.Append(bw.As(ctx, backupWindow, datasize.UnhandledOpts)...)
	if diags.HasError() {
		return nil
	}
	rs := &timeofday.TimeOfDay{
		Hours:   int32(backupWindow.Hours.ValueInt64()),
		Minutes: int32(backupWindow.Minutes.ValueInt64()),
	}
	return rs
}

func ExpandResources[V any, T resourceModel[V]](ctx context.Context, o types.Object, diags *diag.Diagnostics) T {
	if !utils.IsPresent(o) {
		return nil
	}

	d := &Resource{}
	diags.Append(o.As(ctx, d, baseOptions)...)
	if diags.HasError() {
		return nil
	}

	rs := T(new(V))
	rs.SetResourcePresetId(d.ResourcePresetId.ValueString())
	rs.SetDiskSize(datasize.ToBytes(d.DiskSize.ValueInt64()))
	if utils.IsPresent(d.DiskTypeId) {
		rs.SetDiskTypeId(d.DiskTypeId.ValueString())
	}

	return rs
}

var environments = map[string]int32{
	"ENVIRONMENT_UNSPECIFIED": 0,
	"PRODUCTION":              1,
	"PRESTABLE":               2,
}

func ExpandEnvironment[T ~int32](_ context.Context, e types.String, diags *diag.Diagnostics) T {
	if !utils.IsPresent(e) {
		return 0
	}

	v, ok := environments[e.ValueString()]
	if !ok || v == 0 {
		allowedEnvs := make([]string, 0, len(environments))
		for k, v := range environments {
			if v == 0 {
				continue
			}
			allowedEnvs = append(allowedEnvs, k)
		}

		diags.AddError(
			"Failed to parse environment",
			fmt.Sprintf("Error while parsing value for 'environment'. Value must be one of `%s`, not `%s`", strings.Join(allowedEnvs, "`, `"), e),
		)

		return 0
	}
	return T(v)
}

func ExpandLabels(ctx context.Context, labels types.Map, diags *diag.Diagnostics) map[string]string {
	var lMap map[string]string
	if utils.IsPresent(labels) {
		diags.Append(labels.ElementsAs(ctx, &lMap, false)...)
		if diags.HasError() {
			return nil
		}
	}
	return lMap
}

func ExpandClusterMaintenanceWindow[
	V any,
	VW any,
	VA any,

	WD ~int32,

	T maintenanceWindow[
		V,
		VW, VA,
		WD,
		W, A,
	],

	W weeklyMaintenanceWindow[VW, WD],
	A anytimeMaintenanceWindow[VA],
](ctx context.Context, mw types.Object, diags *diag.Diagnostics) T {
	if !utils.IsPresent(mw) {
		return *new(T)
	}

	out := T(new(V))
	var mwConf MaintenanceWindow

	diags.Append(mw.As(ctx, &mwConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return *new(T)
	}

	if mwType := mwConf.Type.ValueString(); mwType == anytimeType {
		out.SetAnytime(new(VA))
	} else if mwType == weeklyType {
		mwDay, mwHour := mwConf.Day.ValueString(), mwConf.Hour.ValueInt64()
		day := weekdayNames[mwDay]

		w := W(new(VW))
		w.SetDay(WD(day))
		w.SetHour(mwHour)
		out.SetWeeklyMaintenanceWindow(w)
	} else {
		diags.AddError(
			"Failed to expand maintenance window.",
			fmt.Sprintf("maintenance_window.type should be %s or %s", anytimeType, weeklyType),
		)
		return *new(T)
	}

	return out
}

func ExpandBoolWrapper(_ context.Context, b types.Bool, _ *diag.Diagnostics) *wrapperspb.BoolValue {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}

	return wrapperspb.Bool(b.ValueBool())
}

func ExpandStringWrapper(_ context.Context, s types.String, _ *diag.Diagnostics) *wrapperspb.StringValue {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}

	return wrapperspb.String(s.ValueString())
}

func ExpandInt64Wrapper(_ context.Context, in types.Int64, _ *diag.Diagnostics) *wrapperspb.Int64Value {
	if in.IsNull() || in.IsUnknown() {
		return nil
	}

	return wrapperspb.Int64(in.ValueInt64())
}

func ExpandDoubleWrapper(_ context.Context, d types.Float64, _ *diag.Diagnostics) *wrapperspb.DoubleValue {
	if d.IsNull() || d.IsUnknown() {
		return nil
	}

	return wrapperspb.Double(d.ValueFloat64())
}

func ExpandSecurityGroupIds(ctx context.Context, sg types.Set, diags *diag.Diagnostics) []string {
	var securityGroupIds []string
	if !(sg.IsUnknown() || sg.IsNull()) {
		securityGroupIds = make([]string, len(sg.Elements()))
		diags.Append(sg.ElementsAs(ctx, &securityGroupIds, false)...)
		if diags.HasError() {
			return nil
		}
	}

	return securityGroupIds
}

func ExpandFolderId(ctx context.Context, f types.String, providerConfig *config.State, diags *diag.Diagnostics) string {
	folderID, d := validate.FolderID(f, providerConfig)
	diags.Append(d)
	return folderID
}

func ExpandAccess[T proto.Message](ctx context.Context, cfgAccess types.Object, diags *diag.Diagnostics) T {
	getBool := func(attrs map[string]attr.Value, key string) bool {
		defer func() {
			delete(attrs, key)
		}()
		v, ok := attrs[key]
		if !ok {
			return false
		}
		boolValue, ok := v.(types.Bool)
		if !ok {
			return false
		}
		return boolValue.ValueBool()
	}

	attributes := cfgAccess.Attributes()

	var acZero T
	ac := reflect.New(reflect.TypeOf(acZero).Elem()).Interface().(T)
	accessReflect := ac.ProtoReflect()

	descriptor := accessReflect.Descriptor().Fields().ByName("yandex_query")
	if descriptor != nil {
		accessReflect.Set(descriptor, protoreflect.ValueOfBool(getBool(attributes, "yandex_query")))
	}
	descriptor = accessReflect.Descriptor().Fields().ByName("data_lens")
	if descriptor != nil {
		accessReflect.Set(descriptor, protoreflect.ValueOfBool(getBool(attributes, "data_lens")))
	}
	descriptor = accessReflect.Descriptor().Fields().ByName("data_transfer")
	if descriptor != nil {
		accessReflect.Set(descriptor, protoreflect.ValueOfBool(getBool(attributes, "data_transfer")))
	}
	descriptor = accessReflect.Descriptor().Fields().ByName("web_sql")
	if descriptor != nil {
		accessReflect.Set(descriptor, protoreflect.ValueOfBool(getBool(attributes, "web_sql")))
	}

	descriptor = accessReflect.Descriptor().Fields().ByName("serverless")
	if descriptor != nil {
		accessReflect.Set(descriptor, protoreflect.ValueOfBool(getBool(attributes, "serverless")))
	}

	if len(attributes) > 0 {
		diags.AddError("left unknown keys", fmt.Sprintf("%v", attributes))
		var s = new(T)
		return *s
	}

	return ac
}
