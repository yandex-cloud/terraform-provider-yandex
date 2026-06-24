package mdb_mysql_user_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type userDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mysql_user_v2"
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf(
				"Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	d.providerConfig = providerConfig
}

func (d *userDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about a Yandex Managed MySQL user.",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Read: true,
			}),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource identifier in format `<cluster_id>:<user_name>`",
			},
			"cluster_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the MySQL cluster",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the user",
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The password of the user",
			},
			"generate_password": schema.BoolAttribute{
				Computed:    true,
				Description: "Generate password using Connection Manager",
			},
			"global_permissions": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of the user's global permissions",
			},
			"authentication_plugin": schema.StringAttribute{
				Computed:    true,
				Description: "Authentication plugin",
			},
			"connection_manager": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Connection Manager connection configuration",
			},
			"deletion_protection_mode": schema.StringAttribute{
				Computed:    true,
				Description: "Deletion Protection inhibits deletion of the user",
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.SetNestedBlock{
				Description: "Set of permissions granted to the user",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"database_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the database that the permission grants access to",
						},
						"roles": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "List of user's roles in the database",
						},
					},
				},
			},
			"connection_limits": schema.ListNestedBlock{
				Description: "User's connection limits",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_questions_per_hour": schema.Int64Attribute{
							Computed:    true,
							Description: "Max questions per hour",
						},
						"max_updates_per_hour": schema.Int64Attribute{
							Computed:    true,
							Description: "Max updates per hour",
						},
						"max_connections_per_hour": schema.Int64Attribute{
							Computed:    true,
							Description: "Max connections per hour",
						},
						"max_user_connections": schema.Int64Attribute{
							Computed:    true,
							Description: "Max user connections",
						},
					},
				},
			},
		},
	}
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state User
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()

	user := ReadUser(ctx, d.providerConfig, &resp.Diagnostics, cid, userName)
	if resp.Diagnostics.HasError() {
		return
	}

	specToState(ctx, user, &state, &resp.Diagnostics)
	state.Password = types.StringNull()
	state.GeneratePassword = types.BoolValue(false)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
