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
			"database": {
				Type:        schema.TypeSet,
				Description: "~> Deprecated! To manage databases, please switch to using a separate resource type `yandex_mdb_postgresql_database`.",
				Computed:    true,
				Set:         mysqlDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: common.ResourceDescriptions["name"],
							Type:        schema.TypeString,
							Required:    true,
						},
						"owner": {
							Type:        schema.TypeString,
							Description: "Name of the user assigned as the owner of the database. Forbidden to change in an existing database.",
							ForceNew:    true,
							Required:    true,
						},
						"lc_collate": {
							Type:        schema.TypeString,
							Description: "POSIX locale for string sorting order. Forbidden to change in an existing database.",
							Optional:    true,
							ForceNew:    true,
							Default:     "C",
						},
						"lc_type": {
							Type:        schema.TypeString,
							Description: "POSIX locale for character classification. Forbidden to change in an existing database.",
							ForceNew:    true,
							Optional:    true,
							Default:     "C",
						},
						"template_db": {
							Type:        schema.TypeString,
							Description: "Name of the template database.",
							ForceNew:    true,
							Optional:    true,
						},
						"extension": {
							Type:        schema.TypeSet,
							Description: "Set of database extensions.",

							Set:      pgExtensionHash,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Description: "Name of the database extension. For more information on available extensions see [the official documentation](https://yandex.cloud/docs/managed-postgresql/operations/cluster-extensions).",
										Required:    true,
									},
									"version": {
										Type:        schema.TypeString,
										Description: "Version of the extension.",
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
			"user": {
				Type:        schema.TypeList,
				Description: "~> Deprecated! To manage users, please switch to using a separate resource type `yandex_mdb_postgresql_user`.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the user.",
							Required:    true,
						},
						"login": {
							Type:        schema.TypeBool,
							Description: "User's ability to login.",
							Optional:    true,
							Default:     true,
						},
						"grants": {
							Type:        schema.TypeList,
							Description: "List of the user's grants.",
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
							Description: "Set of permissions granted to the user.",
							Optional:    true,
							Computed:    true,
							Set:         pgUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:        schema.TypeString,
										Description: "The name of the database that the permission grants access to.",
										Required:    true,
									},
								},
							},
						},
						"conn_limit": {
							Type:        schema.TypeInt,
							Description: "The maximum number of connections per user. (Default 50).",
							Optional:    true,
							Computed:    true,
						},
						"settings": {
							Type:             schema.TypeMap,
							Description:      "Map of user settings. [Full description](https://yandex.cloud/docs/managed-postgresql/api-ref/grpc/Cluster/create#yandex.cloud.mdb.postgresql.v1.UserSettings).\n\n* `default_transaction_isolation` - defines the default isolation level to be set for all new SQL transactions. One of:  - 0: `unspecified`\n  - 1: `read uncommitted`\n  - 2: `read committed`\n  - 3: `repeatable read`\n  - 4: `serializable`\n\n* `lock_timeout` - The maximum time (in milliseconds) for any statement to wait for acquiring a lock on an table, index, row or other database object (default 0)\n\n* `log_min_duration_statement` - This setting controls logging of the duration of statements. (default -1 disables logging of the duration of statements.)\n\n* `synchronous_commit` - This setting defines whether DBMS will commit transaction in a synchronous way. One of:\n  - 0: `unspecified`\n  - 1: `on`\n  - 2: `off`\n  - 3: `local`\n  - 4: `remote write`\n  - 5: `remote apply`\n\n* `temp_file_limit` - The maximum storage space size (in kilobytes) that a single process can use to create temporary files.\n\n* `log_statement` - This setting specifies which SQL statements should be logged (on the user level). One of:\n  - 0: `unspecified`\n  - 1: `none`\n  - 2: `ddl`\n  - 3: `mod`\n  - 4: `all`\n\n* `pool_mode` - Mode that the connection pooler is working in with specified user. One of:\n  - 1: `session`\n  - 2: `transaction`\n  - 3: `statement`\n\n* `prepared_statements_pooling` - This setting allows user to use prepared statements with transaction pooling. Boolean.\n\n* `catchup_timeout` - The connection pooler setting. It determines the maximum allowed replication lag (in seconds). Pooler will reject connections to the replica with a lag above this threshold. Default value is 0, which disables this feature. Integer.\n\n* `wal_sender_timeout` - The maximum time (in milliseconds) to wait for WAL replication (can be set only for PostgreSQL 12+). Terminate replication connections that are inactive for longer than this amount of time. Integer.\n\n* `idle_in_transaction_session_timeout` - Sets the maximum allowed idle time (in milliseconds) between queries, when in a transaction. Value of 0 (default) disables the timeout. Integer.\n\n* `statement_timeout` - The maximum time (in milliseconds) to wait for statement. Value of 0 (default) disables the timeout. Integer\n\n",
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo),
							ValidateFunc:     generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterConfigBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"access": {
				Type:        schema.TypeList,
				Description: "Access policy to the PostgreSQL cluster.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_lens": {
							Type:        schema.TypeBool,
							Description: "Allow access for [Yandex DataLens](https://yandex.cloud/services/datalens).",
							Computed:    true,
						},
						"web_sql": {
							Type:        schema.TypeBool,
							Description: "Allow access for [SQL queries in the management console](https://yandex.cloud/docs/managed-postgresql/operations/web-sql-query).",
							Computed:    true,
						},
						"serverless": {
							Type:        schema.TypeBool,
							Description: "Allow access for [connection to managed databases from functions](https://yandex.cloud/docs/functions/operations/database-connection).",
							Computed:    true,
						},
						"data_transfer": {
							Type:        schema.TypeBool,
							Description: "Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer).",
							Computed:    true,
						},
					},
				},
			},
			"autofailover": {
				Type:        schema.TypeBool,
				Description: "Configuration setting which enables/disables autofailover in cluster.",
				Computed:    true,
			},
			"backup_window_start": {
				Type:        schema.TypeList,
				Description: "Time to start the daily backup, in the UTC timezone.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:        schema.TypeInt,
							Description: "The hour at which backup will be started (UTC).",
							Computed:    true,
						},
						"minutes": {
							Type:        schema.TypeInt,
							Description: "The hour at which backup will be started (UTC).",
							Computed:    true,
						},
					},
				},
			},
			"backup_retain_period_days": {
				Type:        schema.TypeInt,
				Description: "The period in days during which backups are stored.",
				Computed:    true,
			},
			"performance_diagnostics": {
				Type:        schema.TypeList,
				Description: "Cluster performance diagnostics settings. [YC Documentation](https://yandex.cloud/docs/managed-postgresql/api-ref/grpc/cluster_service#PerformanceDiagnostics).",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enable performance diagnostics.",
							Computed:    true,
						},
						"sessions_sampling_interval": {
							Type:        schema.TypeInt,
							Description: "Interval (in seconds) for pg_stat_activity sampling Acceptable values are 1 to 86400, inclusive.",
							Computed:    true,
						},
						"statements_sampling_interval": {
							Type:        schema.TypeInt,
							Description: "Interval (in seconds) for pg_stat_statements sampling Acceptable values are 1 to 86400, inclusive.",
							Computed:    true,
						},
					},
				},
			},
			"disk_size_autoscaling": {
				Type:        schema.TypeList,
				Description: "Cluster disk size autoscaling settings.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size_limit": {
							Type:        schema.TypeInt,
							Description: "Limit of disk size after autoscaling (GiB).",
							Computed:    true,
						},
						"planned_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Maintenance window autoscaling disk usage (percent).",
							Computed:    true,
						},
						"emergency_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Immediate autoscaling disk usage (percent).",
							Computed:    true,
						},
					},
				},
			},
			"pooler_config": {
				Type:        schema.TypeList,
				Description: "Configuration of the connection pooler.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pool_discard": {
							Type:        schema.TypeBool,
							Description: "Setting `pool_discard` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_discard-yesno).",
							Computed:    true,
						},
						"pooling_mode": {
							Type:        schema.TypeString,
							Description: "Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.",
							Computed:    true,
						},
					},
				},
			},
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the PostgreSQL cluster.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size": {
							Type:        schema.TypeInt,
							Description: "The ID of the preset for computational resources available to a PostgreSQL host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/concepts/instance-types).",
							Computed:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Volume of the storage available to a PostgreSQL host, in gigabytes.",
							Computed:    true,
						},
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of PostgreSQL hosts.",
							Computed:    true,
						},
					},
				},
			},
			"version": {
				Type:        schema.TypeString,
				Description: "Version of the PostgreSQL cluster. (allowed versions are: 12, 12-1c, 13, 13-1c, 14, 14-1c, 15, 15-1c, 16, 17).",
				Computed:    true,
			},
			"postgresql_config": {
				Type:        schema.TypeMap,
				Description: "PostgreSQL cluster configuration. For detailed information specific to your PostgreSQL version, please refer to the [API proto specifications](https://github.com/yandex-cloud/cloudapi/tree/master/yandex/cloud/mdb/postgresql/v1/config).",
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
				Description: "Sets whether the host should get a public IP address on creation. It can be changed on the fly only when `name` is set.",
				Computed:    true,
			},
			"fqdn": {
				Type:        schema.TypeString,
				Description: "The fully qualified domain name of the host.",
				Computed:    true,
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
				Computed:    true,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
			},
			"role": {
				Type:        schema.TypeString,
				Description: "Host's role (replica|primary), computed by server.",
				Computed:    true,
			},
			"replication_source": {
				Type:        schema.TypeString,
				Description: "Host replication source (fqdn), when replication_source is empty then host is in HA group.",
				Computed:    true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: "Host priority in HA group. It works only when `name` is set.",
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
				Description: "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
				Computed:    true,
			},
			"day": {
				Type:        schema.TypeString,
				Description: "Day of the week (in `DDD` format). Allowed values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`",
				Computed:    true,
			},
			"hour": {
				Type:        schema.TypeInt,
				Description: "Hour of the day in UTC (in `HH` format). Allowed value is between 1 and 24.",
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
