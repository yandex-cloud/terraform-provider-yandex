package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexMDBMySQLClusterDefaultTimeout = 30 * time.Minute
	yandexMDBMySQLClusterUpdateTimeout  = 60 * time.Minute
)

func resourceYandexMDBMySQLCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a MySQL cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mysql/).",

		Create: resourceYandexMDBMySQLClusterCreate,
		Read:   resourceYandexMDBMySQLClusterRead,
		Update: resourceYandexMDBMySQLClusterUpdate,
		Delete: resourceYandexMDBMySQLClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBMySQLClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBMySQLClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBMySQLClusterDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the MySQL cluster.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseMysqlEnv),
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
				ForceNew:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Description: "Version of the MySQL cluster. (allowed versions are: 5.7, 8.0).",
				Required:    true,
			},
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the MySQL cluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a MySQL host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-mysql/concepts/instance-types).",
							Required:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of MySQL hosts.",
							Required:    true,
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Description: "Volume of the storage available to a MySQL host, in gigabytes.",
							Required:    true,
						},
					},
				},
			},
			"database": {
				Type:        schema.TypeSet,
				Description: "To manage databases, please switch to using a separate resource type `yandex_mdb_mysql_databases`.",
				Optional:    true,
				Set:         mysqlDatabaseHash,
				Deprecated:  useResourceInstead("database", "yandex_mdb_mysql_database"),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the database.",
							Required:    true,
						},
					},
				},
			},
			"user": {
				Type:        schema.TypeList,
				Description: "To manage users, please switch to using a separate resource type `yandex_mdb_mysql_user`.",
				Optional:    true,
				Deprecated:  useResourceInstead("user", "yandex_mdb_mysql_user"),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the user.",
							Required:    true,
						},
						"password": {
							Type:        schema.TypeString,
							Description: "The password of the user.",
							Required:    true,
							Sensitive:   true,
						},
						"permission": {
							Type:        schema.TypeSet,
							Description: "Set of permissions granted to the user.",
							Optional:    true,
							Computed:    true,
							Set:         mysqlUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:        schema.TypeString,
										Description: "The name of the database that the permission grants access to.",
										Required:    true,
									},
									"roles": {
										Type:        schema.TypeList,
										Description: "List user's roles in the database. Allowed roles: `ALL`,`ALTER`,`ALTER_ROUTINE`,`CREATE`,`CREATE_ROUTINE`,`CREATE_TEMPORARY_TABLES`, `CREATE_VIEW`,`DELETE`,`DROP`,`EVENT`,`EXECUTE`,`INDEX`,`INSERT`,`LOCK_TABLES`,`SELECT`,`SHOW_VIEW`,`TRIGGER`,`UPDATE`.",
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
							Description: "List user's global permissions. Allowed permissions: `REPLICATION_CLIENT`, `REPLICATION_SLAVE`, `PROCESS` for clear list use empty list. If the attribute is not specified there will be no changes.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"connection_limits": {
							Type:        schema.TypeList,
							Description: "User's connection limits. If not specified there will be no changes. Default value is -1. When these parameters are set to -1, backend default values will be actually used.",
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_questions_per_hour": {
										Type:        schema.TypeInt,
										Description: "Max questions per hour.",
										Optional:    true,
										Default:     -1,
									},
									"max_updates_per_hour": {
										Type:        schema.TypeInt,
										Description: "Max updates per hour.",
										Optional:    true,
										Default:     -1,
									},
									"max_connections_per_hour": {
										Type:        schema.TypeInt,
										Description: "Max connections per hour.",
										Optional:    true,
										Default:     -1,
									},
									"max_user_connections": {
										Type:        schema.TypeInt,
										Description: "Max user connections.",
										Optional:    true,
										Default:     -1,
									},
								},
							},
						},
						"authentication_plugin": {
							Type:        schema.TypeString,
							Description: "Authentication plugin. Allowed values: `MYSQL_NATIVE_PASSWORD`, `CACHING_SHA2_PASSWORD`, `SHA256_PASSWORD` (for version 5.7 `MYSQL_NATIVE_PASSWORD`, `SHA256_PASSWORD`).",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"host": {
				Type:        schema.TypeList,
				Description: "A host of the MySQL cluster.",
				MinItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:        schema.TypeString,
							Description: common.ResourceDescriptions["zone"],
							Required:    true,
						},
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Description: "Sets whether the host should get a public IP address. It can be changed on the fly only when `name` is set.",
							Optional:    true,
							Default:     false,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
							Optional:    true,
							Computed:    true,
						},
						"fqdn": {
							Type:        schema.TypeString,
							Description: "The fully qualified domain name of the host.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Host state name. It should be set for all hosts or unset for all hosts. This field can be used by another host, to select which host will be its replication source. Please refer to `replication_source_name` parameter.",
							Optional:    true,
						},
						"replication_source": {
							Type:        schema.TypeString,
							Description: "Host replication source (fqdn), when replication_source is empty then host is in HA group.",
							Computed:    true,
						},
						"replication_source_name": {
							Type:        schema.TypeString,
							Description: "Host replication source name points to host's `name` from which this host should replicate. When not set then host in HA group. It works only when `name` is set.",
							Optional:    true,
						},
						"priority": {
							Type:        schema.TypeInt,
							Description: "Host master promotion priority. Value is between 0 and 100, default is 0.",
							Optional:    true,
						},
						"backup_priority": {
							Type:        schema.TypeInt,
							Description: "Host backup priority. Value is between 0 and 100, default is 0.",
							Optional:    true,
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
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
				Description: "Time to start the daily backup, in the UTC.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:         schema.TypeInt,
							Description:  "The hour at which backup will be started.",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minutes": {
							Type:         schema.TypeInt,
							Description:  "The minute at which backup will be started.",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},
			"backup_retain_period_days": {
				Type:        schema.TypeInt,
				Description: "The period in days during which backups are stored.",
				Optional:    true,
				Computed:    true,
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
				Description: "Aggregated health of the cluster.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the cluster.",
				Computed:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
			},
			"restore": {
				Type:        schema.TypeList,
				Description: "The cluster will be created from the specified backup.",
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_id": {
							Type:        schema.TypeString,
							Description: "Backup ID. The cluster will be created from the specified backup. [How to get a list of MySQL backups](https://yandex.cloud/docs/managed-mysql/operations/cluster-backups).",
							Required:    true,
							ForceNew:    true,
						},
						"time": {
							Type:         schema.TypeString,
							Description:  "Timestamp of the moment to which the MySQL cluster should be restored. (Format: `2006-01-02T15:04:05` - UTC). When not set, current time is used.",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: stringToTimeValidateFunc,
						},
					},
				},
			},
			"mysql_config": {
				Type:             schema.TypeMap,
				Description:      "MySQL cluster config block.",
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
				Description: "Access policy to the MySQL cluster.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_lens": {
							Type:        schema.TypeBool,
							Description: "Allow access for [Yandex DataLens](https://yandex.cloud/services/datalens).",
							Optional:    true,
							Default:     false,
						},
						"web_sql": {
							Type:        schema.TypeBool,
							Description: "Allows access for [SQL queries in the management console](https://yandex.cloud/docs/managed-mysql/operations/web-sql-query).",
							Optional:    true,
							Default:     false,
						},
						"data_transfer": {
							Type:        schema.TypeBool,
							Description: "Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer).",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"allow_regeneration_host": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Deprecated field. You can safely remove this option. There is no need to recreate host if assign_public_ip is changed.\nRegenerate hosts after changing the assign_public_ip parameter.",
				Deprecated:  "You can safely remove this option. There is no need to recreate host if assign_public_ip is changed.",
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: "Maintenance policy of the MySQL cluster.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Description:  "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
							ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
							Required:     true,
						},
						"day": {
							Type:         schema.TypeString,
							Description:  "Day of the week (in `DDD` format). Allowed values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`",
							ValidateFunc: mdbMaintenanceWindowSchemaValidateFunc,
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							Description:  "Hour of the day in UTC (in `HH` format). Allowed value is between 1 and 24.",
							ValidateFunc: validation.IntBetween(1, 24),
							Optional:     true,
						},
					},
				},
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
			"performance_diagnostics": {
				Type:        schema.TypeList,
				Description: "Cluster performance diagnostics settings. [YC Documentation](https://yandex.cloud/docs/managed-mysql/api-ref/grpc/cluster_service#PerformanceDiagnostics).",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enable performance diagnostics.",
							Optional:    true,
							Computed:    true,
						},
						"sessions_sampling_interval": {
							Type:        schema.TypeInt,
							Description: "Interval (in seconds) for my_stat_activity sampling Acceptable values are 1 to 86400, inclusive.",
							Required:    true,
						},
						"statements_sampling_interval": {
							Type:        schema.TypeInt,
							Description: "Interval (in seconds) for my_stat_statements sampling Acceptable values are 1 to 86400, inclusive.",
							Required:    true,
						},
					},
				},
			},
			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: "",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceYandexMDBMySQLClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := validateClusterConfig(d)
	if err != nil {
		return err
	}

	req, err := prepareCreateMySQLRequest(d, config)
	if err != nil {
		return err
	}

	if backupID, ok := d.GetOk("restore.0.backup_id"); ok && backupID != "" {
		return resourceYandexMDBMySQLClusterRestore(d, meta, req, backupID.(string))
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.MDB().MySQL().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create MySQL Cluster: %s", err)
	}
	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get MySQL create operation metadata: %s", err)
	}
	md, ok := protoMetadata.(*mysql.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get MySQL Cluster ID from create operation metadata")
	}
	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create MySQL Cluster: %s", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("MySQL Cluster creation failed: %s", err)
	}

	// Update hosts after creation (e.g. configure cascade replicas)
	log.Printf("[INFO] Updating cluster hosts after creation (if needed)...")
	if err := updateMysqlClusterHosts(d, config); err != nil {
		return fmt.Errorf("MySQL Cluster %v update params failed: %s", d.Id(), err)
	}

	return resourceYandexMDBMySQLClusterRead(d, meta)
}

