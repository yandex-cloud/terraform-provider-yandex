package validate

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// UniqueCaseInsensitive validates that the set does not contain two or more equals strings
func UniqueCaseInsensitive() validator.Set {
	return uniqueCaseInsensitive{}
}

// uniqueCaseInsensitive validates that the set does not contain two or more equals strings
type uniqueCaseInsensitive struct {
	validator.Set
}

// Description describes the validation in plain text formatting.
func (v uniqueCaseInsensitive) Description(_ context.Context) string {
	return "set must contain unique elements"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v uniqueCaseInsensitive) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v uniqueCaseInsensitive) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := make([]string, 0, len(req.ConfigValue.Elements()))
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &elements, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := make(map[string]interface{}, len(elements))
	for _, elem := range elements {
		value := strings.ToLower(elem)
		if _, ok := values[value]; ok {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("dublicated value is '%s'", elem),
			))
		} else {
			values[value] = struct{}{}
		}
	}
}
