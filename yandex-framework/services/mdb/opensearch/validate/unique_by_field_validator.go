package validate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// UniqueByField validates that the list does not contain two or more items with same "field" value.
func UniqueByField[T any](fieldName string, getter func(T) string) validator.List {
	return uniqueByFieldValidator[T]{
		fieldName: fieldName,
		getter:    getter,
	}
}

// uniqueByFieldValidator validates that the list does not contain two or more items with same "field" value.
type uniqueByFieldValidator[T any] struct {
	validator.List
	fieldName string
	getter    func(T) string
}

// Description describes the validation in plain text formatting.
func (v uniqueByFieldValidator[T]) Description(_ context.Context) string {
	return fmt.Sprintf("list must contain unique elements by '%s'", v.fieldName)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v uniqueByFieldValidator[T]) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v uniqueByFieldValidator[T]) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := make([]T, 0, len(req.ConfigValue.Elements()))
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &elements, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := make(map[string]interface{}, len(elements))
	for _, elem := range elements {
		value := v.getter(elem)
		if _, ok := values[value]; ok {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("dublicated value is '%s'", value),
			))
		} else {
			values[value] = struct{}{}
		}
	}
}
