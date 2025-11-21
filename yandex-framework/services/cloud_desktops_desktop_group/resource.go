package cloud_desktops_desktop_group

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	cloudDesktopsDesktopGroupCreateTimeout = time.Hour
	cloudDesktopsDesktopGroupDeleteTimeout = time.Hour
	cloudDesktopsDesktopGroupUpdateTimeout = 2 * time.Hour
)

type cloudDesktopDesktopGroupResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &cloudDesktopDesktopGroupResource{}
}

func (r *cloudDesktopDesktopGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_desktops_desktop_group"
}

func (r *cloudDesktopDesktopGroupResource) Configure(_ context.Context,
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

var (
	isMoreThanZero = []validator.Int64{int64validator.AtLeast(1)}
	isPositive     = []validator.Int64{int64validator.AtLeast(0)}
)

func (r *cloudDesktopDesktopGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Cloud Desktops Desktop Group. For more information see [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/concepts/desktops-and-groups)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"desktop_group_id": schema.StringAttribute{
				MarkdownDescription: "The id of the desktop group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: "The folder the dekstop group is in.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"image_id": schema.StringAttribute{
				MarkdownDescription: "The id of the desktop image the group is based on.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the desktop group.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the desktop group.",
				Optional:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"desktop_template": schema.SingleNestedAttribute{
				MarkdownDescription: "The configuration template for the desktop group.",
				Attributes: map[string]schema.Attribute{
					"resources":         getResourcesSpecSchema(),
					"network_interface": getNetworkInterfaceSpecSchema(),
					"boot_disk":         getBootDiskSpecSchema(),
					"data_disk":         getDataDiskSpecSchema(),
				},
				Optional: true,
			},
			"group_config": getGroupConfigSchema(),
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

func getResourcesSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The base resource configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"memory": schema.Int64Attribute{
				MarkdownDescription: "The number of gigabytes of RAM each desktop in this group would have.",
				Optional:            true,
				Validators:          isMoreThanZero,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: "The number of cores each desktop in this group would have.",
				Optional:            true,
				Validators:          []validator.Int64{int64validator.Between(1, 100)},
			},
			"core_fraction": schema.Int64Attribute{
				MarkdownDescription: "The baseline level of CPU performance each desktop in this group would have.",
				Optional:            true,
				Validators:          isMoreThanZero,
			},
		},
		Optional: true,
	}
}

func getNetworkInterfaceSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The base network interface configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The id of the network desktops from the group would use.",
				Required:            true,
			},
			"subnet_ids": schema.ListAttribute{
				MarkdownDescription: "The ids of the subnet networks desktops from the group would use.",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
		Optional: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}
}

func getBootDiskSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The boot disk configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"initialize_params": getDiskSpecSchema(),
		},
		Optional: true,
	}
}

func getDataDiskSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The data disk configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"initialize_params": getDiskSpecSchema(),
		},
		Optional: true,
	}
}

func getDiskSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "General data disk configuration",
		Attributes: map[string]schema.Attribute{
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of disk in gigabytes.",
				Optional:            true,
				Validators:          isMoreThanZero,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of disk. Allowed values: TYPE_UNSPECIFIED, HDD or SDD",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"TYPE_UNSPECIFIED",
						"HDD",
						"SSD",
					),
				},
			},
		},
		Optional: true,
	}
}

func getGroupConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The group configuration.",
		Attributes: map[string]schema.Attribute{
			"min_ready_desktops": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of ready desktops.",
				Optional:            true,
				Validators:          isPositive,
			},
			"max_desktops_amount": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of desktops.",
				Optional:            true,
				Validators:          isMoreThanZero,
			},
			"desktop_type": schema.StringAttribute{
				MarkdownDescription: "The type of the desktop group. Allowed: DESKTOP_TYPE_UNSPECIFIED, PERSISTENT, NON_PERSISTENT",
				Optional:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					"DESKTOP_TYPE_UNSPECIFIED",
					"PERSISTENT",
					"NON_PERSISTENT",
				)},
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "List of members in this desktop group.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The id of the member. More info in [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/DesktopGroup/create#yandex.cloud.access.Subject).",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the member. More info in [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/DesktopGroup/create#yandex.cloud.access.Subject).",
							Required:            true,
							Validators: []validator.String{stringvalidator.OneOf(
								"userAccount",
								"serviceAccount",
								"federatedUser",
								"system",
							)},
						},
					},
				},
			},
		},
		Optional: true,
	}
}