func resourceYandexMDBMySQLClusterRestore(d *schema.ResourceData, meta interface{}, createClusterRequest *mysql.CreateClusterRequest, backupID string) error {
	config := meta.(*Config)
	req, err := prepareCreateMySQLRequest(d, config)
	if err != nil {
		return err
	}

	timeBackup := time.Now()

	if backupTime, ok := d.GetOk("restore.0.time"); ok {
		var err error
		timeBackup, err = parseStringToTime(backupTime.(string))
		if err != nil {
			return fmt.Errorf("Error while parsing restore.0.time to create MySQL Cluster from backup %v, value: %v error: %s", backupID, backupTime, err)
		}

	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.MDB().MySQL().Cluster().Restore(ctx, &mysql.RestoreClusterRequest{
		BackupId: backupID,
		Time: &timestamp.Timestamp{
			Seconds: timeBackup.Unix(),
		},
		Name:              req.Name,
		Description:       req.Description,
		Labels:            req.Labels,
		Environment:       req.Environment,
		ConfigSpec:        req.ConfigSpec,
		HostSpecs:         req.HostSpecs,
		NetworkId:         req.NetworkId,
		FolderId:          req.FolderId,
		SecurityGroupIds:  req.SecurityGroupIds,
		HostGroupIds:      req.HostGroupIds,
		MaintenanceWindow: req.MaintenanceWindow,
	}))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create MySQL Cluster from backup %v: %s", backupID, err)
	}
	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while getting MySQL create operation metadata from backup %v: %s", backupID, err)
	}
	md, ok := protoMetadata.(*mysql.RestoreClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get MySQL Cluster ID from create from backup %v operation metadata", backupID)
	}
	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create MySQL Cluster from backup %v: %s", backupID, err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("MySQL Cluster creation from backup %v failed: %s", backupID, err)
	}

	if err := updateMysqlClusterHosts(d, config); err != nil {
		return fmt.Errorf("MySQL Cluster %v hosts creation from backup %v failed: %s", d.Id(), backupID, err)
	}

	return resourceYandexMDBMySQLClusterRead(d, meta)
}

