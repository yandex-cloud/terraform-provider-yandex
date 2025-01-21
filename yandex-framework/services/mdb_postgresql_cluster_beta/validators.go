package mdb_postgresql_cluster_beta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = &maintenanceWindowStructValidator{}

type maintenanceWindowStructValidator struct{}

func NewMaintenanceWindowStructValidator() *maintenanceWindowStructValidator {
	return &maintenanceWindowStructValidator{}
}

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
