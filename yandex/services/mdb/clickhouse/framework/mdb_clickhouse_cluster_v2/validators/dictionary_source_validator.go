package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// DictionarySourceValidator requires exactly one external dictionary source.
type DictionarySourceValidator struct{}

func (DictionarySourceValidator) Description(context.Context) string {
	return "Exactly one external dictionary source must be specified."
}

func (v DictionarySourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v DictionarySourceValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	count := 0
	unknown := false
	for _, value := range req.ConfigValue.Attributes() {
		if value.IsUnknown() {
			unknown = true
		} else if !value.IsNull() {
			count++
		}
	}
	if count > 1 || (count == 0 && !unknown) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid dictionary source", v.Description(ctx))
	}
}
