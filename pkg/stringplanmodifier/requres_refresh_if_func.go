package stringplanmodifier

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// RequiresRefreshIfFunc is a conditional function used in the RequiresRefreshIf
// plan modifier to determine whether the attribute requires refreshment.
type RequiresRefreshIfFunc func(context.Context, planmodifier.StringRequest, *RequiresRefreshIfFuncResponse)

// RequiresRefreshIfFuncResponse is the response type for a RequiresRefreshIfFunc.
type RequiresRefreshIfFuncResponse struct {
	// Diagnostics report errors or warnings related to this logic. An empty
	// or unset slice indicates success, with no warnings or errors generated.
	Diagnostics diag.Diagnostics

	// RequiresRefresh should be enabled if the resource should be refreshed.
	RequiresRefresh bool
}
