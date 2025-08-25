package mdb_redis_user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	resp.TypeName = req.ProviderTypeName + "_mdb_redis_user"
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
		MarkdownDescription: "Manages a Redis user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-redis/).",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["id"],
			},
			"cluster_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the cluster to which user belongs to.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the user.",
			},
			"passwords": schema.SetAttribute{
				ElementType:         basetypes.StringType{},
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Set of user passwords",
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Is redis user enabled.",
				Computed:            true,
			},
			"acl_options": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Raw ACL string which has been inserted into the Redis",
			},
			"permissions": schema.SingleNestedAttribute{
				MarkdownDescription: "Set of permissions granted to the user.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"commands": schema.StringAttribute{
						MarkdownDescription: "Commands user can execute.",
						Computed:            true,
					},
					"categories": schema.StringAttribute{
						MarkdownDescription: "Command categories user has permissions to.",
						Computed:            true,
					},
					"patterns": schema.StringAttribute{
						MarkdownDescription: "Keys patterns user has permission to.",
						Computed:            true,
					},
					"pub_sub_channels": schema.StringAttribute{
						MarkdownDescription: "Channel patterns user has permissions to.",
						Computed:            true,
					},
					"sanitize_payload": schema.StringAttribute{
						MarkdownDescription: "SanitizePayload parameter.",
						Computed:            true,
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
	userRead(ctx, d.providerConfig.SDK, &resp.Diagnostics, &state)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Id = types.StringValue(resourceid.Construct(cid, userName))

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"delete": types.StringType,
			"update": types.StringType,
		}),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
