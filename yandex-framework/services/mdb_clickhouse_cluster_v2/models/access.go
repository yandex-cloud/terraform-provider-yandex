package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type Access struct {
	DataLens     types.Bool `tfsdk:"data_lens"`
	WebSql       types.Bool `tfsdk:"web_sql"`
	Metrika      types.Bool `tfsdk:"metrika"`
	Serverless   types.Bool `tfsdk:"serverless"`
	DataTransfer types.Bool `tfsdk:"data_transfer"`
	YandexQuery  types.Bool `tfsdk:"yandex_query"`
}

var AccessAttrTypes = map[string]attr.Type{
	"data_lens":     types.BoolType,
	"web_sql":       types.BoolType,
	"metrika":       types.BoolType,
	"serverless":    types.BoolType,
	"data_transfer": types.BoolType,
	"yandex_query":  types.BoolType,
}

func FlattenAccess(ctx context.Context, access *clickhouse.Access, diags *diag.Diagnostics) types.Object {
	if access == nil {
		return types.ObjectNull(AccessAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, AccessAttrTypes, Access{
			DataLens:     types.BoolValue(access.DataLens),
			WebSql:       types.BoolValue(access.WebSql),
			Metrika:      types.BoolValue(access.Metrika),
			Serverless:   types.BoolValue(access.Serverless),
			DataTransfer: types.BoolValue(access.DataTransfer),
			YandexQuery:  types.BoolValue(access.YandexQuery),
		},
	)
	diags.Append(d...)

	return obj
}

func ExpandAccess(ctx context.Context, access types.Object, diags *diag.Diagnostics) *clickhouse.Access {
	if access.IsNull() || access.IsUnknown() {
		return nil
	}

	var obj Access
	if diags.Append(access.As(ctx, &obj, datasize.DefaultOpts)...); diags.HasError() {
		return nil
	}

	return &clickhouse.Access{
		DataLens:     obj.DataLens.ValueBool(),
		WebSql:       obj.WebSql.ValueBool(),
		Metrika:      obj.Metrika.ValueBool(),
		Serverless:   obj.Serverless.ValueBool(),
		DataTransfer: obj.DataTransfer.ValueBool(),
		YandexQuery:  obj.YandexQuery.ValueBool(),
	}
}
