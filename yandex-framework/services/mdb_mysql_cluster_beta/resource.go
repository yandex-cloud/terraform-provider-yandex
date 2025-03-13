package mdb_mysql_cluster_beta

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type clusterResource struct {
	providerConfig *provider_config.Config
}

func NewMySQLClusterResourceBeta() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// TODO we are open for better ideas
	resp.TypeName = req.ProviderTypeName + "_mdb_mysql_cluster_beta"
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
		MarkdownDescription: "Manages a MySQL cluster within the Yandex Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mysql/). [How to connect to the DB](https://yandex.cloud/docs/managed-mysql/quickstart#connect). To connect, use port 6432. The port number is not configurable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the MySQL cluster. Provided by the client when the cluster is created.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the MySQL cluster.",
				Optional:            true,
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "ID of the network that the cluster belongs to.",
				Required:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Deployment environment of the MySQL cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				Optional:            true,
				ElementType:         types.StringType,
			},
			"hosts": schema.MapNestedAttribute{
				MarkdownDescription: "A host configuration of the MySQL cluster.",
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
						"replication_source": schema.StringAttribute{
							MarkdownDescription: "FQDN of the host that is used as a replication source.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "Inhibits deletion of the cluster. Can be either true or false.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the MySQL cluster.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"5.7",
						"8.0",
					),
				},
			},
			"access": schema.SingleNestedAttribute{
				MarkdownDescription: "Access policy to the MySQL cluster.",
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
					"data_transfer": schema.BoolAttribute{
						MarkdownDescription: "Allow access for DataTransfer",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"performance_diagnostics": schema.SingleNestedAttribute{
				MarkdownDescription: "Cluster performance diagnostics settings. The structure is documented below.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable performance diagnostics",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"sessions_sampling_interval": schema.Int64Attribute{
						MarkdownDescription: "Interval (in seconds) for pg_stat_activity sampling Acceptable values are 1 to 86400, inclusive.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.Between(1, 86400),
						},
					},
					"statements_sampling_interval": schema.Int64Attribute{
						MarkdownDescription: "Interval (in seconds) for pg_stat_statements sampling Acceptable values are 60 to 86400, inclusive.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.Between(60, 86400),
						},
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
			"mysql_config": schema.MapAttribute{
				CustomType:          mdbcommon.NewSettingsMapType(msAttrProvider),
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "MySQL cluster config.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_ids": schema.SetAttribute{
				MarkdownDescription: "A set of ids of security groups assigned to hosts of the cluster.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			// Optional nested attribute maintenance_window required all optional nested attributes
			// But if the block is specified explicitly, then the type attribute is required
			"maintenance_window": schema.SingleNestedAttribute{
				MarkdownDescription: "Maintenance policy of the MySQL cluster.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Object{
					NewMaintenanceWindowStructValidator(),
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
		},
		Blocks: map[string]schema.Block{
			"resources": schema.SingleNestedBlock{
				MarkdownDescription: "Resources allocated to hosts of the MySQL cluster.",
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

	tflog.Debug(ctx, "Creating MySQL Cluster")

	hostSpecsSlice, diags := mdbcommon.CreateClusterHosts(ctx, mysqlHostService, plan.HostSpecs)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	// Prepare Create Request
	request, diags := prepareCreateRequest(ctx, &plan, &r.providerConfig.ProviderState)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// Add Hosts to the request
	request.HostSpecs = hostSpecsSlice

	cid := mysqlApi.CreateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, request)
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

	tflog.Debug(ctx, "Updating MySQL Cluster", map[string]interface{}{"id": plan.Id.ValueString()})
	tflog.Debug(ctx, fmt.Sprintf("Update MySQL Cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update MySQL Cluster plan: %+v", plan))

	updateVersionRequest, d := prepareVersionUpdateRequest(&state, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	mysqlApi.UpdateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateVersionRequest)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest, d := prepareUpdateRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	mysqlApi.UpdateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateRequest)
	if resp.Diagnostics.HasError() {
		return
	}

	mdbcommon.UpdateClusterHosts[Host, *mysql.Host, *mysql.HostSpec, mysql.UpdateHostSpec](
		ctx,
		r.providerConfig.SDK,
		&resp.Diagnostics,
		mysqlHostService,
		&mysqlApi,
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
	mysqlApi.DeleteCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid)
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *clusterResource) refreshResourceState(ctx context.Context, state *Cluster, respDiagnostics *diag.Diagnostics) {
	cid := state.Id.ValueString()
	cluster := mysqlApi.GetCluster(ctx, r.providerConfig.SDK, respDiagnostics, cid)
	if respDiagnostics.HasError() {
		return
	}

	entityIdToApiHosts := mdbcommon.ReadHosts(ctx, r.providerConfig.SDK, respDiagnostics, mysqlHostService, &mysqlApi, state.HostSpecs, cid)

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
	state.SecurityGroupIds = flattenSetString(ctx, cluster.SecurityGroupIds, respDiagnostics)

	cfg := flattenConfig(ctx, state.MySQLConfig, cluster.GetConfig(), respDiagnostics)

	state.Version = cfg.Version
	state.Resources = cfg.Resources
	state.Access = cfg.Access
	state.PerformanceDiagnostics = cfg.PerformanceDiagnostics
	state.BackupRetainPeriodDays = cfg.BackupRetainPeriodDays
	state.BackupWindowStart = cfg.BackupWindowStart
	state.MySQLConfig = cfg.MySQLConfig
}
