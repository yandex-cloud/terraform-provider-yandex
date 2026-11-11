package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type PerformanceDiagnostics struct {
	Enabled                  types.Bool   `tfsdk:"enabled"`
	ProcessesRefreshInterval types.String `tfsdk:"processes_refresh_interval"`
}

var PerformanceDiagnosticsAttrTypes = map[string]attr.Type{
	"enabled":                    types.BoolType,
	"processes_refresh_interval": types.StringType,
}

func FlattenPerformanceDiagnostics(ctx context.Context, pd *clickhouse.PerformanceDiagnostics, diags *diag.Diagnostics) types.Object {
	if pd == nil {
		return types.ObjectNull(PerformanceDiagnosticsAttrTypes)
	}

	var interval types.String
	if pd.ProcessesRefreshInterval == nil {
		interval = types.StringNull()
	} else {
		interval = types.StringValue(pd.ProcessesRefreshInterval.AsDuration().String())
	}

	obj, d := types.ObjectValueFrom(
		ctx, PerformanceDiagnosticsAttrTypes, PerformanceDiagnostics{
			Enabled:                  mdbcommon.FlattenBoolWrapper(ctx, pd.Enabled, diags),
			ProcessesRefreshInterval: interval,
		})
	diags.Append(d...)

	return obj
}

func ExpandPerformanceDiagnostics(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *clickhouse.PerformanceDiagnostics {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var pd PerformanceDiagnostics
	diags.Append(obj.As(ctx, &pd, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	interval := converter.ParseDuration(pd.ProcessesRefreshInterval.ValueString(), diags)
	if diags.HasError() {
		return nil
	}

	return &clickhouse.PerformanceDiagnostics{
		Enabled:                  mdbcommon.ExpandBoolWrapper(ctx, pd.Enabled, diags),
		ProcessesRefreshInterval: interval,
	}
}
