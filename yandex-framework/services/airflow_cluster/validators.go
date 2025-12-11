package airflow_cluster

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func airflowConfigValidator() validator.Map {
	return mapvalidator.KeysAre(stringvalidator.RegexMatches(
		regexp.MustCompile(`^[^\.]*$`),
		"must not contain dots",
	))
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

type dagProcessorStructValidator struct{}

func dagProcessorValidator() validator.Object {
	return &dagProcessorStructValidator{}
}

func (d *dagProcessorStructValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	var airflowVersion types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("airflow_version"), &airflowVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if airflowVersion.IsNull() || airflowVersion.IsUnknown() {
		return
	}

	if strings.HasPrefix(airflowVersion.ValueString(), "2.") {
		if req.ConfigValue.IsNull() {
			return
		}
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"dag_processor not supported",
			"dag_processor configuration should not be specified for Airflow 2.x",
		)
	} else {
		if !req.ConfigValue.IsNull() {
			return
		}
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"dag_processor should be specified",
			"dag_processor configuration should be specified for Airflow version 3.0 and above",
		)
	}

}

func (d *dagProcessorStructValidator) Description(_ context.Context) string {
	return "dag_processor configuration should only be specified for Airflow 3.0 and above and should not be specified for Airflow 2.x."
}

func (d *dagProcessorStructValidator) MarkdownDescription(_ context.Context) string {
	return "dag_processor configuration should only be specified for Airflow 3.0 and above and should not be specified for Airflow 2.x"
}

type codeSyncStructValidator struct{}

func codeSyncValidator() validator.Object { return &codeSyncStructValidator{} }

func (c *codeSyncStructValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	var (
		s3Value      S3Value
		gitSyncValue GitSyncValue
	)

	response.Diagnostics.Append(request.Config.GetAttribute(ctx, request.Path.AtName("s3"), &s3Value)...)
	response.Diagnostics.Append(request.Config.GetAttribute(ctx, request.Path.AtName("git_sync"), &gitSyncValue)...)

	if response.Diagnostics.HasError() {
		return
	}

	if (s3Value.IsNull() && gitSyncValue.IsNull()) ||
		(!s3Value.IsNull() && !gitSyncValue.IsNull()) {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid code_sync configuration",
			"The code_sync configuration requires one and only one parameter: 's3' or 'git_sync'.",
		)
		return
	}
}

func (c *codeSyncStructValidator) Description(_ context.Context) string {
	return "code_sync configuration must contains one of 's3' or 'git_sync' and only one of them"
}

func (c *codeSyncStructValidator) MarkdownDescription(_ context.Context) string {
	return "code_sync configuration must contains one of 's3' or 'git_sync' and only one of them"
}
