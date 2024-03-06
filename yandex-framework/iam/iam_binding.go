package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

type bindingResource struct {
	ResourceUpdater ResourceIamUpdater
}

func NewIamBinding(updater ResourceIamUpdater) resource.Resource {
	return &bindingResource{updater}
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.ResourceUpdater.Initialize(ctx, req.Plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	policies := getResourceIamBindings(ctx, req.Plan, &resp.Diagnostics)
	err := iamPolicyReadModifySet(ctx, r.ResourceUpdater, func(ep *Policy) error {
		// Creating a binding does not remove existing members if they are not in the provided members list.
		// This prevents removing existing permission without the user's knowledge.
		// Instead, a diff is shown in that case after creation. Subsequent calls to update will remove any
		// existing members not present in the provided list.
		ep.Bindings = mergeBindings(append(ep.Bindings, policies...))
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Attach Resource Policies",
			fmt.Sprintf("An unexpected error occurred while attempting to attach resource policies"+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	r.RefreshBindingState(ctx, req.Plan, &resp.State, resp.Diagnostics)
}

func (r *bindingResource) RefreshBindingState(ctx context.Context, req Extractable, resp Settable, diag diag.Diagnostics) {
	var role types.String
	diag.Append(req.GetAttribute(ctx, path.Root("role"), &role)...)

	eBindings := getResourceIamBindings(ctx, req, &diag)

	policy, err := r.ResourceUpdater.GetResourceIamPolicy(ctx)
	if err != nil {
		diag.AddError(
			"Unable to Refresh Resource Policies",
			fmt.Sprintf("An unexpected error occurred while refreshing resource policies"+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err))
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Retrieved access bindings of %s: %+v", r.ResourceUpdater.DescribeResource(), policy))

	var mBindings []*access.AccessBinding
	for _, b := range policy.Bindings {
		if b.RoleId != role.ValueString() {
			continue
		}
		if len(eBindings) != 0 {
			for _, e := range eBindings {
				if canonicalMember(e) != canonicalMember(b) {
					continue
				}
				mBindings = append(mBindings, b)
			}
		} else {
			mBindings = append(mBindings, b)
		}
	}

	if len(mBindings) == 0 {
		diag.AddError(
			"Unable to Refresh Resource Policies",
			fmt.Sprintf("An unexpected error occurred while refreshing resource policies"+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err))
		return
	}

	mBindingsSet, diags := types.SetValueFrom(ctx, types.StringType, roleToMembersList(role.ValueString(), mBindings))
	diag.Append(diags...)
	diag.Append(resp.SetAttribute(ctx, path.Root("members"), mBindingsSet)...)
	diag.Append(resp.SetAttribute(ctx, path.Root("role"), role)...)
	diag.Append(resp.SetAttribute(ctx, path.Root(r.ResourceUpdater.GetIdAlias()), r.ResourceUpdater.GetId())...)
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.ResourceUpdater.Initialize(ctx, req.State, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	r.RefreshBindingState(ctx, req.State, &resp.State, resp.Diagnostics)
	return
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.ResourceUpdater.Initialize(ctx, req.Plan, &resp.Diagnostics)
	bindings := getResourceIamBindings(ctx, req.Plan, &resp.Diagnostics)

	var stateRole types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("role"), &stateRole)...)

	err := iamPolicyReadModifySet(ctx, r.ResourceUpdater, func(p *Policy) error {
		p.Bindings = removeRoleFromBindings(stateRole.ValueString(), p.Bindings)
		p.Bindings = append(p.Bindings, bindings...)
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource Policies",
			fmt.Sprintf("An unexpected error occurred while updating resource policies"+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err))
		return
	}
	r.RefreshBindingState(ctx, req.Plan, &resp.State, resp.Diagnostics)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.ResourceUpdater.Initialize(ctx, req.State, &resp.Diagnostics)

	binding := getResourceIamBindings(ctx, req.State, &resp.Diagnostics)
	if len(binding) == 0 {
		tflog.Debug(ctx,
			fmt.Sprintf(
				"Resource %s is missing or deleted, marking policy binding as deleted",
				r.ResourceUpdater.DescribeResource(),
			),
		)
		return
	}
	role := binding[0].RoleId

	err := iamPolicyReadModifySet(ctx, r.ResourceUpdater, func(p *Policy) error {
		p.Bindings = removeRoleFromBindings(role, p.Bindings)
		return nil
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource Policies",
			fmt.Sprintf("An unexpected error occurred while deleting resource policies"+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err))
		return
	}
	r.RefreshBindingState(ctx, req.State, &resp.State, resp.Diagnostics)
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.ResourceUpdater.GetNameSuffix()
}

func (r *bindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.ResourceUpdater.Configure(ctx, req, resp)
}

func (r *bindingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"role":    schema.StringAttribute{Required: true},
			"members": schema.SetAttribute{Required: true, ElementType: types.StringType},
		},
	}
	maps.Copy(resp.Schema.Attributes, r.ResourceUpdater.GetSchemaAttributes())
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: {resource_id},{role}. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(r.ResourceUpdater.GetIdAlias()), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role"), idParts[1])...)
}

// all bindings use same Role
func getResourceIamBindings(ctx context.Context, state Extractable, diag *diag.Diagnostics) []*access.AccessBinding {
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
