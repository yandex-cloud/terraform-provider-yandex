package trino_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func isNullOrUnknown(v attr.Value) bool {
	return v.IsNull() || v.IsUnknown()
}

func allowedLogLevels() []string {
	allowedLevels := make([]string, 0, len(logging.LogLevel_Level_value))
	for levelName, val := range logging.LogLevel_Level_value {
		if val == 0 {
			continue
		}
		allowedLevels = append(allowedLevels, levelName)
	}
	return allowedLevels
}

func logLevelValidator() validator.String {
	return stringvalidator.OneOf(allowedLogLevels()...)
}

func mwTypeValidator() validator.String {
	return stringvalidator.OneOf("ANYTIME", "WEEKLY")
}

func mwHourValidator() validator.Int64 {
	return int64validator.Between(1, 24)
}

func mwDayValidator() validator.String {
	return stringvalidator.OneOf("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN")
}

func mwValidator() validator.Object {
	return &maintenanceWindowStructValidator{}
}

type maintenanceWindowStructValidator struct{}

func (m *maintenanceWindowStructValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var t, d types.String
	var h types.Int64

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("type"), &t)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("day"), &d)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("hour"), &h)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if t.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Failed to validate maintenance_window",
			`Field "type" should be set`,
		)
		return
	}

	if t.ValueString() == "ANYTIME" && (!d.IsNull() || !h.IsNull()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Failed to validate maintenance_window",
			`day and hour should not be set, when using ANYTIME`,
		)
		return
	}

	if t.ValueString() == "WEEKLY" && (d.IsNull() || h.IsNull()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Failed to validate maintenance_window",
			`day and hour should be set, when using WEEKLY`,
		)
	}
}

func (m *maintenanceWindowStructValidator) Description(_ context.Context) string {
	return `
		Maintenance window block validation. 
		Check block structure in general for ANYTIME and WEEKLY maintenance. 
		Attributes hour and day should be set ONLY for WEEKLY maintenance.
	`
}

func (m *maintenanceWindowStructValidator) MarkdownDescription(_ context.Context) string {
	return `
		Maintenance window block validation. 
		Check block structure in general for *ANYTIME* and *WEEKLY* maintenance. 
		Attributes hour and day should be set ONLY for *WEEKLY* maintenance.
	`
}

func workerValidator() validator.Object {
	return &workerStructValidator{}
}

type workerStructValidator struct{}

func (w *workerStructValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var fixedScale FixedScaleValue
	var autoScale AutoScaleValue

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("fixed_scale"), &fixedScale)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("auto_scale"), &autoScale)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isNullOrUnknown(fixedScale) && isNullOrUnknown(autoScale) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Failed to validate worker config",
			`Scale policy for worker must be set`,
		)
		return
	}

	if !isNullOrUnknown(fixedScale) && !isNullOrUnknown(autoScale) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Failed to validate worker config",
			`Only one scale policy can be set`,
		)
		return
	}
}

func (w *workerStructValidator) Description(_ context.Context) string {
	return `
		Worker configuration block validation. 
		Check block structure to make sure, that only one scale policy is set.
	`
}

func (w *workerStructValidator) MarkdownDescription(_ context.Context) string {
	return `
		Worker configuration block validation. 
		Check block structure to make sure, that only one scale policy is set.
	`
}
