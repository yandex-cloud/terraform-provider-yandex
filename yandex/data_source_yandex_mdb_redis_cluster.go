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

		Read: dataSourceYandexMDBRedisClusterRead,
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
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"maxmemory_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"notify_keyspace_events": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slowlog_log_slower_than": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"slowlog_max_len": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"client_output_buffer_limit_normal": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_output_buffer_limit_pubsub": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_luajit": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"io_threads_allowed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"databases": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"maxmemory_percent": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lua_time_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"repl_backlog_size_percent": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cluster_require_full_coverage": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"cluster_allow_reads_when_down": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"cluster_allow_pubsubshard_when_down": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"lfu_decay_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"lfu_log_factor": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"turn_before_switchover": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"allow_data_loss": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"backup_window_start": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disk_type_id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
					},
				},
			},
			"disk_size_autoscaling": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"planned_usage_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"emergency_usage_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"host": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"shard_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replica_priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
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
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hour": {
							Type:     schema.TypeInt,
							Computed: true,
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

	d.SetId(cluster.Id)

	return nil
}
