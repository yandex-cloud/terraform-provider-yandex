package datalens_workbook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens/wire"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var _ datasource.DataSourceWithConfigure = (*workbookDataSource)(nil)

type workbookDataSource struct {
	providerConfig *provider_config.Config
	client         *workbookClient
}

func NewDataSource() datasource.DataSource {
	return &workbookDataSource{}
}

func (d *workbookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_workbook"
}

func (d *workbookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = &workbookClient{client: dlClient}
}

func (d *workbookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a DataLens workbook.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the workbook."},
			"organization_id": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "The organization ID."},
			"collection_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "The parent collection ID, or null if the workbook is at root."},
			"title":           schema.StringAttribute{Computed: true, MarkdownDescription: "The workbook title."},
			"description":     schema.StringAttribute{Computed: true, MarkdownDescription: "The workbook description."},
			"tenant_id":       schema.StringAttribute{Computed: true, MarkdownDescription: "The DataLens tenant ID."},
			"status":          schema.StringAttribute{Computed: true, MarkdownDescription: "The workbook lifecycle status."},
			"created_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "User who created the workbook."},
			"created_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"updated_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "User who last updated the workbook."},
			"updated_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
		},
	}
}

func (d *workbookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens workbook data source")

	var config workbookModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := config.OrganizationId.ValueString()
	if orgID == "" {
		orgID = d.providerConfig.ProviderState.OrganizationID.ValueString()
	}

	apiResp, err := d.client.GetWorkbook(ctx, orgID, config.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens workbook", err.Error())
		return
	}

	config.OrganizationId = types.StringValue(orgID)
	if err := wire.Unmarshal(apiResp, &config); err != nil {
		resp.Diagnostics.AddError("Failed to parse get response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
