package mdb_mysql_database_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type bindingDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &bindingDataSource{}
}

func (d *bindingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mysql_database_v2"
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

func (d *bindingDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about a Yandex Managed MySQL database.",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Read: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the MySQL cluster. Provided by the client when the database is created.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the database.",
				Required:            true,
			},
			"deletion_protection_mode": schema.StringAttribute{
				MarkdownDescription: "Deletion Protection inhibits deletion of the database. Possible values: DELETION_PROTECTION_MODE_DISABLED, DELETION_PROTECTION_MODE_ENABLED, DELETION_PROTECTION_MODE_INHERITED.",
				Computed:            true,
			},
		},
	}
}

func (d *bindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state Database
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	db := ReadDatabase(ctx, d.providerConfig, &resp.Diagnostics, cid, dbName)
	if resp.Diagnostics.HasError() {
		return
	}

	specToState(db, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
