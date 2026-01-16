package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type JdbcBridge struct {
	Host types.String `tfsdk:"host"`
	Port types.Int64  `tfsdk:"port"`
}

var JdbcBridgeAttrTypes = map[string]attr.Type{
	"host": types.StringType,
	"port": types.Int64Type,
}

func flattenJdbcBridge(ctx context.Context, bridge *clickhouseConfig.ClickhouseConfig_JdbcBridge, diags *diag.Diagnostics) types.Object {
	if bridge == nil {
		return types.ObjectNull(JdbcBridgeAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, JdbcBridgeAttrTypes, JdbcBridge{
			Host: types.StringValue(bridge.Host),
			Port: mdbcommon.FlattenInt64Wrapper(ctx, bridge.Port, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func expandJdbcBridge(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_JdbcBridge {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var jdbcBridge JdbcBridge
	diags.Append(c.As(ctx, &jdbcBridge, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_JdbcBridge{
		Host: jdbcBridge.Host.ValueString(),
		Port: mdbcommon.ExpandInt64Wrapper(ctx, jdbcBridge.Port, diags),
	}
}
