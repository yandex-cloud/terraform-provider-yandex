package spark_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type ResourcePool struct {
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
	Size             types.Int64  `tfsdk:"size"`
	MinSize          types.Int64  `tfsdk:"min_size"`
	MaxSize          types.Int64  `tfsdk:"max_size"`
}

type ResourcePools struct {
	Driver   types.Object `tfsdk:"driver"`
	Executor types.Object `tfsdk:"executor"`
}

type Dependencies struct {
	PipPackages types.Set `tfsdk:"pip_packages"`
	DebPackages types.Set `tfsdk:"deb_packages"`
}

type HistoryServer struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type Metastore struct {
	ClusterId types.String `tfsdk:"cluster_id"`
}

var ResourcePoolAttrTypes = map[string]attr.Type{
	"resource_preset_id": types.StringType,
	"size":               types.Int64Type,
	"min_size":           types.Int64Type,
	"max_size":           types.Int64Type,
}

var ResourcePoosAttrTypes = map[string]attr.Type{
	"driver":   types.ObjectType{AttrTypes: ResourcePoolAttrTypes},
	"executor": types.ObjectType{AttrTypes: ResourcePoolAttrTypes},
}

var DependenciesAttrTypes = map[string]attr.Type{
	"pip_packages": types.SetType{ElemType: types.StringType},
	"deb_packages": types.SetType{ElemType: types.StringType},
}

var HistoryServerAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
}

var MetastoreAttrTypes = map[string]attr.Type{
	"cluster_id": types.StringType,
}

func nullableStringSliceToSet(ctx context.Context, s []string) (types.Set, diag.Diagnostics) {
	if s == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	return types.SetValueFrom(ctx, types.StringType, s)
}

func setsAreEqual(set1, set2 types.Set) bool {
	// if one of sets is null and the other is empty then we assume that they are equal
	if len(set1.Elements()) == 0 && len(set2.Elements()) == 0 {
		return true
	}
	if set1.Equal(set2) {
		return true
	}
	return false
}

func mapsAreEqual(map1, map2 types.Map) bool {
	// if one of map is null and the other is empty then we assume that they are equal
	if len(map1.Elements()) == 0 && len(map2.Elements()) == 0 {
		return true
	}
	if map1.Equal(map2) {
		return true
	}
	return false
}

func stringsAreEqual(str1, str2 types.String) bool {
	// if one of strings is null and the other is empty then we assume that they are equal
	if str1.ValueString() == "" && str2.ValueString() == "" {
		return true
	}
	if str1.Equal(str2) {
		return true
	}
	return false
}

func (v LoggingValue) isExplicitlyDisabled() bool {
	return !v.IsNull() && !v.Enabled.ValueBool()
}

func loggingValuesAreEqual(val1, val2 LoggingValue) bool {
	// if one of values is null and the other is empty then we assume that they are equal
	if (val1.isExplicitlyDisabled() && val2.IsNull()) || (val1.IsNull() && val2.isExplicitlyDisabled()) {
		return true
	}
	if val1.Equal(val2) {
		return true
	}
	return false
}

func clusterConfigsAreEqual(ctx context.Context, val1, val2 ConfigValue, diags *diag.Diagnostics) bool {
	if val1.Equal(val2) {
		return true
	}

	if val1.ResourcePools.Equal(val2.ResourcePools) && val1.HistoryServer.Equal(val2.HistoryServer) && val1.Metastore.Equal(val2.Metastore) {
		var dependenciesVal1 Dependencies
		diags.Append(val1.Dependencies.As(ctx, &dependenciesVal1, datasize.DefaultOpts)...)

		var dependenciesVal2 Dependencies
		diags.Append(val2.Dependencies.As(ctx, &dependenciesVal2, datasize.DefaultOpts)...)

		if setsAreEqual(dependenciesVal1.PipPackages, dependenciesVal2.PipPackages) && setsAreEqual(dependenciesVal1.DebPackages, dependenciesVal2.DebPackages) {
			return true
		}
	}

	return false
}

func networkConfigsAreEqual(val1, val2 NetworkValue) bool {
	if val1.IsNull() && val2.IsNull() {
		return true
	}
	if val1.IsNull() != val2.IsNull() {
		return false
	}
	if val1.Equal(val2) {
		return true
	}
	if setsAreEqual(val1.SubnetIds, val2.SubnetIds) && setsAreEqual(val1.SecurityGroupIds, val2.SecurityGroupIds) {
		return true
	}
	return false
}

func maintenanceWindowsAreEqual(val1, val2 MaintenanceWindowValue) bool {
	if val1.IsNull() && val2.IsNull() {
		return true
	}
	if val1.IsNull() != val2.IsNull() {
		return false
	}
	return val1.Equal(val2)
}

func maintenanceWindowDefault() defaults.Object {
	return objectdefault.StaticValue(
		types.ObjectValueMust(
			map[string]attr.Type{
				"type": types.StringType,
				"day":  types.StringType,
				"hour": types.Int64Type,
			},
			map[string]attr.Value{
				"type": types.StringValue("ANYTIME"),
				"day":  types.StringNull(),
				"hour": types.Int64Null(),
			},
		),
	)
}

func dependenciesDefault() defaults.Object {
	return objectdefault.StaticValue(
		types.ObjectValueMust(
			DependenciesAttrTypes,
			map[string]attr.Value{
				"pip_packages": types.SetValueMust(types.StringType, []attr.Value{}),
				"deb_packages": types.SetValueMust(types.StringType, []attr.Value{}),
			},
		),
	)
}

func historyServerDefault() defaults.Object {
	return objectdefault.StaticValue(
		types.ObjectValueMust(
			HistoryServerAttrTypes,
			map[string]attr.Value{
				"enabled": types.BoolValue(true),
			},
		),
	)
}

func metastoreDefault() defaults.Object {
	return objectdefault.StaticValue(
		types.ObjectValueMust(
			MetastoreAttrTypes,
			map[string]attr.Value{
				"cluster_id": types.StringValue(""),
			},
		),
	)
}

func stringSetDefault() defaults.Set {
	return setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{}))
}
