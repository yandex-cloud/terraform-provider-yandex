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
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	common_schema "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/schema/descriptions"
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
		MarkdownDescription: descriptions.Datasource,
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"config": schema.SingleNestedBlock{
				MarkdownDescription: descriptions.Config,
				Blocks: map[string]schema.Block{
					"opensearch": schema.SingleNestedBlock{
						MarkdownDescription: descriptions.Opensearch,
						Attributes: map[string]schema.Attribute{
							"plugins": schema.SetAttribute{
								MarkdownDescription: descriptions.Plugins,
								Computed:            true,
								Optional:            true,
								ElementType:         types.StringType,
							},
							"config": common_schema.OpenSearchConfig2(),
						},
						Blocks: map[string]schema.Block{
							"node_groups": schema.ListNestedBlock{
								MarkdownDescription: descriptions.NodeGroups,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"resources": common_schema.NodeResource(),
									},
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: descriptions.NodeGroupName,
											Computed:            true,
										},
										"hosts_count": schema.Int64Attribute{
											MarkdownDescription: descriptions.HostsCount,
											Computed:            true,
										},
										"zone_ids": schema.SetAttribute{
											MarkdownDescription: descriptions.ZoneIDs,
											Optional:            true,
											Computed:            true,
											ElementType:         types.StringType,
										},
										"subnet_ids": schema.ListAttribute{
											MarkdownDescription: descriptions.SubnetIDs,
											Optional:            true,
											Computed:            true,
											ElementType:         types.StringType,
										},

										"assign_public_ip": schema.BoolAttribute{
											MarkdownDescription: descriptions.AssignPublicIP,
											Computed:            true,
											Optional:            true,
										},
										"roles": schema.SetAttribute{
											MarkdownDescription: descriptions.Roles,
											Optional:            true,
											Computed:            true,
											ElementType:         types.StringType,
										},
										"disk_size_autoscaling": schema.SingleNestedAttribute{
											Description: "Node group disk size autoscaling settings.",
											Optional:    true,
											Computed:    true,
											Attributes: map[string]schema.Attribute{
												"disk_size_limit": schema.Int64Attribute{
													Description: "The overall maximum for disk size that limit all autoscaling iterations. See the [documentation](https://yandex.cloud/en/docs/managed-opensearch/concepts/storage#auto-rescale) for details.",
													Computed:    true,
												},
												"planned_usage_threshold": schema.Int64Attribute{
													Description: "Threshold of storage usage (in percent) that triggers automatic scaling of the storage during the maintenance window. Zero value means disabled threshold.",
													Optional:    true,
													Computed:    true,
												},
												"emergency_usage_threshold": schema.Int64Attribute{
													Description: "Threshold of storage usage (in percent) that triggers immediate automatic scaling of the storage. Zero value means disabled threshold.",
													Optional:    true,
													Computed:    true,
												},
											},
										},
									},
								},
							},
						},
					},
					"dashboards": schema.SingleNestedBlock{
						MarkdownDescription: descriptions.Dashboards,
						Blocks: map[string]schema.Block{
							"node_groups": schema.ListNestedBlock{
								MarkdownDescription: descriptions.DashboardNodeGroups,
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"resources": common_schema.NodeResource(),
									},
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: descriptions.NodeGroupName,
											Required:            true,
										},
										"hosts_count": schema.Int64Attribute{
											MarkdownDescription: descriptions.HostsCount,
											Required:            true,
										},
										"zone_ids": schema.SetAttribute{
											MarkdownDescription: descriptions.ZoneIDs,
											Optional:            true,
											Computed:            true,
											ElementType:         types.StringType,
										},
										"subnet_ids": schema.ListAttribute{
											MarkdownDescription: descriptions.SubnetIDs,
											Optional:            true,
											Computed:            true,
											ElementType:         types.StringType,
										},

										"assign_public_ip": schema.BoolAttribute{
											MarkdownDescription: descriptions.AssignPublicIP,
											Computed:            true,
											Optional:            true,
										},
									},
								},
							},
						},
					},
					"access": schema.SingleNestedBlock{
						MarkdownDescription: descriptions.Access,
						Attributes: map[string]schema.Attribute{
							"data_transfer": schema.BoolAttribute{
								MarkdownDescription: descriptions.DataTransfer,
								Computed:            true,
							},
							"serverless": schema.BoolAttribute{
								MarkdownDescription: descriptions.Serverless,
								Computed:            true,
							},
						},
					},
				},
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						MarkdownDescription: descriptions.Version,
						Computed:            true,
						Optional:            true,
					},
					"admin_password": schema.StringAttribute{
						MarkdownDescription: descriptions.AdminPassword,
						Computed:            true,
						Optional:            true,
						Sensitive:           true,
					},
				},
			},
			"maintenance_window": schema.SingleNestedBlock{
				MarkdownDescription: descriptions.MaintenanceWindow,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: descriptions.MaintenanceType,
						Computed:            true,
					},
					"day": schema.StringAttribute{
						MarkdownDescription: descriptions.MaintenanceDay,
						Computed:            true,
					},
					"hour": schema.Int64Attribute{
						MarkdownDescription: descriptions.MaintenanceHour,
						Computed:            true,
					},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: descriptions.ClusterID,
				Computed:            true,
				Optional:            true,
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
				Computed:            true,
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["created_at"],
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: descriptions.Name,
				Computed:            true,
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["description"],
				Computed:            true,
				Optional:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				Computed:            true,
				Optional:            true,
				ElementType:         types.StringType,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: descriptions.Environment,
				Computed:            true,
			},
			"hosts": common_schema.Hosts(),
			"network_id": schema.StringAttribute{
				MarkdownDescription: descriptions.NetworkID,
				Computed:            true,
			},
			"health": schema.StringAttribute{
				MarkdownDescription: descriptions.Health,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: descriptions.Status,
				Computed:            true,
			},
			"security_group_ids": schema.SetAttribute{
				MarkdownDescription: descriptions.SecurityGroupIDs,
				Computed:            true,
				Optional:            true,
				ElementType:         types.StringType,
			},
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: descriptions.ServiceAccountID,
				Computed:            true,
				Optional:            true,
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: common.ResourceDescriptions["deletion_protection"],
				Computed:            true,
				Optional:            true,
			},
			"disk_encryption_key_id": schema.StringAttribute{
				MarkdownDescription: descriptions.DiskEncryptionKeyID,
				Computed:            true,
				Optional:            true,
			},
			"auth_settings": schema.SingleNestedAttribute{
				MarkdownDescription: descriptions.AuthSettings,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"saml": schema.SingleNestedAttribute{
						MarkdownDescription: descriptions.SAML,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: descriptions.SAMLEnabled,
								Computed:            true,
							},
							"idp_entity_id": schema.StringAttribute{
								MarkdownDescription: descriptions.SAMLIdpEntityID,
								Computed:            true,
							},
							"idp_metadata_file_content": schema.StringAttribute{
								MarkdownDescription: descriptions.SAMLIdpMetadataFileContent,
								Computed:            true,
							},
							"sp_entity_id": schema.StringAttribute{
								MarkdownDescription: descriptions.SAMLSpEntityID,
								Computed:            true,
							},
							"dashboards_url": schema.StringAttribute{
								MarkdownDescription: descriptions.SAMLDashboardsURL,
								Computed:            true,
							},
							"roles_key": schema.StringAttribute{
								MarkdownDescription: descriptions.SAMLRolesKey,
								Optional:            true,
								Computed:            true,
							},
							"subject_key": schema.StringAttribute{
								MarkdownDescription: descriptions.SAMLSubjectKey,
								Optional:            true,
								Computed:            true,
							},
						},
					},
				},
			},
		},
	}
}
