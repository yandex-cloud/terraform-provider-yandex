package adapters

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type Adapter interface {
	Fill(ctx context.Context, target any, attributes map[string]attr.Value, diags *diag.Diagnostics)
	Extract(ctx context.Context, src any, diags *diag.Diagnostics) map[string]attr.Value
}
