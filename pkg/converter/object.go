package converter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func ExpandObject[T any](ctx context.Context, obj types.Object, target T, diags *diag.Diagnostics) any {
	if !obj.IsNull() && !obj.IsUnknown() {
		diags.Append(obj.As(ctx, &target, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	}
	return target
}
