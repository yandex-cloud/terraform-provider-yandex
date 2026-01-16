package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
)

type Macro struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

var MacroAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}

func flattenMacro(ctx context.Context, macro *clickhouseConfig.ClickhouseConfig_Macro, diags *diag.Diagnostics) types.Object {
	if macro == nil {
		return types.ObjectNull(MacroAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, MacroAttrTypes, Macro{
			Name:  types.StringValue(macro.Name),
			Value: types.StringValue(macro.Value),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListMacro(ctx context.Context, macros []*clickhouseConfig.ClickhouseConfig_Macro, diags *diag.Diagnostics) types.List {
	if macros == nil {
		return types.ListNull(types.ObjectType{AttrTypes: MacroAttrTypes})
	}

	tfMacros := make([]types.Object, len(macros))
	for i, m := range macros {
		tfMacros[i] = flattenMacro(ctx, m, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MacroAttrTypes}, tfMacros)
	diags.Append(d...)

	return list
}

func expandListMacro(ctx context.Context, list types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_Macro {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	result := make([]*clickhouseConfig.ClickhouseConfig_Macro, 0, len(list.Elements()))
	macros := make([]Macro, 0, len(list.Elements()))
	diags.Append(list.ElementsAs(ctx, &macros, false)...)
	if diags.HasError() {
		return nil
	}

	for _, macro := range macros {
		result = append(result, &clickhouseConfig.ClickhouseConfig_Macro{
			Name:  macro.Name.ValueString(),
			Value: macro.Value.ValueString(),
		})
	}

	return result
}
