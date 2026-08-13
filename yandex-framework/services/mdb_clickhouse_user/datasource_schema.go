package mdb_clickhouse_user

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon/usersettings"
)

func DataSourceUserSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a ClickHouse user within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the ClickHouse cluster. Provided by the client when the user is created.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the ClickHouse user. Provided by the client when the user is created.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the ClickHouse user. Provided by the client when the user is created.",
				Computed:            true,
				Sensitive:           true,
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method for the user. Possible values are `password`, `iam`. Default is `password`.",
				Computed:            true,
			},
			"connection_manager": DataSourceConnectionManagerSchema(),
		},
		Blocks: map[string]schema.Block{
			"permission": DataSourcePermissionSchema(),
			"quota":      DataSourceQuotasSchema(),
			"settings":   DataSourceSettingsSchema(),
		},
	}
}

func DataSourcePermissionSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"database_name": schema.StringAttribute{
					Computed: true,
				},
			},
		},
	}
}

func DataSourceQuotasSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"interval_duration": schema.Int64Attribute{
					Computed: true,
				},
				"queries": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"errors": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"result_rows": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"read_rows": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"execution_time": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func DataSourceSettingsSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: usersettings.DataSourceAttributes(),
	}
}

func DataSourceConnectionManagerSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Connection Manager connection configuration. Filled in by the server automatically.",
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "ID of Connection Manager connection. Filled in by the server automatically. String.",
				Computed:            true,
			},
		},
		Computed: true,
	}
}
