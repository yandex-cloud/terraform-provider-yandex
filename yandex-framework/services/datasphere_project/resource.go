package datasphere_project

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type projectResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &projectResource{}
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating project resource")

	var plannedProject projectDataModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plannedProject)...)

	createTimeout, timeoutInitError := plannedProject.Timeouts.Create(ctx, provider_config.DefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	createProjectRequestData := datasphere.CreateProjectRequest{
		Name:        plannedProject.Name.ValueString(),
		CommunityId: plannedProject.CommunityId.ValueString(),
		Description: plannedProject.Description.ValueString(),
	}
	if !plannedProject.Labels.IsNull() && !plannedProject.Labels.IsUnknown() {
		labels := make(map[string]string, len(plannedProject.Labels.Elements()))
		resp.Diagnostics.Append(plannedProject.Labels.ElementsAs(ctx, &labels, false)...)
		createProjectRequestData.SetLabels(labels)

	}

	if plannedProject.Settings.IsNull() || plannedProject.Settings.IsUnknown() {
		tflog.Info(ctx, "Settings field is not set, would be used default settings")
	} else {
		var settings settingsObjectModel

		tflog.Debug(ctx, "Converting settings from terraform plan to proto model")
		resp.Diagnostics.Append(plannedProject.Settings.As(ctx, &settings, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		createProjectSettingsRequestData := datasphere.Project_Settings{
			SubnetId:          settings.SubnetId.ValueString(),
			ServiceAccountId:  settings.ServiceAccountId.ValueString(),
			DataProcClusterId: settings.DataProcClusterId.ValueString(),
			DefaultFolderId:   settings.DefaultFolderId.ValueString(),
		}

		if !settings.SecurityGroupIds.IsNull() && !settings.SecurityGroupIds.IsUnknown() {
			settingsSecurityGroups := make([]string, 0, len(settings.SecurityGroupIds.Elements()))
			resp.Diagnostics.Append(settings.SecurityGroupIds.ElementsAs(ctx, &settingsSecurityGroups, false)...)
			createProjectSettingsRequestData.SetSecurityGroupIds(settingsSecurityGroups)
		}
		if !settings.StaleExecTimeoutMode.IsNull() && !settings.StaleExecTimeoutMode.IsUnknown() {
			createProjectSettingsRequestData.SetStaleExecTimeoutMode(datasphere.Project_Settings_StaleExecutionTimeoutMode(
				datasphere.Project_Settings_StaleExecutionTimeoutMode_value[settings.StaleExecTimeoutMode.ValueString()]))
		}
		createProjectRequestData.SetSettings(&createProjectSettingsRequestData)

	}

	if plannedProject.Limits.IsNull() || plannedProject.Limits.IsUnknown() {
		tflog.Debug(ctx, "Limits are not set, should be used default limits")
	} else {
		var limits limitsObjectModel
		createProjectLimitsRequestData := datasphere.Project_Limits{}

		tflog.Debug(ctx, "Converting limits from terraform plan to proto model")
		resp.Diagnostics.Append(plannedProject.Limits.As(ctx, &limits, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !limits.MaxUnitsPerHour.IsNull() && !limits.MaxUnitsPerHour.IsUnknown() {
			createProjectLimitsRequestData.SetMaxUnitsPerHour(wrapperspb.Int64(limits.MaxUnitsPerHour.ValueInt64()))
		}
		if !limits.MaxUnitsPerExecution.IsNull() && !limits.MaxUnitsPerExecution.IsUnknown() {
			createProjectLimitsRequestData.SetMaxUnitsPerExecution(wrapperspb.Int64(limits.MaxUnitsPerExecution.ValueInt64()))
		}

		createProjectRequestData.SetLimits(&createProjectLimitsRequestData)

	}

	tflog.Info(ctx,
		fmt.Sprintf("Making API call to create new project with parameters %+v", &createProjectRequestData),
	)
	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.Datasphere().Project().Create(ctx, &createProjectRequestData))
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
	createdProject := protoResponse.(*datasphere.Project)
	plannedProject.Id = types.StringValue(createdProject.Id)

	// Balance has his descriptors methods
	var plannedBalance types.Int64
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("limits").AtName("balance"), &plannedBalance)...)

	var updatedBalance *wrapperspb.Int64Value

	if !plannedBalance.IsNull() && !plannedBalance.IsUnknown() {
		tflog.Info(ctx, "Making additional request to update project balance")
		setProjectBalanceRequest := datasphere.SetUnitBalanceRequest{
			ProjectId:   plannedProject.Id.ValueString(),
			UnitBalance: wrapperspb.Int64(plannedBalance.ValueInt64()),
		}
		opBalance, errBalance := r.providerConfig.SDK.WrapOperation(
			r.providerConfig.SDK.Datasphere().Project().SetUnitBalance(ctx, &setProjectBalanceRequest),
		)
		if errBalance != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				fmt.Sprintf("An unexpected error occurred while updating project balance. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: %s", errBalance),
			)
			return
		}
		errBalance = opBalance.Wait(ctx)
		if errBalance != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				fmt.Sprintf("An unexpected error occurred while updating project balance. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: %s", errBalance),
			)
			return
		}
		updatedBalance = wrapperspb.Int64(plannedBalance.ValueInt64())
	}

	tflog.Info(ctx, fmt.Sprintf("Project with following id %s was created", createdProject.Id))
	convertToTerraformModel(ctx, &plannedProject, createdProject, &resp.Diagnostics, updatedBalance)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plannedProject)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading project resource")
	var stateProject projectDataModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateProject)...)

	existingProject, err := r.providerConfig.SDK.Datasphere().Project().Get(ctx,
		&datasphere.GetProjectRequest{ProjectId: stateProject.Id.ValueString()})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to refresh resource state. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)

		return
	}
	unitBalance, err := r.providerConfig.SDK.Datasphere().Project().GetUnitBalance(
		ctx,
		&datasphere.GetUnitBalanceRequest{ProjectId: stateProject.Id.ValueString()},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			fmt.Sprintf("An unexpected error occurred while attempting to refresh resource state. "+
				"Please retry the operation or report this issue to the provider developers.\n\nError: %s", err),
		)
		return
	}

	convertToTerraformModel(ctx, &stateProject, existingProject, &resp.Diagnostics, unitBalance.UnitBalance)

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateProject)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating project resource")

	var planProject, stateProject projectDataModel

	updateTimeout, timeoutInitError := planProject.Timeouts.Update(ctx, provider_config.DefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planProject)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateProject)...)

	var updatePaths []string
	updateProjectRequest := &datasphere.UpdateProjectRequest{
		Name:        planProject.Name.ValueString(),
		Description: planProject.Description.ValueString(),
		ProjectId:   planProject.Id.ValueString(),
	}

	// Compare name attribute value between plan and prior state
	if !planProject.Description.Equal(stateProject.Description) {
		updatePaths = append(updatePaths, "description")
	}
	if !planProject.Name.Equal(stateProject.Name) {
		updatePaths = append(updatePaths, "name")
	}
	if !planProject.Labels.Equal(stateProject.Labels) {
		updatePaths = append(updatePaths, "labels")
		labels := make(map[string]string, len(planProject.Labels.Elements()))
		resp.Diagnostics.Append(planProject.Labels.ElementsAs(ctx, &labels, false)...)
		updateProjectRequest.SetLabels(labels)
	}
	if !planProject.Settings.Equal(stateProject.Settings) {
		var planProjectSettings, stateProjectSettings settingsObjectModel
		const pathPrefix = "settings."
		resp.Diagnostics.Append(planProject.Settings.As(ctx, &planProjectSettings, basetypes.ObjectAsOptions{})...)
		resp.Diagnostics.Append(stateProject.Settings.As(ctx, &stateProjectSettings, basetypes.ObjectAsOptions{})...)

		updateProjectSettingsRequestData := datasphere.Project_Settings{
			SubnetId:          planProjectSettings.SubnetId.ValueString(),
			ServiceAccountId:  planProjectSettings.ServiceAccountId.ValueString(),
			DataProcClusterId: planProjectSettings.DataProcClusterId.ValueString(),
			DefaultFolderId:   planProjectSettings.DefaultFolderId.ValueString(),
		}

		if !planProjectSettings.ServiceAccountId.Equal(stateProjectSettings.ServiceAccountId) {
			updatePaths = append(updatePaths, pathPrefix+"service_account_id")
		}
		if !planProjectSettings.SubnetId.Equal(stateProjectSettings.SubnetId) {
			updatePaths = append(updatePaths, pathPrefix+"subnet_id")
		}
		if !planProjectSettings.DataProcClusterId.Equal(stateProjectSettings.DataProcClusterId) {
			updatePaths = append(updatePaths, pathPrefix+"data_proc_cluster_id")
		}
		if !planProjectSettings.SecurityGroupIds.Equal(stateProjectSettings.SecurityGroupIds) {
			updatePaths = append(updatePaths, pathPrefix+"security_group_ids")
			settingsSecurityGroups := make([]string, 0, len(planProjectSettings.SecurityGroupIds.Elements()))
			resp.Diagnostics.Append(
				planProjectSettings.SecurityGroupIds.ElementsAs(ctx, &settingsSecurityGroups, false)...)
			updateProjectSettingsRequestData.SetSecurityGroupIds(settingsSecurityGroups)
		}
		if !planProjectSettings.DefaultFolderId.Equal(stateProjectSettings.DefaultFolderId) {
			updatePaths = append(updatePaths, pathPrefix+"default_folder_id")
		}
		if !planProjectSettings.StaleExecTimeoutMode.Equal(stateProjectSettings.StaleExecTimeoutMode) {
			updatePaths = append(updatePaths, pathPrefix+"stale_exec_timeout_mode")
			updateProjectSettingsRequestData.SetStaleExecTimeoutMode(
				datasphere.Project_Settings_StaleExecutionTimeoutMode(
					datasphere.Project_Settings_StaleExecutionTimeoutMode_value[planProjectSettings.StaleExecTimeoutMode.ValueString()]))
		}
		updateProjectRequest.SetSettings(&updateProjectSettingsRequestData)

	}
	if !planProject.Limits.Equal(stateProject.Limits) {
		var planProjectLimits, stateProjectLimits limitsObjectModel
		const pathPrefix = "limits."
		updateProjectLimitsRequestData := datasphere.Project_Limits{}

		resp.Diagnostics.Append(planProject.Limits.As(ctx, &planProjectLimits, basetypes.ObjectAsOptions{})...)

		resp.Diagnostics.Append(planProject.Limits.As(ctx, &planProjectLimits, basetypes.ObjectAsOptions{})...)
		resp.Diagnostics.Append(stateProject.Limits.As(ctx, &stateProjectLimits, basetypes.ObjectAsOptions{})...)
		if !planProjectLimits.MaxUnitsPerHour.Equal(stateProjectLimits.MaxUnitsPerHour) {
			updatePaths = append(updatePaths, pathPrefix+"max_units_per_hour")
			updateProjectLimitsRequestData.SetMaxUnitsPerExecution(
				wrapperspb.Int64(planProjectLimits.MaxUnitsPerExecution.ValueInt64()))

		}
		if !planProjectLimits.MaxUnitsPerExecution.Equal(stateProjectLimits.MaxUnitsPerExecution) {
			updatePaths = append(updatePaths, pathPrefix+"max_units_per_execution")
			updateProjectLimitsRequestData.SetMaxUnitsPerHour(
				wrapperspb.Int64(planProjectLimits.MaxUnitsPerHour.ValueInt64()))

		}
		updateProjectRequest.SetLimits(&updateProjectLimitsRequestData)
	}

	if len(updatePaths) == 0 || resp.Diagnostics.HasError() {
		return
	}

	updateProjectRequest.SetUpdateMask(&field_mask.FieldMask{Paths: updatePaths})
	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.Datasphere().Project().Update(ctx, updateProjectRequest))
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
	updatedProject, ok := protoResponse.(*datasphere.Project)

	if !ok {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			fmt.Sprintf("Expected *datasphere.Project, got: %T. "+
				"Please report this issue to the provider developers.", updatedProject),
		)
		return
	}

	// Balance has his descriptors methods
	var plannedBalance, stateBalance types.Int64
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx,
		path.Root("limits").AtName("balance"), &plannedBalance)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx,
		path.Root("limits").AtName("balance"), &stateBalance)...)

	var updatedBalance *wrapperspb.Int64Value

	if !plannedBalance.Equal(stateBalance) {
		setProjectBalanceRequest := datasphere.SetUnitBalanceRequest{
			ProjectId:   planProject.Id.ValueString(),
			UnitBalance: wrapperspb.Int64(plannedBalance.ValueInt64()),
		}
		opBalance, errBalance := r.providerConfig.SDK.WrapOperation(
			r.providerConfig.SDK.Datasphere().Project().SetUnitBalance(ctx, &setProjectBalanceRequest))
		if errBalance != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				fmt.Sprintf("An unexpected error occurred while attempting to update the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: %s", err),
			)
			return
		}
		errBalance = opBalance.Wait(ctx)
		if errBalance != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				fmt.Sprintf("An unexpected error occurred while attempting to update the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: %s", err),
			)
			return
		}
		updatedBalance = wrapperspb.Int64(plannedBalance.ValueInt64())
	}

	tflog.Debug(ctx,
		fmt.Sprintf(
			"Project was update with following parameters %+v and balance %s",
			updatedProject,
			updatedBalance,
		),
	)
	convertToTerraformModel(ctx, &planProject, updatedProject, &resp.Diagnostics, updatedBalance)

	resp.Diagnostics.Append(resp.State.Set(ctx, &planProject)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting project resource")
	var stateProject projectDataModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateProject)...)

	removeTimeout, timeoutInitError := stateProject.Timeouts.Delete(ctx, provider_config.DefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, removeTimeout)
	defer cancel()
	tflog.Info(ctx,
		fmt.Sprintf("Make API call to delete project with following id: %s", stateProject.Id.ValueString()),
	)

	deleteProjectRequest := datasphere.DeleteProjectRequest{ProjectId: stateProject.Id.ValueString()}
	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.Datasphere().Project().Delete(ctx, &deleteProjectRequest))

	timoutErr := op.Wait(ctx)
	if timoutErr != nil {
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
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasphere_project"
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "Initializing DatasphereProject schema.")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Allows management of Yandex Cloud Datasphere Projects.",
		Attributes: map[string]schema.Attribute{
			"id":         defaultschema.Id(),
			"created_at": defaultschema.CreatedAt(),
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Creator account ID of the Datasphere Project.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"community_id": schema.StringAttribute{
				MarkdownDescription: "Community ID where project would be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["name"],
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(regexp.MustCompile(`[a-zA-Z0-9\x{0401}\x{0451}\x{0410}-\x{044F}]\S{1,61}[a-zA-Z0-9\x{0401}\x{0451}\x{0410}-\x{044F}]`),
						"Can contain lowercase and uppercase letters of the Latin and Russian alphabets, "+
							"numbers, hyphens, underscores and spaces. The first character must be a letter. "+
							"The last character must not be a hyphen, underscore or space."),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["name"],
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"labels": schema.MapAttribute{Optional: true,
				MarkdownDescription: common.ResourceDescriptions["labels"],
				ElementType:         types.StringType,
				Validators: []validator.Map{
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
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Datasphere Project settings configuration.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"service_account_id": schema.StringAttribute{
						MarkdownDescription: common.ResourceDescriptions["service_account_id"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"subnet_id": schema.StringAttribute{
						MarkdownDescription: "ID of the subnet where the DataProcessing cluster resides. Currently only subnets created in the availability zone `ru-central1-a` are supported.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"data_proc_cluster_id": schema.StringAttribute{
						MarkdownDescription: "ID of the DataProcessing cluster.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"security_group_ids": schema.SetAttribute{
						MarkdownDescription: common.ResourceDescriptions["security_group_ids"],
						Optional:            true,
						ElementType:         types.StringType,
					},
					"default_folder_id": schema.StringAttribute{
						MarkdownDescription: "Default project folder ID.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"stale_exec_timeout_mode": schema.StringAttribute{
						MarkdownDescription: "The timeout to automatically stop stale executions. The following modes can be used:\n * `ONE_HOUR`: Setting to automatically stop stale execution after one hour with low consumption.\n  * `THREE_HOURS`: Setting to automatically stop stale execution after three hours with low consumption.\n  * `NO_TIMEOUT`: Setting to never automatically stop stale executions.\n",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("STALE_EXECUTION_TIMEOUT_MODE_UNSPECIFIED",
								"ONE_HOUR", "THREE_HOURS", "NO_TIMEOUT"),
						},
					},
				},
			},
			"limits": schema.SingleNestedAttribute{
				MarkdownDescription: "Datasphere Project limits configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"max_units_per_hour": schema.Int64Attribute{
						MarkdownDescription: "The number of units that can be spent per hour.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"max_units_per_execution": schema.Int64Attribute{
						MarkdownDescription: "The number of units that can be spent on the one execution.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"balance": schema.Int64Attribute{
						MarkdownDescription: "The number of units available to the project.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}
