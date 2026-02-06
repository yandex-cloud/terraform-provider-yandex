package iam_policy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

// ResourceIamPolicyBindingUpdater manages IAM Policy Bindings for a resource
type ResourceIamPolicyBindingUpdater interface {
	// GetResourcePolicyBindings fetches the existing IAM policy bindings attached to a resource
	GetResourcePolicyBindings(ctx context.Context) ([]*access.AccessPolicyBinding, error)

	// BindAccessPolicy binds an access policy to the resource
	BindAccessPolicy(ctx context.Context, binding *access.AccessPolicyBinding) error

	// UnbindAccessPolicy unbinds an access policy from the resource
	UnbindAccessPolicy(ctx context.Context, accessPolicyTemplateID string) error

	// UpdateAccessPolicyBindingParameters updates parameters of an existing policy binding (optional)
	UpdateAccessPolicyBindingParameters(ctx context.Context, binding *access.AccessPolicyBinding) error
}

type Extractable interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

type Settable interface {
	SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics
	RemoveResource(ctx context.Context)
}

func GetResourceIamPolicyBindingFromState(ctx context.Context, state Extractable, diag *diag.Diagnostics) *access.AccessPolicyBinding {
	var templateID types.String
	var params types.Map

	diag.Append(state.GetAttribute(ctx, path.Root("access_policy_template_id"), &templateID)...)
	diag.Append(state.GetAttribute(ctx, path.Root("parameters"), &params)...)

	paramsParsed := make(map[string]string, len(params.Elements()))
	diag.Append(params.ElementsAs(ctx, &paramsParsed, false)...)

	return &access.AccessPolicyBinding{
		AccessPolicyTemplateId: templateID.ValueString(),
		Parameters:             paramsParsed,
	}
}

// CalculatePolicyBindingChanges compares current and desired policy bindings
// Returns: bindings to add, template IDs to remove, bindings to update
func CalculatePolicyBindingChanges(current, desired []*access.AccessPolicyBinding) (
	toAdd []*access.AccessPolicyBinding,
	toRemove []string,
	toUpdate []*access.AccessPolicyBinding,
) {
	currentMap := make(map[string]*access.AccessPolicyBinding)
	for _, binding := range current {
		currentMap[binding.AccessPolicyTemplateId] = binding
	}

	desiredMap := make(map[string]*access.AccessPolicyBinding)
	for _, binding := range desired {
		desiredMap[binding.AccessPolicyTemplateId] = binding
	}

	// Find bindings to add or update
	for templateID, desiredBinding := range desiredMap {
		if currentBinding, exists := currentMap[templateID]; !exists {
			// New binding
			toAdd = append(toAdd, desiredBinding)
		} else if !parametersEqual(currentBinding.Parameters, desiredBinding.Parameters) {
			// Existing binding with different parameters
			toUpdate = append(toUpdate, desiredBinding)
		}
	}

	// Find bindings to remove
	for templateID := range currentMap {
		if _, exists := desiredMap[templateID]; !exists {
			toRemove = append(toRemove, templateID)
		}
	}

	return toAdd, toRemove, toUpdate
}
