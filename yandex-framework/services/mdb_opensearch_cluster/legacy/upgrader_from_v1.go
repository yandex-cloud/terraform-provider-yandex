package legacy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	common_schema "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/validate"
)

type openSearchV1 struct {
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	ID                 types.String   `tfsdk:"id"`
	ClusterID          types.String   `tfsdk:"cluster_id"`
	FolderID           types.String   `tfsdk:"folder_id"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	Name               types.String   `tfsdk:"name"`
	Description        types.String   `tfsdk:"description"`
	Labels             types.Map      `tfsdk:"labels"`
	Environment        types.String   `tfsdk:"environment"`
	Config             types.Object   `tfsdk:"config"`
	Hosts              types.Set      `tfsdk:"hosts"`
	NetworkID          types.String   `tfsdk:"network_id"`
	Health             types.String   `tfsdk:"health"`
	Status             types.String   `tfsdk:"status"`
	SecurityGroupIDs   types.Set      `tfsdk:"security_group_ids"`
	ServiceAccountID   types.String   `tfsdk:"service_account_id"`
	DeletionProtection types.Bool     `tfsdk:"deletion_protection"`
	MaintenanceWindow  types.Object   `tfsdk:"maintenance_window"`
}

type hostWithoutNodeGroup struct {
	FQDN           types.String `tfsdk:"fqdn"`
	Type           types.String `tfsdk:"type"`
	Roles          types.Set    `tfsdk:"roles"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
	Zone           types.String `tfsdk:"zone"`
	SubnetID       types.String `tfsdk:"subnet_id"`
}

func hostsWithoutNodeGroup() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Computed:    true,
		Description: "Current nodes in the cluster",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"fqdn": schema.StringAttribute{Computed: true},
				"zone": schema.StringAttribute{Computed: true},
				"type": schema.StringAttribute{Computed: true},
				"roles": schema.SetAttribute{
					Computed:    true,
					ElementType: types.StringType,
				},
				"assign_public_ip": schema.BoolAttribute{Computed: true, Optional: true},
				"subnet_id":        schema.StringAttribute{Computed: true, Optional: true},
			},
		},
	}
}

func transformHosts(ctx context.Context, hostsAttr basetypes.SetValue) (basetypes.ListValue, diag.Diagnostics) {
	current := make([]hostWithoutNodeGroup, 0, len(hostsAttr.Elements()))
	diags := hostsAttr.ElementsAs(ctx, &current, false)
	if diags.HasError() {
		return types.ListUnknown(model.HostType), diags
	}

	target := make([]model.Host, len(current))
	for i := range current {
		target[i] = model.Host{
			FQDN:           current[i].FQDN,
			Type:           current[i].Type,
			Roles:          current[i].Roles,
			AssignPublicIP: current[i].AssignPublicIP,
			Zone:           current[i].Zone,
			SubnetID:       current[i].SubnetID,
			NodeGroup:      types.StringNull(),
		}
	}

	return types.ListValueFrom(ctx, model.HostType, target)
}

// return StateUpgrader implementation from 1 (prior state version) to 2 (Schema.Version)
func NewUpgraderFromV1(ctx context.Context) resource.StateUpgrader {
	return resource.StateUpgrader{
		PriorSchema: &schema.Schema{
			Version: 1,
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
							Validators: []validator.Object{
								objectvalidator.IsRequired(),
							},
							Attributes: map[string]schema.Attribute{
								"plugins": schema.SetAttribute{
									Computed:    true,
									Optional:    true,
									ElementType: types.StringType,
								},
							},
							Blocks: map[string]schema.Block{
								//NOTE: changed "set" to "list+customValidator" because https://github.com/hashicorp/terraform-plugin-sdk/issues/1210
								"node_groups": schema.ListNestedBlock{
									Validators: []validator.List{
										listvalidator.IsRequired(),
										listvalidator.SizeAtLeast(1),
										validate.UniqueByField("name", func(x model.OpenSearchNode) string { return x.Name.ValueString() }),
									},
									NestedObject: schema.NestedBlockObject{
										Blocks: map[string]schema.Block{
											"resources": common_schema.NodeResource(),
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
												Validators: []validator.Set{
													validate.UniqueCaseInsensitive(),
												},
											},
										},
									},
								},
							},
						},
						"dashboards": schema.SingleNestedBlock{
							Validators: []validator.Object{
								objectvalidator.AlsoRequires(
									path.MatchRoot("config").AtName("dashboards").AtName("node_groups"),
								),
							},
							Blocks: map[string]schema.Block{
								//NOTE: changed "set" to "list+customValidator" because https://github.com/hashicorp/terraform-plugin-sdk/issues/1210
								"node_groups": schema.ListNestedBlock{
									Validators: []validator.List{
										listvalidator.SizeAtLeast(1),
										validate.UniqueByField("name", func(x model.DashboardNode) string { return x.Name.ValueString() }),
									},
									NestedObject: schema.NestedBlockObject{
										Blocks: map[string]schema.Block{
											"resources": common_schema.NodeResource(),
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
				"maintenance_window": schema.SingleNestedBlock{
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
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"cluster_id": schema.StringAttribute{
					Computed: true,
					Optional: true,
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
			oldModel := openSearchV1{}
			resp.Diagnostics.Append(req.State.Get(ctx, &oldModel)...)
			if resp.Diagnostics.HasError() {
				return
			}

			tflog.Debug(ctx, fmt.Sprintf("UpgraderFromV1.OldModel: %+v\n", oldModel))

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
				Config:             oldModel.Config,
				Hosts:              newHosts,
				NetworkID:          oldModel.NetworkID,
				Health:             oldModel.Health,
				Status:             oldModel.Status,
				SecurityGroupIDs:   oldModel.SecurityGroupIDs,
				ServiceAccountID:   oldModel.ServiceAccountID,
				DeletionProtection: oldModel.DeletionProtection,
				MaintenanceWindow:  oldModel.MaintenanceWindow,
				AuthSettings:       newAuthSettings,
				Timeouts:           oldModel.Timeouts,
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, newModel)...)
		},
	}
}
