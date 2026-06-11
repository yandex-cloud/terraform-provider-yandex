package datalens_dataset

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens/wire"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var _ datasource.DataSourceWithConfigure = (*datasetDataSource)(nil)

type datasetDataSource struct {
	providerConfig *provider_config.Config
	client         *datasetClient
}

func NewDataSource() datasource.DataSource {
	return &datasetDataSource{}
}

func (d *datasetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_dataset"
}

func (d *datasetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T.", req.ProviderData),
		)
		return
	}
	d.providerConfig = providerConfig

	dlClient, err := datalens.NewClient(datalens.Config{
		Endpoint: providerConfig.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			t, err := providerConfig.SDK.CreateIAMToken(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to get IAM token: %w", err)
			}
			return t.IamToken, nil
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create DataLens client", err.Error())
		return
	}
	d.client = &datasetClient{client: dlClient}
}

func (d *datasetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema()
}

func (d *datasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens dataset data source")

	var config datasetModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := config.OrganizationId.ValueString()
	if orgID == "" {
		orgID = d.providerConfig.ProviderState.OrganizationID.ValueString()
	}

	apiResp, err := d.client.GetDataset(ctx, orgID, config.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens dataset", err.Error())
		return
	}

	config.OrganizationId = types.StringValue(orgID)
	if err := wire.Unmarshal(apiResp, &config); err != nil {
		resp.Diagnostics.AddError("Failed to populate dataset state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
