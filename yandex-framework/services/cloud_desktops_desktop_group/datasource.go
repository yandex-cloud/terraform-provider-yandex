package cloud_desktops_desktop_group

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type cloudDesktopDesktopGroupDatasource struct {
	providerConfig *provider_config.Config
}

func NewDatasource() datasource.DataSource {
	return &cloudDesktopDesktopGroupDatasource{}
}

func (d *cloudDesktopDesktopGroupDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_desktops_desktop_group"
}

func (d *cloudDesktopDesktopGroupDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}

func (d *cloudDesktopDesktopGroupDatasource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Cloud Desktops Desktop Group. For more information see [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/concepts/desktops-and-groups)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
			},
			"desktop_group_id": schema.StringAttribute{
				MarkdownDescription: "The id of the desktop group.",
				Computed:            true,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("desktop_group_id"),
						path.MatchRoot("name"),
					),
				},
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: "The folder the dekstop group is in.",
				Computed:            true,
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the desktop group.",
				Computed:            true,
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the desktop group.",
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				Computed:            true,
				ElementType:         types.StringType,
			},
			"desktop_template": schema.SingleNestedAttribute{
				MarkdownDescription: "The configuration template for the desktop group.",
				Attributes: map[string]schema.Attribute{
					"resources":         getDatasourceResourcesSpecSchema(),
					"network_interface": getDatasourceNetworkInterfaceSpecSchema(),
					"boot_disk":         getDatasourceBootDiskSpecSchema(),
					"data_disk":         getDatasourceDataDiskSpecSchema(),
				},
				Optional: true,
				Computed: true,
			},
			"group_config": getDatasourceGroupConfigSchema(),
		},
	}
}

func getDatasourceResourcesSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The base resource configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"memory": schema.Int64Attribute{
				MarkdownDescription: "The number of gigabytes of RAM each desktop in this group would have.",
				Computed:            true,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: "The number of cores each desktop in this group would have.",
				Computed:            true,
			},
			"core_fraction": schema.Int64Attribute{
				MarkdownDescription: "The baseline level of CPU performance each desktop in this group would have.",
				Computed:            true,
			},
		},
		Optional: true,
		Computed: true,
	}
}

func getDatasourceNetworkInterfaceSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The base network interface configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The id of the network desktops from the group would use.",
				Computed:            true,
			},
			"subnet_ids": schema.ListAttribute{
				MarkdownDescription: "The ids of the subnet networks desktops from the group would use.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
		Optional: true,
		Computed: true,
	}
}

func getDatasourceGroupConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The group configuration.",
		Attributes: map[string]schema.Attribute{
			"min_ready_desktops": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of ready desktops.",
				Computed:            true,
			},
			"max_desktops_amount": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of desktops.",
				Computed:            true,
			},
			"desktop_type": schema.StringAttribute{
				MarkdownDescription: "The type of the desktop group. Allowed: DESKTOP_TYPE_UNSPECIFIED, PERSISTENT, NON_PERSISTENT",
				Computed:            true,
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "List of members in this desktop group.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The id of the member. More info in [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/DesktopGroup/create#yandex.cloud.access.Subject).",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the member. More info in [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/DesktopGroup/create#yandex.cloud.access.Subject).",
							Computed:            true,
						},
					},
				},
			},
		},
		Optional: true,
		Computed: true,
	}
}

func getDatasourceBootDiskSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The boot disk configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"initialize_params": getDatasourceDiskSpecSchema(),
		},
		Optional: true,
		Computed: true,
	}
}

func getDatasourceDataDiskSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The data disk configuration for each desktop in the group.",
		Attributes: map[string]schema.Attribute{
			"initialize_params": getDatasourceDiskSpecSchema(),
		},
		Optional: true,
		Computed: true,
	}
}

func getDatasourceDiskSpecSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "General data disk configuration",
		Attributes: map[string]schema.Attribute{
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of disk in gigabytes.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of disk. Allowed values: TYPE_UNSPECIFIED, HDD or SDD",
				Computed:            true,
			},
		},
		Optional: true,
		Computed: true,
	}
}
func (d *cloudDesktopDesktopGroupDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DesktopGroupDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var desktopProto *clouddesktop.DesktopGroup
	isNotFound := true
	if !(state.DesktopGroupID.IsUnknown() || state.DesktopGroupID.IsNull()) {
		desktopProto, isNotFound = readDesktopGroupByID(ctx, d.providerConfig.SDKv2, &resp.Diagnostics, state.DesktopGroupID.ValueString())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	if isNotFound {
		folderId := converter.GetFolderID(state.FolderID.ValueString(), d.providerConfig, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		desktopProto, isNotFound, err = readDesktopGroupByNameAndFolderID(ctx, d.providerConfig.SDKv2, &resp.Diagnostics, state.Name.ValueString(), folderId)
	}
	if isNotFound {
		resp.Diagnostics.AddError(
			"Failed to Read resource",
			"Error while requesting API to List Desktop Groups: "+err.Error(),
		)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(desktopGroupToDataSourceState(ctx, desktopProto, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(ConstructID(state.Name.ValueString(), state.FolderID.ValueString(), ""))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
