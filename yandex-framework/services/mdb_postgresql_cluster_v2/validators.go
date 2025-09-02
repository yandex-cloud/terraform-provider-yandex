package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

// atLeastIfConfiguredValidator verifies that at least one value from the expressions is specified in the configuration if the parent field is also specified.
type atLeastIfConfiguredValidator struct {
	PathExpressions []path.Expression
}

func NewAtLeastIfConfiguredValidator(expressions ...path.Expression) *atLeastIfConfiguredValidator {
	return &atLeastIfConfiguredValidator{
		PathExpressions: expressions,
	}
}

func (at *atLeastIfConfiguredValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(at.PathExpressions...)

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			var mpVal attr.Value
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if !mpVal.IsNull() && !mpVal.IsUnknown() {
				return
			}
		}
	}

	resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
		req.Path,
		fmt.Sprintf("At least one attribute out of %s must be configured", expressions),
	))
}

func (at *atLeastIfConfiguredValidator) Description(_ context.Context) string {
	return "At least one of the nested attributes should be configured"
}

func (at *atLeastIfConfiguredValidator) MarkdownDescription(_ context.Context) string {
	return "At least one of the nested attributes should be configured"
}
