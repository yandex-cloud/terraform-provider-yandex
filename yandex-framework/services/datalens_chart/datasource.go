package datalens_chart

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var _ datasource.DataSourceWithConfigure = (*chartDataSource)(nil)

type chartDataSource struct {
	providerConfig *provider_config.Config
	client         *chartClient
}

func NewDataSource() datasource.DataSource {
	return &chartDataSource{}
}

func (d *chartDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_chart"
}

func (d *chartDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.AddWarning(
		"Experimental data source",
		"yandex_datalens_chart wraps DataLens chart endpoints that are marked "+
			"Experimental in the upstream API. The schema and behavior may "+
			"change in future provider versions.",
	)
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
	d.client = &chartClient{client: dlClient}
}

func (d *chartDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = chartDataSourceSchema()
}

func (d *chartDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens chart data source")

	var config chartModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := config.OrganizationId.ValueString()
	if orgID == "" {
		orgID = d.providerConfig.ProviderState.OrganizationID.ValueString()
	}

	apiResp, err := d.client.GetChart(ctx, orgID, config.Type.ValueString(), config.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens chart", err.Error())
		return
	}

	config.OrganizationId = types.StringValue(orgID)
	if err := unmarshalChartResponse(&config, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate chart state from response", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
