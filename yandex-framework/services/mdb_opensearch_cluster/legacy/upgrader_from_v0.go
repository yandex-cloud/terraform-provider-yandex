package legacy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	common_schema "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/schema"
)

type openSearchModel struct {
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	ID                 types.String   `tfsdk:"id"`
	FolderID           types.String   `tfsdk:"folder_id"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	Name               types.String   `tfsdk:"name"`
	Description        types.String   `tfsdk:"description"`
	Labels             types.Map      `tfsdk:"labels"`
	Environment        types.String   `tfsdk:"environment"`
	Config             types.List     `tfsdk:"config"`
	Hosts              types.Set      `tfsdk:"hosts"`
	NetworkID          types.String   `tfsdk:"network_id"`
	Health             types.String   `tfsdk:"health"`
	Status             types.String   `tfsdk:"status"`
	SecurityGroupIDs   types.Set      `tfsdk:"security_group_ids"`
	ServiceAccountID   types.String   `tfsdk:"service_account_id"`
	DeletionProtection types.Bool     `tfsdk:"deletion_protection"`
	MaintenanceWindow  types.List     `tfsdk:"maintenance_window"`
}

type config struct {
	Version       types.String `tfsdk:"version"`
	AdminPassword types.String `tfsdk:"admin_password"`
	OpenSearch    types.List   `tfsdk:"opensearch"`
	Dashboards    types.List   `tfsdk:"dashboards"`
	Access        types.Object `tfsdk:"access"`
}

type openSearchSubConfig struct {
	NodeGroups types.Set `tfsdk:"node_groups"`
	Plugins    types.Set `tfsdk:"plugins"`
}

type dashboardSubConfig struct {
	NodeGroups types.Set `tfsdk:"node_groups"`
}

type openSearchNode struct {
	Name           types.String `tfsdk:"name"`
	Resources      types.List   `tfsdk:"resources"`
	HostsCount     types.Int64  `tfsdk:"hosts_count"`
	ZoneIDs        types.Set    `tfsdk:"zone_ids"`
	SubnetIDs      types.List   `tfsdk:"subnet_ids"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
	Roles          types.Set    `tfsdk:"roles"`
}

