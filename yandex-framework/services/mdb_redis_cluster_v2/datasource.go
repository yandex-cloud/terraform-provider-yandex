package mdb_redis_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func NewDataSource() datasource.DataSource {
	return &redisClusterDataSource{}
}

type redisClusterDataSource struct {
	providerConfig *provider_config.Config
}

// Configure implements datasource.DataSource.
func (o *redisClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (o *redisClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_redis_cluster_v2"
}

// Read implements datasource.DataSource.
func (o *redisClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	clusterId, diagnostics := mdbcommon.GetClusterIdForDatasource(ctx, o.providerConfig, req.Config)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config Cluster
	config.ID = types.StringValue(clusterId)
	clusterRead(ctx, o.providerConfig.SDK, &resp.Diagnostics, &config)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *redisClusterDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Info(ctx, "Initializing Redis data source schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Redis cluster within the Yandex Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/). [How to connect to the DB](https://yandex.cloud/docs/managed-redis/quickstart#connect). To connect, use port 6379. The port number is not configurable.",
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
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the Redis cluster. This ID is assigned by MDB at creation time.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["name"],
			},
			"network_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["network_id"],
			},
			"environment": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Deployment environment of the Redis cluster.",
			},
			"hosts": schema.MapNestedAttribute{
				Computed:            true,
				MarkdownDescription: "A hosts of the Redis cluster as label:host_info pairs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: common.ResourceDescriptions["zone"],
						},
						"shard_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Shard Name of the host in the cluster.",
						},
						"subnet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the subnet where the host is located.",
						},
						"fqdn": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Fully Qualified Domain Name. In other words, hostname.",
						},
						"assign_public_ip": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Assign a public IP address to the host. Can be either true or false.",
						},
						"replica_priority": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "A replica with a low priority number is considered better for promotion.",
						},
					},
				},
			},

			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["description"],
			},
			"labels": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: common.ResourceDescriptions["labels"],
			},
			"sharded": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Redis sharded mode. Can be either true or false.",
			},
			"tls_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "TLS port and functionality. Can be either true or false.",
			},
			"persistence_mode": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Persistence mode.",
			},
			"announce_hostnames": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Announce fqdn instead of ip address. Can be either true or false.",
			},
			"folder_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["created_at"],
			},
			"security_group_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["security_group_ids"],
			},
			"deletion_protection": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["deletion_protection"],
			},
			"auth_sentinel": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Allows to use ACL users to auth in sentinel",
			},
			"disk_encryption_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the symmetric encryption key used to encrypt the disk of the cluster.",
			},
			"resources": schema.SingleNestedAttribute{
				MarkdownDescription: "Resources allocated to hosts of the Redis cluster.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"resource_preset_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "ID of the resource preset that determines the number of CPU cores and memory size for the host.",
					},
					"disk_size": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Size of the disk in bytes.",
					},
					"disk_type_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "ID of the disk type that determines the disk performance characteristics.",
					},
				},
			},
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of the Redis cluster.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"timeout": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Time that Redis keeps the connection open while the client is idle.",
					},
					"password": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "Authentication password.",
					},
					"maxmemory_policy": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Redis key eviction policy for a dataset that reaches maximum memory, available to the host.",
					},
					"notify_keyspace_events": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: `String setting for pub\sub functionality.`,
					},
					"slowlog_log_slower_than": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Threshold for logging slow requests to server in microseconds (log only slower than it).",
					},
					"slowlog_max_len": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Max slow requests number to log.",
					},
					"databases": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Number of database buckets on a single redis-server process.",
					},
					"maxmemory_percent": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Redis maxmemory percent",
					},
					"client_output_buffer_limit_normal": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Redis connection output buffers limits for clients.",
					},
					"client_output_buffer_limit_pubsub": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Redis connection output buffers limits for pubsub operations.",
					},
					"use_luajit": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Use JIT for lua scripts and functions. Can be either true or false.",
					},
					"io_threads_allowed": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Allow redis to use io-threads. Can be either true or false.",
					},
					"version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: `Redis version.`,
					},
					"lua_time_limit": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Maximum time in milliseconds for Lua scripts, 0 - disabled mechanism.",
					},
					"repl_backlog_size_percent": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Replication backlog size as a percentage of flavor maxmemory.",
					},
					"cluster_require_full_coverage": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Controls whether all hash slots must be covered by nodes. Can be either true or false.",
					},
					"cluster_allow_reads_when_down": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Allows read operations when cluster is down. Can be either true or false.",
					},
					"cluster_allow_pubsubshard_when_down": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: `Permits Pub/Sub shard operations when cluster is down. Can be either true or false.`,
					},
					"lfu_decay_time": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The time, in minutes, that must elapse in order for the key counter to be divided by two (or decremented if it has a value less <= 10).",
					},
					"lfu_log_factor": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Determines how the frequency counter represents key hits.",
					},
					"turn_before_switchover": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Allows to turn before switchover in RDSync. Can be either true or false.",
					},
					"allow_data_loss": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: `Allows some data to be lost in favor of faster switchover/restart. Can be either true or false.`,
					},
					"backup_retain_period_days": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Retain period of automatically created backup in days.",
					},
					"zset_max_listpack_entries": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Controls max number of entries in zset before conversion from memory-efficient listpack to CPU-efficient hash table and skiplist",
					},
					"backup_window_start": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Time to start the daily backup, in the UTC timezone.",

						Attributes: map[string]schema.Attribute{
							"hours": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "The hour at which backup will be started.",
							},
							"minutes": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "The minute at which backup will be started.",
							},
						},
					},
				},
			},

			"access": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Access policy to the Redis cluster.",

				Attributes: map[string]schema.Attribute{
					"data_lens": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Allow access for Yandex DataLens. Can be either true or false.",
					},
					"web_sql": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Allow access for SQL queries in the management console. Can be either true or false.",
					},
				},
			},

			"disk_size_autoscaling": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Disk size autoscaling settings.",

				Attributes: map[string]schema.Attribute{
					"disk_size_limit": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Limit of disk size after autoscaling in bytes.",
					},
					"planned_usage_threshold": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Maintenance window autoscaling disk usage (percent).",
					},
					"emergency_usage_threshold": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Immediate autoscaling disk usage (percent).",
					},
				},
			},
			"maintenance_window": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Maintenance window settings of the Redis cluster.",

				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Type of maintenance window.",
					},
					"day": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Day of week for maintenance window if window type is weekly.",
					},
					"hour": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
					},
				},
			},
		},
	}
}
