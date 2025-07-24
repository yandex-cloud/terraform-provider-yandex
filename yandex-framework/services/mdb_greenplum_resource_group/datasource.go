package mdb_greenplum_resource_group

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
	resp.TypeName = req.ProviderTypeName + "_mdb_greenplum_resource_group"
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
		MarkdownDescription: "Get information about a greenplum resource group.",
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
			"is_user_defined": schema.BoolAttribute{
				MarkdownDescription: resourceSchema.Attributes["is_user_defined"].GetMarkdownDescription(),
				Computed:            true,
			},
			"concurrency": schema.Int64Attribute{
				MarkdownDescription: resourceSchema.Attributes["concurrency"].GetMarkdownDescription(),
				Computed:            true,
			},
			"cpu_rate_limit": schema.Int64Attribute{
				MarkdownDescription: resourceSchema.Attributes["cpu_rate_limit"].GetMarkdownDescription(),
				Computed:            true,
			},
			"memory_limit": schema.Int64Attribute{
				MarkdownDescription: resourceSchema.Attributes["memory_limit"].GetMarkdownDescription(),
				Computed:            true,
			},
			"memory_shared_quota": schema.Int64Attribute{
				MarkdownDescription: resourceSchema.Attributes["memory_shared_quota"].GetMarkdownDescription(),
				Computed:            true,
			},
			"memory_spill_ratio": schema.Int64Attribute{
				MarkdownDescription: resourceSchema.Attributes["memory_spill_ratio"].GetMarkdownDescription(),
				Computed:            true,
			},
		},
	}
}

func (d *bindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ResourceGroup
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	rgName := state.Name.ValueString()
	rg := readResourceGroup(ctx, d.providerConfig.SDK, &resp.Diagnostics, cid, rgName)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Id = types.StringValue(resourceid.Construct(cid, rgName))

	resourceGroupToState(rg, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
