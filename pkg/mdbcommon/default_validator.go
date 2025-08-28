package mdbcommon

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stringToTimeValidator struct{}

func NewStringToTimeValidator() *stringToTimeValidator {
	return &stringToTimeValidator{}
}

func (m *stringToTimeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var t types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path, &t)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := ParseStringToTime(t.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Failed to validate time",
			fmt.Sprintf(`Cant cast string %s to time: %v`, t.ValueString(), err),
		)
	}

}

func (m *stringToTimeValidator) Description(_ context.Context) string {
	return `
		Time string format validation.
		Check that string in a right format (e.g. "2006-01-02T15:04:05") and can be cast to Time.
	`
}

func (m *stringToTimeValidator) MarkdownDescription(_ context.Context) string {
	return `
		Time string format validation.
		Check that string in a right format (e.g. "2006-01-02T15:04:05") and can be cast to Time.
	`
}
