package organizationmanager_mfa_enforcement_audience

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	sdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type audience struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &audience{}
}

func (r *audience) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizationmanager_mfa_enforcement_audience"
}

func (r *audience) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *audience) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *audience) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data audienceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := organizationmanager.UpdateAudienceRequest{
		MfaEnforcementId: data.MfaEnforcementId.ValueString(),
		AudienceDeltas: []*organizationmanager.AudienceDelta{
			{
				SubjectId: data.SubjectId.ValueString(),
				Action:    organizationmanager.AudienceDelta_ACTION_ADD,
			},
		},
	}
	op, err := sdk.NewMfaEnforcementClient(r.providerConfig.SDKv2).UpdateAudience(ctx, &updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to create MFA enforcement audience resource: "+err.Error(),
		)
		return
	}
	if _, err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create MFA enforcement audience resource: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *audience) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data audienceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	client := sdk.NewMfaEnforcementClient(r.providerConfig.SDKv2)
	exists := audienceExists(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *audience) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Failed to Update resource",
		"MFA enforcement audience update is not allowed",
	)
}

func (r *audience) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data audienceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	client := sdk.NewMfaEnforcementClient(r.providerConfig.SDKv2)
	exists := audienceExists(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !exists {
		resp.Diagnostics.AddWarning(
			"Failed to Delete resource",
			"MFA enforcement audience resource not found",
		)
		return
	}

	updateReq := organizationmanager.UpdateAudienceRequest{
		MfaEnforcementId: data.MfaEnforcementId.ValueString(),
		AudienceDeltas: []*organizationmanager.AudienceDelta{
			{
				SubjectId: data.SubjectId.ValueString(),
				Action:    organizationmanager.AudienceDelta_ACTION_REMOVE,
			},
		},
	}
	op, err := client.UpdateAudience(ctx, &updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete MFA enforcement audience resource: "+err.Error(),
		)
		return
	}
	if _, err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete MFA enforcement audience resource: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func audienceExists(ctx context.Context, client sdk.MfaEnforcementClient, data *audienceModel, diag *diag.Diagnostics) bool {
	pageToken := ""
	for {
		resp, err := client.ListAudience(ctx, &organizationmanager.ListAudienceRequest{
			MfaEnforcementId: data.MfaEnforcementId.ValueString(),
			PageSize:         100,
			PageToken:        pageToken,
		})
		if err != nil {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to get MFA enforcement audience: "+err.Error(),
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
