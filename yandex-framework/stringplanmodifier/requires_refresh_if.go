package stringplanmodifier

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RequiresRefreshIf returns a plan modifier that conditionally requires
// attribute refreshment if:
//
//   - The resource is planned for update.
//   - The given function returns true. Returning false will not unset any
//     prior resource refreshment.
func RequiresRefreshIf(f RequiresRefreshIfFunc, description, markdownDescription string) planmodifier.String {
	return requiresRefreshIfModifier{
		ifFunc:              f,
		description:         description,
		markdownDescription: markdownDescription,
	}
}

// requiresRefreshIfModifier is a plan modifier that sets RequiresRefresh
// on the attribute if a given function is true.
type requiresRefreshIfModifier struct {
	ifFunc              RequiresRefreshIfFunc
	description         string
	markdownDescription string
}

// Description returns a human-readable description of the plan modifier.
func (m requiresRefreshIfModifier) Description(_ context.Context) string {
	return m.description
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m requiresRefreshIfModifier) MarkdownDescription(_ context.Context) string {
	return m.markdownDescription
}

// PlanModifyString implements the plan modification logic.
func (m requiresRefreshIfModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do not refresh on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not refresh on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	ifFuncResp := &RequiresRefreshIfFuncResponse{}

	m.ifFunc(ctx, req, ifFuncResp)

	resp.Diagnostics.Append(ifFuncResp.Diagnostics...)
	if ifFuncResp.RequiresRefresh {
		resp.PlanValue = types.StringUnknown()
	}
}