func prepareCreateMySQLRequest(d *schema.ResourceData, meta *Config) (*mysql.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, fmt.Errorf("Error while expanding labels on MySQL Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating MySQL Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseMysqlEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating MySQL Cluster: %s", err)
	}

	dbSpecs, err := expandMysqlDatabases(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding databases on Mysql Cluster create: %s", err)
	}

	hostSpecs, err := expandMysqlHostSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding hostsFromScheme on MySQL Cluster create: %s", err)
	}
	// It is not possible to specify replication-source during cluster creation (host names are unknown)
	// so, create all hosts as HA-hosts, and then reconfigure it
	for _, hostSpec := range hostSpecs {
		hostSpec.ReplicationSource = ""
	}

	users, err := expandMySQLUsers(nil, d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding user specs on MySQL Cluster create: %s", err)
	}

	configSpec, err := expandMySQLConfigSpec(d)
	if err != nil {
		return nil, err
	}

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on MySQL Cluster create: %s", err)
	}

	maintenanceWindow, err := expandMySQLMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding maintenance_window on MySQL Cluster create: %s", err)
	}

	return &mysql.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Environment:        env,
		ConfigSpec:         configSpec,
		DatabaseSpecs:      dbSpecs,
		UserSpecs:          users,
		HostSpecs:          hostSpecs,
		Labels:             labels,
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		DeletionProtection: d.Get("deletion_protection").(bool),
		HostGroupIds:       expandHostGroupIds(d.Get("host_group_ids")),
		MaintenanceWindow:  maintenanceWindow,
	}, nil
}

func resourceYandexMDBMySQLClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().MySQL().Cluster().Get(ctx, &mysql.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	d.Set("folder_id", cluster.GetFolderId())
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

	hosts, err := listMysqlHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	fHosts, err := flattenMysqlHosts(d, hosts, false)
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

	stateUser := d.Get("user").([]interface{})
	if len(stateUser) == 0 {
		if err := d.Set("user", []map[string]interface{}{}); err != nil {
			return err
		}
	} else {
		users, err := listMysqlUsers(ctx, config, d.Id())
		if err != nil {
			return err
		}
		userSpecs, err := expandMySQLUsers(nil, d)
		if err != nil {
			return err
		}
		passwords := mysqlUsersPasswords(userSpecs)

		fUsers, err := flattenMysqlUsers(users, passwords)
		if err != nil {
			return err
		}

		sortInterfaceListByResourceData(fUsers, d, "user", "name")

		if err := d.Set("user", fUsers); err != nil {
			return err
		}
	}

	stateDatabases := d.Get("database").(*schema.Set).List()
	if len(stateDatabases) == 0 {
		if err := d.Set("database", []map[string]interface{}{}); err != nil {
			return err
		}
	} else {
		databases, err := listMysqlDatabases(ctx, config, d.Id())
		if err != nil {
			return err
		}

		fDatabases := flattenMysqlDatabases(databases)
		if err := d.Set("database", fDatabases); err != nil {
			return err
		}
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

	if err := d.Set("deletion_protection", cluster.DeletionProtection); err != nil {
		return err
	}

	perfDiag, err := flattenMyPerformanceDiagnostics(cluster.Config.PerformanceDiagnostics)
	if err != nil {
		return err
	}

	if err := d.Set("performance_diagnostics", perfDiag); err != nil {
		return err
	}

	if err = d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	if err = d.Set("backup_retain_period_days", int(cluster.Config.BackupRetainPeriodDays.Value)); err != nil {
		return err
	}

	return d.Set("created_at", getTimestamp(cluster.CreatedAt))
}

func resourceYandexMDBMySQLClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	d.Partial(true)

	err := validateClusterConfig(d)
	if err != nil {
		return err
	}

	if err = setMySQLFolderID(d, meta); err != nil {
		return err
	}

	if err := updateMysqlClusterParams(d, meta); err != nil {
		return err
	}

	stateDatabase := d.Get("database").(*schema.Set).List()
	if d.HasChange("database") && len(stateDatabase) > 0 {
		if err := updateMysqlClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	stateUser := d.Get("user").([]interface{})
	if d.HasChange("user") && len(stateUser) > 0 {
		if err := updateMysqlClusterUsers(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updateMysqlClusterHosts(d, config); err != nil {
			return err
		}
	}

	d.Partial(false)
	return resourceYandexMDBMySQLClusterRead(d, meta)
}

var mdbMysqlUpdateFieldsMap = map[string]string{
	"name":                      "name",
	"description":               "description",
	"labels":                    "labels",
	"network_id":                "network_id",
	"access":                    "config_spec.access",
	"backup_window_start":       "config_spec.backup_window_start",
	"resources":                 "config_spec.resources",
	"backup_retain_period_days": "config_spec.backup_retain_period_days",
	"version":                   "config_spec.version",
	"performance_diagnostics":   "config_spec.performance_diagnostics",
	"security_group_ids":        "security_group_ids",
	"maintenance_window":        "maintenance_window",
	"deletion_protection":       "deletion_protection",
}

func updateMysqlClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	request, err := prepareMySQLClusterUpdateRequest(d, config)
	if err != nil {
		return err
	}

	if len(request.UpdateMask.Paths) == 0 {
		return nil
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MySQL().Cluster().Update(ctx, request))
	if err != nil {
		return fmt.Errorf("error while requesting API to update MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating MySQL Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func prepareMySQLClusterUpdateRequest(d *schema.ResourceData, config *Config) (*mysql.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating MySQL cluster: %s", err)
	}

	if d.HasChange("host_group_ids") {
		return nil, fmt.Errorf("host_group_ids change is not supported yet")
	}

	maintenanceWindow, err := expandMySQLMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding maintenance_window while updating MySQL cluster: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, config)
	if err != nil {
		return nil, fmt.Errorf("error expanding network_id while updating MySQL cluster: %s", err)
	}

	configSpec, err := expandMySQLConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding mysql_config while updating MySQL cluster: %s", err)
	}

	updatePaths := []string{}
	for field, path := range mdbMysqlUpdateFieldsMap {
		if d.HasChange(field) {
			updatePaths = append(updatePaths, path)
		}
	}

	if d.HasChange("mysql_config") {
		version := d.Get("version").(string)
		updatePaths = append(updatePaths, "config_spec."+getMySQLConfigFieldName(version))
	}

	return &mysql.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		NetworkId:          networkID,
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		MaintenanceWindow:  maintenanceWindow,
		DeletionProtection: d.Get("deletion_protection").(bool),
		ConfigSpec:         configSpec,
		UpdateMask:         &field_mask.FieldMask{Paths: updatePaths},
	}, nil
}

