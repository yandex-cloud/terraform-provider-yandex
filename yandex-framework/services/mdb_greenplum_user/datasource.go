package mdb_greenplum_user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type bindingDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &bindingDataSource{}
}

func (d *bindingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_greenplum_user"
}

func (d *bindingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bindingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about a greenplum user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: resourceSchema.Attributes["id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: resourceSchema.Attributes["cluster_id"].GetMarkdownDescription(),
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: resourceSchema.Attributes["name"].GetMarkdownDescription(),
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: resourceSchema.Attributes["password"].GetMarkdownDescription(),
				Sensitive:           true,
				Computed:            true,
			},
			"resource_group": schema.StringAttribute{
				MarkdownDescription: resourceSchema.Attributes["resource_group"].GetMarkdownDescription(),
				Computed:            true,
			},
		},
	}
}

func (d *bindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state User
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()
	user := readUser(ctx, d.providerConfig.SDK, &resp.Diagnostics, cid, userName)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Id = types.StringValue(resourceid.Construct(cid, userName))

	userToState(user, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
