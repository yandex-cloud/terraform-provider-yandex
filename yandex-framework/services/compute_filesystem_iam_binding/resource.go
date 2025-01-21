// Code generated with blueprint. You can edit it, based on your certain requirements.

package compute_filesystem_iam_binding

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/accessbinding"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type IAMUpdater struct {
	FilesystemId   string
	ProviderConfig *provider_config.Config
}

func NewIamBinding() resource.Resource {
	return accessbinding.NewIamBinding(newIAMUpdater())
}

func newIAMUpdater() accessbinding.ResourceIamUpdater {
	return &IAMUpdater{}
}

func (u *IAMUpdater) GetResourceIamPolicy(ctx context.Context) (*accessbinding.Policy, error) {
	bindings, err := u.getAccessBindings(ctx, u.FilesystemId)
	if err != nil {
		return nil, err
	}
	return &accessbinding.Policy{Bindings: bindings}, nil
}

func (u *IAMUpdater) getAccessBindings(ctx context.Context, id string) ([]*access.AccessBinding, error) {
	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := u.ProviderConfig.SDK.Compute().Filesystem().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: id,
			PageSize:   accessbinding.DefaultPageSize,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, err
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}

func (u *IAMUpdater) SetResourceIamPolicy(ctx context.Context, policy *accessbinding.Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.FilesystemId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, provider_config.DefaultTimeout)
	defer cancel()

	op, err := u.ProviderConfig.SDK.WrapOperation(u.ProviderConfig.SDK.Compute().Filesystem().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *IAMUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *accessbinding.PolicyDelta) error {
	var (
		bSize  = 1000
		deltas = policy.Deltas
		dLen   = len(deltas)
	)

	for i := 0; i < accessbinding.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.FilesystemId,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.ProviderConfig.SDK.WrapOperation(u.ProviderConfig.SDK.Compute().Filesystem().UpdateAccessBindings(ctx, req))
		if err != nil {
			return fmt.Errorf("error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
	}

	return nil
}

func (u *IAMUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-compute-filesystem-%s", u.FilesystemId)
}

func (u *IAMUpdater) DescribeResource() string {
	return fmt.Sprintf("compute-filesystem '%s'", u.FilesystemId)
}

func (u *IAMUpdater) GetNameSuffix() string {
	return "compute_filesystem_iam_binding"
}

func (u *IAMUpdater) GetSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		u.GetIdAlias(): schema.StringAttribute{
			MarkdownDescription: "The ID of the compute Filesystem to attach the policy to.",
			Required:            true,
		},
	}
}

func (u *IAMUpdater) GetIdAlias() string {
	return "filesystem_id"
}

func (u *IAMUpdater) GetId() string {
	return u.FilesystemId
}

func (u *IAMUpdater) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. "+
				"Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	u.ProviderConfig = providerConfig
}

func (u *IAMUpdater) Initialize(ctx context.Context, state accessbinding.Extractable, diag *diag.Diagnostics) {
	var id types.String

	diag.Append(state.GetAttribute(ctx, path.Root("filesystem_id"), &id)...)
	u.FilesystemId = id.ValueString()
}
