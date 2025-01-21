package model

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type MaintenanceWindow struct {
	Type types.String `tfsdk:"type"`
	Day  types.String `tfsdk:"day"`
	Hour types.Int64  `tfsdk:"hour"`
}

var MaintenanceWindowAttrTypes = map[string]attr.Type{
	"type": types.StringType,
	"day":  types.StringType,
	"hour": types.Int64Type,
}

func maintenanceWindowToObject(ctx context.Context, mw *opensearch.MaintenanceWindow) (types.Object, diag.Diagnostics) {
	var res basetypes.ObjectValue
	var diags diag.Diagnostics
	if val := mw.GetAnytime(); val != nil {
		res, diags = types.ObjectValueFrom(ctx, MaintenanceWindowAttrTypes, MaintenanceWindow{
			Type: types.StringValue("ANYTIME"),
		})
	}

	if val := mw.GetWeeklyMaintenanceWindow(); val != nil {
		res, diags = types.ObjectValueFrom(ctx, MaintenanceWindowAttrTypes, MaintenanceWindow{
			Type: types.StringValue("WEEKLY"),
			Day:  types.StringValue(val.GetDay().String()),
			Hour: types.Int64Value(val.GetHour()),
		})
	}

	if diags.HasError() {
		return types.ObjectUnknown(MaintenanceWindowAttrTypes), diags
	}

	return res, diags
}

func ParseMaintenanceWindow(ctx context.Context, model *OpenSearch) (*MaintenanceWindow, diag.Diagnostics) {
	res := &MaintenanceWindow{}
	diags := model.MaintenanceWindow.As(ctx, res, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}