func (r *cloudDesktopDesktopGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DesktopGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var desktopProto *clouddesktop.DesktopGroup
	isNotFound := true
	if !state.DesktopGroupID.IsNull() {
		desktopGroupID := state.DesktopGroupID.ValueString()
		desktopProto, isNotFound = readDesktopGroupByID(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopGroupID)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if isNotFound {
		name := state.Name.ValueString()
		folderID := state.FolderID.ValueString()
		desktopProto, isNotFound, _ = readDesktopGroupByNameAndFolderID(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, name, folderID)
	}
	if isNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(desktopGroupToState(ctx, desktopProto, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudDesktopDesktopGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DesktopGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.FolderID = types.StringValue(converter.GetFolderID(plan.FolderID.ValueString(), r.providerConfig, &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, cloudDesktopsDesktopGroupCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	desktopGroup, diags := planToCreateRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	desktopGroupID := createDesktopGroup(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopGroup)
	if resp.Diagnostics.HasError() {
		return
	}
	defer func() {
		if resp.Diagnostics.HasError() {
			deleteDesktopGroup(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopGroupID)
		}
	}()

	// need to do all this, because sdk doesn't support setting up labels when creating the Desktop Group
	updateDesktopGroupReq, diag := planToUpdateRequest(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateDesktopGroup(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopGroupID, updateDesktopGroupReq, []string{"labels"})
	if resp.Diagnostics.HasError() {
		return
	}

	plan.DesktopGroupID = types.StringValue(desktopGroupID)
	plan.Id = types.StringValue(ConstructID(desktopGroup.Name, desktopGroup.FolderId, desktopGroup.DesktopImageId))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func getUpdatePaths(plan, state *clouddesktop.UpdateDesktopGroupRequest) []string {
	var updatePaths []string
	if state.DesktopImageId != plan.DesktopImageId {
		updatePaths = append(updatePaths, "desktop_image_id")
	}
	if state.Name != plan.Name {
		updatePaths = append(updatePaths, "name")
	}
	if state.Description != plan.Description {
		updatePaths = append(updatePaths, "description")
	}
	if !reflect.DeepEqual(state.Labels, plan.Labels) {
		updatePaths = append(updatePaths, "labels")
	}

	updatePaths = getResourcesUpdatePaths(plan.ResourcesSpec, state.ResourcesSpec, updatePaths)
	updatePaths = getGroupConfigUpdatePaths(plan.GroupConfig, state.GroupConfig, updatePaths)
	updatePaths = getDiskSpecUpdatePaths(plan.BootDiskSpec, state.BootDiskSpec, updatePaths, "boot_disk_spec")
	updatePaths = getDiskSpecUpdatePaths(plan.DataDiskSpec, state.DataDiskSpec, updatePaths, "data_disk_spec")
	// I don't know why I can't update network specs through API

	return updatePaths
}

func getResourcesUpdatePaths(plan, state *clouddesktop.ResourcesSpec, updatePaths []string) []string {
	prefix := "resources_spec"

	if state == nil && plan == nil {
		return updatePaths
	} else if state == nil || plan == nil {
		return append(updatePaths, prefix)
	}

	if state.Memory != plan.Memory {
		updatePaths = append(updatePaths, prefix+".memory")
	}
	if state.Cores != plan.Cores {
		updatePaths = append(updatePaths, prefix+".cores")
	}
	if state.CoreFraction != plan.CoreFraction {
		updatePaths = append(updatePaths, prefix+".core_fraction")
	}

	return updatePaths
}

func getGroupConfigUpdatePaths(plan, state *clouddesktop.DesktopGroupConfiguration, updatePaths []string) []string {
	prefix := "group_config"

	if state == nil && plan == nil {
		return updatePaths
	} else if state == nil || plan == nil {
		return append(updatePaths, prefix)
	}

	if state.MinReadyDesktops != plan.MinReadyDesktops {
		updatePaths = append(updatePaths, prefix+".min_ready_desktops")
	}
	if state.MaxDesktopsAmount != plan.MaxDesktopsAmount {
		updatePaths = append(updatePaths, prefix+".max_desktops_amount")
	}
	if state.DesktopType != plan.DesktopType {
		updatePaths = append(updatePaths, prefix+".desktop_type")
	}
	if !reflect.DeepEqual(state.Members, plan.Members) {
		updatePaths = append(updatePaths, prefix+".members")
	}

	return updatePaths
}

func getDiskSpecUpdatePaths(plan, state *clouddesktop.DiskSpec, updatePaths []string, prefix string) []string {
	if state == nil && plan == nil {
		return updatePaths
	} else if state == nil || plan == nil {
		return append(updatePaths, prefix)
	}

	if state.Size != plan.Size {
		updatePaths = append(updatePaths, prefix+".size")
	}
	if state.Type != plan.Type {
		updatePaths = append(updatePaths, prefix+".type")
	}

	return updatePaths
}

func (r *cloudDesktopDesktopGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DesktopGroup
	var state DesktopGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.FolderID = types.StringValue(converter.GetFolderID(plan.FolderID.ValueString(), r.providerConfig, &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, cloudDesktopsDesktopGroupUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	desktopGroupID := state.DesktopGroupID.ValueString()
	planDesktopGroup, diags := planToUpdateRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	stateDesktopGroup, diags := planToUpdateRequest(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatePaths := getUpdatePaths(planDesktopGroup, stateDesktopGroup)

	if len(updatePaths) > 0 {
		updateDesktopGroup(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopGroupID, planDesktopGroup, updatePaths)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(ConstructID(plan.Name.ValueString(), plan.FolderID.ValueString(), plan.DesktopImageID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *cloudDesktopDesktopGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DesktopGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, cloudDesktopsDesktopGroupDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	desktopGroupID := state.DesktopGroupID.ValueString()
	deleteDesktopGroup(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, desktopGroupID)
}

func (r *cloudDesktopDesktopGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name, folderID, imageID, err := DeconstructID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}

	desktopGroup, isNotFound, _ := readDesktopGroupByNameAndFolderID(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, name, folderID)
	if resp.Diagnostics.HasError() {
		return
	}
	if isNotFound {
		resp.Diagnostics.AddError(
			"Failed to Import resource",
			"No resource with such Name and FolderID",
		)
		return
	}

	var state DesktopGroup
	resp.Diagnostics.Append(desktopGroupToState(ctx, desktopGroup, &state)...)
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

	state.DesktopImageID = types.StringValue(imageID)
	state.Id = types.StringValue(req.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
