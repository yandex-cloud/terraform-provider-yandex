package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

type MonitoringModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Link        types.String `tfsdk:"link"`
}

var MonitoringAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"description": types.StringType,
	"link":        types.StringType,
}

func FlattenListMonitoring(ctx context.Context, monitoring []*clickhouse.Monitoring, diags *diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: MonitoringAttrTypes}

	if len(monitoring) == 0 {
		return types.ListNull(elemType)
	}

	items := make([]attr.Value, 0, len(monitoring))
	for _, m := range monitoring {
		obj, d := types.ObjectValueFrom(ctx, MonitoringAttrTypes, MonitoringModel{
			Name:        types.StringValue(m.Name),
			Description: types.StringValue(m.Description),
			Link:        types.StringValue(m.Link),
		})
		diags.Append(d...)
		items = append(items, obj)
	}

	result, d := types.ListValue(elemType, items)
	diags.Append(d...)
	return result
}
