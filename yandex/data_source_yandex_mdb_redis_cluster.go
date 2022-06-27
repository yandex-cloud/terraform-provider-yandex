package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBRedisCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBRedisClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Computed: true,
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
						"databases": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
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
							Optional: true,
							Default:  defaultReplicaPriority,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"sharded": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"persistence_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
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
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
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
	err = d.Set("persistence_mode", cluster.GetPersistenceMode().String())
	if err != nil {
		return err
	}

	conf := extractRedisConfig(cluster.Config)
	err = d.Set("config", []map[string]interface{}{
		{
			"timeout":                           conf.timeout,
			"maxmemory_policy":                  conf.maxmemoryPolicy,
			"version":                           conf.version,
			"notify_keyspace_events":            conf.notifyKeyspaceEvents,
			"slowlog_log_slower_than":           conf.slowlogLogSlowerThan,
			"slowlog_max_len":                   conf.slowlogMaxLen,
			"databases":                         conf.databases,
			"client_output_buffer_limit_normal": conf.clientOutputBufferLimitNormal,
			"client_output_buffer_limit_pubsub": conf.clientOutputBufferLimitPubsub,
		},
	})
	if err != nil {
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
