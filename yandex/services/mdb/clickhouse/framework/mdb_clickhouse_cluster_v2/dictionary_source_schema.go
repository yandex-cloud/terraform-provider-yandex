package mdb_clickhouse_cluster_v2

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2/validators"
)

func dictionaryPostgresqlSourceV2Schema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "PostgreSQL source using the native ClickHouse integration. Use this source for new PostgreSQL dictionaries instead of the deprecated ODBC-based postgresql_source.",
		Optional:    true,
		Validators:  []validator.Object{validators.DictionaryPostgresqlReplicasValidator{}},
		Attributes: map[string]schema.Attribute{
			"db": schema.StringAttribute{
				Description: "PostgreSQL database name.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"table": schema.StringAttribute{
				Description: "PostgreSQL table name. Specify exactly one of table or query. The source user must have SELECT access to the table.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("query")),
				},
			},
			"query": schema.StringAttribute{
				Description: "Custom SQL query returning the dictionary key and attributes. Specify exactly one of table or query. Cannot be combined with where.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"replicas": schema.ListNestedAttribute{
				Description: "PostgreSQL replicas used as dictionary sources. The API requires at least one replica. Replicas inherit port, user and password from the source unless overridden.",
				Required:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Description: "PostgreSQL replica host.",
							Required:    true,
							Validators:  []validator.String{stringvalidator.LengthBetween(1, 253)},
						},
						"priority": schema.Int64Attribute{
							Description: "Replica priority. Lower values have higher priority. Defaults to 0 when omitted.",
							Optional:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Replica port. Inherits the source port when omitted.",
							Optional:    true,
							Validators:  []validator.Int64{int64validator.Between(1, 65535)},
						},
						"user": schema.StringAttribute{
							Description: "Replica user. Inherits the source user when omitted.",
							Optional:    true,
						},
						"password": schema.StringAttribute{
							Description: "Replica password. Inherits the source password when omitted. Sensitive; stored in Terraform state. The API does not return passwords.",
							Optional:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"port": schema.Int64Attribute{
				Description: "Default PostgreSQL port inherited by replicas without a port override. Defaults to 5432.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(5432),
				Validators:  []validator.Int64{int64validator.Between(1, 65535)},
			},
			"user": schema.StringAttribute{
				Description: "Default PostgreSQL user inherited by replicas without a user override. Set this value or specify a user for every replica.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Default PostgreSQL password inherited by replicas without a password override. Sensitive; stored in Terraform state. The API does not return passwords.",
				Optional:    true,
				Sensitive:   true,
			},
			"invalidate_query": schema.StringAttribute{
				Description: "Query to check if the dictionary data has changed.",
				Optional:    true,
			},
			"ssl_mode": schema.StringAttribute{
				Description: "SSL mode for the PostgreSQL connection. When omitted, the current value is preserved. Valid values: DISABLE, ALLOW, PREFER, VERIFY_CA, VERIFY_FULL, REQUIRE.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("DISABLE", "ALLOW", "PREFER", "VERIFY_CA", "VERIFY_FULL", "REQUIRE"),
				},
			},
			"where": schema.StringAttribute{
				Description: "Selection criteria (WHERE clause) for data in the specified table. Cannot be combined with query.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("query")),
				},
			},
			"background_reconnect": schema.BoolAttribute{
				Description: "Reconnect to unavailable replicas in the background.",
				Optional:    true,
			},
		},
	}
}

func dictionaryYtSourceV2Schema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "YTsaurus source using the native ClickHouse integration. Available in public and internal clouds. Use this source for YTsaurus dictionaries; the legacy library-based source is unavailable starting with ClickHouse 26.8.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"http_proxy_urls": schema.ListAttribute{
				Description: "URLs of the YTsaurus HTTP proxies, reachable from the ClickHouse cluster. At least one non-empty URL is required.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"cypress_path": schema.StringAttribute{
				Description: "Cypress path to the YTsaurus table used as the dictionary source, for example //home/user/cities.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"oauth_token": schema.StringAttribute{
				Description: "OAuth token with read access to the YTsaurus table. Sensitive; stored in Terraform state. The API does not return this token.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func dataSourceDictionaryPostgresqlSourceV2Schema() datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		Description: "PostgreSQL source using the native ClickHouse integration. Use this source for new PostgreSQL dictionaries instead of the deprecated ODBC-based postgresql_source.",
		Computed:    true,
		Attributes: map[string]datasourceschema.Attribute{
			"db":    datasourceschema.StringAttribute{Computed: true, Description: "PostgreSQL database name."},
			"table": datasourceschema.StringAttribute{Computed: true, Description: "PostgreSQL table name."},
			"query": datasourceschema.StringAttribute{Computed: true, Description: "Custom SQL query."},
			"replicas": datasourceschema.ListNestedAttribute{
				Description: "PostgreSQL replicas used as dictionary sources. The API requires at least one replica. Replicas inherit port, user and password from the source unless overridden.",
				Computed:    true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: map[string]datasourceschema.Attribute{
						"host":     datasourceschema.StringAttribute{Computed: true, Description: "PostgreSQL replica host."},
						"priority": datasourceschema.Int64Attribute{Computed: true, Description: "Replica priority. Lower values have higher priority. Defaults to 0 when omitted."},
						"port":     datasourceschema.Int64Attribute{Computed: true, Description: "Replica port."},
						"user":     datasourceschema.StringAttribute{Computed: true, Description: "Replica user."},
						"password": datasourceschema.StringAttribute{Computed: true, Sensitive: true, Description: "Replica password. Sensitive; the API does not return this value."},
					},
				},
			},
			"port":                 datasourceschema.Int64Attribute{Computed: true, Description: "Default PostgreSQL port inherited by replicas without a port override. Defaults to 5432."},
			"user":                 datasourceschema.StringAttribute{Computed: true, Description: "Default PostgreSQL user for replicas without a user override."},
			"password":             datasourceschema.StringAttribute{Computed: true, Sensitive: true, Description: "Default PostgreSQL password for replicas. Sensitive; the API does not return this value."},
			"invalidate_query":     datasourceschema.StringAttribute{Computed: true, Description: "Query to check if the dictionary data has changed."},
			"ssl_mode":             datasourceschema.StringAttribute{Computed: true, Description: "SSL mode for the PostgreSQL connection. Valid values: DISABLE, ALLOW, PREFER, VERIFY_CA, VERIFY_FULL, REQUIRE."},
			"where":                datasourceschema.StringAttribute{Computed: true, Description: "Selection criteria (WHERE clause) for data in the specified table. Cannot be combined with query."},
			"background_reconnect": datasourceschema.BoolAttribute{Computed: true, Description: "Reconnect to unavailable replicas in the background."},
		},
	}
}

func dataSourceDictionaryYtSourceV2Schema() datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		Description: "YTsaurus source using the native ClickHouse integration. Available in public and internal clouds. Use this source for YTsaurus dictionaries; the legacy library-based source is unavailable starting with ClickHouse 26.8.",
		Computed:    true,
		Attributes: map[string]datasourceschema.Attribute{
			"http_proxy_urls": datasourceschema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "URLs of the YTsaurus HTTP proxies, reachable from the ClickHouse cluster. At least one non-empty URL is required."},
			"cypress_path":    datasourceschema.StringAttribute{Computed: true, Description: "Cypress path to the YTsaurus table used as the dictionary source, for example //home/user/cities."},
			"oauth_token":     datasourceschema.StringAttribute{Computed: true, Sensitive: true, Description: "OAuth token for accessing the YTsaurus table. Sensitive; the API does not return this value."},
		},
	}
}
