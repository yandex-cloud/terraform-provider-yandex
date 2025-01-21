package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func prepareMaintenanceWindow(ctx context.Context, m *model.OpenSearch) (*opensearch.MaintenanceWindow, diag.Diagnostics) {
	mw, diags := model.ParseMaintenanceWindow(ctx, m)
	if diags.HasError() {
		return nil, diags
	}

	result := &opensearch.MaintenanceWindow{}

	switch mw.Type.ValueString() {
	case "ANYTIME":
		if mw.Day.ValueString() != "" || !(mw.Hour.IsUnknown() || mw.Hour.IsNull()) {
			diags.Append(diag.NewErrorDiagnostic(
				"Failed to parse OpenSearch maintenance window",
				"Error while parsing value for 'maintenance_window'. With ANYTIME type of maintenance window both DAY and HOUR should be omitted"))
			return nil, diags
		}

		result.SetAnytime(&opensearch.AnytimeMaintenanceWindow{})
	case "WEEKLY":
		weekly := &opensearch.WeeklyMaintenanceWindow{}
		if mw.Day.ValueString() != "" {
			day, d := toWeekDay(mw.Day)
			diags.Append(d)
			if diags.HasError() {
				return nil, diags
			}
			weekly.Day = day
		}

		if !(mw.Hour.IsUnknown() || mw.Hour.IsNull()) {
			weekly.Hour = mw.Hour.ValueInt64()
		}

		result.SetWeeklyMaintenanceWindow(weekly)
	default:
		diags.Append(diag.NewErrorDiagnostic(
			"Failed to parse OpenSearch maintenance window",
			fmt.Sprintf("Error while parsing value for 'maintenance_window'. Unknown type '%s'", mw.Type.ValueString())))
		return nil, diags
	}

	return result, diags
}

func toWeekDay(e basetypes.StringValue) (opensearch.WeeklyMaintenanceWindow_WeekDay, diag.Diagnostic) {
	v, ok := opensearch.WeeklyMaintenanceWindow_WeekDay_value[e.ValueString()]
	if !ok || v == 0 {
		allowedDays := make([]string, 0, len(opensearch.WeeklyMaintenanceWindow_WeekDay_value))
		for k, v := range opensearch.WeeklyMaintenanceWindow_WeekDay_value {
			if v == 0 {
				continue
			}
			allowedDays = append(allowedDays, k)
		}

		return 0, diag.NewErrorDiagnostic(
			"Failed to parse OpenSearch maintenance window",
			fmt.Sprintf("Error while parsing value for 'maintenance_window'. Value for 'day' should be one of `%s`, not `%s`", strings.Join(allowedDays, "`, `"), e),
		)
	}
	return opensearch.WeeklyMaintenanceWindow_WeekDay(v), nil
}
