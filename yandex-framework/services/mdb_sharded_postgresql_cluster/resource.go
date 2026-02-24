package mdb_sharded_postgresql_cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBShardedPostgreSQLClusterCreateTimeout = 30 * time.Minute
	yandexMDBShardedPostgreSQLClusterDeleteTimeout = 15 * time.Minute
	yandexMDBShardedPostgreSQLClusterUpdateTimeout = 60 * time.Minute
)

type clusterResource struct {
	providerConfig *provider_config.Config
}

func NewShardedPostgreSQLClusterResource() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_sharded_postgresql_cluster"
}

func (r *clusterResource) Configure(_ context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.providerConfig = providerConfig
}

func (r *clusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sharded Postgresql cluster within the Yandex Cloud.",
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
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Sharded PostgreSQL cluster. Provided by the client when the cluster is created.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Sharded PostgreSQL cluster.",
				Optional:            true,
			},
			"folder_id":  defaultschema.FolderId(),
			"network_id": defaultschema.NetworkId(),
			"environment": schema.StringAttribute{
				MarkdownDescription: "Deployment environment of the Sharded PostgreSQL cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": defaultschema.Labels(),
			"hosts": schema.MapNestedAttribute{
				MarkdownDescription: "A host configuration of the Sharded PostgreSQL cluster.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							MarkdownDescription: "The availability zone where the host is located.",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"subnet_id": schema.StringAttribute{
							MarkdownDescription: "ID of the subnet where the host is located.",
							Optional:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"assign_public_ip": schema.BoolAttribute{
							MarkdownDescription: "Assign a public IP address to the host.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"fqdn": schema.StringAttribute{
							MarkdownDescription: "The fully qualified domain name of the host.",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
					},
				},
			},
			"deletion_protection": defaultschema.DeletionProtection(),
			"security_group_ids":  defaultschema.SecurityGroupIds(),
			"maintenance_window": schema.SingleNestedAttribute{
				MarkdownDescription: "Maintenance policy of the Sharded PostgreSQL cluster.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of maintenance window. Can be either ANYTIME or WEEKLY. A day and hour of window need to be specified with weekly window.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANYTIME", "WEEKLY"),
						},
					},
					"day": schema.StringAttribute{
						MarkdownDescription: "Day of the week (in DDD format). Allowed values: \"MON\", \"TUE\", \"WED\", \"THU\", \"FRI\", \"SAT\",\"SUN\"",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"MON", "TUE",
								"WED", "THU",
								"FRI", "SAT",
								"SUN",
							),
						},
					},
					"hour": schema.Int64Attribute{
						MarkdownDescription: "Hour of the day in UTC (in HH format). Allowed value is between 1 and 24.",
						Optional:            true,
						Validators: []validator.Int64{
							int64validator.Between(1, 24),
						},
					},
				},
			},
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of the Sharded PostgreSQL cluster.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"access": schema.SingleNestedAttribute{
						MarkdownDescription: "Access policy to the Sharded PostgreSQL cluster.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"data_lens": schema.BoolAttribute{
								MarkdownDescription: "Allow access for Yandex DataLens.",
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
							"web_sql": schema.BoolAttribute{
								MarkdownDescription: "Allow access for SQL queries in the management console",
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
							"serverless": schema.BoolAttribute{
								MarkdownDescription: "Allow access for connection to managed databases from functions",
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
							"data_transfer": schema.BoolAttribute{
								MarkdownDescription: "Allow access for DataTransfer",
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
						},
					},
					"backup_retain_period_days": schema.Int64Attribute{
						MarkdownDescription: "The period in days during which backups are stored.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(7),
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"backup_window_start": schema.SingleNestedAttribute{
						MarkdownDescription: "Time to start the daily backup, in the UTC timezone.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"hours": schema.Int64Attribute{
								MarkdownDescription: "The hour at which backup will be started (UTC).",
								Computed:            true,
								Optional:            true,
								Default:             int64default.StaticInt64(0),
								Validators: []validator.Int64{
									int64validator.Between(0, 23),
								},
							},
							"minutes": schema.Int64Attribute{
								MarkdownDescription: "The minute at which backup will be started (UTC).",
								Computed:            true,
								Optional:            true,
								Default:             int64default.StaticInt64(0),
								Validators: []validator.Int64{
									int64validator.Between(0, 59),
								},
							},
						},
					},
					"sharded_postgresql_config": schema.SingleNestedAttribute{
						MarkdownDescription: "Sharded PostgreSQL cluster configuration.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"common": schema.MapAttribute{
								CustomType:          mdbcommon.NewSettingsMapType(attrProvider),
								MarkdownDescription: "General settings for all types of hosts.",
								PlanModifiers: []planmodifier.Map{
									mapplanmodifier.UseStateForUnknown(),
								},
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"router": schema.SingleNestedAttribute{
								MarkdownDescription: "Router specific configuration.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"config": schema.MapAttribute{
										CustomType:          mdbcommon.NewSettingsMapType(attrProvider),
										MarkdownDescription: "Router settings.",
										PlanModifiers: []planmodifier.Map{
											mapplanmodifier.UseStateForUnknown(),
										},
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"resources": schema.SingleNestedAttribute{
										MarkdownDescription: "Resources allocated to routers of the Sharded PostgreSQL cluster.",
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"resource_preset_id": schema.StringAttribute{
												MarkdownDescription: "ID of the resource preset that determines the number of CPU cores and memory size for the host.",
												Required:            true,
											},
											"disk_type_id": schema.StringAttribute{
												MarkdownDescription: "ID of the disk type that determines the disk performance characteristics.",
												Required:            true,
											},
											"disk_size": schema.Int64Attribute{
												MarkdownDescription: "Size of the disk in bytes.",
												Required:            true,
											},
										},
									},
								},
							},
							"coordinator": schema.SingleNestedAttribute{
								MarkdownDescription: "Coordinator specific configuration.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"config": schema.MapAttribute{
										CustomType:          mdbcommon.NewSettingsMapType(attrProvider),
										MarkdownDescription: "Coordinator settings.",
										PlanModifiers: []planmodifier.Map{
											mapplanmodifier.UseStateForUnknown(),
										},
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"resources": schema.SingleNestedAttribute{
										MarkdownDescription: "Resources allocated to routers of the Sharded PostgreSQL cluster.",
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"resource_preset_id": schema.StringAttribute{
												MarkdownDescription: "ID of the resource preset that determines the number of CPU cores and memory size for the host.",
												Required:            true,
											},
											"disk_type_id": schema.StringAttribute{
												MarkdownDescription: "ID of the disk type that determines the disk performance characteristics.",
												Required:            true,
											},
											"disk_size": schema.Int64Attribute{
												MarkdownDescription: "Size of the disk in bytes.",
												Required:            true,
											},
										},
									},
								},
							},
							"infra": schema.SingleNestedAttribute{
								MarkdownDescription: "",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"router": schema.MapAttribute{
										CustomType:          mdbcommon.NewSettingsMapType(attrProvider),
										MarkdownDescription: "Router settings.",
										PlanModifiers: []planmodifier.Map{
											mapplanmodifier.UseStateForUnknown(),
										},
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"coordinator": schema.MapAttribute{
										CustomType:          mdbcommon.NewSettingsMapType(attrProvider),
										MarkdownDescription: "Coordinator settings.",
										PlanModifiers: []planmodifier.Map{
											mapplanmodifier.UseStateForUnknown(),
										},
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"resources": schema.SingleNestedAttribute{
										MarkdownDescription: "Resources allocated to routers of the Sharded PostgreSQL cluster.",
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"resource_preset_id": schema.StringAttribute{
												MarkdownDescription: "ID of the resource preset that determines the number of CPU cores and memory size for the host.",
												Required:            true,
											},
											"disk_type_id": schema.StringAttribute{
												MarkdownDescription: "ID of the disk type that determines the disk performance characteristics.",
												Required:            true,
											},
											"disk_size": schema.Int64Attribute{
												MarkdownDescription: "Size of the disk in bytes.",
												Required:            true,
											},
										},
									},
								},
							},
							"balancer": schema.MapAttribute{
								CustomType:          mdbcommon.NewSettingsMapType(attrProvider),
								MarkdownDescription: "Balancer specific configuration.",
								PlanModifiers: []planmodifier.Map{
									mapplanmodifier.UseStateForUnknown(),
								},
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Load the current state of the resource
	var state Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Resource State
	r.refreshResourceState(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	d := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(d...)
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Cluster
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBShardedPostgreSQLClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	tflog.Debug(ctx, "Creating Sharded Postgresql Cluster")

	hostSpecsSlice, diags := mdbcommon.CreateClusterHosts(ctx, spqrHostService, plan.HostSpecs)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	// Prepare Create Request
	request, diags := prepareCreateRequest(ctx, &plan, &r.providerConfig.ProviderState)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	// Add Hosts to the request
	request.HostSpecs = hostSpecsSlice

	cid := shardedPostgreSQLAPI.CreateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, request)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(cid)

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBShardedPostgreSQLClusterDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid := state.Id.ValueString()
	shardedPostgreSQLAPI.DeleteCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Cluster
	var state Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBShardedPostgreSQLClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, "Updating Sharded Postgresql Cluster", map[string]interface{}{"id": plan.Id.ValueString()})
	tflog.Debug(ctx, fmt.Sprintf("Update Sharded Postgresql Cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update Sharded Postgresql Cluster plan: %+v", plan))

	updateRequest, d := prepareUpdateRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	shardedPostgreSQLAPI.UpdateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateRequest)
	if resp.Diagnostics.HasError() {
		return
	}

	config := Config{}
	diags = state.Config.As(ctx, &config, datasize.DefaultOpts)
	resp.Diagnostics.Append(diags...)
	updateHosts(
		ctx,
		r.providerConfig.SDK,
		&resp.Diagnostics,
		spqrHostService,
		&shardedPostgreSQLAPI,
		plan.Id.ValueString(),
		plan.HostSpecs,
		state.HostSpecs,
		&config,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *clusterResource) refreshResourceState(ctx context.Context, state *Cluster, respDiagnostics *diag.Diagnostics) {
	cid := state.Id.ValueString()
	cluster := shardedPostgreSQLAPI.GetCluster(ctx, r.providerConfig.SDK, respDiagnostics, cid)
	if respDiagnostics.HasError() {
		return
	}

	entityIdToApiHosts := mdbcommon.ReadHosts(ctx, r.providerConfig.SDK, respDiagnostics, spqrHostService, &shardedPostgreSQLAPI, state.HostSpecs, cid)

	var diags diag.Diagnostics
	state.HostSpecs, diags = types.MapValueFrom(ctx, hostType, entityIdToApiHosts)
	respDiagnostics.Append(diags...)
	if respDiagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(cluster.Id)
	state.FolderId = types.StringValue(cluster.FolderId)
	state.NetworkId = types.StringValue(cluster.NetworkId)
	state.Name = types.StringValue(cluster.Name)
	state.Description = types.StringValue(cluster.Description)
	state.Environment = types.StringValue(cluster.Environment.String())
	state.Labels = flattenMapString(ctx, cluster.Labels, respDiagnostics)
	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())
	state.MaintenanceWindow = flattenMaintenanceWindow(ctx, cluster.MaintenanceWindow, respDiagnostics)
	state.SecurityGroupIds = mdbcommon.FlattenSetString(ctx, cluster.SecurityGroupIds, respDiagnostics)

	var cfgState Config
	diags.Append(state.Config.As(ctx, &cfgState, datasize.DefaultOpts)...)
	state.Config = flattenConfig(ctx, cfgState, cluster.GetConfig(), respDiagnostics)
}

