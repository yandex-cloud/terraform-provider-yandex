package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexMDBMySQLCluster() *schema.Resource {
	resourceSchema := resourceYandexMDBMySQLCluster().Schema
	backupWindowStartElem := resourceSchema["backup_window_start"].Elem.(*schema.Resource)
	resourcesElem := resourceSchema["resources"].Elem.(*schema.Resource)
	databaseElem := resourceSchema["database"].Elem.(*schema.Resource)
	userElem := resourceSchema["user"].Elem.(*schema.Resource)
	userPermissionElem := userElem.Schema["permission"].Elem.(*schema.Resource)
	userConnectionLimitsElem := userElem.Schema["connection_limits"].Elem.(*schema.Resource)
	hostElem := resourceSchema["host"].Elem.(*schema.Resource)
	accessElem := resourceSchema["access"].Elem.(*schema.Resource)
	maintenanceWindowElem := resourceSchema["maintenance_window"].Elem.(*schema.Resource)
	performanceDiagnosticsElem := resourceSchema["performance_diagnostics"].Elem.(*schema.Resource)
	diskSizeAutoscalingElem := resourceSchema["disk_size_autoscaling"].Elem.(*schema.Resource)

	return &schema.Resource{
		Description: "Get information about a Yandex Managed MySQL cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mysql/).\n\n~> Either `cluster_id` or `name` should be specified.\n",

		Read: dataSourceYandexMDBMySQLClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the MySQL cluster.",
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Computed:    true,
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBMySQLCluster().Schema["environment"].Description,
				Computed:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Computed:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBMySQLCluster().Schema["version"].Description,
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"backup_window_start": {
				Type:        schema.TypeList,
				Description: resourceSchema["backup_window_start"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:        schema.TypeInt,
							Description: backupWindowStartElem.Schema["hours"].Description,
							Optional:    true,
							Default:     0,
						},
						"minutes": {
							Type:        schema.TypeInt,
							Description: backupWindowStartElem.Schema["minutes"].Description,
							Optional:    true,
							Default:     0,
						},
					},
				},
			},
			"resources": {
				Type:        schema.TypeList,
				Description: resourceSchema["resources"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: resourcesElem.Schema["resource_preset_id"].Description,
							Computed:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: resourcesElem.Schema["disk_type_id"].Description,
							Computed:    true,
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Description: resourcesElem.Schema["disk_size"].Description,
							Computed:    true,
						},
					},
				},
			},
			"database": {
				Type:        schema.TypeSet,
				Description: resourceSchema["database"].Description,
				Computed:    true,
				Set:         mysqlDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: databaseElem.Schema["name"].Description,
							Computed:    true,
						},
					},
				},
			},
			"user": {
				Type:        schema.TypeList,
				Description: resourceSchema["user"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: userElem.Schema["name"].Description,
							Computed:    true,
						},
						"password": {
							Type:        schema.TypeString,
							Description: userElem.Schema["password"].Description,
							Computed:    true,
							Sensitive:   true,
						},
						"permission": {
							Type:        schema.TypeSet,
							Description: userElem.Schema["permission"].Description,
							Optional:    true,
							Computed:    true,
							Set:         mysqlUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:        schema.TypeString,
										Description: userPermissionElem.Schema["database_name"].Description,
										Computed:    true,
									},
									"roles": {
										Type:        schema.TypeList,
										Description: userPermissionElem.Schema["roles"].Description,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
									},
								},
							},
						},
						"global_permissions": {
							Type:        schema.TypeSet,
							Description: userElem.Schema["global_permissions"].Description,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},
						"connection_limits": {
							Type:        schema.TypeList,
							Description: userElem.Schema["connection_limits"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_questions_per_hour": {
										Type:        schema.TypeInt,
										Description: userConnectionLimitsElem.Schema["max_questions_per_hour"].Description,
										Computed:    true,
									},
									"max_updates_per_hour": {
										Type:        schema.TypeInt,
										Description: userConnectionLimitsElem.Schema["max_updates_per_hour"].Description,
										Computed:    true,
									},
									"max_connections_per_hour": {
										Type:        schema.TypeInt,
										Description: userConnectionLimitsElem.Schema["max_connections_per_hour"].Description,
										Computed:    true,
									},
									"max_user_connections": {
										Type:        schema.TypeInt,
										Description: userConnectionLimitsElem.Schema["max_user_connections"].Description,
										Computed:    true,
									},
								},
							},
						},
						"authentication_plugin": {
							Type:        schema.TypeString,
							Description: userElem.Schema["authentication_plugin"].Description,
							Computed:    true,
						},
					},
				},
			},
			"host": {
				Type:        schema.TypeList,
				Description: resourceSchema["host"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:        schema.TypeString,
							Description: hostElem.Schema["zone"].Description,
							Computed:    true,
						},
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Description: hostElem.Schema["assign_public_ip"].Description,
							Optional:    true,
							Default:     false,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Description: hostElem.Schema["subnet_id"].Description,
							Optional:    true,
							Computed:    true,
						},
						"fqdn": {
							Type:        schema.TypeString,
							Description: hostElem.Schema["fqdn"].Description,
							Computed:    true,
						},
						"replication_source": {
							Type:        schema.TypeString,
							Description: hostElem.Schema["replication_source"].Description,
							Computed:    true,
						},
						"priority": {
							Type:        schema.TypeInt,
							Description: hostElem.Schema["priority"].Description,
							Optional:    true,
						},
						"backup_priority": {
							Type:        schema.TypeInt,
							Description: hostElem.Schema["backup_priority"].Description,
							Optional:    true,
						},
					},
				},
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBMySQLCluster().Schema["health"].Description,
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBMySQLCluster().Schema["status"].Description,
				Computed:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"mysql_config": {
				Type:             schema.TypeMap,
				Description:      resourceSchema["mysql_config"].Description,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbMySQLSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbMySQLSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"access": {
				Type:        schema.TypeList,
				Description: resourceSchema["access"].Description,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_lens": {
							Type:        schema.TypeBool,
							Description: accessElem.Schema["data_lens"].Description,
							Computed:    true,
							Optional:    true,
						},
						"web_sql": {
							Type:        schema.TypeBool,
							Description: accessElem.Schema["web_sql"].Description,
							Computed:    true,
							Optional:    true,
						},
						"data_transfer": {
							Type:        schema.TypeBool,
							Description: accessElem.Schema["data_transfer"].Description,
							Computed:    true,
							Optional:    true,
						},
					},
				},
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: resourceSchema["maintenance_window"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Description: maintenanceWindowElem.Schema["type"].Description,
							Computed:    true,
						},
						"day": {
							Type:        schema.TypeString,
							Description: maintenanceWindowElem.Schema["day"].Description,
							Computed:    true,
						},
						"hour": {
							Type:        schema.TypeInt,
							Description: maintenanceWindowElem.Schema["hour"].Description,
							Computed:    true,
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
				Description: resourceYandexMDBMySQLCluster().Schema["disk_encryption_key_id"].Description,
				Computed:    true,
				Optional:    true,
			},
			"performance_diagnostics": {
				Type:        schema.TypeList,
				Description: resourceSchema["performance_diagnostics"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: performanceDiagnosticsElem.Schema["enabled"].Description,
							Computed:    true,
						},
						"sessions_sampling_interval": {
							Type:        schema.TypeInt,
							Description: performanceDiagnosticsElem.Schema["sessions_sampling_interval"].Description,
							Computed:    true,
						},
						"statements_sampling_interval": {
							Type:        schema.TypeInt,
							Description: performanceDiagnosticsElem.Schema["statements_sampling_interval"].Description,
							Computed:    true,
						},
					},
				},
			},
			"disk_size_autoscaling": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBMySQLCluster().Schema["disk_size_autoscaling"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size_limit": {
							Type:        schema.TypeInt,
							Description: diskSizeAutoscalingElem.Schema["disk_size_limit"].Description,
							Computed:    true,
						},
						"planned_usage_threshold": {
							Type:        schema.TypeInt,
							Description: diskSizeAutoscalingElem.Schema["planned_usage_threshold"].Description,
							Computed:    true,
						},
						"emergency_usage_threshold": {
							Type:        schema.TypeInt,
							Description: diskSizeAutoscalingElem.Schema["emergency_usage_threshold"].Description,
							Computed:    true,
						},
					},
				},
			},
			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: resourceSchema["host_group_ids"].Description,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"backup_retain_period_days": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBMySQLCluster().Schema["backup_retain_period_days"].Description,
				Computed:    true,
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

	backupWindowStart := flattenMDBBackupWindowStart(cluster.GetConfig().GetBackupWindowStart())
	if err := d.Set("backup_window_start", backupWindowStart); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	clusterConfig, err := flattenMySQLConfig(cluster.Config)
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

	if cluster.DiskEncryptionKeyId != nil {
		if err = d.Set("disk_encryption_key_id", cluster.DiskEncryptionKeyId.GetValue()); err != nil {
			return err
		}
	}

	if err := d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	if err := d.Set("backup_retain_period_days", cluster.Config.BackupRetainPeriodDays.Value); err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.SetId(clusterID)
	return nil
}
