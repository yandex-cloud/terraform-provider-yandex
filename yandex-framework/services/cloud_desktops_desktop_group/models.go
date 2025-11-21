package cloud_desktops_desktop_group

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type DesktopGroup struct {
	Id types.String `tfsdk:"id"`

	DesktopGroupID types.String `tfsdk:"desktop_group_id"`
	FolderID       types.String `tfsdk:"folder_id"`
	DesktopImageID types.String `tfsdk:"image_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Labels         types.Map    `tfsdk:"labels"`

	DesktopTemplate DesktopTemplate `tfsdk:"desktop_template"`
	GroupConfig     GroupConfig     `tfsdk:"group_config"`
	Timeouts        timeouts.Value  `tfsdk:"timeouts"`
}

type DesktopGroupDataSource struct {
	Id types.String `tfsdk:"id"`

	DesktopGroupID types.String `tfsdk:"desktop_group_id"`
	FolderID       types.String `tfsdk:"folder_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Labels         types.Map    `tfsdk:"labels"`

	DesktopTemplate *DesktopTemplate `tfsdk:"desktop_template"`
	GroupConfig     *GroupConfig     `tfsdk:"group_config"`
}

type DesktopTemplate struct {
	Networks  NetworkInterface `tfsdk:"network_interface"`
	BootDisk  DiskEncased      `tfsdk:"boot_disk"`
	DataDisk  DiskEncased      `tfsdk:"data_disk"`
	Resources Resources        `tfsdk:"resources"`
}

type NetworkInterface struct {
	NetworkID types.String `tfsdk:"network_id"`
	SubnetIDs types.List   `tfsdk:"subnet_ids"`
}

type DiskEncased struct {
	Disk Disk `tfsdk:"initialize_params"`
}

type Disk struct {
	Size types.Int64  `tfsdk:"size"`
	Type types.String `tfsdk:"type"`
}

type GroupConfig struct {
	MinReadyDesktops  types.Int64  `tfsdk:"min_ready_desktops"`
	MaxDesktopsAmount types.Int64  `tfsdk:"max_desktops_amount"`
	DesktopType       types.String `tfsdk:"desktop_type"`
	Members           types.List   `tfsdk:"members"`
}

type GroupMember struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

type Resources struct {
	Memory       types.Int64 `tfsdk:"memory"`
	Cores        types.Int64 `tfsdk:"cores"`
	CoreFraction types.Int64 `tfsdk:"core_fraction"`
}

var memberType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":   types.StringType,
		"type": types.StringType,
	},
}

func desktopGroupToState(ctx context.Context, desktopGroup *clouddesktop.DesktopGroup, state *DesktopGroup) diag.Diagnostics {
	state.DesktopGroupID = types.StringValue(desktopGroup.Id)
	state.FolderID = types.StringValue(desktopGroup.FolderId)
	state.Name = types.StringValue(desktopGroup.Name)
	state.Description = types.StringValue(desktopGroup.Description)

	var diag diag.Diagnostics
	state.Labels, diag = types.MapValueFrom(ctx, types.StringType, desktopGroup.Labels)
	if diag.HasError() {
		return diag
	}

	state.DesktopTemplate.Resources.Memory = types.Int64Value(datasize.ToGigabytes(desktopGroup.ResourcesSpec.Memory))
	state.DesktopTemplate.Resources.Cores = types.Int64Value(desktopGroup.ResourcesSpec.Cores)
	state.DesktopTemplate.Resources.CoreFraction = types.Int64Value(desktopGroup.ResourcesSpec.CoreFraction)

	diskToState(desktopGroup.BootDiskSpec, &state.DesktopTemplate.BootDisk)
	diskToState(desktopGroup.DataDiskSpec, &state.DesktopTemplate.DataDisk)

	diag = networkInterfacesToState(ctx, desktopGroup.NetworkInterfaceSpec, &state.DesktopTemplate.Networks)
	if diag.HasError() {
		return diag
	}

	return groupConfigToState(desktopGroup.GroupConfig, &state.GroupConfig)
}

