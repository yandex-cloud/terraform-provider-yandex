package organizationmanager_idp_application_saml_application_assignment

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/saml"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/saml"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type assignment struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &assignment{}
}

func (r *assignment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizationmanager_idp_application_saml_application_assignment"
}

func (r *assignment) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *assignment) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *assignment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data assignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := saml.UpdateAssignmentsRequest{
		ApplicationId: data.ApplicationId.ValueString(),
		AssignmentDeltas: []*saml.AssignmentDelta{
			{
				Action: saml.AssignmentAction_ADD,
				Assignment: &saml.Assignment{
					SubjectId: data.SubjectId.ValueString(),
				},
			},
		},
	}
	op, err := samlsdk.NewApplicationClient(r.providerConfig.SDKv2).UpdateAssignments(ctx, &updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to create SAML application assignment resource: "+err.Error(),
		)
		return
	}
	if _, err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create SAML application assignment resource: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *assignment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data assignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	client := samlsdk.NewApplicationClient(r.providerConfig.SDKv2)
	exists := assignmentExists(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *assignment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Failed to Update resource",
		"SAML application assignment update is not allowed",
	)
}

func (r *assignment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data assignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	client := samlsdk.NewApplicationClient(r.providerConfig.SDKv2)
	exists := assignmentExists(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !exists {
		resp.Diagnostics.AddWarning(
			"Failed to Delete resource",
			"SAML application assignment resource not found",
		)
		return
	}

	updateReq := saml.UpdateAssignmentsRequest{
		ApplicationId: data.ApplicationId.ValueString(),
		AssignmentDeltas: []*saml.AssignmentDelta{
			{
				Action: saml.AssignmentAction_REMOVE,
				Assignment: &saml.Assignment{
					SubjectId: data.SubjectId.ValueString(),
				},
			},
		},
	}
	op, err := client.UpdateAssignments(ctx, &updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete SAML application assignment resource: "+err.Error(),
		)
		return
	}
	if _, err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete SAML application assignment resource: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func assignmentExists(ctx context.Context, client samlsdk.ApplicationClient, data *assignmentModel, diag *diag.Diagnostics) bool {
	pageToken := ""
	for {
		resp, err := client.ListAssignments(ctx, &saml.ListAssignmentsRequest{
			ApplicationId: data.ApplicationId.ValueString(),
			PageSize:      100,
			PageToken:     pageToken,
		})
		if err != nil {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to get SAML application assignments: "+err.Error(),
			)
			return false
		}
		for _, assignment := range resp.Assignments {
			if assignment.SubjectId == data.SubjectId.ValueString() {
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
