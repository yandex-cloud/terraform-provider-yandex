package kubernetes_marketplace_helm_release

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	marketplace "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/marketplace/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	yandexMarketplaceHelmReleaseTimeout = 10 * time.Minute

	installationErrorWaitTime = 1 * time.Minute // when install fails, HelmRelease may still be created, but not appear immediately

	nameValue      = "applicationName"
	namespaceValue = "namespace"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &helmReleaseResource{}
	_ resource.ResourceWithConfigure   = &helmReleaseResource{}
	_ resource.ResourceWithImportState = &helmReleaseResource{}
)

func NewResource() resource.Resource {
	return &helmReleaseResource{}
}

// helmReleaseResource is the resource implementation.
type helmReleaseResource struct {
	providerConfig *config.Config
}

// Metadata returns the resource type name.
func (r *helmReleaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_marketplace_helm_release"
}

// Schema defines the schema for the resource.
func (r *helmReleaseResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Allows management of Kubernetes product installed from Yandex Cloud Marketplace.\nFor more information, see [official documentation](https://yandex.cloud/marketplace?type=K8S).",
		Attributes: map[string]schema.Attribute{
			"product_version": schema.StringAttribute{
				MarkdownDescription: "The ID of the product version to be installed.",
				Required:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Kubernetes cluster where the product will be installed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the deployment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The Kubernetes namespace where the product will be installed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_values": schema.MapAttribute{
				MarkdownDescription: "Values to be passed for the installation of the product. " +
					"The block consists of attributes that accept string values. " +
					"The exact structure depends on the particular product and may differ for different versions of the same product. " +
					"Depending on the product, some values may be required, and the installation may fail if they are not provided.\n" +
					"~> `applicationName` and `namespace`, if provided in this block, override `name` and `namespace` arguments, respectively.\n",
				ElementType: types.StringType,
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Marketplace product.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product_name": schema.StringAttribute{
				MarkdownDescription: "The name of the Marketplace product.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the deployment.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The Helm Release creation (first installation) timestamp.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Version: 1,
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				CustomType: timeouts.Type{},
			},
		},
	}
	resp.Schema.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})
}

// Configure implements resource.Resource.
func (r *helmReleaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

// Create a new resource.
func (r *helmReleaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan helmReleaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate InstallHelmReleaseRequest from plan
	installHelmReleaseRequest := &marketplace.InstallHelmReleaseRequest{
		ClusterId:        plan.ClusterID.ValueString(),
		ProductVersionId: plan.ProductVersionID.ValueString(),
	}

	installHelmReleaseRequest.UserValues, diags = userValuesFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Install Helm Release request: %+v", installHelmReleaseRequest))

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMarketplaceHelmReleaseTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Install requested release
	id, diags := installHelmRelease(ctx, r.providerConfig.SDK, installHelmReleaseRequest)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		if id == "" {
			return
		}

		tflog.Debug(ctx, "Waiting to see if failed Helm Release gets created anyway", helmReleaseIDLogField(id))
		time.Sleep(installationErrorWaitTime)
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(id)
	diags = updateState(ctx, r.providerConfig.SDK, &plan)
	if diags.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Finished installing Helm Release", helmReleaseIDLogField(id))
}

// Read implements resource.Resource.
func (r *helmReleaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state helmReleaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, "Reading Helm Release", helmReleaseIDLogField(id))
	hr, d := getHelmRelease(ctx, r.providerConfig.SDK, id)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	if hr == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = helmReleaseToModel(ctx, hr, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Helm Release", helmReleaseIDLogField(id))
}

// Update implements resource.Resource.
func (r *helmReleaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan helmReleaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state helmReleaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Helm Release", helmReleaseIDLogField(state.ID.ValueString()))

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMarketplaceHelmReleaseTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Update Helm Release state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update Helm Release plan: %+v", plan))

	updateHelmReleaseRequest := &marketplace.UpdateHelmReleaseRequest{
		Id:               state.ID.ValueString(),
		ProductVersionId: plan.ProductVersionID.ValueString(),
	}

	updateHelmReleaseRequest.UserValues, diags = userValuesFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Update Helm Release request: %+v", updateHelmReleaseRequest))

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Update release
	d := updateHelmRelease(ctx, r.providerConfig.SDK, updateHelmReleaseRequest)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = updateState(ctx, r.providerConfig.SDK, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "Finished updating Helm release", helmReleaseIDLogField(state.ID.ValueString()))
}

// Delete implements resource.Resource.
func (r *helmReleaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state helmReleaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, "Deleting Helm Release", helmReleaseIDLogField(id))

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMarketplaceHelmReleaseTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	d := uninstallHelmRelease(ctx, r.providerConfig.SDK, &marketplace.UninstallHelmReleaseRequest{
		Id: id,
	})
	resp.Diagnostics.Append(d)

	tflog.Debug(ctx, "Finished deleting Helm Release", helmReleaseIDLogField(id))
}

// ImportState implements resource.ResourceWithImportState.
func (r *helmReleaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func updateState(ctx context.Context, sdk *ycsdk.SDK, state *helmReleaseResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	id := state.ID.ValueString()
	tflog.Debug(ctx, "Reading Helm Release", helmReleaseIDLogField(id))
	hr, d := getHelmRelease(ctx, sdk, id)
	diags.Append(d)
	if diags.HasError() {
		return diags
	}

	if hr == nil {
		diags.AddError(
			"Helm Release not found",
			fmt.Sprintf("Helm Release with id %s not found", id))
		return diags
	}

	dd := helmReleaseToModel(ctx, hr, state)
	diags.Append(dd...)
	return diags
}

func helmReleaseIDLogField(id string) map[string]interface{} {
	return map[string]interface{}{
		"helm_release_id": id,
	}
}

func userValuesFromPlan(ctx context.Context, plan helmReleaseResourceModel) ([]*marketplace.ValueWithKey, diag.Diagnostics) {
	var planValues map[string]string

	var diags diag.Diagnostics

	if !plan.UserValues.IsNull() {
		diags = plan.UserValues.ElementsAs(ctx, &planValues, false)
		if diags.HasError() {
			return nil, diags
		}
	}

	if planValues == nil {
		planValues = make(map[string]string)
	}

	if _, ok := planValues[nameValue]; !ok {
		planValues[nameValue] = plan.Name.ValueString()
	}

	if _, ok := planValues[namespaceValue]; !ok {
		planValues[namespaceValue] = plan.Namespace.ValueString()
	}

	values := make([]*marketplace.ValueWithKey, 0, len(planValues))

	for k, v := range planValues {
		value := &marketplace.Value{}
		value.SetValue(
			&marketplace.Value_TypedValue{
				TypedValue: v,
			},
		)
		values = append(values, &marketplace.ValueWithKey{
			Key:   k,
			Value: value,
		})
	}

	return values, diags
}
