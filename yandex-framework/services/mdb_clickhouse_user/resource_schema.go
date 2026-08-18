package mdb_clickhouse_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon/usersettings"
)

func UserSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a ClickHouse user within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the ClickHouse cluster. Provided by the client when the user is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the ClickHouse user. Provided by the client when the user is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the ClickHouse user. Provided by the client when the user is created.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
				},
			},
			"password_wo": schema.StringAttribute{
				MarkdownDescription: "Password of the ClickHouse user. This attribute is write-only and is not stored in state. Requires `password_wo_version` to trigger updates. Write-only arguments are supported in Terraform 1.11 and later.",
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				MarkdownDescription: "A version number for the write-only password. Increment this to trigger a password update.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo")),
				},
			},
			"generate_password": schema.BoolAttribute{
				MarkdownDescription: "Generate password using Connection Manager. Allowed values: `true` or `false`.\n\n~> **For password authentication, must specify exactly one of password, password_wo, or generate_password**.\n",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method for the user. Possible values are `password`, `iam`. Default is `password`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultUserAuthMethod),
				Validators:          UserAuthMethod_validator,
			},
			"connection_manager": ConnectionManagerSchema(),
		},
		Blocks: map[string]schema.Block{
			"permission": PermissionSchema(),
			"quota":      QuotasSchema(),
			"settings":   SettingsSchema(),
		},
	}
}

func PermissionSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "Block represents databases that are permitted to user.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"database_name": schema.StringAttribute{
					MarkdownDescription: "Name of the database that the permission grants access to.",
					Required:            true,
				},
			},
		},
	}
}

func QuotasSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "ClickHouse quota representation. Each quota associated with an user and limits it resource usage for an interval. For more information, see [the official documentation](https://clickhouse.com/docs/en/operations/quotas)",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"interval_duration": schema.Int64Attribute{MarkdownDescription: "Duration of interval for quota in milliseconds.", Required: true},
				"queries":           schema.Int64Attribute{MarkdownDescription: "The total number of queries. 0 - unlimited.", Optional: true},
				"errors":            schema.Int64Attribute{MarkdownDescription: "The number of queries that threw exception. 0 - unlimited.", Optional: true},
				"result_rows":       schema.Int64Attribute{MarkdownDescription: "The total number of rows given as the result. 0 - unlimited.", Optional: true},
				"read_rows":         schema.Int64Attribute{MarkdownDescription: "The total number of source rows read from tables for running the query, on all remote servers. 0 - unlimited.", Optional: true},
				"execution_time":    schema.Int64Attribute{MarkdownDescription: "The total query execution time, in milliseconds (wall time). 0 - unlimited.", Optional: true},
			},
		},
	}
}

func SettingsSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "Block represents ClickHouse user settings. For more information, see [the official documentation](https://clickhouse.com/docs/ru/operations/settings/settings)",
		Attributes:          usersettings.ResourceAttributes(),
	}
}

func ConnectionManagerSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Connection Manager connection configuration. Filled in by the server automatically.",
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "ID of Connection Manager connection. Filled in by the server automatically. String.",
				Computed:            true,
			},
		},
		Computed: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
	}
}
