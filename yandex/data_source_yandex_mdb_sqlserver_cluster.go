package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/sqlserver/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBSQLServerCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBSQLServerClusterRead,
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
			"environment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disk_type_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disk_size": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"permission": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      sqlserverUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"roles": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"host_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
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
						"assign_public_ip": {
							Type:     schema.TypeBool,
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
			"sqlserver_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbSQLServerSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbSQLServerSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"sqlcollation": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceYandexMDBSQLServerClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.SQLServerClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source SQLServer Cluster by name: %v", err)
		}
	}
	cluster, err := config.sdk.MDB().SQLServer().Cluster().Get(ctx, &sqlserver.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	d.Set("folder_id", cluster.GetFolderId())
	d.Set("cluster_id", cluster.Id)
	d.Set("name", cluster.GetName())
	d.Set("description", cluster.GetDescription())
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("network_id", cluster.GetNetworkId())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("version", cluster.GetConfig().GetVersion())

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if err := d.Set("resources", flattenSQLServerResources(cluster.Config.Resources)); err != nil {
		return err
	}

	usersSpec, err := listSQLServerUsers(ctx, config, cluster.Id)
	if err != nil {
		return err
	}

	passwords := expandSQLServerUserPasswords(d)

	users, err := flattenSQLServerUsers(usersSpec, passwords)

	if err != nil {
		return err
	}

	if err = d.Set("user", users); err != nil {
		return err
	}
	if err = d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	if err = d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	hostsSpec, err := listSQLServerHosts(ctx, config, cluster.Id)
	if err != nil {
		return err
	}
	hosts, err := flattenSQLServerHosts(d, hostsSpec)
	if err != nil {
		return err
	}
	if err = d.Set("host", hosts); err != nil {
		return err
	}

	databasesSpec, err := listSQLServerDatabases(ctx, config, cluster.Id)
	if err != nil {
		return err
	}

	databases := flattenSQLServerDatabases(databasesSpec)

	if err = d.Set("database", databases); err != nil {
		return err
	}

	backupWindowStart := flattenMDBBackupWindowStart(cluster.GetConfig().GetBackupWindowStart())
	if err = d.Set("backup_window_start", backupWindowStart); err != nil {
		return err
	}

	clusterConfig, err := flattenSQLServerSettings(cluster.Config)
	if err != nil {
		return err
	}

	if err := d.Set("sqlserver_config", clusterConfig); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)
	d.Set("sqlcollation", cluster.Sqlcollation)

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.SetId(cluster.Id)
	return nil
}