func desktopGroupToDataSourceState(ctx context.Context, desktopGroup *clouddesktop.DesktopGroup, state *DesktopGroupDataSource) diag.Diagnostics {
	state.DesktopGroupID = types.StringValue(desktopGroup.Id)
	state.FolderID = types.StringValue(desktopGroup.FolderId)
	state.Name = types.StringValue(desktopGroup.Name)
	state.Description = types.StringValue(desktopGroup.Description)

	var diag diag.Diagnostics
	state.Labels, diag = types.MapValueFrom(ctx, types.StringType, desktopGroup.Labels)
	if diag.HasError() {
		return diag
	}

	if state.DesktopTemplate == nil {
		state.DesktopTemplate = &DesktopTemplate{}
	}
	state.DesktopTemplate.Resources.Memory = types.Int64Value(datasize.ToGigabytes(desktopGroup.ResourcesSpec.Memory))
	state.DesktopTemplate.Resources.Cores = types.Int64Value(desktopGroup.ResourcesSpec.Cores)
	state.DesktopTemplate.Resources.CoreFraction = types.Int64Value(desktopGroup.ResourcesSpec.CoreFraction)

	diskToState(desktopGroup.BootDiskSpec, &state.DesktopTemplate.BootDisk)
	diskToState(desktopGroup.DataDiskSpec, &state.DesktopTemplate.DataDisk)

	diag = networkInterfacesToState(ctx, desktopGroup.NetworkInterfaceSpec, &state.DesktopTemplate.Networks)
	if diag.HasError() {
		return diag
	}

	if state.GroupConfig == nil {
		state.GroupConfig = &GroupConfig{}
	}
	return groupConfigToState(desktopGroup.GroupConfig, state.GroupConfig)
}

func diskToState(disk *clouddesktop.DiskSpec, state *DiskEncased) {
	state.Disk.Size = types.Int64Value(datasize.ToGigabytes(disk.Size))
	state.Disk.Type = types.StringValue(disk.Type.String())
}

func networkInterfacesToState(ctx context.Context, spec *clouddesktop.NetworkInterfaceSpec, state *NetworkInterface) diag.Diagnostics {
	state.NetworkID = types.StringValue(spec.NetworkId)

	var diag diag.Diagnostics
	state.SubnetIDs, diag = types.ListValueFrom(ctx, types.StringType, spec.SubnetIds)
	return diag
}

func groupConfigToState(group *clouddesktop.DesktopGroupConfiguration, state *GroupConfig) diag.Diagnostics {
	state.MinReadyDesktops = types.Int64Value(group.MinReadyDesktops)
	state.MaxDesktopsAmount = types.Int64Value(group.MaxDesktopsAmount)
	state.DesktopType = types.StringValue(group.DesktopType.String())

	var diag diag.Diagnostics
	var memberValues []attr.Value
	for _, member := range group.Members {
		memberValue, diagInt := types.ObjectValue(memberType.AttrTypes, map[string]attr.Value{
			"id":   types.StringValue(member.Id),
			"type": types.StringValue(member.Type),
		})

		memberValues = append(memberValues, memberValue)
		diag.Append(diagInt...)
	}
	if diag.HasError() {
		return diag
	}

	state.Members, diag = types.ListValue(memberType, memberValues)
	return diag
}

func planToCreateRequest(ctx context.Context, plan *DesktopGroup) (*clouddesktop.CreateDesktopGroupRequest, diag.Diagnostics) {
	request := &clouddesktop.CreateDesktopGroupRequest{}

	request.Name = plan.Name.ValueString()
	request.FolderId = plan.FolderID.ValueString()
	request.Description = plan.Description.ValueString()
	request.DesktopImageId = plan.DesktopImageID.ValueString()

	request.NetworkInterfaceSpec = &clouddesktop.NetworkInterfaceSpec{}
	diag := planToNetworkInterfaces(ctx, &plan.DesktopTemplate.Networks, request.NetworkInterfaceSpec)
	if diag.HasError() {
		return nil, diag
	}

	request.ResourcesSpec = &clouddesktop.ResourcesSpec{}
	planToResources(&plan.DesktopTemplate.Resources, request.ResourcesSpec)

	request.BootDiskSpec = &clouddesktop.DiskSpec{}
	diag = planToDisk(&plan.DesktopTemplate.BootDisk.Disk, request.BootDiskSpec)
	if diag.HasError() {
		return nil, diag
	}

	request.DataDiskSpec = &clouddesktop.DiskSpec{}
	diag = planToDisk(&plan.DesktopTemplate.DataDisk.Disk, request.DataDiskSpec)
	if diag.HasError() {
		return nil, diag
	}

	request.GroupConfig = &clouddesktop.DesktopGroupConfiguration{}
	diag = planToGroupConfig(ctx, &plan.GroupConfig, request.GroupConfig)
	if diag.HasError() {
		return nil, diag
	}

	return request, diag
}

