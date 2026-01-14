package trino_catalog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CatalogDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"clickhouse": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"connection_manager":    connectionManagerDataSourceSchema(),
					"on_premise":            onPremiseDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Clickhouse connector.",
				MarkdownDescription: "Configuration for Clickhouse connector.",
			},
			"delta_lake": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"file_system":           fileSystemDataSourceSchema(),
					"metastore":             metastoreDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Delta Lake connector.",
				MarkdownDescription: "Configuration for Delta Lake connector.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The resource description.",
				MarkdownDescription: "The resource description.",
			},
			"hive": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"file_system":           fileSystemDataSourceSchema(),
					"metastore":             metastoreDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Hive connector.",
				MarkdownDescription: "Configuration for Hive connector.",
			},
			"hudi": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"file_system":           fileSystemDataSourceSchema(),
					"metastore":             metastoreDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Hudi connector.",
				MarkdownDescription: "Configuration for Hudi connector.",
			},
			"iceberg": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"file_system":           fileSystemDataSourceSchema(),
					"metastore":             metastoreDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Iceberg connector.",
				MarkdownDescription: "Configuration for Iceberg connector.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The resource identifier.",
				MarkdownDescription: "The resource identifier.",
			},
			"cluster_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the Trino cluster.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "A set of key/value label pairs which assigned to resource.",
				MarkdownDescription: "A set of key/value label pairs which assigned to resource.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The resource name.",
				MarkdownDescription: "The resource name.",
			},
			"mysql": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"connection_manager":    mysqlConnectionManagerDataSourceSchema(),
					"on_premise":            onPremiseDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for MySQL connector.",
				MarkdownDescription: "Configuration for MySQL connector.",
			},
			"oracle": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"on_premise":            onPremiseDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Oracle connector.",
				MarkdownDescription: "Configuration for Oracle connector.",
			},
			"postgresql": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"connection_manager":    connectionManagerDataSourceSchema(),
					"on_premise":            onPremiseDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Postgresql connector.",
				MarkdownDescription: "Configuration for Postgresql connector.",
			},
			"greenplum": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"connection_manager":    connectionManagerDataSourceSchema(),
					"on_premise":            onPremiseDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for Greenplum/Cloudberry connector.",
				MarkdownDescription: "Configuration for Greenplum/Cloudberry connector.",
			},
			"sqlserver": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
					"on_premise":            onPremiseDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for SQLServer connector.",
				MarkdownDescription: "Configuration for SQLServer connector.",
			},
			"tpcds": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for TPCDS connector.",
				MarkdownDescription: "Configuration for TPCDS connector.",
			},
			"tpch": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesDataSourceSchema(),
				},
				Computed:            true,
				Description:         "Configuration for TPCH connector.",
				MarkdownDescription: "Configuration for TPCH connector.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Read: true,
			}),
		},
		Description: "Catalog for Managed Trino cluster.",
	}
}

func additionalPropertiesDataSourceSchema() schema.MapAttribute {
	return schema.MapAttribute{
		ElementType:         types.StringType,
		Computed:            true,
		Description:         "Additional properties.",
		MarkdownDescription: "Additional properties.",
	}
}

func connectionManagerDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				Computed:            true,
				Description:         "Connection ID.",
				MarkdownDescription: "Connection ID.",
			},
			"connection_properties": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "Additional connection properties.",
				MarkdownDescription: "Additional connection properties.",
			},
			"database": schema.StringAttribute{
				Computed:            true,
				Description:         "Database.",
				MarkdownDescription: "Database.",
			},
		},
		Computed:            true,
		Description:         "Configuration for connection manager connection.",
		MarkdownDescription: "Configuration for connection manager connection.",
	}
}

func onPremiseDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"connection_url": schema.StringAttribute{
				Computed:            true,
				Description:         "Connection URL.",
				MarkdownDescription: "Connection URL.",
			},
			"user_name": schema.StringAttribute{
				Computed:            true,
				Description:         "Name of the user.",
				MarkdownDescription: "Name of the user.",
			},
			"password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "Password of the user.",
				MarkdownDescription: "Password of the user.",
			},
		},
		Computed:            true,
		Description:         "Configuration for on-premise connection.",
		MarkdownDescription: "Configuration for on-premise connection.",
	}
}

func fileSystemDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"external_s3": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"aws_endpoint": schema.StringAttribute{
						Computed:            true,
						Description:         "AWS S3 compatible endpoint URL.",
						MarkdownDescription: "AWS S3 compatible endpoint URL.",
					},
					"aws_region": schema.StringAttribute{
						Computed:            true,
						Description:         "AWS region for S3 storage.",
						MarkdownDescription: "AWS region for S3 storage.",
					},
					"aws_access_key": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						Description:         "AWS access key ID for S3 authentication.",
						MarkdownDescription: "AWS access key ID for S3 authentication.",
					},
					"aws_secret_key": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						Description:         "AWS secret access key for S3 authentication.",
						MarkdownDescription: "AWS secret access key for S3 authentication.",
					},
				},
				Computed:            true,
				Description:         "Describes External S3 compatible file system.",
				MarkdownDescription: "Describes External S3 compatible file system.",
			},
			"s3": schema.SingleNestedAttribute{
				Attributes:          map[string]schema.Attribute{},
				Computed:            true,
				Description:         "Describes YandexCloud native S3 file system.",
				MarkdownDescription: "Describes YandexCloud native S3 file system.",
			},
		},
		Computed:            true,
		Description:         "File system configuration.",
		MarkdownDescription: "File system configuration.",
	}
}

func metastoreDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				Computed:            true,
				Description:         "The resource description.",
				MarkdownDescription: "The resource description.",
			},
		},
		Computed:            true,
		Description:         "Metastore configuration.",
		MarkdownDescription: "Metastore configuration.",
	}
}

func mysqlConnectionManagerDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				Computed:            true,
				Description:         "Connection ID.",
				MarkdownDescription: "Connection ID.",
			},
			"connection_properties": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "Additional connection properties.",
				MarkdownDescription: "Additional connection properties.",
			},
		},
		Computed:            true,
		Description:         "Configuration for MySQL connection manager connection.",
		MarkdownDescription: "Configuration for MySQL connection manager connection.",
	}
}