func prepareUpdateRequest(ctx context.Context, state, plan *Cluster) (*spqr.UpdateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	request := &spqr.UpdateClusterRequest{
		ClusterId:  state.Id.ValueString(),
		UpdateMask: &field_mask.FieldMask{},
	}

	if !plan.Name.Equal(state.Name) {
		request.SetName(plan.Name.ValueString())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "name")
	}

	if !plan.Description.Equal(state.Description) {
		request.SetDescription(plan.Description.ValueString())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "description")
	}

	if !plan.Labels.Equal(state.Labels) {
		request.SetLabels(expandLabels(ctx, plan.Labels, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "labels")
	}

	var planConfig Config
	diags = plan.Config.As(ctx, &planConfig, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}
	var stateConfig Config
	diags = state.Config.As(ctx, &stateConfig, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	config, updateMaskPaths, diags := prepareConfigChange(ctx, &planConfig, &stateConfig)
	if diags.HasError() {
		return nil, diags
	}

	request.SetConfigSpec(config)
	request.UpdateMask.Paths = append(request.UpdateMask.Paths, updateMaskPaths...)

	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		request.SetDeletionProtection(plan.DeletionProtection.ValueBool())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "deletion_protection")
	}

	if !plan.SecurityGroupIds.Equal(state.SecurityGroupIds) {
		request.SetSecurityGroupIds(mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "security_group_ids")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		request.SetMaintenanceWindow(expandClusterMaintenanceWindow(ctx, plan.MaintenanceWindow, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "maintenance_window")
	}

	return request, diags
}
