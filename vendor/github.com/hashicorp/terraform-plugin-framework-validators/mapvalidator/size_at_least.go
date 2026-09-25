package mapvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Map = sizeAtLeastValidator{}

// sizeAtLeastValidator validates that map contains at least min elements.
type sizeAtLeastValidator struct {
	min int
}

// Description describes the validation in plain text formatting.
func (v sizeAtLeastValidator) Description(_ context.Context) string {
	return fmt.Sprintf("map must contain at least %d elements", v.min)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v sizeAtLeastValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v sizeAtLeastValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elems := req.ConfigValue.Elements()

	if len(elems) < v.min {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("%d", len(elems)),
		))
	}
}

// SizeAtLeast returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a Map.
//   - Contains at least min elements.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func SizeAtLeast(min int) validator.Map {
	return sizeAtLeastValidator{
		min: min,
	}
}