type dashboardNode struct {
	Name           types.String `tfsdk:"name"`
	Resources      types.List   `tfsdk:"resources"`
	HostsCount     types.Int64  `tfsdk:"hosts_count"`
	ZoneIDs        types.Set    `tfsdk:"zone_ids"`
	SubnetIDs      types.List   `tfsdk:"subnet_ids"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
}

// return StateUpgrader implementation from 0 (prior state version) to 2 (Schema.Version)
func NewUpgraderFromV0(ctx context.Context) resource.StateUpgrader {
	return resource.StateUpgrader{
		PriorSchema: &schema.Schema{
			Version: 0,
			Blocks: map[string]schema.Block{
				"timeouts": timeouts.Block(ctx, timeouts.Opts{
					Create: true,
					Update: true,
					Delete: true,
				}),
				"config": schema.ListNestedBlock{
					Description: "Configuration of the OpenSearch cluster.",
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"opensearch": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
									listvalidator.IsRequired(),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"plugins": schema.SetAttribute{
											Computed:    true,
											Optional:    true,
											ElementType: types.StringType,
										},
									},
									Blocks: map[string]schema.Block{
										"node_groups": schema.SetNestedBlock{
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
												setvalidator.IsRequired(),
											},
											NestedObject: schema.NestedBlockObject{
												Blocks: map[string]schema.Block{
													"resources": schema.ListNestedBlock{
														NestedObject: schema.NestedBlockObject{
															Attributes: common_schema.NodeResourceAttributes(),
														},
													},
												},
												Attributes: map[string]schema.Attribute{
													"name":        schema.StringAttribute{Required: true},
													"hosts_count": schema.Int64Attribute{Required: true},
													"zone_ids": schema.SetAttribute{
														Required:    true,
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
							},
							"dashboards": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.AlsoRequires(
										path.MatchRoot("config").AtName("dashboards").AtName("node_groups"),
									),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"node_groups": schema.SetNestedBlock{
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedBlockObject{
												Blocks: map[string]schema.Block{
													"resources": schema.ListNestedBlock{
														NestedObject: schema.NestedBlockObject{
															Attributes: common_schema.NodeResourceAttributes(),
														},
													},
												},
												Attributes: map[string]schema.Attribute{
													"name":        schema.StringAttribute{Required: true},
													"hosts_count": schema.Int64Attribute{Required: true},
													"zone_ids": schema.SetAttribute{
														Required:    true,
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
							},
							"access": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"data_transfer": schema.BoolAttribute{Optional: true},
									"serverless":    schema.BoolAttribute{Optional: true},
								},
							},
						},
						Attributes: map[string]schema.Attribute{
							"version": schema.StringAttribute{Computed: true, Optional: true},
							"admin_password": schema.StringAttribute{
								Required:  true,
								Sensitive: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
				"maintenance_window": schema.ListNestedBlock{
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("ANYTIME", "WEEKLY"),
								},
							},
							"day": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.OneOf("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"),
								},
							},
							"hour": schema.Int64Attribute{
								Optional: true,
								Validators: []validator.Int64{
									int64validator.Between(1, 24),
								},
							},
						},
					},
				},
			},
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"folder_id": schema.StringAttribute{
					Computed:    true,
					Optional:    true,
					Description: "ID of the folder that the OpenSearch cluster belongs to.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				"created_at": schema.StringAttribute{
					Computed:    true,
					Description: "Creation timestamp",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"name":        schema.StringAttribute{Required: true, Description: "Name of the OpenSearch cluster. The name must be unique within the folder."},
				"description": schema.StringAttribute{Optional: true, Description: "Description of the OpenSearch cluster"},
				"labels": schema.MapAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Description: "Custom labels for the OpenSearch cluster as `key:value` pairs.",
				},
				"environment": schema.StringAttribute{
					Computed: true,
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Description: "Deployment environment of the OpenSearch cluster.",
				},
				"hosts": hostsWithoutNodeGroup(),
				"network_id": schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Description: "ID of the network that the cluster belongs to.",
				},
				"health": schema.StringAttribute{Computed: true, Description: "Aggregated cluster health."},
				"status": schema.StringAttribute{Computed: true, Description: "Current state of the cluster."},
				"security_group_ids": schema.SetAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Description: "User security groups",
				},
				"service_account_id":  schema.StringAttribute{Optional: true},
				"deletion_protection": schema.BoolAttribute{Computed: true, Optional: true},
			},
		},
		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			oldModel := openSearchModel{}
			resp.Diagnostics.Append(req.State.Get(ctx, &oldModel)...)
			if resp.Diagnostics.HasError() {
				return
			}

			tflog.Debug(ctx, fmt.Sprintf("UpgraderFromV0.OldModel: %+v\n", oldModel))

			oldConfigs := make([]config, 0, 1)
			resp.Diagnostics.Append(oldModel.Config.ElementsAs(ctx, &oldConfigs, false)...)
			oldConfig := oldConfigs[0]

			newConfig := model.Config{
				Version:       oldConfig.Version,
				AdminPassword: oldConfig.AdminPassword,
				Access:        oldConfig.Access,
			}

			openSearchSubConfigs := make([]openSearchSubConfig, 0, 1)
			resp.Diagnostics.Append(oldConfig.OpenSearch.ElementsAs(ctx, &openSearchSubConfigs, false)...)
			oldOpenSearchSubConfig := openSearchSubConfigs[0]

			oldOpenSearchNodeGroups := make([]openSearchNode, 0, len(oldOpenSearchSubConfig.NodeGroups.Elements()))
			resp.Diagnostics.Append(oldOpenSearchSubConfig.NodeGroups.ElementsAs(ctx, &oldOpenSearchNodeGroups, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			openSearchNodes := make([]model.OpenSearchNode, 0, len(oldOpenSearchNodeGroups))
			for _, oldOpenSearchNodeGroup := range oldOpenSearchNodeGroups {
				resources := make([]model.NodeResource, 0, 1)
				resp.Diagnostics.Append(oldOpenSearchNodeGroup.Resources.ElementsAs(ctx, &resources, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				newResource, diags := types.ObjectValueFrom(ctx, model.NodeResourceAttrTypes, resources[0])
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				openSearchNodes = append(openSearchNodes, model.OpenSearchNode{
					Name:           oldOpenSearchNodeGroup.Name,
					Resources:      newResource,
					HostsCount:     oldOpenSearchNodeGroup.HostsCount,
					ZoneIDs:        oldOpenSearchNodeGroup.ZoneIDs,
					SubnetIDs:      oldOpenSearchNodeGroup.SubnetIDs,
					AssignPublicIP: oldOpenSearchNodeGroup.AssignPublicIP,
					Roles:          oldOpenSearchNodeGroup.Roles,
				})
			}

			newOpenSearchNodeGroups, diags := types.ListValueFrom(ctx, model.OpenSearchNodeType, openSearchNodes)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			newOpenSearchSubConfig := model.OpenSearchSubConfig{
				NodeGroups: newOpenSearchNodeGroups,
				Plugins:    oldOpenSearchSubConfig.Plugins,
			}

			newOpenSearchSubConfigObj, diags := types.ObjectValueFrom(ctx, model.OpenSearchSubConfigAttrTypes, newOpenSearchSubConfig)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			newConfig.OpenSearch = newOpenSearchSubConfigObj

			if len(oldConfig.Dashboards.Elements()) > 0 {
				dashboardsSubConfigs := make([]dashboardSubConfig, 0, 1)
				resp.Diagnostics.Append(oldConfig.Dashboards.ElementsAs(ctx, &dashboardsSubConfigs, false)...)
				oldDashboardsSubConfig := dashboardsSubConfigs[0]

				oldDashboardsNodeGroups := make([]dashboardNode, 0, len(oldDashboardsSubConfig.NodeGroups.Elements()))
				resp.Diagnostics.Append(oldDashboardsSubConfig.NodeGroups.ElementsAs(ctx, &oldDashboardsNodeGroups, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				dashboardsNodes := make([]model.DashboardNode, 0, len(oldDashboardsNodeGroups))
				for _, oldDashboardsNodeGroup := range oldDashboardsNodeGroups {
					resources := make([]model.NodeResource, 0, 1)
					resp.Diagnostics.Append(oldDashboardsNodeGroup.Resources.ElementsAs(ctx, &resources, false)...)
					if resp.Diagnostics.HasError() {
						return
					}

					newResource, diags := types.ObjectValueFrom(ctx, model.NodeResourceAttrTypes, resources[0])
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					dashboardsNodes = append(dashboardsNodes, model.DashboardNode{
						Name:           oldDashboardsNodeGroup.Name,
						Resources:      newResource,
						HostsCount:     oldDashboardsNodeGroup.HostsCount,
						ZoneIDs:        oldDashboardsNodeGroup.ZoneIDs,
						SubnetIDs:      oldDashboardsNodeGroup.SubnetIDs,
						AssignPublicIP: oldDashboardsNodeGroup.AssignPublicIP,
					})
				}

				newDashboardsNodeGroups, diags := types.ListValueFrom(ctx, model.DashboardNodeType, dashboardsNodes)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				newDashboardsSubConfig := model.DashboardsSubConfig{
					NodeGroups: newDashboardsNodeGroups,
				}

				newDashboardsSubConfigObj, diags := types.ObjectValueFrom(ctx, model.DashboardsSubConfigAttrTypes, newDashboardsSubConfig)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				newConfig.Dashboards = newDashboardsSubConfigObj
			} else {
				newConfig.Dashboards = types.ObjectNull(model.DashboardsSubConfigAttrTypes)
			}

			maintenanceWindows := make([]model.MaintenanceWindow, 0, 1)
			resp.Diagnostics.Append(oldModel.MaintenanceWindow.ElementsAs(ctx, &maintenanceWindows, false)...)
			oldMaintenanceWindow := maintenanceWindows[0]

			newMaintenanceWindow, diags := types.ObjectValueFrom(ctx, model.MaintenanceWindowAttrTypes, oldMaintenanceWindow)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			newConfigObj, diags := types.ObjectValueFrom(ctx, model.ConfigAttrTypes, newConfig)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			newHosts, diags := transformHosts(ctx, oldModel.Hosts)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			newAuthSettings := types.ObjectNull(model.AuthSettingsAttrTypes)

			newModel := model.OpenSearch{
				ID:                 oldModel.ID,
				ClusterID:          oldModel.ID,
				FolderID:           oldModel.FolderID,
				CreatedAt:          oldModel.CreatedAt,
				Name:               oldModel.Name,
				Labels:             oldModel.Labels,
				Environment:        oldModel.Environment,
				Config:             newConfigObj,
				Hosts:              newHosts,
				NetworkID:          oldModel.NetworkID,
				Health:             oldModel.Health,
				Status:             oldModel.Status,
				SecurityGroupIDs:   oldModel.SecurityGroupIDs,
				ServiceAccountID:   oldModel.ServiceAccountID,
				DeletionProtection: oldModel.DeletionProtection,
				MaintenanceWindow:  newMaintenanceWindow,
				AuthSettings:       newAuthSettings,
				Timeouts:           oldModel.Timeouts,
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, newModel)...)
		},
	}
}
