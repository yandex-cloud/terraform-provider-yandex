package datalens_connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
)

func ResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a DataLens connection resource. " +
			"For more information, see [the official documentation](https://yandex.cloud/ru/docs/datalens/operations/api-start).",
		Attributes: map[string]schema.Attribute{
			"id": defaultschema.Id(),
			"type": schema.StringAttribute{
				MarkdownDescription: "The connection type. Currently supported: `ydb`. " +
					"Changing this attribute forces recreation of the resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("ydb"),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["name"] +
					" Changing this attribute forces recreation of the resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["description"],
				Optional:            true,
			},
			"created_at": defaultschema.CreatedAt(),
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last update timestamp of the resource.",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID for the DataLens instance. " +
					"If not specified, the provider-level `organization_id` is used.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Connection-type-specific nested attributes
			"ydb": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for YDB connection type. " +
					"Must be specified when `type` is `ydb`.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"workbook_id": schema.StringAttribute{
						MarkdownDescription: "The workbook ID where the connection will be created. " +
							"Either `workbook_id` or `dir_path` must be specified. " +
							"Changing this attribute forces recreation of the resource.",
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"dir_path": schema.StringAttribute{
						MarkdownDescription: "The directory path where the connection entry will be stored " +
							"(used when connections are organized in folders instead of workbooks). " +
							"Either `workbook_id` or `dir_path` must be specified. " +
							"Changing this attribute forces recreation of the resource.",
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},

					// YDB-specific required fields
					"host": schema.StringAttribute{
						MarkdownDescription: "The hostname of the YDB database endpoint.",
						Required:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "The port number of the YDB database endpoint.",
						Required:            true,
					},
					"db_name": schema.StringAttribute{
						MarkdownDescription: "The YDB database name (path).",
						Required:            true,
					},
					"cloud_id": schema.StringAttribute{
						MarkdownDescription: "The cloud ID where the YDB database is located.",
						Required:            true,
					},
					"folder_id": schema.StringAttribute{
						MarkdownDescription: "The folder ID where the YDB database is located.",
						Required:            true,
					},
					"service_account_id": schema.StringAttribute{
						MarkdownDescription: "The service account ID used to access the YDB database.",
						Required:            true,
					},

					// YDB-specific optional fields
					"auth_type": schema.StringAttribute{
						MarkdownDescription: "The authentication type for the connection. " +
							"Possible values: `anonymous`, `password`, `oauth`.",
						Optional: true,
						Computed: true,
						Validators: []validator.String{
							stringvalidator.OneOf("anonymous", "password", "oauth"),
						},
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "The username for authentication (used when `auth_type` is `password`).",
						Optional:            true,
					},
					"token": schema.StringAttribute{
						MarkdownDescription: "The OAuth token for authentication (used when `auth_type` is `oauth`). " +
							"This is a write-only field and will not be returned by the API.",
						Optional:  true,
						Sensitive: true,
					},
					"ssl_ca": schema.StringAttribute{
						MarkdownDescription: "The SSL CA certificate for secure connections.",
						Optional:            true,
						Sensitive:           true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"ssl_enable": schema.StringAttribute{
						MarkdownDescription: "Whether SSL is enabled for the connection. " +
							"Possible values: `on`, `off`. Defaults to `on`.",
						Optional: true,
						Computed: true,
					},
					"raw_sql_level": schema.StringAttribute{
						MarkdownDescription: "The level of raw SQL queries allowed. " +
							"Possible values: `off`, `subselect`, `template`, `dashsql`. Defaults to `off`.",
						Optional: true,
						Computed: true,
						Validators: []validator.String{
							stringvalidator.OneOf("off", "subselect", "template", "dashsql"),
						},
					},
					"cache_ttl_sec": schema.Int64Attribute{
						MarkdownDescription: "The cache TTL in seconds. `null` means default caching behavior.",
						Optional:            true,
					},
					"data_export_forbidden": schema.StringAttribute{
						MarkdownDescription: "Whether data export is forbidden. " +
							"Possible values: `on`, `off`. Defaults to `off`.",
						Optional: true,
						Computed: true,
					},
					"mdb_cluster_id": schema.StringAttribute{
						MarkdownDescription: "The Managed Databases cluster ID (for managed YDB instances).",
						Optional:            true,
					},
					"mdb_folder_id": schema.StringAttribute{
						MarkdownDescription: "The folder ID for Managed Databases cluster lookup.",
						Optional:            true,
					},
					"delegation_is_set": schema.BoolAttribute{
						MarkdownDescription: "Whether delegation is configured for the connection.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}
