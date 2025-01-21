package mdb_opensearch_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	common_schema "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/schema"
)

func NewDataSource() datasource.DataSource {
	return &openSearchClusterDataSource{}
}

// ex dataSourceYandexMDBOpenSearchCluster
type openSearchClusterDataSource struct {
	providerConfig *provider_config.Config
}

// Configure implements datasource.DataSource.
func (o *openSearchClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	o.providerConfig = providerConfig
}

// Metadata implements datasource.DataSource.
func (o *openSearchClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_opensearch_cluster"
}

// Read implements datasource.DataSource.
func (o *openSearchClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config model.OpenSearch
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ClusterID.ValueString() == "" && config.Name.ValueString() == "" {
		resp.Diagnostics.AddError(
			"At least one of cluster_id or name is required",
			"The cluster ID or Name must be specified in the configuration",
		)
		return
	}

	clusterID := config.ClusterID.ValueString()
	if clusterID == "" {
		folderID, d := validate.FolderID(config.FolderID, &o.providerConfig.ProviderState)
		resp.Diagnostics.Append(d)
		if resp.Diagnostics.HasError() {
			return
		}

		clusterID, d = objectid.ResolveByNameAndFolderID(ctx, o.providerConfig.SDK, folderID, config.Name.ValueString(), sdkresolvers.OpenSearchClusterResolver)
		resp.Diagnostics.Append(d)
		if resp.Diagnostics.HasError() {
			return
		}

		config.ClusterID = types.StringValue(clusterID)
	}

	config.ID = types.StringValue(clusterID)
	updateState(ctx, o.providerConfig.SDK, &config, &resp.Diagnostics, false)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// Schema implements datasource.DataSource.
func (o *openSearchClusterDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Info(ctx, "Initializing opensearch data source schema")
	resp.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"config": schema.SingleNestedBlock{
				Description: "Configuration of the OpenSearch cluster.",
				Blocks: map[string]schema.Block{
					"opensearch": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"plugins": schema.SetAttribute{
								Computed:    true,
								Optional:    true,
								ElementType: types.StringType,
							},
						},
						Blocks: map[string]schema.Block{
							"node_groups": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"resources": common_schema.NodeResource(),
									},
									Attributes: map[string]schema.Attribute{
										"name":        schema.StringAttribute{Computed: true},
										"hosts_count": schema.Int64Attribute{Computed: true},
										"zone_ids": schema.SetAttribute{
											Optional:    true,
											Computed:    true,
											ElementType: types.StringType,
										},
										"subnet_ids": schema.ListAttribute{
											Optional:    true,
											Computed:    true,
											ElementType: types.StringType,
										},

										"assign_public_ip": schema.BoolAttribute{Computed: true, Optional: true},
										"roles": schema.SetAttribute{
											Optional:    true,
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
							},
						},
					},
					"dashboards": schema.SingleNestedBlock{
						Blocks: map[string]schema.Block{
							"node_groups": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"resources": common_schema.NodeResource(),
									},
									Attributes: map[string]schema.Attribute{
										"name":        schema.StringAttribute{Required: true},
										"hosts_count": schema.Int64Attribute{Required: true},
										"zone_ids": schema.SetAttribute{
											Optional:    true,
											Computed:    true,
											ElementType: types.StringType,
										},
										"subnet_ids": schema.ListAttribute{
											Optional:    true,
											Computed:    true,
											ElementType: types.StringType,
										},

										"assign_public_ip": schema.BoolAttribute{Computed: true, Optional: true},
									},
								},
							},
						},
					},
					"access": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"data_transfer": schema.BoolAttribute{Computed: true},
							"serverless":    schema.BoolAttribute{Computed: true},
						},
					},
				},
				Attributes: map[string]schema.Attribute{
					"version":        schema.StringAttribute{Computed: true, Optional: true},
					"admin_password": schema.StringAttribute{Computed: true, Optional: true, Sensitive: true},
				},
			},
			"maintenance_window": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Computed: true},
					"day":  schema.StringAttribute{Computed: true},
					"hour": schema.Int64Attribute{Computed: true},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"cluster_id":  schema.StringAttribute{Computed: true, Optional: true},
			"folder_id":   schema.StringAttribute{Computed: true, Optional: true},
			"created_at":  schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true, Optional: true},
			"description": schema.StringAttribute{Computed: true, Optional: true},
			"labels": schema.MapAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"environment": schema.StringAttribute{Computed: true},
			"hosts":       common_schema.Hosts(),
			"network_id":  schema.StringAttribute{Computed: true},
			"health":      schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
			"security_group_ids": schema.SetAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"service_account_id":  schema.StringAttribute{Computed: true, Optional: true},
			"deletion_protection": schema.BoolAttribute{Computed: true, Optional: true},
			"auth_settings": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"saml": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"enabled":                   schema.BoolAttribute{Computed: true},
							"idp_entity_id":             schema.StringAttribute{Computed: true},
							"idp_metadata_file_content": schema.StringAttribute{Computed: true},
							"sp_entity_id":              schema.StringAttribute{Computed: true},
							"dashboards_url":            schema.StringAttribute{Computed: true},
							"roles_key":                 schema.StringAttribute{Optional: true, Computed: true},
							"subject_key":               schema.StringAttribute{Optional: true, Computed: true},
						},
					},
				},
			},
		},
	}
}
