package datalens_dashboard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var _ datasource.DataSourceWithConfigure = (*dashboardDataSource)(nil)

type dashboardDataSource struct {
	providerConfig *provider_config.Config
	client         *dashboardClient
}

func NewDataSource() datasource.DataSource {
	return &dashboardDataSource{}
}

func (d *dashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_dashboard"
}

func (d *dashboardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.AddWarning(
		"Experimental data source",
		"yandex_datalens_dashboard wraps DataLens dashboard endpoints that are "+
			"marked Experimental in the upstream API. The schema and behavior "+
			"may change in future provider versions.",
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
	d.client = &dashboardClient{client: dlClient}
}

func (d *dashboardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dashboardDataSourceSchema()
}

func (d *dashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens dashboard data source")

	var config dashboardModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := config.OrganizationId.ValueString()
	if orgID == "" {
		orgID = d.providerConfig.ProviderState.OrganizationID.ValueString()
	}

	apiResp, err := d.client.GetDashboard(ctx, orgID, config.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens dashboard", err.Error())
		return
	}

	config.OrganizationId = types.StringValue(orgID)
	if err := unmarshalDashboardResponse(&config, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate dashboard state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
