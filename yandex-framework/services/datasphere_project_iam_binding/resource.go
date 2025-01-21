package datasphere_project_iam_binding

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

type ProjectIAMUpdater struct {
	ProjectId      string
	ProviderConfig *provider_config.Config
}

func NewIamBinding() resource.Resource {
	return accessbinding.NewIamBinding(newProjectIamUpdater())
}

func newProjectIamUpdater() accessbinding.ResourceIamUpdater {
	return &ProjectIAMUpdater{}
}

func (u *ProjectIAMUpdater) GetNameSuffix() string {
	return "datasphere_project_iam_binding"
}

func (u *ProjectIAMUpdater) GetSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		u.GetIdAlias(): schema.StringAttribute{
			MarkdownDescription: "The ID of the Datasphere Project to attach the policy to.",
			Required:            true,
		},
	}
}

func (u *ProjectIAMUpdater) GetIdAlias() string {
	return "project_id"
}

func (u *ProjectIAMUpdater) GetId() string {
	return u.ProjectId
}

func (u *ProjectIAMUpdater) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (u *ProjectIAMUpdater) Initialize(ctx context.Context, state accessbinding.Extractable, diag *diag.Diagnostics) {
	var id types.String
	diag.Append(state.GetAttribute(ctx, path.Root("project_id"), &id)...)
	u.ProjectId = id.ValueString()
}

func (u *ProjectIAMUpdater) GetResourceIamPolicy(ctx context.Context) (*accessbinding.Policy, error) {
	bindings, err := u.GeAccessBindings(ctx, u.ProjectId)
	if err != nil {
		return nil, err
	}
	return &accessbinding.Policy{Bindings: bindings}, nil
}

func (u *ProjectIAMUpdater) SetResourceIamPolicy(ctx context.Context, policy *accessbinding.Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.ProjectId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, provider_config.DefaultTimeout)
	defer cancel()

	op, err := u.ProviderConfig.SDK.WrapOperation(u.ProviderConfig.SDK.Datasphere().Project().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *ProjectIAMUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *accessbinding.PolicyDelta) error {
	bSize := 1000
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < accessbinding.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.ProjectId,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}
		op, err := u.ProviderConfig.SDK.WrapOperation(u.ProviderConfig.SDK.Datasphere().Project().UpdateAccessBindings(ctx, req))
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

func (u *ProjectIAMUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-datasphere-project-%s", u.ProjectId)
}

func (u *ProjectIAMUpdater) DescribeResource() string {
	return fmt.Sprintf("datasphere-project '%s'", u.ProjectId)
}

func (u *ProjectIAMUpdater) GeAccessBindings(ctx context.Context, id string) ([]*access.AccessBinding, error) {
	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := u.ProviderConfig.SDK.Datasphere().Project().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: id,
			PageSize:   accessbinding.DefaultPageSize,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, fmt.Errorf("error retrieving access bindings of project %s: %w", id, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
