package trino_catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func isNullOrUnknown(v attr.Value) bool {
	return v.IsNull() || v.IsUnknown()
}

func onlyOneOptionValidator(name string, attributes ...string) validator.Object {
	return &onlyOneOptionSetStructValidator{
		attributes: attributes,
		name:       name,
	}
}

type onlyOneOptionSetStructValidator struct {
	name       string
	attributes []string
}

func (o *onlyOneOptionSetStructValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	count := 0
	for _, name := range o.attributes {
		val, ok := req.ConfigValue.Attributes()[name]
		if ok && !isNullOrUnknown(val) {
			count++
		}
	}

	if count == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("Failed to validate %s config", o.name),
			fmt.Sprintf("One of [%s] must be set", strings.Join(o.attributes, ", ")),
		)
		return
	}

	if count > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("Failed to validate %s config", o.name),
			fmt.Sprintf("Only one of [%s] can be set", strings.Join(o.attributes, ", ")),
		)
		return
	}
}

func (o *onlyOneOptionSetStructValidator) Description(_ context.Context) string {
	return fmt.Sprintf(`
		%s configuration block validation. 
		Check block structure to make sure, that only one option is set.
	`, o.name)
}

func (o *onlyOneOptionSetStructValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf(`
		%s configuration block validation. 
		Check block structure to make sure, that only one option is set.
	`, o.name)
}