func updateMysqlClusterDatabases(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currDBs, err := listMysqlDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandMysqlDatabases(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mysqlDatabasesDiff(currDBs, targetDBs)

	for _, db := range toDelete {
		err := deleteMysqlDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createMysqlDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	return nil
}

// Takes the current list of dbs and the desirable list of dbs.
// Returns the slice of dbs to delete and the slice of dbs to add.
func mysqlDatabasesDiff(currDBs []*mysql.Database, targetDBs []*mysql.DatabaseSpec) ([]string, []string) {
	m := map[string]bool{}
	toAdd := []string{}
	toDelete := map[string]bool{}
	for _, db := range currDBs {
		toDelete[db.Name] = true
		m[db.Name] = true
	}

	for _, db := range targetDBs {
		delete(toDelete, db.Name)
		if _, ok := m[db.Name]; !ok {
			toAdd = append(toAdd, db.Name)
		}
	}

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func deleteMysqlDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Database().Delete(ctx, &mysql.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMysqlDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Database().Create(ctx, &mysql.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &mysql.DatabaseSpec{
				Name: dbName,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateMysqlClusterUsers(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	currUsers, err := listMysqlUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetUsers, err := expandMySQLUsers(currUsers, d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mysqlUsersDiff(currUsers, targetUsers)

	for _, u := range toDelete {
		err := deleteMysqlUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}
	for _, u := range toAdd {
		err := createMysqlUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	currUsers, err = listMysqlUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	changedUsers, err := mysqlChangedUsers(currUsers, d)
	if err != nil {
		return err
	}

	for _, u := range changedUsers {
		err := updateMysqlUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	return nil
}

// Takes the current list of users and the desirable list of users.
// Returns the slice of usernames to delete and the slice of users to add.
func mysqlUsersDiff(currUsers []*mysql.User, targetUsers []*mysql.UserSpec) ([]string, []*mysql.UserSpec) {
	m := map[string]bool{}
	toDelete := map[string]bool{}
	toAdd := []*mysql.UserSpec{}

	for _, u := range currUsers {
		toDelete[u.Name] = true
		m[u.Name] = true
	}

	for _, u := range targetUsers {
		delete(toDelete, u.Name)
		if _, ok := m[u.Name]; !ok {
			toAdd = append(toAdd, u)
		}
	}

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func deleteMysqlUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().User().Delete(ctx, &mysql.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMysqlUser(ctx context.Context, config *Config, d *schema.ResourceData, user *mysql.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().User().Create(ctx, &mysql.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

// Takes the old set of user specs and the new set of user specs.
// Returns the slice of user specs which have changed.
func mysqlChangedUsers(users []*mysql.User, d *schema.ResourceData) ([]*mysql.UserSpec, error) {

	oldSpecs, newSpecs := d.GetChange("user")
	result := []*mysql.UserSpec{}
	oldPwd := make(map[string]string)

	for _, spec := range oldSpecs.([]interface{}) {
		m := spec.(map[string]interface{})
		oldPwd[m["name"].(string)] = m["password"].(string)
	}

	usersMap := make(map[string]*mysql.User)

	for _, u := range users {
		usersMap[u.Name] = u
	}

	for _, u := range newSpecs.([]interface{}) {
		m := u.(map[string]interface{})
		user, isDiff, err := expandMysqlUser(m, usersMap[(m["name"]).(string)])
		if err != nil {
			return nil, err
		}
		if isDiff || oldPwd[user.Name] != user.Password {
			result = append(result, user)
		}
	}

	return result, nil
}

func updateMysqlUser(ctx context.Context, config *Config, d *schema.ResourceData, user *mysql.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().User().Update(ctx, &mysql.UpdateUserRequest{
			ClusterId:            d.Id(),
			UserName:             user.Name,
			Password:             user.Password,
			Permissions:          user.Permissions,
			AuthenticationPlugin: user.AuthenticationPlugin,
			ConnectionLimits:     user.ConnectionLimits,
			GlobalPermissions:    user.GlobalPermissions,
			UpdateMask:           &field_mask.FieldMask{Paths: []string{"authentication_plugin", "password", "permissions", "connection_limits", "global_permissions"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateMysqlClusterHosts(d *schema.ResourceData, meta interface{}) error {
	// Ideas:
	// 1. In order to do it safely for clients: firstly add new hosts and only then delete unneeded hosts
	// 2. Batch Add/Update operations are not supported, so we should update hosts one by one
	//    It may produce issues with cascade replicas: we should change replication-source in such way, that
	//    there is no attempts to create replication loop
	//    Solution: update HA-replicas first, then use BFS (using `compareMySQLHostsInfoResult.hierarchyExists`)

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// Step 1: Add new hosts (as HA-hosts):
	err := createMysqlClusterHosts(ctx, config, d)
	if err != nil {
		return err
	}

	// Step 2: update hosts:
	currHosts, err := listMysqlHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	compareHostsInfo, err := compareMySQLHostsInfo(d, currHosts, true)
	if err != nil {
		return err
	}

	for _, hostInfo := range compareHostsInfo.hostsInfo {
		if hostInfo.inTargetSet {
			var maskPaths []string
			if hostInfo.oldReplicationSource != hostInfo.newReplicationSource {
				maskPaths = append(maskPaths, "replication_source")
			}
			if hostInfo.oldAssignPublicIP != hostInfo.newAssignPublicIP {
				maskPaths = append(maskPaths, "assign_public_ip")
			}
			if hostInfo.oldBackupPriority != hostInfo.newBackupPriority {
				maskPaths = append(maskPaths, "backup_priority")
			}
			if hostInfo.oldPriority != hostInfo.newPriority {
				maskPaths = append(maskPaths, "priority")
			}
			if len(maskPaths) > 0 {
				log.Printf("[DEBUG] Updating host (change paths: %v)", maskPaths)
				if err := updateMySQLHost(ctx, config, d, &mysql.UpdateHostSpec{
					HostName:          hostInfo.fqdn,
					ReplicationSource: hostInfo.newReplicationSource,
					AssignPublicIp:    hostInfo.newAssignPublicIP,
					BackupPriority:    hostInfo.newBackupPriority,
					Priority:          hostInfo.newPriority,
					UpdateMask:        &field_mask.FieldMask{Paths: maskPaths},
				}); err != nil {
					return err
				}
			}
		}
	}

	// Step 3: delete hosts:
	for _, hostInfo := range compareHostsInfo.hostsInfo {
		if !hostInfo.inTargetSet {
			log.Printf("[DEBUG] Deleting host %v", hostInfo.fqdn)
			if err := deleteMysqlHost(ctx, config, d, hostInfo.fqdn); err != nil {
				return err
			}
		}
	}

	return nil
}

func printCompareHostInfo(compareHostInfo compareMySQLHostsInfoResult) {
	log.Printf("[DEBUG] Current cluster hosts view:")
	for _, hi := range compareHostInfo.hostsInfo {
		log.Printf("[DEBUG] %s -> %s", hi.name, hi.fqdn)
	}
	for _, chi := range compareHostInfo.createHostsInfo {
		log.Printf("[DEBUG] new %s", chi.name)
	}
}

func createMysqlClusterHosts(ctx context.Context, config *Config, d *schema.ResourceData) error {
	currHosts, err := listMysqlHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	compareHostsInfo, err := compareMySQLHostsInfo(d, currHosts, true)
	if err != nil {
		return err
	}
	printCompareHostInfo(compareHostsInfo)

	if compareHostsInfo.hierarchyExists && len(compareHostsInfo.createHostsInfo) == 0 {
		return fmt.Errorf("Create cluster hosts error. Exists host with replication source, which can't be created. Possibly there is a loop")
	}

	var newHosts []*mysql.HostSpec
	for _, newHostInfo := range compareHostsInfo.createHostsInfo {
		newHosts = append(newHosts, &mysql.HostSpec{
			ZoneId:         newHostInfo.zone,
			SubnetId:       newHostInfo.subnetID,
			AssignPublicIp: newHostInfo.newAssignPublicIP,
			BackupPriority: newHostInfo.newBackupPriority,
			Priority:       newHostInfo.newPriority,
		})
	}

	for _, newHost := range newHosts { // batch operations are not supported
		log.Printf("[DEBUG] Add new host: %+v", newHost)
		err = addMySQLHost(ctx, config, d, newHost)
		if err != nil {
			return err
		}
	}

	if compareHostsInfo.hierarchyExists {
		return createMysqlClusterHosts(ctx, config, d)
	}

	return nil
}

func deleteMysqlHost(ctx context.Context, config *Config, d *schema.ResourceData, fqdn string) error {
	op, err := config.sdk.WrapOperation(
		// FYI: Deleting multiple hosts at once is not supported yet
		config.sdk.MDB().MySQL().Cluster().DeleteHosts(ctx, &mysql.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: []string{fqdn},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from MySQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting host from MySQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func resourceYandexMDBMySQLClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting MySQL Cluster %q", d.Id())

	req := &mysql.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MySQL().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("MySQL Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting MySQL Cluster %q", d.Id())
	return nil
}

func validateClusterConfig(d *schema.ResourceData) error {
	targetHosts, err := expandEnrichedMySQLHostSpec(d)
	if err != nil {
		return err
	}
	err = validateMysqlReplicationReferences(targetHosts)
	if err != nil {
		return err
	}

	return nil
}

func setMySQLFolderID(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().MySQL().Cluster().Get(ctx, &mysql.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	folderID, ok := d.GetOk("folder_id")
	if !ok {
		return nil
	}
	if folderID == "" {
		return nil
	}

	if cluster.FolderId != folderID {
		request := &mysql.MoveClusterRequest{
			ClusterId:           d.Id(),
			DestinationFolderId: folderID.(string),
		}
		op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending MySQL cluster move request: %+v", request)
			return config.sdk.MDB().MySQL().Cluster().Move(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("error while requesting API to move MySQL Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while moving MySQL Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("moving MySQL Cluster %q to folder %v failed: %s", d.Id(), folderID, err)
		}

	}

	return nil
}
