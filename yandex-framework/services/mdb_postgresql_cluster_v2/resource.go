package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"golang.org/x/exp/maps"
)

type clusterResource struct {
	providerConfig *provider_config.Config
}

func NewPostgreSQLClusterResourceV2() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// TODO we are open for better ideas
	resp.TypeName = req.ProviderTypeName + "_mdb_postgresql_cluster_v2"
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

func (r *clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PostgreSQL cluster within the Yandex Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/). [How to connect to the DB](https://yandex.cloud/docs/managed-postgresql/quickstart#connect). To connect, use port 6432. The port number is not configurable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: common.ResourceDescriptions["id"],
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the PostgreSQL cluster. Provided by the client when the cluster is created.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the PostgreSQL cluster.",
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Optional:    true,
			},
			"folder_id": schema.StringAttribute{
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.StringAttribute{
				Description: "ID of the network that the cluster belongs to.",
				Required:    true,
			},
			"environment": schema.StringAttribute{
				Description: "Deployment environment of the PostgreSQL cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.MapAttribute{
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				ElementType: types.StringType,
			},
			"hosts": schema.MapNestedAttribute{
				Description: "A host configuration of the PostgreSQL cluster.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Description: "The availability zone where the host is located.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"subnet_id": schema.StringAttribute{
							Description: "ID of the subnet where the host is located.",
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"assign_public_ip": schema.BoolAttribute{
							Description: "Whether the host should get a public IP address.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"fqdn": schema.StringAttribute{
							Description: "The fully qualified domain name of the host.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"replication_source": schema.StringAttribute{
							Description: "FQDN of the host that is used as a replication source.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
					},
				},
			},
			"deletion_protection": schema.BoolAttribute{
				Description: "Inhibits deletion of the cluster. Can be either true or false.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"disk_encryption_key_id": schema.StringAttribute{
				Description: "ID of the KMS key for cluster disk encryption.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_ids": defaultschema.SecurityGroupIds(),
			"restore": schema.SingleNestedAttribute{
				Description: "The cluster will be created from the specified backup.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"backup_id": schema.StringAttribute{
						Description: "Backup ID. The cluster will be created from the specified backup. [How to get a list of PostgreSQL backups](https://yandex.cloud/docs/managed-postgresql/operations/cluster-backups).",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"time_inclusive": schema.BoolAttribute{
						Description: "Flag that indicates whether a database should be restored to the first backup point available just after the timestamp specified in the [time] field instead of just before. Possible values:\n* `false` (default) — the restore point refers to the first backup moment before [time].\n* `true` — the restore point refers to the first backup point after [time].",
						Optional:    true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"time": schema.StringAttribute{
						Description: "Timestamp of the moment to which the PostgreSQL cluster should be restored. (Format: `2006-01-02T15:04:05` - UTC). When not set, current time is used.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Validators: []validator.String{
							mdbcommon.NewStringToTimeValidator(),
						},
					},
				},
			},
			// Optional nested attribute maintenance_window required all optional nested attributes
			// But if the block is specified explicitly, then the type attribute is required
			"maintenance_window": schema.SingleNestedAttribute{
				Description: "Maintenance policy of the PostgreSQL cluster.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Object{
					NewMaintenanceWindowStructValidator(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Type of maintenance window. Can be either ANYTIME or WEEKLY. A day and hour of window need to be specified with weekly window.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANYTIME", "WEEKLY"),
						},
					},
					"day": schema.StringAttribute{
						Description: "Day of the week (in DDD format). Allowed values: \"MON\", \"TUE\", \"WED\", \"THU\", \"FRI\", \"SAT\",\"SUN\"",
						Optional:    true,
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
						Description: "Hour of the day in UTC (in HH format). Allowed value is between 1 and 24.",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.Between(1, 24),
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.SingleNestedBlock{
				Description: "Configuration of the PostgreSQL cluster.",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Description: "Version of the PostgreSQL cluster.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"13", "13-1c",
								"14", "14-1c",
								"15", "15-1c",
								"16", "17",
							),
						},
					},
					"autofailover": schema.BoolAttribute{
						Description: "Configuration setting which enables/disables automatic failover in the cluster.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"access": schema.SingleNestedAttribute{
						Description: "Access policy to the PostgreSQL cluster.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"data_lens": schema.BoolAttribute{
								Description: "Allow access for Yandex DataLens.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"web_sql": schema.BoolAttribute{
								Description: "Allow access for SQL queries in the management console",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"serverless": schema.BoolAttribute{
								Description: "Allow access for connection to managed databases from functions",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"data_transfer": schema.BoolAttribute{
								Description: "Allow access for DataTransfer",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
						},
					},
					"performance_diagnostics": schema.SingleNestedAttribute{
						Description: "Cluster performance diagnostics settings. The structure is documented below.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Enable performance diagnostics",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"sessions_sampling_interval": schema.Int64Attribute{
								Description: "Interval (in seconds) for pg_stat_activity sampling. Acceptable values are 1 to 86400, inclusive.",
								Required:    true,
								Validators: []validator.Int64{
									int64validator.Between(1, 86400),
								},
							},
							"statements_sampling_interval": schema.Int64Attribute{
								Description: "Interval (in seconds) for pg_stat_statements sampling. Acceptable values are 60 to 86400, inclusive.",
								Required:    true,
								Validators: []validator.Int64{
									int64validator.Between(60, 86400),
								},
							},
						},
					},
					"backup_retain_period_days": schema.Int64Attribute{
						Description: "The period in days during which backups are stored.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(7),
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"backup_window_start": schema.SingleNestedAttribute{
						Description: "Time to start the daily backup, in the UTC timezone.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"hours": schema.Int64Attribute{
								Description: "The hour at which backup will be started (UTC).",
								Computed:    true,
								Optional:    true,
								Default:     int64default.StaticInt64(0),
								Validators: []validator.Int64{
									int64validator.Between(0, 23),
								},
							},
							"minutes": schema.Int64Attribute{
								Description: "The minute at which backup will be started.",
								Computed:    true,
								Optional:    true,
								Default:     int64default.StaticInt64(0),
								Validators: []validator.Int64{
									int64validator.Between(0, 59),
								},
							},
						},
					},
					"pooler_config": schema.SingleNestedAttribute{
						Description: "Configuration of the connection pooler.",
						Optional:    true,
						Computed:    true,
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(PoolerConfigAttrTypes, map[string]attr.Value{
								"pooling_mode": types.StringValue(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
								"pool_discard": types.BoolNull(),
							}),
						),
						Validators: []validator.Object{
							NewAtLeastIfConfiguredValidator(
								path.MatchRelative().AtName("pooling_mode"),
								path.MatchRelative().AtName("pool_discard"),
							),
						},
						Attributes: map[string]schema.Attribute{
							"pooling_mode": schema.StringAttribute{
								Description: "Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.)",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf(
										maps.Keys(postgresql.ConnectionPoolerConfig_PoolingMode_value)...,
									),
								},
								Default: stringdefault.StaticString(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
							},
							"pool_discard": schema.BoolAttribute{
								Description: "Setting pool_discard parameter in Odyssey.",
								Optional:    true,
							},
						},
					},
					"disk_size_autoscaling": schema.SingleNestedAttribute{
						Description: "Cluster disk size autoscaling settings.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"disk_size_limit": schema.Int64Attribute{
								Description: "The overall maximum for disk size that limit all autoscaling iterations. See the [documentation](https://yandex.cloud/en/docs/managed-postgresql/concepts/storage#auto-rescale) for details.",
								Required:    true,
								Validators: []validator.Int64{
									Int64GreaterValidator(path.MatchRoot("config").AtName("resources").AtName("disk_size")),
								},
							},
							"planned_usage_threshold": schema.Int64Attribute{
								Description: "Threshold of storage usage (in percent) that triggers automatic scaling of the storage during the maintenance window. Zero value means disabled threshold.",
								Optional:    true,
								Computed:    true,
								Validators: []validator.Int64{
									int64validator.Any(
										int64validator.OneOf(0),
										int64validator.AlsoRequires(
											path.MatchRoot("maintenance_window"),
											path.MatchRoot("maintenance_window").AtName("type"),
											path.MatchRoot("maintenance_window").AtName("hour"),
											path.MatchRoot("maintenance_window").AtName("day"),
										),
									),
								},
								Default: int64default.StaticInt64(0),
							},
							"emergency_usage_threshold": schema.Int64Attribute{
								Description: "Threshold of storage usage (in percent) that triggers immediate automatic scaling of the storage. Zero value means disabled threshold.",
								Validators: []validator.Int64{
									int64validator.Any(
										Int64GreaterValidator(path.MatchRoot("config").AtName("disk_size_autoscaling").AtName("planned_usage_threshold")),
										int64validator.OneOf(0),
									),
								},
								Optional: true,
								Computed: true,
								Default:  int64default.StaticInt64(0),
							},
						},
					},
					"postgresql_config": schema.MapAttribute{
						CustomType: mdbcommon.NewSettingsMapType(pgAttrProvider),
						PlanModifiers: []planmodifier.Map{
							mapplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "PostgreSQL cluster configuration. For detailed information specific to your PostgreSQL version, please refer to the [API proto specifications](https://github.com/yandex-cloud/cloudapi/tree/master/yandex/cloud/mdb/postgresql/v1/config).",
						Optional:            true,
						Computed:            true,
					},
				},
				Blocks: map[string]schema.Block{
					"resources": schema.SingleNestedBlock{
						Description: "Resources allocated to hosts of the PostgreSQL cluster.",
						Attributes: map[string]schema.Attribute{
							"resource_preset_id": schema.StringAttribute{
								Description: "ID of the resource preset that determines the number of CPU cores and memory size for the host.",
								Required:    true,
							},
							"disk_type_id": schema.StringAttribute{
								Description: "ID of the disk type that determines the disk performance characteristics.",
								Required:    true,
							},
							"disk_size": schema.Int64Attribute{
								Description: "Size of the disk in bytes.",
								Required:    true,
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

func (r *clusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	var plan Cluster
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Modifying plan for cluster", map[string]interface{}{"id": plan.Id.ValueString()})

	var cfgState Config
	var cfgPlan Config
	resp.Diagnostics.Append(state.Config.As(ctx, &cfgState, datasize.DefaultOpts)...)
	resp.Diagnostics.Append(plan.Config.As(ctx, &cfgPlan, datasize.DefaultOpts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	autoscalingOn := utils.IsPresent(attr.Value(cfgState.DiskSizeAutoscaling))

	// remove changes on disk_size from plan if enabled autoscaling
	cfgPlan.Resources = mdbcommon.FixDiskSizeOnAutoscalingChanges(ctx, cfgPlan.Resources, cfgState.Resources, autoscalingOn, &resp.Diagnostics)
	cfgPlanObj, d := types.ObjectValueFrom(ctx, ConfigAttrTypes, cfgPlan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Config = cfgPlanObj
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Cluster
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating PostgreSQL Cluster")

	hostSpecsSlice, diags := mdbcommon.CreateClusterHosts(ctx, postgresqlHostService, plan.HostSpecs)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	if utils.IsPresent(plan.Restore) {
		r.restoreCluster(ctx, diags, plan, hostSpecsSlice, resp)
	} else {
		r.createCluster(ctx, plan, hostSpecsSlice, resp)
	}
}

func (r *clusterResource) createCluster(
	ctx context.Context,
	plan Cluster,
	hostSpecsSlice []*postgresql.HostSpec,
	resp *resource.CreateResponse,
) {
	request, diags := prepareCreateRequest(ctx, &plan, &r.providerConfig.ProviderState, hostSpecsSlice)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	cid := postgresqlApi.CreateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, request)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = types.StringValue(cid)

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterResource) restoreCluster(
	ctx context.Context,
	diags diag.Diagnostics,
	plan Cluster,
	hostSpecsSlice []*postgresql.HostSpec,
	resp *resource.CreateResponse,
) {
	request, diags := prepareRestoreRequest(ctx, &plan, &r.providerConfig.ProviderState, hostSpecsSlice)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	cid := postgresqlApi.RestoreCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, request)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = types.StringValue(cid)

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Load the current plan
	// We shouldnt read the state because we shouldn't use the state in the host update method
	// The plan and the Api response should be enough
	var plan Cluster
	var state Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating PostgreSQL Cluster", map[string]interface{}{"id": plan.Id.ValueString()})
	tflog.Debug(ctx, fmt.Sprintf("Update PostgreSQL Cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update PostgreSQL Cluster plan: %+v", plan))

	updateVersionRequest, d := prepareVersionUpdateRequest(&state, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgresqlApi.UpdateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateVersionRequest)

	updateRequest, d := prepareUpdateRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgresqlApi.UpdateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateRequest)
	if resp.Diagnostics.HasError() {
		return
	}

	mdbcommon.UpdateClusterHosts(
		ctx,
		r.providerConfig.SDK,
		&resp.Diagnostics,
		postgresqlHostService,
		&postgresqlApi,
		plan.Id.ValueString(),
		plan.HostSpecs,
		state.HostSpecs,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.Id.ValueString()
	postgresqlApi.DeleteCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid)
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *clusterResource) refreshResourceState(ctx context.Context, state *Cluster, respDiagnostics *diag.Diagnostics) {
	cid := state.Id.ValueString()
	cluster := postgresqlApi.GetCluster(ctx, r.providerConfig.SDK, respDiagnostics, cid)
	if respDiagnostics.HasError() {
		return
	}

	entityIdToApiHosts := mdbcommon.ReadHosts(ctx, r.providerConfig.SDK, respDiagnostics, postgresqlHostService, &postgresqlApi, state.HostSpecs, cid)

	var diags diag.Diagnostics
	state.HostSpecs, diags = types.MapValueFrom(ctx, hostType, entityIdToApiHosts)
	respDiagnostics.Append(diags...)
	if respDiagnostics.HasError() {
		return
	}

	var cfgState Config
	diags.Append(state.Config.As(ctx, &cfgState, datasize.DefaultOpts)...)

	state.Id = types.StringValue(cluster.Id)
	state.FolderId = types.StringValue(cluster.FolderId)
	state.NetworkId = types.StringValue(cluster.NetworkId)
	state.Name = types.StringValue(cluster.Name)
	state.Description = types.StringValue(cluster.Description)
	state.Environment = types.StringValue(cluster.Environment.String())
	state.Labels = flattenMapString(ctx, cluster.Labels, &diags)

	state.Config = flattenConfig(ctx, cfgState.PostgtgreSQLConfig, cluster.GetConfig(), respDiagnostics)

	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())
	state.MaintenanceWindow = mdbcommon.FlattenMaintenanceWindow[
		postgresql.MaintenanceWindow,
		postgresql.WeeklyMaintenanceWindow,
		postgresql.AnytimeMaintenanceWindow,
		postgresql.WeeklyMaintenanceWindow_WeekDay,
	](ctx, cluster.MaintenanceWindow, respDiagnostics)
	state.SecurityGroupIds = mdbcommon.FlattenSetString(ctx, cluster.SecurityGroupIds, respDiagnostics)
	state.DiskEncryptionKeyId = flattenStringWrapper(ctx, cluster.DiskEncryptionKeyId, respDiagnostics)
}
