package datasphere_community_iam_binding

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

type CommunityIAMUpdater struct {
	CommunityId    string
	ProviderConfig *provider_config.Config
}

func NewIamBinding() resource.Resource {
	return accessbinding.NewIamBinding(newCommunityIamUpdater())
}

func newCommunityIamUpdater() accessbinding.ResourceIamUpdater {
	return &CommunityIAMUpdater{}
}

func (u *CommunityIAMUpdater) GetNameSuffix() string {
	return "datasphere_community_iam_binding"
}

func (u *CommunityIAMUpdater) GetSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		u.GetIdAlias(): schema.StringAttribute{
			MarkdownDescription: "The ID of the Datasphere Community to attach the policy to.",
			Required:            true,
		},
	}
}

func (u *CommunityIAMUpdater) GetIdAlias() string {
	return "community_id"
}

func (u *CommunityIAMUpdater) GetId() string {
	return u.CommunityId
}

func (u *CommunityIAMUpdater) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (u *CommunityIAMUpdater) Initialize(ctx context.Context, state accessbinding.Extractable, diag *diag.Diagnostics) {
	var id types.String
	diag.Append(state.GetAttribute(ctx, path.Root("community_id"), &id)...)
	u.CommunityId = id.ValueString()
}

func (u *CommunityIAMUpdater) GetResourceIamPolicy(ctx context.Context) (*accessbinding.Policy, error) {
	bindings, err := u.GeAccessBindings(ctx, u.CommunityId)
	if err != nil {
		return nil, err
	}
	return &accessbinding.Policy{Bindings: bindings}, nil
}

func (u *CommunityIAMUpdater) SetResourceIamPolicy(ctx context.Context, policy *accessbinding.Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.CommunityId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, provider_config.DefaultTimeout)
	defer cancel()

	op, err := u.ProviderConfig.SDK.WrapOperation(
		u.ProviderConfig.SDK.Datasphere().Community().SetAccessBindings(ctx, req),
	)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *CommunityIAMUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *accessbinding.PolicyDelta) error {
	bSize := 1000
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < accessbinding.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.CommunityId,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.ProviderConfig.SDK.WrapOperation(
			u.ProviderConfig.SDK.Datasphere().Community().UpdateAccessBindings(ctx, req),
		)
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

func (u *CommunityIAMUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-datasphere-community-%s", u.CommunityId)
}

func (u *CommunityIAMUpdater) DescribeResource() string {
	return fmt.Sprintf("datasphere-community '%s'", u.CommunityId)
}

func (u *CommunityIAMUpdater) GeAccessBindings(ctx context.Context, id string) ([]*access.AccessBinding, error) {
	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := u.ProviderConfig.SDK.Datasphere().Community().ListAccessBindings(
			ctx,
			&access.ListAccessBindingsRequest{
				ResourceId: id,
				PageSize:   accessbinding.DefaultPageSize,
				PageToken:  pageToken,
			},
		)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, fmt.Errorf("error retrieving access bindings of function %s: %w", id, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