func planToUpdateRequest(ctx context.Context, plan *DesktopGroup) (*clouddesktop.UpdateDesktopGroupRequest, diag.Diagnostics) {
	request := &clouddesktop.UpdateDesktopGroupRequest{}

	request.DesktopGroupId = plan.DesktopGroupID.ValueString()
	request.Name = plan.Name.ValueString()
	request.Description = plan.Description.ValueString()

	request.Labels = make(map[string]string, 0)
	diag := plan.Labels.ElementsAs(ctx, &request.Labels, false)
	if diag.HasError() {
		return nil, diag
	}

	request.ResourcesSpec = &clouddesktop.ResourcesSpec{}
	planToResources(&plan.DesktopTemplate.Resources, request.ResourcesSpec)

	request.BootDiskSpec = &clouddesktop.DiskSpec{}
	diag = planToDisk(&plan.DesktopTemplate.BootDisk.Disk, request.BootDiskSpec)
	if diag.HasError() {
		return nil, diag
	}

	request.DataDiskSpec = &clouddesktop.DiskSpec{}
	planToDisk(&plan.DesktopTemplate.DataDisk.Disk, request.DataDiskSpec)
	if diag.HasError() {
		return nil, diag
	}

	request.GroupConfig = &clouddesktop.DesktopGroupConfiguration{}
	diag = planToGroupConfig(ctx, &plan.GroupConfig, request.GroupConfig)
	if diag.HasError() {
		return nil, diag
	}

	return request, diag
}

func planToNetworkInterfaces(ctx context.Context, plan *NetworkInterface, spec *clouddesktop.NetworkInterfaceSpec) diag.Diagnostics {
	spec.NetworkId = plan.NetworkID.ValueString()
	spec.SubnetIds = make([]string, 0)

	return plan.SubnetIDs.ElementsAs(ctx, &spec.SubnetIds, false)
}

func planToResources(plan *Resources, spec *clouddesktop.ResourcesSpec) {
	spec.Memory = datasize.ToBytes(plan.Memory.ValueInt64())
	spec.Cores = plan.Cores.ValueInt64()
	spec.CoreFraction = plan.CoreFraction.ValueInt64()
}

func planToDisk(plan *Disk, spec *clouddesktop.DiskSpec) diag.Diagnostics {
	var diag diag.Diagnostics
	spec.Size = datasize.ToBytes(plan.Size.ValueInt64())

	num, ok := clouddesktop.DiskSpec_Type_value[plan.Type.ValueString()]
	if !ok {
		diag.AddError(
			"Error converting",
			"Unsupported disk type: "+plan.Type.String(),
		)
		return diag
	}

	spec.Type = clouddesktop.DiskSpec_Type(num)
	return diag
}

func planToGroupConfig(ctx context.Context, plan *GroupConfig, spec *clouddesktop.DesktopGroupConfiguration) diag.Diagnostics {
	var diag diag.Diagnostics
	spec.MinReadyDesktops = plan.MinReadyDesktops.ValueInt64()
	spec.MaxDesktopsAmount = plan.MaxDesktopsAmount.ValueInt64()

	num, ok := clouddesktop.DesktopGroupConfiguration_DesktopType_value[plan.DesktopType.ValueString()]
	if !ok {
		diag.AddError(
			"Error converting",
			"Unsupported desktop type: "+plan.DesktopType.String(),
		)
		return diag
	}

	spec.DesktopType = clouddesktop.DesktopGroupConfiguration_DesktopType(num)

	members := []GroupMember{}
	diag = plan.Members.ElementsAs(ctx, &members, false)
	if diag.HasError() {
		return diag
	}

	for _, member := range members {
		spec.Members = append(spec.Members, &access.Subject{
			Id:   member.ID.ValueString(),
			Type: member.Type.ValueString(),
		})
	}

	return diag
}

func ConstructID(name, folderID, desktopImageID string) string {
	return strings.Join([]string{name, folderID, desktopImageID}, ",")
}

func DeconstructID(ID string) (string, string, string, error) {
	vals := strings.Split(ID, ",")
	if len(vals) != 3 {
		return "", "", "", fmt.Errorf("Invalid resource id format: %q", ID)
	}

	return vals[0], vals[1], vals[2], nil
}
