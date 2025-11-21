package cloud_desktops_desktop

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

type cloudDesktopDesktopDatasource struct {
	providerConfig *provider_config.Config
}

func NewDatasource() datasource.DataSource {
	return &cloudDesktopDesktopDatasource{}
}

func (d *cloudDesktopDesktopDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_desktops_desktop"
}

func (d *cloudDesktopDesktopDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cloudDesktopDesktopDatasource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Cloud Desktops Desktop. For more information see [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/concepts/desktops-and-groups)",
		Attributes: map[string]schema.Attribute{
			"desktop_id": schema.StringAttribute{
				MarkdownDescription: "The id of the Desktop",
				Computed:            true,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("desktop_id"),
						path.MatchRoot("name"),
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Desktop",
				Optional:            true,
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: "The folder containing the Desktop",
				Computed:            true,
				Optional:            true,
			},
			"desktop_group_id": schema.StringAttribute{
				MarkdownDescription: "The id of the Desktop Group to which the Desktop belongs",
				Computed:            true,
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "The list of members which can use the Desktop",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"subject_id": schema.StringAttribute{
							MarkdownDescription: "Identity of the access binding. See [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/Desktop/create#yandex.cloud.clouddesktop.v1.api.User)",
							Computed:            true,
						},
						"subject_type": schema.StringAttribute{
							MarkdownDescription: "Type of the access binding. See [the official documentation](https://yandex.cloud/ru/docs/cloud-desktop/api-ref/grpc/Desktop/create#yandex.cloud.clouddesktop.v1.api.User)",
							Computed:            true,
						},
					},
				},
				Computed: true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *cloudDesktopDesktopDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DesktopDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var desktopProto *clouddesktop.Desktop
	isNotFound := true
	if !(state.DesktopId.IsUnknown() || state.DesktopId.IsNull()) {
		desktopProto, isNotFound = readDesktopByID(ctx, d.providerConfig.SDKv2, &resp.Diagnostics, state.DesktopId.ValueString())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if isNotFound {
		folderId := converter.GetFolderID(state.FolderId.ValueString(), d.providerConfig, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		desktopProto = readDesktopByNameAndFolderID(ctx, d.providerConfig.SDKv2, &resp.Diagnostics, state.Name.ValueString(), folderId)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(desktopToDataSourceState(ctx, desktopProto, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
