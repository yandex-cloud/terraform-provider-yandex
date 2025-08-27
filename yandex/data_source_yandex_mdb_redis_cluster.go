package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexMDBRedisCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed Redis cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-redis/concepts).\n\n~> Either `cluster_id` or `name` should be specified.\n",
		Read:        dataSourceYandexMDBRedisClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Redis cluster.",
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the Redis cluster.",
				Computed:    true,
				Optional:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Computed:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBRedisCluster().Schema["environment"].Description,
				Computed:    true,
			},
			"config": {
				Description: resourceYandexMDBRedisCluster().Schema["config"].Description,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeout": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Close the connection after a client is idle for N seconds.",
						},
						"maxmemory_policy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Redis key eviction policy for a dataset that reaches maximum memory. Can be any of the listed in [the official RedisDB documentation](https://docs.redislabs.com/latest/rs/administering/database-operations/eviction-policy/).",
						},
						"notify_keyspace_events": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Select the events that Redis will notify among a set of classes.",
						},
						"slowlog_log_slower_than": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Log slow queries below this number in microseconds.",
						},
						"slowlog_max_len": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Slow queries log length.",
						},
						"client_output_buffer_limit_normal": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Normal clients output buffer limits. See [redis config file](https://github.com/redis/redis/blob/6.2/redis.conf#L1841).",
						},
						"client_output_buffer_limit_pubsub": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Pubsub clients output buffer limits. See [redis config file](https://github.com/redis/redis/blob/6.2/redis.conf#L1843).",
						},
						"use_luajit": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Use JIT for lua scripts and functions.",
						},
						"io_threads_allowed": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow Redis to use io-threads.",
						},
						"databases": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of databases (changing requires redis-server restart).",
						},
						"maxmemory_percent": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Redis maxmemory usage in percent",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Version of Redis",
						},
						"lua_time_limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum time in milliseconds for Lua scripts.",
						},
						"repl_backlog_size_percent": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Replication backlog size as a percentage of flavor maxmemory.",
						},
						"cluster_require_full_coverage": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Controls whether all hash slots must be covered by nodes.",
						},
						"cluster_allow_reads_when_down": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allows read operations when cluster is down.",
						},
						"cluster_allow_pubsubshard_when_down": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Permits Pub/Sub shard operations when cluster is down.",
						},
						"lfu_decay_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The time, in minutes, that must elapse in order for the key counter to be divided by two (or decremented if it has a value less <= 10).",
						},
						"lfu_log_factor": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Determines how the frequency counter represents key hits.",
						},
						"turn_before_switchover": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allows to turn before switchover in RDSync.",
						},
						"allow_data_loss": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allows some data to be lost in favor of faster switchover/restart by RDSync.",
						},
						"zset_max_listpack_entries": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Controls max number of entries in zset before conversion from memory-efficient listpack to CPU-efficient hash table and skiplist",
						},
						"backup_window_start": {
							Description: "Time to start the daily backup, in the UTC timezone.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Description: "The hour at which backup will be started.",
										Type:        schema.TypeInt,
										Computed:    true,
									},
									"minutes": {
										Description: "The minute at which backup will be started.",
										Type:        schema.TypeInt,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"resources": {
				Description: resourceYandexMDBRedisCluster().Schema["resources"].Description,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-redis/concepts).",
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Volume of the storage available to a host, in gigabytes.",
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "Type of the storage of Redis hosts - environment default is used if missing.",
						},
					},
				},
			},
			"disk_size_autoscaling": {
				Description: resourceYandexMDBRedisCluster().Schema["disk_size_autoscaling"].Description,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size_limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Limit of disk size after autoscaling (GiB).",
						},
						"planned_usage_threshold": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maintenance window autoscaling disk usage (percent).",
						},
						"emergency_usage_threshold": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Immediate autoscaling disk usage (percent).",
						},
					},
				},
			},
			"host": {
				Description: resourceYandexMDBRedisCluster().Schema["host"].Description,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: common.ResourceDescriptions["zone"],
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
						},
						"shard_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the shard to which the host belongs.",
						},
						"fqdn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The fully qualified domain name of the host.",
						},
						"replica_priority": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Replica priority of a current replica (usable for non-sharded only).",
						},
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Sets whether the host should get a public IP address or not.",
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"sharded": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBRedisCluster().Schema["sharded"].Description,
				Computed:    true,
			},
			"tls_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBRedisCluster().Schema["tls_enabled"].Description,
				Computed:    true,
			},
			"persistence_mode": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBRedisCluster().Schema["persistence_mode"].Description,
				Computed:    true,
			},
			"announce_hostnames": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBRedisCluster().Schema["announce_hostnames"].Description,
				Computed:    true,
			},
			"auth_sentinel": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBRedisCluster().Schema["auth_sentinel"].Description,
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBRedisCluster().Schema["health"].Description,
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBRedisCluster().Schema["status"].Description,
				Computed:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"maintenance_window": {
				Description: resourceYandexMDBRedisCluster().Schema["maintenance_window"].Description,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
						},
						"day": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
						},
						"hour": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
						},
					},
				},
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
				Optional:    true,
			},
			"disk_encryption_key_id": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBRedisCluster().Schema["disk_encryption_key_id"].Description,
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

func dataSourceYandexMDBRedisClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.RedisClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Redis Cluster by name: %v", err)
		}
	}

	cluster, err := config.sdk.MDB().Redis().Cluster().Get(ctx, &redis.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	hosts := []*redis.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Redis().Cluster().ListHosts(ctx, &redis.ListClusterHostsRequest{
			ClusterId: clusterID,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return fmt.Errorf("Error while getting list of hosts for '%s': %s", clusterID, err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("cluster_id", cluster.Id)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)
	d.Set("sharded", cluster.Sharded)
	d.Set("tls_enabled", cluster.TlsEnabled)
	d.Set("announce_hostnames", cluster.AnnounceHostnames)
	d.Set("auth_sentinel", cluster.AuthSentinel)
	err = d.Set("persistence_mode", cluster.GetPersistenceMode().String())
	if err != nil {
		return err
	}

	conf := extractRedisConfig(cluster.Config)
	err = d.Set("config", []map[string]interface{}{
		{
			"timeout":                             conf.timeout,
			"maxmemory_policy":                    conf.maxmemoryPolicy,
			"version":                             conf.version,
			"notify_keyspace_events":              conf.notifyKeyspaceEvents,
			"slowlog_log_slower_than":             conf.slowlogLogSlowerThan,
			"slowlog_max_len":                     conf.slowlogMaxLen,
			"databases":                           conf.databases,
			"maxmemory_percent":                   conf.maxmemoryPercent,
			"client_output_buffer_limit_normal":   conf.clientOutputBufferLimitNormal,
			"client_output_buffer_limit_pubsub":   conf.clientOutputBufferLimitPubsub,
			"lua_time_limit":                      conf.luaTimeLimit,
			"repl_backlog_size_percent":           conf.replBacklogSizePercent,
			"cluster_require_full_coverage":       conf.clusterRequireFullCoverage,
			"cluster_allow_reads_when_down":       conf.clusterAllowReadsWhenDown,
			"cluster_allow_pubsubshard_when_down": conf.clusterAllowPubsubshardWhenDown,
			"lfu_decay_time":                      conf.lfuDecayTime,
			"lfu_log_factor":                      conf.lfuLogFactor,
			"turn_before_switchover":              conf.turnBeforeSwitchover,
			"allow_data_loss":                     conf.allowDataLoss,
			"use_luajit":                          conf.useLuajit,
			"zset_max_listpack_entries":           conf.zsetMaxListpackEntries,
			"io_threads_allowed":                  conf.ioThreadsAllowed,
			"backup_window_start":                 flattenMDBBackupWindowStart(cluster.GetConfig().GetBackupWindowStart()),
		},
	})
	if err != nil {
		return err
	}

	dsa, err := flattenRedisDiskSizeAutoscaling(cluster.Config.DiskSizeAutoscaling)
	if err != nil {
		return err
	}

	if err := d.Set("disk_size_autoscaling", dsa); err != nil {
		return err
	}

	resources, err := flattenRedisResources(cluster.Config.Resources)
	if err != nil {
		return err
	}

	hs, err := flattenRedisHosts(cluster.Sharded, hosts)
	if err != nil {
		return err
	}

	if err := d.Set("resources", resources); err != nil {
		return err
	}

	if err := d.Set("host", hs); err != nil {
		return err
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	mw := flattenRedisMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	if cluster.DiskEncryptionKeyId != nil {
		if err = d.Set("disk_encryption_key_id", cluster.DiskEncryptionKeyId.GetValue()); err != nil {
			return err
		}
	}

	d.SetId(cluster.Id)

	return nil
}
