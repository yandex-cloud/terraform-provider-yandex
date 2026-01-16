package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
)

// GraphiteRollup

type GraphiteRollup struct {
	Name              types.String `tfsdk:"name"`
	Patterns          types.List   `tfsdk:"patterns"`
	PathColumnName    types.String `tfsdk:"path_column_name"`
	TimeColumnName    types.String `tfsdk:"time_column_name"`
	ValueColumnName   types.String `tfsdk:"value_column_name"`
	VersionColumnName types.String `tfsdk:"version_column_name"`
}

var GraphiteRollupAttrTypes = map[string]attr.Type{
	"name":                types.StringType,
	"patterns":            types.ListType{ElemType: types.ObjectType{AttrTypes: PatternAttrTypes}},
	"path_column_name":    types.StringType,
	"time_column_name":    types.StringType,
	"value_column_name":   types.StringType,
	"version_column_name": types.StringType,
}

func flattenGraphiteRollup(ctx context.Context, graphiteRollup *clickhouseConfig.ClickhouseConfig_GraphiteRollup, diags *diag.Diagnostics) types.Object {
	if graphiteRollup == nil {
		return types.ObjectNull(GraphiteRollupAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, GraphiteRollupAttrTypes, GraphiteRollup{
			Name:              types.StringValue(graphiteRollup.Name),
			Patterns:          flattenListGraphiteRollupPattern(ctx, graphiteRollup.Patterns, diags),
			PathColumnName:    types.StringValue(graphiteRollup.PathColumnName),
			TimeColumnName:    types.StringValue(graphiteRollup.TimeColumnName),
			ValueColumnName:   types.StringValue(graphiteRollup.ValueColumnName),
			VersionColumnName: types.StringValue(graphiteRollup.VersionColumnName),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListGraphiteRollup(ctx context.Context, graphiteRollups []*clickhouseConfig.ClickhouseConfig_GraphiteRollup, diags *diag.Diagnostics) types.List {
	if graphiteRollups == nil {
		return types.ListNull(types.ObjectType{AttrTypes: GraphiteRollupAttrTypes})
	}

	tfGraphiteRollups := make([]types.Object, len(graphiteRollups))
	for i, gr := range graphiteRollups {
		tfGraphiteRollups[i] = flattenGraphiteRollup(ctx, gr, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: GraphiteRollupAttrTypes}, tfGraphiteRollups)
	diags.Append(d...)

	return list
}

func expandListGraphiteRollup(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_GraphiteRollup {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	result := make([]*clickhouseConfig.ClickhouseConfig_GraphiteRollup, 0, len(c.Elements()))
	rollups := make([]GraphiteRollup, 0, len(c.Elements()))
	diags.Append(c.ElementsAs(ctx, &rollups, false)...)
	if diags.HasError() {
		return nil
	}

	for _, rollup := range rollups {
		result = append(result, &clickhouseConfig.ClickhouseConfig_GraphiteRollup{
			Name:              rollup.Name.ValueString(),
			Patterns:          expandListGraphiteRollupPattern(ctx, rollup.Patterns, diags),
			PathColumnName:    rollup.PathColumnName.ValueString(),
			TimeColumnName:    rollup.TimeColumnName.ValueString(),
			ValueColumnName:   rollup.ValueColumnName.ValueString(),
			VersionColumnName: rollup.VersionColumnName.ValueString(),
		})
	}

	return result
}

// Pattern

type Pattern struct {
	Regexp     types.String `tfsdk:"regexp"`
	Function   types.String `tfsdk:"function"`
	Retentions types.List   `tfsdk:"retention"`
}

var PatternAttrTypes = map[string]attr.Type{
	"regexp":    types.StringType,
	"function":  types.StringType,
	"retention": types.ListType{ElemType: types.ObjectType{AttrTypes: RetentionAttrTypes}},
}

func flattenGraphiteRollupPattern(ctx context.Context, pattern *clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern, diags *diag.Diagnostics) types.Object {
	if pattern == nil {
		return types.ObjectNull(PatternAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, PatternAttrTypes, Pattern{
			Regexp:     types.StringValue(pattern.Regexp),
			Function:   types.StringValue(pattern.Function),
			Retentions: flattenListGraphiteRollupPatternRetention(ctx, pattern.Retention, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListGraphiteRollupPattern(ctx context.Context, patterns []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern, diags *diag.Diagnostics) types.List {
	if patterns == nil {
		return types.ListNull(types.ObjectType{AttrTypes: PatternAttrTypes})
	}

	tfPatterns := make([]types.Object, len(patterns))
	for i, p := range patterns {
		tfPatterns[i] = flattenGraphiteRollupPattern(ctx, p, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: PatternAttrTypes}, tfPatterns)
	diags.Append(d...)

	return list
}

func expandListGraphiteRollupPattern(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	result := make([]*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern, 0, len(c.Elements()))
	patterns := make([]Pattern, 0, len(c.Elements()))
	diags.Append(c.ElementsAs(ctx, &patterns, false)...)
	if diags.HasError() {
		return nil
	}

	for _, pattern := range patterns {
		result = append(result, &clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern{
			Regexp:    pattern.Regexp.ValueString(),
			Function:  pattern.Function.ValueString(),
			Retention: expandListGraphiteRollupPatternRetention(ctx, pattern.Retentions, diags),
		})
	}

	return result
}

// Retention

type Retention struct {
	Age       types.Int64 `tfsdk:"age"`
	Precision types.Int64 `tfsdk:"precision"`
}

var RetentionAttrTypes = map[string]attr.Type{
	"age":       types.Int64Type,
	"precision": types.Int64Type,
}

func flattenGraphiteRollupPatternRetention(ctx context.Context, retention *clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention, diags *diag.Diagnostics) types.Object {
	if retention == nil {
		return types.ObjectNull(RetentionAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, RetentionAttrTypes, Retention{
			Age:       types.Int64Value(retention.Age),
			Precision: types.Int64Value(retention.Precision),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListGraphiteRollupPatternRetention(ctx context.Context, retentions []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention, diags *diag.Diagnostics) types.List {
	if retentions == nil {
		return types.ListNull(types.ObjectType{AttrTypes: RetentionAttrTypes})
	}

	tfRetentions := make([]types.Object, len(retentions))
	for i, r := range retentions {
		tfRetentions[i] = flattenGraphiteRollupPatternRetention(ctx, r, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: RetentionAttrTypes}, tfRetentions)
	diags.Append(d...)

	return list
}

func expandListGraphiteRollupPatternRetention(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	result := make([]*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention, 0, len(c.Elements()))
	retentions := make([]Retention, 0, len(c.Elements()))
	diags.Append(c.ElementsAs(ctx, &retentions, false)...)
	if diags.HasError() {
		return nil
	}

	for _, retention := range retentions {
		result = append(result, &clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
			Age:       retention.Age.ValueInt64(),
			Precision: retention.Precision.ValueInt64(),
		})
	}

	return result
}
