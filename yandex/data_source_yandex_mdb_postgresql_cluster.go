package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexMDBPostgreSQLCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed PostgreSQL cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/). [How to connect to the DB](https://yandex.cloud/docs/managed-postgresql/quickstart#connect). To connect, use port 6432. The port number is not configurable.\n\n~> Either `cluster_id` or `name` should be specified.\n",

		Read: dataSourceYandexMDBPostgreSQLClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the PostgreSQL cluster.",
				Computed:    true,
				Optional:    true,
			},
			"config": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["config"].Description,
				Computed:    true,
				Elem:        dataSourceYandexMDBPostgreSQLClusterConfigBlock(),
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["environment"].Description,
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["health"].Description,
				Computed:    true,
			},
			"host": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["host"].Description,
				Computed:    true,
				Elem:        dataSourceYandexMDBPostgreSQLClusterHostBlock(),
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the PostgreSQL cluster.",
				Computed:    true,
				Optional:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["status"].Description,
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
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["maintenance_window"].Description,
				Computed:    true,
				Elem:        dataSourceYandexMDBPostgreSQLClusterMaintenanceWindowBlock(),
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
				Optional:    true,
			},
			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["host_group_ids"].Description,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"disk_encryption_key_id": {
				Type:        schema.TypeString,
				Description: "ID of the KMS key for cluster disk encryption.",
				Computed:    true,
				Optional:    true,
			},
			"database": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["database"].Description,
				Computed:    true,
				Set:         mysqlDatabaseHash,
				Elem:        dataSourceYandexMDBPostgreSQLClusterDatabaseBlock(),
			},
			"user": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLCluster().Schema["user"].Description,
				Computed:    true,
				Elem:        dataSourceYandexMDBPostgreSQLClusterUserBlock(),
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterDatabaseBlock() *schema.Resource {
	extensionElem := (resourceYandexMDBPostgreSQLClusterDatabaseBlock().Schema["extension"].Elem).(*schema.Resource)
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterDatabaseBlock().Schema["name"].Description,
				Required:    true,
			},
			"owner": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterDatabaseBlock().Schema["owner"].Description,
				ForceNew:    true,
				Required:    true,
			},
			"lc_collate": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterDatabaseBlock().Schema["lc_collate"].Description,
				Optional:    true,
				ForceNew:    true,
				Default:     "C",
			},
			"lc_type": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterDatabaseBlock().Schema["lc_type"].Description,
				ForceNew:    true,
				Optional:    true,
				Default:     "C",
			},
			"template_db": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterDatabaseBlock().Schema["template_db"].Description,
				ForceNew:    true,
				Optional:    true,
			},
			"extension": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBPostgreSQLClusterDatabaseBlock().Schema["extension"].Description,
				Set:         pgExtensionHash,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: extensionElem.Schema["name"].Description,
							Required:    true,
						},
						"version": {
							Type:        schema.TypeString,
							Description: extensionElem.Schema["version"].Description,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterUserBlock() *schema.Resource {
	permissionElem := (resourceYandexMDBPostgreSQLClusterUserBlock().Schema["permission"].Elem).(*schema.Resource)
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterUserBlock().Schema["name"].Description,
				Required:    true,
			},
			"login": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBPostgreSQLClusterUserBlock().Schema["login"].Description,
				Optional:    true,
				Default:     true,
			},
			"grants": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLClusterUserBlock().Schema["grants"].Description,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
			// TODO change to permissions
			"permission": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBPostgreSQLClusterUserBlock().Schema["permission"].Description,
				Optional:    true,
				Computed:    true,
				Set:         pgUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:        schema.TypeString,
							Description: permissionElem.Schema["database_name"].Description,
							Required:    true,
						},
					},
				},
			},
			"conn_limit": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBPostgreSQLClusterUserBlock().Schema["conn_limit"].Description,
				Optional:    true,
				Computed:    true,
			},
			"settings": {
				Type:             schema.TypeMap,
				Description:      resourceYandexMDBPostgreSQLClusterUserBlock().Schema["settings"].Description,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterConfigBlock() *schema.Resource {
	accessElem := (resourceYandexMDBPostgreSQLClusterConfig().Schema["access"].Elem).(*schema.Resource)
	performanceDiagnosticsElem := (resourceYandexMDBPostgreSQLClusterConfig().Schema["performance_diagnostics"].Elem).(*schema.Resource)
	backupWindowStart := (resourceYandexMDBPostgreSQLClusterConfig().Schema["backup_window_start"].Elem).(*schema.Resource)
	diskSizeAutoscalingElem := (resourceYandexMDBPostgreSQLClusterConfig().Schema["disk_size_autoscaling"].Elem).(*schema.Resource)
	poolerConfigElem := (resourceYandexMDBPostgreSQLClusterConfig().Schema["pooler_config"].Elem).(*schema.Resource)
	resourcesElem := (resourceYandexMDBPostgreSQLClusterConfig().Schema["resources"].Elem).(*schema.Resource)

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"access": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["access"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_lens": {
							Type:        schema.TypeBool,
							Description: accessElem.Schema["data_lens"].Description,
							Computed:    true,
						},
						"web_sql": {
							Type:        schema.TypeBool,
							Description: accessElem.Schema["web_sql"].Description,
							Computed:    true,
						},
						"serverless": {
							Type:        schema.TypeBool,
							Description: accessElem.Schema["serverless"].Description,
							Computed:    true,
						},
						"data_transfer": {
							Type:        schema.TypeBool,
							Description: accessElem.Schema["data_transfer"].Description,
							Computed:    true,
						},
					},
				},
			},
			"autofailover": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["autofailover"].Description,
				Computed:    true,
			},
			"backup_window_start": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["backup_window_start"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:        schema.TypeInt,
							Description: backupWindowStart.Schema["hours"].Description,
							Computed:    true,
						},
						"minutes": {
							Type:        schema.TypeInt,
							Description: backupWindowStart.Schema["minutes"].Description,
							Computed:    true,
						},
					},
				},
			},
			"backup_retain_period_days": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["backup_retain_period_days"].Description,
				Computed:    true,
			},
			"performance_diagnostics": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["performance_diagnostics"].Description,
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
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["disk_size_autoscaling"].Description,
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
			"pooler_config": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["pooler_config"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pool_discard": {
							Type:        schema.TypeBool,
							Description: poolerConfigElem.Schema["pool_discard"].Description,
							Computed:    true,
						},
						"pooling_mode": {
							Type:        schema.TypeString,
							Description: poolerConfigElem.Schema["pooling_mode"].Description,
							Computed:    true,
						},
					},
				},
			},
			"resources": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["resources"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size": {
							Type:        schema.TypeInt,
							Description: resourcesElem.Schema["disk_size"].Description,
							Computed:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: resourcesElem.Schema["disk_type_id"].Description,
							Computed:    true,
						},
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: resourcesElem.Schema["resource_preset_id"].Description,
							Computed:    true,
						},
					},
				},
			},
			"version": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["version"].Description,
				Computed:    true,
			},
			"postgresql_config": {
				Type:        schema.TypeMap,
				Description: resourceYandexMDBPostgreSQLClusterConfig().Schema["postgresql_config"].Description,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterHostBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"assign_public_ip": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBPostgreSQLClusterHost().Schema["assign_public_ip"].Description,
				Computed:    true,
			},
			"fqdn": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterHost().Schema["fqdn"].Description,
				Computed:    true,
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterHost().Schema["subnet_id"].Description,
				Computed:    true,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterHost().Schema["zone"].Description,
				Computed:    true,
			},
			"role": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterHost().Schema["role"].Description,
				Computed:    true,
			},
			"replication_source": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterHost().Schema["replication_source"].Description,
				Computed:    true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBPostgreSQLClusterHost().Schema["priority"].Description,
				Computed:    true,
				Deprecated:  "The field has not affected anything. You can safely delete it.",
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterMaintenanceWindowBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterMaintenanceWindow().Schema["type"].Description,
				Computed:    true,
			},
			"day": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLClusterMaintenanceWindow().Schema["day"].Description,
				Computed:    true,
			},
			"hour": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBPostgreSQLClusterMaintenanceWindow().Schema["hour"].Description,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.PostgreSQLClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source PostgreSQL Cluster by name: %v", err)
		}
	}

	cluster, err := config.sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", clusterID))
	}

	databases, err := listPGDatabases(ctx, config, clusterID)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", clusterID))
	}
	if err := d.Set("database", flattenPGDatabases(databases)); err != nil {
		return err
	}

	passwords := pgUsersPasswords(make([]*postgresql.UserSpec, 0))
	users, err := listPGUsers(ctx, config, clusterID)
	if err != nil {
		return err
	}
	fUsers, err := flattenPGUsers(users, passwords, mdbPGUserSettingsFieldsInfo)
	if err != nil {
		return err
	}
	if err := d.Set("user", fUsers); err != nil {
		return err
	}

	pgClusterConfig, err := flattenPGClusterConfig(cluster.Config)
	if err != nil {
		return err
	}
	if err := d.Set("config", pgClusterConfig); err != nil {
		return err
	}

	hosts, err := retryListPGHostsWrapper(ctx, config, clusterID)
	if err != nil {
		return err
	}

	orderedHostInfos, err := flattenPGHostsInfo(d, hosts)
	if err != nil {
		return err
	}

	hs := flattenPGHostsFromHostInfos(d, orderedHostInfos, true)
	if err := d.Set("host", hs); err != nil {
		return err
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	maintenanceWindow, err := flattenPGMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}

	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
		return err
	}

	if err = d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	if cluster.DiskEncryptionKeyId != nil {
		if err = d.Set("disk_encryption_key_id", cluster.DiskEncryptionKeyId.String()); err != nil {
			return err
		}
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
	d.Set("deletion_protection", cluster.DeletionProtection)

	d.SetId(cluster.Id)
	return nil
}
