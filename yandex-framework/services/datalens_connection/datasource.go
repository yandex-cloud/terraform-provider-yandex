package datalens_connection

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSourceWithConfigure = (*connectionDataSource)(nil)

type connectionDataSource struct {
	providerConfig *provider_config.Config
	client         *connectionClient
}

func NewDataSource() datasource.DataSource {
	return &connectionDataSource{}
}

func (d *connectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_connection"
}

func (d *connectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig

	dlClient, err := datalens.NewClient(datalens.Config{
		Endpoint: providerConfig.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			resp, err := providerConfig.SDK.CreateIAMToken(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to get IAM token: %w", err)
			}
			return resp.IamToken, nil
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create DataLens client",
			fmt.Sprintf("Error creating the DataLens API client: %s", err),
		)
		return
	}
	d.client = &connectionClient{client: dlClient}
}

func (d *connectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a DataLens connection. " +
			"For more information, see [the official documentation](https://yandex.cloud/ru/docs/datalens/operations/api-start).",
		Attributes: map[string]schema.Attribute{
			// Required input
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the connection.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID for the DataLens instance. " +
					"If not specified, the provider-level `organization_id` is used.",
				Optional: true,
				Computed: true,
			},

			// Computed top-level fields
			"type": schema.StringAttribute{
				MarkdownDescription: "The connection type.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the connection.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the connection.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the resource.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last update timestamp of the resource.",
				Computed:            true,
			},

			// Connection-type-specific nested attributes (computed)
			"ydb": schema.SingleNestedAttribute{
				MarkdownDescription: "YDB connection configuration. Populated when `type` is `ydb`.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"workbook_id": schema.StringAttribute{
						MarkdownDescription: "The workbook ID where the connection is stored.",
						Computed:            true,
					},
					"dir_path": schema.StringAttribute{
						MarkdownDescription: "The directory path where the connection entry is stored.",
						Computed:            true,
					},
					"host": schema.StringAttribute{
						MarkdownDescription: "The hostname of the YDB database endpoint.",
						Computed:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "The port number of the YDB database endpoint.",
						Computed:            true,
					},
					"db_name": schema.StringAttribute{
						MarkdownDescription: "The YDB database name (path).",
						Computed:            true,
					},
					"cloud_id": schema.StringAttribute{
						MarkdownDescription: "The cloud ID where the YDB database is located.",
						Computed:            true,
					},
					"folder_id": schema.StringAttribute{
						MarkdownDescription: "The folder ID where the YDB database is located.",
						Computed:            true,
					},
					"service_account_id": schema.StringAttribute{
						MarkdownDescription: "The service account ID used to access the YDB database.",
						Computed:            true,
					},
					"auth_type": schema.StringAttribute{
						MarkdownDescription: "The authentication type for the connection.",
						Computed:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "The username for authentication.",
						Computed:            true,
					},
					"ssl_enable": schema.StringAttribute{
						MarkdownDescription: "Whether SSL is enabled for the connection.",
						Computed:            true,
					},
					"raw_sql_level": schema.StringAttribute{
						MarkdownDescription: "The level of raw SQL queries allowed.",
						Computed:            true,
					},
					"cache_ttl_sec": schema.Int64Attribute{
						MarkdownDescription: "The cache TTL in seconds.",
						Computed:            true,
					},
					"data_export_forbidden": schema.StringAttribute{
						MarkdownDescription: "Whether data export is forbidden.",
						Computed:            true,
					},
					"mdb_cluster_id": schema.StringAttribute{
						MarkdownDescription: "The Managed Databases cluster ID.",
						Computed:            true,
					},
					"mdb_folder_id": schema.StringAttribute{
						MarkdownDescription: "The folder ID for Managed Databases cluster lookup.",
						Computed:            true,
					},
					"delegation_is_set": schema.BoolAttribute{
						MarkdownDescription: "Whether delegation is configured for the connection.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens connection data source")

	var config connectionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := ""
	if !config.OrganizationId.IsNull() && !config.OrganizationId.IsUnknown() {
		orgID = config.OrganizationId.ValueString()
	} else {
		orgID = d.providerConfig.ProviderState.OrganizationID.ValueString()
	}

	connectionID := config.Id.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Fetching DataLens connection %s", connectionID))

	apiResponse, err := d.client.GetConnection(ctx, orgID, connectionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DataLens Connection",
			fmt.Sprintf("An unexpected error occurred while reading the connection data source.\n\n"+
				"Error: %s", err),
		)
		return
	}

	config.Id = types.StringValue(connectionID)
	config.OrganizationId = types.StringValue(orgID)

	// The API accepts "type" on create but returns "db_type" on read.
	if v, ok := apiResponse["db_type"].(string); ok {
		config.Type = types.StringValue(v)
	} else if v, ok := apiResponse["type"].(string); ok {
		config.Type = types.StringValue(v)
	}
	if v, ok := apiResponse["name"].(string); ok {
		config.Name = types.StringValue(v)
	}
	if v, ok := apiResponse["description"]; ok {
		if v == nil || v == "" {
			config.Description = types.StringNull()
		} else if s, ok := v.(string); ok {
			config.Description = types.StringValue(s)
		}
	}
	if v, ok := apiResponse["created_at"].(string); ok {
		config.CreatedAt = types.StringValue(v)
	}
	if v, ok := apiResponse["updated_at"].(string); ok {
		config.UpdatedAt = types.StringValue(v)
	}

	connType := config.Type.ValueString()
	switch connType {
	case "ydb":
		ydb := &ydbDataSourceConfigModel{}
		populateYdbDataSourceFromResponse(ydb, apiResponse)
		config.Ydb = ydb
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
