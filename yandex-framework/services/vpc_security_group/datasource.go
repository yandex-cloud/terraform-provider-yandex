package vpc_security_group

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	RuleDataSourceAttributes = map[string]schema.Attribute{
		"id":                schema.StringAttribute{Computed: true},
		"description":       schema.StringAttribute{Computed: true},
		"labels":            schema.MapAttribute{Computed: true, ElementType: types.StringType},
		"protocol":          schema.StringAttribute{Computed: true},
		"port":              schema.Int64Attribute{Computed: true},
		"from_port":         schema.Int64Attribute{Computed: true},
		"to_port":           schema.Int64Attribute{Computed: true},
		"v4_cidr_blocks":    schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"v6_cidr_blocks":    schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"security_group_id": schema.StringAttribute{Computed: true},
		"predefined_target": schema.StringAttribute{Computed: true},
	}

	_ datasource.DataSource              = &securityGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &securityGroupDataSource{}
)

type securityGroupDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &securityGroupDataSource{}
}

func (g *securityGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_security_group"
}

func (g *securityGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Debug(ctx, "Initializing VPC SecurityGroup schema")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Optional: true, Computed: true},
			"security_group_id": schema.StringAttribute{
				Optional: true,
				Computed: false,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
			},
			"created_at":  schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Optional: true, Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"labels":      schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"folder_id":   schema.StringAttribute{Optional: true, Computed: true},
			"network_id":  schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
		},
		Blocks: map[string]schema.Block{
			"ingress": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: RuleDataSourceAttributes,
				},
			},
			"egress": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: RuleDataSourceAttributes,
				},
			},
		},
	}
}

func (g *securityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state securityGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := state.SecurityGroupID.ValueString()
	if sgID == "" {
		folderID, d := validate.FolderID(state.FolderID, &g.providerConfig.ProviderState)
		resp.Diagnostics.Append(d)
		if resp.Diagnostics.HasError() {
			return
		}

		sgID, d = objectid.ResolveByNameAndFolderID(ctx, g.providerConfig.SDK, folderID, state.Name.ValueString(), sdkresolvers.SecurityGroupResolver)
		resp.Diagnostics.Append(d)
		if resp.Diagnostics.HasError() {
			return
		}

		state.SecurityGroupID = types.StringValue(sgID)
	}

	state.ID = types.StringValue(sgID)
	updateState(ctx, g.providerConfig.SDK, &state.securityGroupModel, &resp.Diagnostics, false)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *securityGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	g.providerConfig = providerConfig
}
