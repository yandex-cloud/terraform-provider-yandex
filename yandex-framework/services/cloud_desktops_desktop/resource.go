package cloud_desktops_desktop

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	cloudDesktopsDesktopCreateTimeout = time.Hour
	cloudDesktopsDesktopDeleteTimeout = time.Hour
	cloudDesktopsDesktopUpdateTimeout = 2 * time.Hour
)

type cloudDesktopDesktopResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &cloudDesktopDesktopResource{}
}

func (r *cloudDesktopDesktopResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_desktops_desktop"
}

func (r *cloudDesktopDesktopResource) Configure(_ context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cloudDesktopDesktopResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Cloud Desktops Desktop. For more information see [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/concepts/desktops-and-groups)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Import ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"desktop_id": schema.StringAttribute{
				MarkdownDescription: "The id of the Desktop",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Desktop",
				Optional:            true,
			},
			"desktop_group_id": schema.StringAttribute{
				MarkdownDescription: "The id of the Desktop Group to which the Desktop belongs",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "The list of members which can use the Desktop",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"subject_id": schema.StringAttribute{
							MarkdownDescription: "Identity of the access binding. See [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/Desktop/create#yandex.cloud.clouddesktop.v1.api.User)",
							Required:            true,
						},
						"subject_type": schema.StringAttribute{
							MarkdownDescription: "Type of the access binding. See [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/Desktop/create#yandex.cloud.clouddesktop.v1.api.User)",
							Required:            true,
						},
					},
				},
				Optional: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"network_interface": schema.SingleNestedAttribute{
				MarkdownDescription: "The specification of the Desktop network interface",
				Attributes: map[string]schema.Attribute{
					"subnet_id": schema.StringAttribute{
						MarkdownDescription: "ID of the subnet for desktop",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
				},
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *cloudDesktopDesktopResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Desktop
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desktopProto, isNotFound := readDesktopByID(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, state.DesktopId.ValueString())
	if isNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(desktopToState(ctx, desktopProto, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(ConstructID(state.DesktopId.ValueString(), state.NetworkInterface.SubnetId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudDesktopDesktopResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, state Desktop
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	state.Labels, diags = types.MapValue(types.StringType, make(map[string]attr.Value, 0))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, cloudDesktopsDesktopCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	desktop, diags := planToCreateDesktop(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	update, diags := planToUpdateDesktop(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	currentState, diags := planToUpdateDesktop(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updatePaths := getUpdatePaths(update, currentState)

	desktopID := createDesktop(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktop)
	if resp.Diagnostics.HasError() {
		return
	}
	defer func() {
		if resp.Diagnostics.HasError() {
			deleteDesktop(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopID)
		}
	}()

	if len(updatePaths) > 0 {
		updateDesktop(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopID, update, updatePaths)
	}

	plan.DesktopId = types.StringValue(desktopID)
	plan.Id = types.StringValue(ConstructID(plan.DesktopId.ValueString(), plan.NetworkInterface.SubnetId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func getUpdatePaths(plan, state *clouddesktop.UpdatePropertiesRequest) []string {
	var updatePaths []string
	if !reflect.DeepEqual(state.Labels, plan.Labels) {
		updatePaths = append(updatePaths, "labels")
	}
	if state.Name != plan.Name {
		updatePaths = append(updatePaths, "name")
	}

	return updatePaths
}

func (r *cloudDesktopDesktopResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Desktop
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, cloudDesktopsDesktopUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	desktopID := state.DesktopId.ValueString()
	planDesktop, diags := planToUpdateDesktop(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	stateDesktop, diags := planToUpdateDesktop(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatePaths := getUpdatePaths(planDesktop, stateDesktop)
	if len(updatePaths) > 0 {
		updateDesktop(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopID, planDesktop, updatePaths)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(ConstructID(plan.DesktopId.ValueString(), plan.NetworkInterface.SubnetId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *cloudDesktopDesktopResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Desktop
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, cloudDesktopsDesktopDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	desktopID := state.DesktopId.ValueString()
	deleteDesktop(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopID)
}

func (r *cloudDesktopDesktopResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	desktopId, subnetID, err := DeconstructID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}

	desktop, isNotFound := readDesktopByID(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopId)
	if resp.Diagnostics.HasError() {
		return
	}
	if isNotFound {
		resp.Diagnostics.AddError(
			"Failed to Import resource",
			"No resource with such ID found",
		)
		return
	}

	var state Desktop
	resp.Diagnostics.Append(desktopToState(ctx, desktop, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"delete": types.StringType,
			"update": types.StringType,
		}),
	}

	state.DesktopId = types.StringValue(desktopId)
	state.NetworkInterface.SubnetId = types.StringValue(subnetID)
	state.Id = types.StringValue(req.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
