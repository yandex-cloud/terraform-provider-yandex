package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
)

type QueryMaskingRule struct {
	Name    types.String `tfsdk:"name"`
	Regexp  types.String `tfsdk:"regexp"`
	Replace types.String `tfsdk:"replace"`
}

var QueryMaskingRuleAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"regexp":  types.StringType,
	"replace": types.StringType,
}

func flattenQueryMaskingRule(ctx context.Context, rule *clickhouseConfig.ClickhouseConfig_QueryMaskingRule, diags *diag.Diagnostics) types.Object {
	if rule == nil {
		return types.ObjectNull(QueryMaskingRuleAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, QueryMaskingRuleAttrTypes, QueryMaskingRule{
			Name:    types.StringValue(rule.Name),
			Regexp:  types.StringValue(rule.Regexp),
			Replace: types.StringValue(rule.Replace),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListQueryMaskingRule(ctx context.Context, rules []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule, diags *diag.Diagnostics) types.List {
	if rules == nil {
		return types.ListNull(types.ObjectType{AttrTypes: QueryMaskingRuleAttrTypes})
	}

	tfRules := make([]types.Object, len(rules))
	for i, r := range rules {
		tfRules[i] = flattenQueryMaskingRule(ctx, r, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: QueryMaskingRuleAttrTypes}, tfRules)
	diags.Append(d...)

	return list
}

func expandListQueryMaskingRule(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	result := make([]*clickhouseConfig.ClickhouseConfig_QueryMaskingRule, 0, len(c.Elements()))
	rules := make([]QueryMaskingRule, 0, len(c.Elements()))
	diags.Append(c.ElementsAs(ctx, &rules, false)...)
	if diags.HasError() {
		return nil
	}

	for _, rule := range rules {
		result = append(result, &clickhouseConfig.ClickhouseConfig_QueryMaskingRule{
			Name:    rule.Name.ValueString(),
			Regexp:  rule.Regexp.ValueString(),
			Replace: rule.Replace.ValueString(),
		})
	}

	return result
}
