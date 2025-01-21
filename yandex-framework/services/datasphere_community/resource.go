package datasphere_community

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
)

type communityResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &communityResource{}
}

func (r *communityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating community resource")
	var plannedCommunity communityDataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plannedCommunity)...)

	// organization ID could be set on provider level or by terraform resource configuration
	if plannedCommunity.OrganizationId.IsNull() || plannedCommunity.OrganizationId.IsUnknown() {
		if len(r.providerConfig.ProviderState.OrganizationID.ValueString()) == 0 {
			resp.Diagnostics.AddError("Failed to Create Resource",
				"Error getting organization ID while creating Community: cannot determine organization_id: "+
					"please set 'organization_id' key in this resource or at provider level")
			return
		} else {
			plannedCommunity.OrganizationId = r.providerConfig.ProviderState.OrganizationID
		}
	}

	createTimeout, timeoutInitError := plannedCommunity.Timeouts.Create(ctx, provider_config.DefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	createCommunityRequestData := &datasphere.CreateCommunityRequest{
		Name:             plannedCommunity.Name.ValueString(),
		Description:      plannedCommunity.Description.ValueString(),
		OrganizationId:   plannedCommunity.OrganizationId.ValueString(),
		BillingAccountId: plannedCommunity.BillingAccountId.ValueString(),
	}
	if !plannedCommunity.Labels.IsNull() && !plannedCommunity.Labels.IsUnknown() {
		labels := make(map[string]string, len(plannedCommunity.Labels.Elements()))
		resp.Diagnostics.Append(plannedCommunity.Labels.ElementsAs(ctx, &labels, false)...)
		createCommunityRequestData.SetLabels(labels)

	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx,
		fmt.Sprintf("Making API call to create new community with parameters %+v", createCommunityRequestData),
	)
	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.Datasphere().Community().Create(ctx, createCommunityRequestData),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to create the resource."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to create the resource."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	protoResponse, err := op.Response()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			fmt.Sprintf("An unexpected error occurred while parsing API create response. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	createdCommunity, ok := protoResponse.(*datasphere.Community)
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			fmt.Sprintf("Expected *datasphere.Community, got: %T. "+
				"Please report this issue to the provider developers.", createdCommunity),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Community with id `%s` was created", createdCommunity.Id))

	plannedCommunity.Id = types.StringValue(createdCommunity.Id)

	convertToTerraformModel(ctx, &plannedCommunity, createdCommunity, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plannedCommunity)...)
}

func (r *communityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading community resource")
	var stateCommunity communityDataModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateCommunity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx,
		fmt.Sprintf("Making API call to retrieve community with id %s", stateCommunity.Id.ValueString()),
	)

	existingCommunity, err := r.providerConfig.SDK.Datasphere().Community().Get(ctx,
		&datasphere.GetCommunityRequest{CommunityId: stateCommunity.Id.ValueString()},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to refresh resource state. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)

		return
	}

	convertToTerraformModel(ctx, &stateCommunity, existingCommunity, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateCommunity)...)
}

func (r *communityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating community resource")

	var stateCommunity, plannedCommunity communityDataModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plannedCommunity)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateCommunity)...)

	updateTimeout, timeoutInitError := plannedCommunity.Timeouts.Update(ctx, provider_config.DefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	var updatePaths []string
	updateCommunityRequest := &datasphere.UpdateCommunityRequest{
		Name:        plannedCommunity.Name.ValueString(),
		Description: plannedCommunity.Description.ValueString(),
		CommunityId: plannedCommunity.Id.ValueString(),
	}

	// Compare name attribute value between plan and prior state
	if !plannedCommunity.Description.Equal(stateCommunity.Description) {
		updatePaths = append(updatePaths, "description")
	}
	if !plannedCommunity.Name.Equal(stateCommunity.Name) {
		updatePaths = append(updatePaths, "name")
	}
	if !plannedCommunity.Labels.Equal(stateCommunity.Labels) {
		updatePaths = append(updatePaths, "labels")
		labels := make(map[string]string, len(plannedCommunity.Labels.Elements()))
		resp.Diagnostics.Append(plannedCommunity.Labels.ElementsAs(ctx, &labels, false)...)
		updateCommunityRequest.SetLabels(labels)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if len(updatePaths) == 0 {
		return
	}
	updateCommunityRequest.SetUpdateMask(&field_mask.FieldMask{Paths: updatePaths})

	tflog.Info(ctx,
		fmt.Sprintf("Make API call to update community with following parameters: %+v", updateCommunityRequest),
	)

	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.Datasphere().Community().Update(ctx, updateCommunityRequest),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}
	err = op.Wait(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)

		return
	}

	protoResponse, err := op.Response()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			fmt.Sprintf("An unexpected error occurred while parsing update response API. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)

		return
	}
	updatedCommunity, ok := protoResponse.(*datasphere.Community)
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			fmt.Sprintf("Expected *datasphere.Community, got: %T. "+
				"Please report this issue to the provider developers.", updatedCommunity),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Community was update with following parameters %+v", updatedCommunity))
	convertToTerraformModel(ctx, &plannedCommunity, updatedCommunity, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plannedCommunity)...)
}

func (r *communityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting community resource")

	var stateCommunity communityDataModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateCommunity)...)

	removeTimeout, timeoutInitError := stateCommunity.Timeouts.Delete(ctx, provider_config.DefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, removeTimeout)
	defer cancel()

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx,
		fmt.Sprintf("Make API call to delete community with following id: %s", stateCommunity.Id.ValueString()),
	)
	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.Datasphere().Community().Delete(
			ctx,
			&datasphere.DeleteCommunityRequest{CommunityId: stateCommunity.Id.ValueString()},
		),
	)

	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}
	tflog.Info(ctx,
		fmt.Sprintf("Community was deleted id: %s", stateCommunity.Id.ValueString()),
	)

}

func (r *communityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}

func (r *communityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasphere_community"
}

func (r *communityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.providerConfig = providerConfig
}

func (r *communityResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "Initializing DatasphereCommunity resource schema.")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Allows management of Yandex Cloud Datasphere Communities.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				}},
			"name": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["name"],
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`[a-zA-Z0-9\x{0401}\x{0451}\x{0410}-\x{044F}]\S{1,61}[a-zA-Z0-9\x{0401}\x{0451}\x{0410}-\x{044F}]`),
						"Can contain lowercase and uppercase letters of the Latin and Russian alphabets, "+
							"numbers, hyphens, underscores and spaces. The first character must be a letter. "+
							"The last character must not be a hyphen, underscore or space."),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
				},
			},
			"labels": schema.MapAttribute{Optional: true,
				ElementType: types.StringType,
				Validators: []validator.Map{
					mapvalidator.SizeAtMost(64),
					mapvalidator.KeysAre(
						stringvalidator.LengthBetween(1, 63),
						stringvalidator.RegexMatches(regexp.MustCompile(`[a-z][-_0-9a-z]*`),
							"It can contain lowercase letters of the Latin alphabet, numbers, "+
								"hyphens and underscores. And first character must be letter."),
					),
					mapvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 63),
						stringvalidator.RegexMatches(regexp.MustCompile(`[-_0-9a-z]*`),
							"It can contain lowercase letters of the Latin alphabet, numbers, "+
								"hyphens and underscores."),
					),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["created_at"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Creator account ID of the Datasphere Community",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization ID where community would be created",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"billing_account_id": schema.StringAttribute{
				MarkdownDescription: "Billing account ID to associated with community",
				Optional:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}
