package validate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

/*---------------------------------------------------------------------------------------------*/

var _ validator.Int64 = &int64GreaterValidator{}

type int64GreaterValidator struct {
	expressions path.Expressions
}

func (v *int64GreaterValidator) Description(_ context.Context) string {
	return fmt.Sprintf("If configured, must be greater than %s attributes", v.expressions)
}

func (v *int64GreaterValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *int64GreaterValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {

	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(v.expressions...)

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, matchedPath := range matchedPaths {
			var matchedPathValue attr.Value

			diags := req.Config.GetAttribute(ctx, matchedPath, &matchedPathValue)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if matchedPathValue.IsNull() || matchedPathValue.IsUnknown() {
				continue
			}

			var matchedPathConfig types.Int64
			diags = tfsdk.ValueAs(ctx, matchedPathValue, &matchedPathConfig)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if matchedPathConfig.ValueInt64() >= req.ConfigValue.ValueInt64() {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid Attribute Value",
					fmt.Sprintf("%s must be greater than %s value: %d", req.Path, matchedPath.String(), matchedPathConfig.ValueInt64()),
				)

				return
			}
		}
	}
}

func Int64GreaterValidator(expressions ...path.Expression) *int64GreaterValidator {
	return &int64GreaterValidator{
		expressions: expressions,
	}
}
