package mapvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Map = sizeAtMostValidator{}

// sizeAtMostValidator validates that map contains at most max elements.
type sizeAtMostValidator struct {
	max int
}

// Description describes the validation in plain text formatting.
func (v sizeAtMostValidator) Description(_ context.Context) string {
	return fmt.Sprintf("map must contain at most %d elements", v.max)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v sizeAtMostValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v sizeAtMostValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elems := req.ConfigValue.Elements()

	if len(elems) > v.max {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("%d", len(elems)),
		))
	}
}

// SizeAtMost returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a Map.
//   - Contains at most max elements.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func SizeAtMost(max int) validator.Map {
	return sizeAtMostValidator{
		max: max,
	}
}
