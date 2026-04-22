package organizationmanager_mfa_enforcement_excluded_audience

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	sdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type excludedAudience struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &excludedAudience{}
}

func (r *excludedAudience) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizationmanager_mfa_enforcement_excluded_audience"
}

func (r *excludedAudience) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.providerConfig = providerConfig
}

func (r *excludedAudience) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *excludedAudience) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data excludedAudienceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := organizationmanager.UpdateExcludedAudienceRequest{
		MfaEnforcementId: data.MfaEnforcementId.ValueString(),
		AudienceDeltas: []*organizationmanager.AudienceDelta{
			{
				SubjectId: data.SubjectId.ValueString(),
				Action:    organizationmanager.AudienceDelta_ACTION_ADD,
			},
		},
	}
	op, err := sdk.NewMfaEnforcementClient(r.providerConfig.SDKv2).UpdateExcludedAudience(ctx, &updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to create MFA enforcement excluded audience resource: "+err.Error(),
		)
		return
	}
	if _, err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create MFA enforcement excluded audience resource: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *excludedAudience) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data excludedAudienceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	client := sdk.NewMfaEnforcementClient(r.providerConfig.SDKv2)
	exists := excludedAudienceExists(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *excludedAudience) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Failed to Update resource",
		"MFA enforcement excluded audience update is not allowed",
	)
}

func (r *excludedAudience) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data excludedAudienceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	client := sdk.NewMfaEnforcementClient(r.providerConfig.SDKv2)
	exists := excludedAudienceExists(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !exists {
		resp.Diagnostics.AddWarning(
			"Failed to Delete resource",
			"MFA enforcement excluded audience resource not found",
		)
		return
	}

	updateReq := organizationmanager.UpdateExcludedAudienceRequest{
		MfaEnforcementId: data.MfaEnforcementId.ValueString(),
		AudienceDeltas: []*organizationmanager.AudienceDelta{
			{
				SubjectId: data.SubjectId.ValueString(),
				Action:    organizationmanager.AudienceDelta_ACTION_REMOVE,
			},
		},
	}
	op, err := client.UpdateExcludedAudience(ctx, &updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete MFA enforcement excluded audience resource: "+err.Error(),
		)
		return
	}
	if _, err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete MFA enforcement excluded audience resource: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func excludedAudienceExists(ctx context.Context, client sdk.MfaEnforcementClient, data *excludedAudienceModel, diag *diag.Diagnostics) bool {
	pageToken := ""
	for {
		resp, err := client.ListExcludedAudience(ctx, &organizationmanager.ListExcludedAudienceRequest{
			MfaEnforcementId: data.MfaEnforcementId.ValueString(),
			PageSize:         100,
			PageToken:        pageToken,
		})
		if err != nil {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to get MFA enforcement excluded audience: "+err.Error(),
			)
			return false
		}
		for _, s := range resp.Subjects {
			if s.Id == data.SubjectId.ValueString() {
				return true
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return false
}
