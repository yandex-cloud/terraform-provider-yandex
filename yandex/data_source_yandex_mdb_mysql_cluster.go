package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBMySQLCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBMySQLClusterRead,
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
			"description": {
				Type:     schema.TypeString,
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
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
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
							Optional: true,
							Default:  0,
						},
						"minutes": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
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
						"disk_type_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      mysqlDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
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
							Optional: true,
							Computed: true,
							Set:      mysqlUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"roles": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
									},
								},
							},
						},
						"global_permissions": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},
						"connection_limits": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_questions_per_hour": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_updates_per_hour": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_connections_per_hour": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_user_connections": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"authentication_plugin": {
							Type:     schema.TypeString,
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
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"backup_priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
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
			"mysql_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbMySQLSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbMySQLSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"access": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_lens": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"web_sql": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
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

func dataSourceYandexMDBMySQLClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.MySQLClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source MySQL Cluster by name: %v", err)
		}
	}
	cluster, err := config.sdk.MDB().MySQL().Cluster().Get(ctx, &mysql.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	d.Set("folder_id", cluster.GetFolderId())
	d.Set("name", cluster.GetName())
	d.Set("cluster_id", cluster.Id)
	d.Set("description", cluster.GetDescription())
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("network_id", cluster.GetNetworkId())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("version", cluster.GetConfig().GetVersion())

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	hosts, err := listMysqlHosts(ctx, config, clusterID)
	if err != nil {
		return err
	}

	fHosts, err := flattenMysqlHosts(d, hosts, true)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] reading cluster:")
	for i, h := range fHosts {
		log.Printf("[DEBUG] match [%d]: %s -> %s", i, h["name"], h["fqdn"])
	}

	if err := d.Set("host", fHosts); err != nil {
		return err
	}

	userSpecs, err := expandMySQLUsers(nil, d)
	if err != nil {
		return err
	}
	passwords := mysqlUsersPasswords(userSpecs)
	users, err := listMysqlUsers(ctx, config, clusterID)
	if err != nil {
		return err
	}
	fUsers, err := flattenMysqlUsers(users, passwords)
	if err != nil {
		return err
	}

	if err := d.Set("user", fUsers); err != nil {
		return err
	}

	databases, err := listMysqlDatabases(ctx, config, clusterID)
	if err != nil {
		return err
	}

	fDatabases := flattenMysqlDatabases(databases)
	if err := d.Set("database", fDatabases); err != nil {
		return err
	}

	mysqlResources, err := flattenMysqlResources(cluster.GetConfig().GetResources())
	if err != nil {
		return err
	}
	err = d.Set("resources", mysqlResources)
	if err != nil {
		return err
	}

	backupWindowStart, err := flattenMysqlBackupWindowStart(cluster.GetConfig().GetBackupWindowStart())
	if err != nil {
		return err
	}
	if err := d.Set("backup_window_start", backupWindowStart); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	clusterConfig, err := flattenMySQLSettings(cluster.Config)
	if err != nil {
		return err
	}

	if err := d.Set("mysql_config", clusterConfig); err != nil {
		return err
	}

	access, err := flattenMySQLAccess(cluster.Config.Access)
	if err != nil {
		return err
	}

	if err := d.Set("access", access); err != nil {
		return err
	}

	maintenanceWindow, err := flattenMysqlMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}

	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.SetId(clusterID)
	return nil
}
