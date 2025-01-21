package datasphere_community

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type communityDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &communityDataSource{}
}

func (d *communityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasphere_community"
}

func (d *communityDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Info(ctx, "Initializing datasphere community schema")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Required: true},
			"organization_id":    schema.StringAttribute{Computed: true},
			"created_at":         schema.StringAttribute{Computed: true},
			"created_by":         schema.StringAttribute{Computed: true},
			"billing_account_id": schema.StringAttribute{Optional: true},
			"name":               schema.StringAttribute{Computed: true},
			"description":        schema.StringAttribute{Computed: true},
			"labels":             schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}

}

func (d *communityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading community data source")

	var configCommunity communityDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configCommunity)...)

	tflog.Info(ctx, fmt.Sprintf("Making API call to fetch community data for community %s", configCommunity.Id.ValueString()))
	existingCommunity, err := d.providerConfig.SDK.Datasphere().Community().Get(ctx,
		&datasphere.GetCommunityRequest{CommunityId: configCommunity.Id.ValueString()},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch DataSource",
			fmt.Sprintf("An unexpected error occurred while attempting to fetch datasource "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	convertToTerraformModel(ctx, &configCommunity, existingCommunity, &resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &configCommunity)...)
}

func (d *communityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. "+
				"Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}
