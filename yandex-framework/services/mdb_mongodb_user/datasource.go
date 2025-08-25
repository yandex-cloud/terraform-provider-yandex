package mdb_mongodb_user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type bindingDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &bindingDataSource{}
}

func (d *bindingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mongodb_user"
}

func (d *bindingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}

func (d *bindingDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MongoDB user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/).",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to which user belongs to.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password of the user.",
				Computed:            true,
				Sensitive:           true,
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.SetNestedBlock{
				MarkdownDescription: "Set of permissions granted to the user.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"database_name": schema.StringAttribute{
							MarkdownDescription: "The name of the database that the permission grants access to.",
							Required:            true,
						},
						"roles": schema.SetAttribute{
							MarkdownDescription: "The roles of the user in this database. For more information see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts/users-and-roles).",
							Optional:            true,
							ElementType:         basetypes.StringType{},
						},
					},
				},
			},
		},
	}
}

func (d *bindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state User
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()
	user := readUser(ctx, d.providerConfig.SDK, &resp.Diagnostics, cid, userName)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Id = types.StringValue(resourceid.Construct(cid, userName))

	resp.Diagnostics.Append(userToState(user, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
