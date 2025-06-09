package gitlab_instance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ycsdk "github.com/yandex-cloud/go-sdk"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	InstanceCreateTimeout = 60 * time.Minute
	InstanceDeleteTimeout = 60 * time.Minute
	InstanceUpdateTimeout = 60 * time.Minute
)

var (
	_ resource.Resource                = &gitlabInstanceResource{}
	_ resource.ResourceWithConfigure   = &gitlabInstanceResource{}
	_ resource.ResourceWithImportState = &gitlabInstanceResource{}
)

type gitlabInstanceResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &gitlabInstanceResource{}
}

func (r *gitlabInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab_instance"
}

func (r *gitlabInstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitlabInstanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = InstanceResourceSchema(ctx)
	resp.Schema.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})
}

func (r *gitlabInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *gitlabInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InstanceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createInstanceRequest, diags := BuildCreateInstanceRequest(ctx, &plan, &r.providerConfig.ProviderState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create Gitlab instance request: %+v", createInstanceRequest))

	createTimeout, diags := plan.Timeouts.Create(ctx, InstanceCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	instanceID, d := CreateInstance(ctx, r.providerConfig.SDK, &resp.Diagnostics, createInstanceRequest)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(instanceID)
	diags = updateState(ctx, r.providerConfig.SDK, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Finished creating Gitlab instance", instanceIDLogField(instanceID))
}

func (r *gitlabInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InstanceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Gitlab instance", instanceIDLogField(instanceID))
	instance, d := GetInstanceByID(ctx, r.providerConfig.SDK, instanceID)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	if instance == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = InstanceToState(ctx, instance, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Gitlab instance", instanceIDLogField(instanceID))
}

func (r *gitlabInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InstanceModel
	var state InstanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Gitlab instance", instanceIDLogField(state.Id.ValueString()))

	updateTimeout, diags := plan.Timeouts.Update(ctx, InstanceUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, fmt.Sprintf("Update Gitlab instance state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update Gitlab instance plan: %+v", plan))

	tflog.Error(ctx, "Update operation is not implemented")
}

func (r *gitlabInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InstanceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting Gitalb instance", instanceIDLogField(instanceID))

	deleteTimeout, diags := state.Timeouts.Delete(ctx, InstanceDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	d := DeleteInstance(ctx, r.providerConfig.SDK, instanceID)
	resp.Diagnostics.Append(d)

	tflog.Debug(ctx, "Finished deleting Gitlab instance", instanceIDLogField(instanceID))
}

func updateState(ctx context.Context, sdk *ycsdk.SDK, state *InstanceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	instanceId := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Gitlab instance", instanceIDLogField(instanceId))
	instance, d := GetInstanceByID(ctx, sdk, instanceId)
	diags.Append(d)
	if diags.HasError() {
		return diags
	}

	if instance == nil {
		diags.AddError(
			"Gitlab instance not found",
			fmt.Sprintf("Gitlab instance with id %s not found", instanceId))
		return diags
	}

	dd := InstanceToState(ctx, instance, state)
	diags.Append(dd...)
	return diags
}
