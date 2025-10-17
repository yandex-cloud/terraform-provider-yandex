package mdb_redis_cluster_v2

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"golang.org/x/exp/maps"
)

const (
	yandexMDBRedisClusterCreateTimeout = 45 * time.Minute
	yandexMDBRedisClusterUpdateTimeout = 60 * time.Minute
	yandexMDBRedisClusterDeleteTimeout = 20 * time.Minute
	defaultReplicaPriority             = 100
)

var (
	baseOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}
)

type redisClusterResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &redisClusterResource{}
}

func (r *redisClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_redis_cluster_v2"
}

func (r *redisClusterResource) Configure(_ context.Context,
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

func (r *redisClusterResource) Schema(ctx context.Context,
	_ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Redis cluster within the Yandex Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/). [How to connect to the DB](https://yandex.cloud/docs/managed-redis/quickstart#connect). To connect, use port 6379. The port number is not configurable.",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": defaultschema.Id(),
			"cluster_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the Redis cluster. This ID is assigned by MDB at creation time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: common.ResourceDescriptions["name"],
			},
			"network_id": defaultschema.NetworkId(stringplanmodifier.RequiresReplace()),
			"environment": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators:          []validator.String{stringvalidator.OneOf(maps.Keys(redis.Cluster_Environment_value)...)},
				MarkdownDescription: "Deployment environment of the Redis cluster.",
			},
			"hosts": schema.MapNestedAttribute{
				Required:            true,
				MarkdownDescription: "A hosts of the Redis cluster as label:host_info pairs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: common.ResourceDescriptions["zone"],
						},
						"shard_name": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Shard Name of the host in the cluster.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"subnet_id": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "ID of the subnet where the host is located.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"fqdn": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Fully Qualified Domain Name. In other words, hostname.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"assign_public_ip": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Assign a public IP address to the host. Can be either true or false.",
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"replica_priority": schema.Int64Attribute{
							Optional:            true,
							MarkdownDescription: "A replica with a low priority number is considered better for promotion.",
							Computed:            true,
							Default:             int64default.StaticInt64(defaultReplicaPriority),
						},
					},
				},
				PlanModifiers: []planmodifier.Map{
					MapWarningHostsChangedAfterImport(),
				},
			},

			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: common.ResourceDescriptions["description"],
			},
			"labels": defaultschema.Labels(),
			"sharded": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Redis sharded mode. Can be either true or false.",
			},
			"tls_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "TLS port and functionality. Can be either true or false.",
			},
			"persistence_mode": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf(maps.Keys(redis.Cluster_PersistenceMode_value)...)},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Persistence mode.",
			},
			"announce_hostnames": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Announce fqdn instead of ip address. Can be either true or false.",
			},
			"folder_id":           defaultschema.FolderId(),
			"created_at":          defaultschema.CreatedAt(),
			"security_group_ids":  defaultschema.SecurityGroupIds(),
			"deletion_protection": defaultschema.DeletionProtection(),
			"auth_sentinel": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Allows to use ACL users to auth in sentinel",
			},
			"disk_encryption_key_id": schema.StringAttribute{
				Description: "ID of the symmetric encryption key used to encrypt the disk of the cluster.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resources": schema.SingleNestedAttribute{
				MarkdownDescription: "Resources allocated to hosts of the Redis cluster.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"resource_preset_id": schema.StringAttribute{
						MarkdownDescription: "ID of the resource preset that determines the number of CPU cores and memory size for the host.",
						Required:            true,
					},
					"disk_size": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "Size of the disk in bytes.",
					},
					"disk_type_id": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "ID of the disk type that determines the disk performance characteristics.",
					},
				},
			},
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of the Redis cluster.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"password": schema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						MarkdownDescription: "Authentication password.",
					},
					"timeout": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Time that Redis keeps the connection open while the client is idle.",
					},
					"maxmemory_policy": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Redis key eviction policy for a dataset that reaches maximum memory, available to the host.",
					},
					"notify_keyspace_events": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: `String setting for pub\sub functionality.`,
					},
					"slowlog_log_slower_than": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Threshold for logging slow requests to server in microseconds (log only slower than it).",
					},
					"slowlog_max_len": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Max slow requests number to log.",
					},
					"databases": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Number of database buckets on a single redis-server process.",
					},
					"maxmemory_percent": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Redis maxmemory percent",
					},
					"client_output_buffer_limit_normal": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Redis connection output buffers limits for clients.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"client_output_buffer_limit_pubsub": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Redis connection output buffers limits for pubsub operations.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"use_luajit": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Use JIT for lua scripts and functions. Can be either true or false.",
					},
					"io_threads_allowed": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Allow redis to use io-threads. Can be either true or false.",
					},
					"version": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: `Redis version.`,
					},
					"lua_time_limit": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Maximum time in milliseconds for Lua scripts, 0 - disabled mechanism.",
					},
					"repl_backlog_size_percent": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Replication backlog size as a percentage of flavor maxmemory.",
					},
					"cluster_require_full_coverage": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Controls whether all hash slots must be covered by nodes. Can be either true or false.",
					},
					"cluster_allow_reads_when_down": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Allows read operations when cluster is down. Can be either true or false.",
					},
					"cluster_allow_pubsubshard_when_down": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: `Permits Pub/Sub shard operations when cluster is down. Can be either true or false.`,
					},
					"lfu_decay_time": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The time, in minutes, that must elapse in order for the key counter to be divided by two (or decremented if it has a value less <= 10).",
					},
					"lfu_log_factor": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Determines how the frequency counter represents key hits.",
					},
					"turn_before_switchover": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Allows to turn before switchover in RDSync. Can be either true or false.",
					},
					"allow_data_loss": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: `Allows some data to be lost in favor of faster switchover/restart. Can be either true or false.`,
					},
					"backup_retain_period_days": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Retain period of automatically created backup in days.",
					},
					"zset_max_listpack_entries": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Controls max number of entries in zset before conversion from memory-efficient listpack to CPU-efficient hash table and skiplist",
					},
					"backup_window_start": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Time to start the daily backup, in the UTC timezone.",
						Attributes: map[string]schema.Attribute{
							"hours": schema.Int64Attribute{
								Required:            true,
								Validators:          []validator.Int64{int64validator.Between(0, 23)},
								MarkdownDescription: "The hour at which backup will be started.",
							},
							"minutes": schema.Int64Attribute{
								Required:            true,
								Validators:          []validator.Int64{int64validator.Between(0, 59)},
								MarkdownDescription: "The minute at which backup will be started.",
							},
						},
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},

			"access": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Access policy to the Redis cluster.",
				Attributes: map[string]schema.Attribute{
					"data_lens": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Allow access for Yandex DataLens. Can be either true or false.",
					},
					"web_sql": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Allow access for SQL queries in the management console. Can be either true or false.",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},

			"disk_size_autoscaling": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Disk size autoscaling settings.",
				Attributes: map[string]schema.Attribute{
					"disk_size_limit": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "Limit of disk size after autoscaling in bytes.",
					},
					"planned_usage_threshold": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Maintenance window autoscaling disk usage (percent).",
					},
					"emergency_usage_threshold": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Immediate autoscaling disk usage (percent).",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"modules": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Valkey modules.",
				Attributes: map[string]schema.Attribute{
					"valkey_search": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Valkey search module settings.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
								MarkdownDescription: "Enable Valkey search module.",
							},
							"reader_threads": schema.Int64Attribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
								MarkdownDescription: "Number of reader threads.",
							},
							"writer_threads": schema.Int64Attribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
								MarkdownDescription: "Number of writer threads.",
							},
						},
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
					},
					"valkey_json": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Valkey json module settings.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
								MarkdownDescription: "Enable Valkey json module.",
							},
						},
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
					},
					"valkey_bloom": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Valkey bloom module settings.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
								MarkdownDescription: "Enable Valkey bloom module.",
							},
						},
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					modulesUseStateForUnknownWhenNotConfigured(),
				},
			},
			"maintenance_window": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Maintenance window settings of the Redis cluster.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						Validators:          []validator.String{stringvalidator.OneOf("ANYTIME", "WEEKLY")},
						MarkdownDescription: "Type of maintenance window.",
					},
					"day": schema.StringAttribute{
						Optional:            true,
						Validators:          []validator.String{stringvalidator.OneOf(maps.Keys(redis.WeeklyMaintenanceWindow_WeekDay_value)...)},
						MarkdownDescription: "Day of week for maintenance window if window type is weekly.",
					},
					"hour": schema.Int64Attribute{
						Optional:            true,
						Validators:          []validator.Int64{int64validator.Between(1, 24)},
						MarkdownDescription: "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *redisClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	clusterRead(ctx, r.providerConfig.SDK, &resp.Diagnostics, &state)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *redisClusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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

	autoscalingOn := utils.IsPresent(state.DiskSizeAutoscaling)
	// remove changes on disk_size from plan if enabled autoscaling
	plan.Resources = mdbcommon.FixDiskSizeOnAutoscalingChanges(ctx, plan.Resources, state.Resources, autoscalingOn, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

func (r *redisClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBRedisClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	hostSpecsSlice, diags := mdbcommon.CreateClusterHosts(ctx, redisHostService, plan.HostSpecs)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	request := prepareCreateRedisRequest(ctx, r.providerConfig, &resp.Diagnostics, &plan, hostSpecsSlice)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := redisAPI.CreateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, request)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(cid)

	clusterRead(ctx, r.providerConfig.SDK, &resp.Diagnostics, &plan)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *redisClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Cluster
	var state Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBRedisClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	if !plan.FolderID.Equal(state.FolderID) {
		redisAPI.MoveCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, plan.ID.ValueString(), plan.FolderID.ValueString())
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.Sharded.Equal(state.Sharded) {
		if !plan.Sharded.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("sharded"),
				"Wrong state",
				fmt.Sprintf("Disabling sharding on Redis Cluster is not supported, Id: %q", plan.ID.ValueString()),
			)
			return
		}
		redisAPI.EnableShardingRedis(ctx, r.providerConfig.SDK, &resp.Diagnostics, state.ID.ValueString())
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateRedisClusterParams(ctx, r.providerConfig.SDK, &resp.Diagnostics, &plan, &state)
	if resp.Diagnostics.HasError() {
		return
	}

	mdbcommon.UpdateClusterHostsWithShards[Host, *redis.Host, *redis.HostSpec, redis.UpdateHostSpec](
		ctx,
		r.providerConfig.SDK,
		&resp.Diagnostics,
		redisHostService,
		&redisAPI,
		plan.ID.ValueString(),
		plan.HostSpecs,
		state.HostSpecs,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterRead(ctx, r.providerConfig.SDK, &resp.Diagnostics, &plan)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *redisClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	deleteTimeout, diags := state.Timeouts.Update(ctx, yandexMDBRedisClusterDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	redisAPI.DeleteCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, state.ID.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *redisClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddWarning(
		"Not completed resource",
		"you need to run `terraform apply` to fully",
	)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
