package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBClickHouseCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBClickHouseClusterRead,
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
			"clickhouse": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							MaxItems: 1,
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
									},
								},
							},
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      clickHouseUserHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"permission": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      clickHouseUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      clickHouseDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"host": {
				Type:     schema.TypeList,
				MinItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"shard_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_window_start": {
				Type:     schema.TypeList,
				MaxItems: 1,
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
			"access": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"web_sql": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"data_lens": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"metrika": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"serverless": {
							Type:     schema.TypeBool,
							Computed: true,
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
			"zookeeper": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							MaxItems: 1,
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
									},
								},
							},
						},
					},
				},
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
		},
	}
}

func dataSourceYandexMDBClickHouseClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.ClickhouseClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source ClickHouse Cluster by name: %v", err)
		}
	}

	cluster, err := config.sdk.MDB().Clickhouse().Cluster().Get(ctx, &clickhouse.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	chResources, err := flattenClickHouseResources(cluster.Config.Clickhouse.Resources)
	if err != nil {
		return err
	}
	d.Set("clickhouse", []map[string]interface{}{
		{
			"resources": chResources,
		},
	})

	zkResources, err := flattenClickHouseResources(cluster.Config.Zookeeper.Resources)
	if err != nil {
		return err
	}
	d.Set("zookeeper", []map[string]interface{}{
		{
			"resources": zkResources,
		},
	})

	bws := flattenClickHouseBackupWindowStart(cluster.Config.BackupWindowStart)
	if err := d.Set("backup_window_start", bws); err != nil {
		return err
	}

	ac := flattenClickHouseAccess(cluster.Config.Access)
	if err := d.Set("access", ac); err != nil {
		return err
	}

	hosts, err := listClickHouseHosts(ctx, config, clusterID)
	if err != nil {
		return err
	}
	hs, err := flattenClickHouseHosts(hosts)
	if err != nil {
		return err
	}
	if err := d.Set("host", hs); err != nil {
		return err
	}

	databases, err := listClickHouseDatabases(ctx, config, clusterID)
	if err != nil {
		return err
	}
	dbs := flattenClickHouseDatabases(databases)
	if err := d.Set("database", dbs); err != nil {
		return err
	}

	users, err := listClickHouseUsers(ctx, config, clusterID)
	if err != nil {
		return err
	}
	us := flattenClickHouseUsers(users, nil)
	if err := d.Set("user", us); err != nil {
		return err
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("cluster_id", cluster.Id)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)

	d.SetId(cluster.Id)
	return nil
}
