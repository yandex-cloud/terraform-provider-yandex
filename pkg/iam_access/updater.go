package iam_access

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

type Policy struct {
	Bindings []*access.AccessBinding
}

type PolicyDelta struct {
	Deltas []*access.AccessBindingDelta
}

type ResourceIamUpdater interface {
	// GetResourceIamPolicy Fetch the existing IAM policy attached to a resource.
	GetResourceIamPolicy(ctx context.Context) (*Policy, error)

	// SetResourceIamPolicy Replaces the existing IAM Policy attached to a resource.
	// Useful for `iam_binding` and `iam_policy` resources
	SetResourceIamPolicy(ctx context.Context, policy *Policy) error

	// UpdateResourceIamPolicy Updates the existing IAM Policy attached to a resource.
	// Useful for `iam_member` resources
	UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error

	// Initialize Initializing resource from given state
	Initialize(ctx context.Context, state Extractable, diag *diag.Diagnostics)

	// Configure Configurate resource with provider data etc.
	Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse)
}

type Extractable interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

type Settable interface {
	SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics
	RemoveResource(ctx context.Context)
}

func GetResourceIamBindingsFromState(ctx context.Context, state Extractable, diag *diag.Diagnostics) []*access.AccessBinding {
	var role types.String
	var members types.Set

	diag.Append(state.GetAttribute(ctx, path.Root("role"), &role)...)
	diag.Append(state.GetAttribute(ctx, path.Root("members"), &members)...)

	membersString := make([]string, 0, len(members.Elements()))
	diag.Append(members.ElementsAs(ctx, &membersString, false)...)

	result := make([]*access.AccessBinding, len(membersString))

	for i, member := range membersString {
		result[i] = roleMemberToAccessBinding(role.ValueString(), member)
	}
	return result
}

func GetResourceIamMemberFromState(ctx context.Context, state Extractable, diag *diag.Diagnostics) *access.AccessBinding {
	var role types.String
	var member types.String

	diag.Append(state.GetAttribute(ctx, path.Root("role"), &role)...)
	diag.Append(state.GetAttribute(ctx, path.Root("member"), &member)...)

	if !strings.ContainsRune(member.ValueString(), ':') {
		return nil
	}

	return roleMemberToAccessBinding(role.ValueString(), member.ValueString())
}
