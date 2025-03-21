package yandex

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexMDBPostgreSQLClusterCreateTimeout = 30 * time.Minute
	yandexMDBPostgreSQLClusterDeleteTimeout = 15 * time.Minute
	yandexMDBPostgreSQLClusterUpdateTimeout = 60 * time.Minute
)

func resourceYandexMDBPostgreSQLCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a PostgreSQL cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/). [How to connect to the DB](https://yandex.cloud/docs/managed-postgresql/quickstart#connect). To connect, use port 6432. The port number is not configurable.\n\n~> Historically, `user` and `database` blocks of the `yandex_mdb_postgresql_cluster` resource were used to manage users and databases of the PostgreSQL cluster. However, this approach has many disadvantages. In particular, adding and removing a resource from the terraform recipe worked wrong because terraform misleads the user about the planned changes. Now, the recommended way to manage databases and users is using `yandex_mdb_postgresql_user` and `yandex_mdb_postgresql_database` resources.\n",

		Create: resourceYandexMDBPostgreSQLClusterCreate,
		Read:   resourceYandexMDBPostgreSQLClusterRead,
		Update: resourceYandexMDBPostgreSQLClusterUpdate,
		Delete: resourceYandexMDBPostgreSQLClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBPostgreSQLClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBPostgreSQLClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBPostgreSQLClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: "Deployment environment of the PostgreSQL cluster.",
				Required:    true,
				ForceNew:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
			},
			"config": {
				Type:        schema.TypeList,
				Description: "Configuration of the PostgreSQL cluster.",
				Required:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBPostgreSQLClusterConfig(),
			},
			"database": {
				Type:        schema.TypeList,
				Description: "~> Deprecated! To manage databases, please switch to using a separate resource type `yandex_mdb_postgresql_database`.",
				Optional:    true,
				Elem:        resourceYandexMDBPostgreSQLClusterDatabaseBlock(),
				Deprecated:  useResourceInstead("database", "yandex_mdb_postgresql_database"),
			},
			"user": {
				Type:        schema.TypeList,
				Description: "~> Deprecated! To manage users, please switch to using a separate resource type `yandex_mdb_postgresql_user`.",
				Optional:    true,
				Elem:        resourceYandexMDBPostgreSQLClusterUserBlock(),
				Deprecated:  useResourceInstead("user", "yandex_mdb_postgresql_user"),
			},
			"host": {
				Type:        schema.TypeList,
				Description: "A host of the PostgreSQL cluster.",
				MinItems:    1,
				Required:    true,
				Elem:        resourceYandexMDBPostgreSQLClusterHost(),
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
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
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
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
				Computed:    true,
			},
			"host_master_name": {
				Type:       schema.TypeString,
				Optional:   true,
				Computed:   true,
				Deprecated: "It sets name of master host. It works only when `host.name` is set. This field does not guarantee that a specific host will always be the master. We do not recommend using it. This functionality will be removed in future versions. If you are absolutely certain that you need this functionality, please contact technical support.",
			},
			"restore": {
				Type:        schema.TypeList,
				Description: "The cluster will be created from the specified backup.",
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Elem:        resourceYandexMDBPostgreSQLClusterRestoreBlock(),
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: "Maintenance policy of the PostgreSQL cluster.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        resourceYandexMDBPostgreSQLClusterMaintenanceWindow(),
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: "Host Group IDs.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceYandexMDBPostgreSQLClusterConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"version": {
				Type:        schema.TypeString,
				Description: "Version of the PostgreSQL cluster. (allowed versions are: 10, 10-1c, 11, 11-1c, 12, 12-1c, 13, 13-1c, 14, 14-1c, 15, 15-1c, 16, 17).",
				Required:    true,
			},
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the PostgreSQL cluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a PostgreSQL host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/concepts/instance-types).",
							Required:    true,
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Description: "Volume of the storage available to a PostgreSQL host, in gigabytes.",
							Required:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of PostgreSQL hosts.",
							Optional:    true,
						},
					},
				},
			},
			"autofailover": {
				Type:        schema.TypeBool,
				Description: "Configuration setting which enables/disables autofailover in cluster.",
				Optional:    true,
				Computed:    true,
			},
			"pooler_config": {
				Type:        schema.TypeList,
				Description: "Configuration of the connection pooler.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pooling_mode": {
							Type:        schema.TypeString,
							Description: "Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.",
							Optional:    true,
						},
						"pool_discard": {
							Type:        schema.TypeBool,
							Description: "Setting `pool_discard` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_discard-yesno).",
							Optional:    true,
						},
					},
				},
			},
			"backup_window_start": {
				Type:        schema.TypeList,
				Description: "Time to start the daily backup, in the UTC timezone.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:         schema.TypeInt,
							Description:  "The hour at which backup will be started (UTC).",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minutes": {
							Type:         schema.TypeInt,
							Description:  "The hour at which backup will be started (UTC).",
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
			"performance_diagnostics": {
				Type:        schema.TypeList,
				Description: "Cluster performance diagnostics settings. [YC Documentation](https://yandex.cloud/docs/managed-postgresql/api-ref/grpc/cluster_service#PerformanceDiagnostics).",
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
							Description: "Interval (in seconds) for pg_stat_activity sampling Acceptable values are 1 to 86400, inclusive.",
							Required:    true,
						},
						"statements_sampling_interval": {
							Type:        schema.TypeInt,
							Description: "Interval (in seconds) for pg_stat_statements sampling Acceptable values are 1 to 86400, inclusive.",
							Required:    true,
						},
					},
				},
			},
			"disk_size_autoscaling": {
				Type:        schema.TypeList,
				Description: "Cluster disk size autoscaling settings.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size_limit": {
							Type:        schema.TypeInt,
							Description: "Limit of disk size after autoscaling (GiB).",
							Required:    true,
						},
						"planned_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Maintenance window autoscaling disk usage (percent).",
							Optional:    true,
						},
						"emergency_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Immediate autoscaling disk usage (percent).",
							Optional:    true,
						},
					},
				},
			},
			"access": {
				Type:        schema.TypeList,
				Description: "Access policy to the PostgreSQL cluster.",
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
							Description: "Allow access for [SQL queries in the management console](https://yandex.cloud/docs/managed-postgresql/operations/web-sql-query).",
							Optional:    true,
							Computed:    true,
						},
						"serverless": {
							Type:        schema.TypeBool,
							Description: "Allow access for [connection to managed databases from functions](https://yandex.cloud/docs/functions/operations/database-connection).",
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
			"postgresql_config": {
				Type:             schema.TypeMap,
				Description:      "PostgreSQL cluster config. Detail info in `postresql config` section.",
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbPGSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceYandexMDBPostgreSQLClusterDatabaseBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lc_collate": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "C",
			},
			"lc_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "C",
			},
			"template_db": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"extension": {
				Type:     schema.TypeSet,
				Set:      pgExtensionHash,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexMDBPostgreSQLClusterUserBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"login": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"grants": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      pgUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"conn_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"settings": {
				Type:             schema.TypeMap,
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

func resourceYandexMDBPostgreSQLClusterHost() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Required:    true,
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
				Optional:    true,
			},
			"assign_public_ip": {
				Type:        schema.TypeBool,
				Description: "Sets whether the host should get a public IP address on creation. It can be changed on the fly only when `name` is set.",
				Optional:    true,
			},
			"fqdn": {
				Type:        schema.TypeString,
				Description: "The fully qualified domain name of the host.",
				Computed:    true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Host state name. It should be set for all hosts or unset for all hosts. This field can be used by another host, to select which host will be its replication source. Please see `replication_source_name` parameter.",
				Optional:    true,
			},
			"replication_source": {
				Type:        schema.TypeString,
				Description: "Host replication source (fqdn), when replication_source is empty then host is in HA group.",
				Computed:    true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: "Host priority in HA group. It works only when `name` is set.",
				Optional:    true,
				Deprecated:  "The field has not affected anything. You can safely delete it.",
			},
			"replication_source_name": {
				Type:        schema.TypeString,
				Description: "Host replication source name points to host's `name` from which this host should replicate. When not set then host in HA group. It works only when `name` is set.",
				Optional:    true,
			},
		},
	}
}

func resourceYandexMDBPostgreSQLClusterRestoreBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"backup_id": {
				Type:        schema.TypeString,
				Description: "Backup ID. The cluster will be created from the specified backup. [How to get a list of PostgreSQL backups](https://yandex.cloud/docs/managed-postgresql/operations/cluster-backups).",
				Required:    true,
				ForceNew:    true,
			},
			"time_inclusive": {
				Type:        schema.TypeBool,
				Description: "Flag that indicates whether a database should be restored to the first backup point available just after the timestamp specified in the [time] field instead of just before. Possible values:\n* `false` (default) — the restore point refers to the first backup moment before [time].\n* `true` — the restore point refers to the first backup point after [time].\n",
				Optional:    true,
				ForceNew:    true,
			},
			"time": {
				Type:         schema.TypeString,
				Description:  "Timestamp of the moment to which the PostgreSQL cluster should be restored. (Format: `2006-01-02T15:04:05` - UTC). When not set, current time is used.",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: stringToTimeValidateFunc,
			},
		},
	}
}

func resourceYandexMDBPostgreSQLClusterMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
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
				ValidateFunc: validation.IntBetween(1, 24),
				Description:  "Hour of the day in UTC (in `HH` format). Allowed value is between 1 and 24.",
				Optional:     true,
			},
		},
	}
}

func resourceYandexMDBPostgreSQLClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("folder_id", cluster.GetFolderId())
	d.Set("name", cluster.GetName())
	d.Set("description", cluster.GetDescription())
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("network_id", cluster.GetNetworkId())

	if err := d.Set("labels", cluster.GetLabels()); err != nil {
		return err
	}

	pgClusterConf, err := flattenPGClusterConfig(cluster.Config)
	if err != nil {
		return err
	}

	if err := d.Set("config", pgClusterConf); err != nil {
		return err
	}

	stateDatabases := d.Get("database").([]interface{})
	if len(stateDatabases) == 0 {
		if err := d.Set("database", []map[string]interface{}{}); err != nil {
			return err
		}
	} else {
		databases, err := listPGDatabases(ctx, config, d.Id())
		if err != nil {
			return err
		}

		databaseSpecs, err := expandPGDatabaseSpecs(d)
		if err != nil {
			return err
		}
		sortPGDatabases(databases, databaseSpecs)

		if err := d.Set("database", flattenPGDatabases(databases)); err != nil {
			return err
		}
	}

	stateUsers := d.Get("user").([]interface{})
	if len(stateUsers) == 0 {
		if err := d.Set("user", []map[string]interface{}{}); err != nil {
			return err
		}
	} else {
		userSpecs, err := expandPGUserSpecs(d)
		if err != nil {
			return err
		}
		passwords := pgUsersPasswords(userSpecs)
		users, err := listPGUsers(ctx, config, d.Id())
		if err != nil {
			return err
		}
		sortPGUsers(users, userSpecs)

		fUsers, err := flattenPGUsers(users, passwords, mdbPGUserSettingsFieldsInfo)
		if err != nil {
			return err
		}
		if err := d.Set("user", fUsers); err != nil {
			return err
		}
	}

	hosts, err := retryListPGHostsWrapper(ctx, config, d.Id())
	if err != nil {
		return err
	}

	orderedHostInfos, err := flattenPGHostsInfo(d, hosts)
	if err != nil {
		return err
	}

	fHosts := flattenPGHostsFromHostInfos(d, orderedHostInfos, false)
	masterHostname := getMasterHostname(orderedHostInfos)

	if err := d.Set("host", fHosts); err != nil {
		return err
	}
	if err := d.Set("host_master_name", masterHostname); err != nil {
		return err
	}

	maintenanceWindow, err := flattenPGMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}

	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	if err = d.Set("deletion_protection", cluster.DeletionProtection); err != nil {
		return err
	}

	if err = d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	return nil
}

func sortPGUsers(users []*postgresql.User, specs []*postgresql.UserSpec) {
	for i, spec := range specs {
		for j := i + 1; j < len(users); j++ {
			if spec.Name == users[j].Name {
				users[i], users[j] = users[j], users[i]
				break
			}
		}
	}
}

func sortPGDatabases(databases []*postgresql.Database, specs []*postgresql.DatabaseSpec) {
	for i, spec := range specs {
		for j := i + 1; j < len(databases); j++ {
			if spec.Name == databases[j].Name {
				databases[i], databases[j] = databases[j], databases[i]
				break
			}
		}
	}
}

func resourceYandexMDBPostgreSQLClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	request, err := prepareCreatePostgreSQLRequest(d, config)
	if err != nil {
		return err
	}

	if backupID, ok := d.GetOk("restore.0.backup_id"); ok && backupID != "" {
		return resourceYandexMDBPostgreSQLClusterRestore(d, meta, request, backupID.(string))
	}

	// This is a dirty hack to avoid the issue with the timeout of the create operation.
	// We are investigating the issue on the MDB side
	createTimeout := d.Timeout(schema.TimeoutCreate)
	if createTimeout < 5*time.Minute {
		createTimeout = 5 * time.Minute
	}
	ctx, cancel := config.ContextWithTimeout(createTimeout)
	defer cancel()

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster create request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().Create(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("Error while requesting API to create PostgreSQL Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get PostgreSQL Cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*postgresql.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get PostgreSQL Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create PostgreSQL Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("PostgreSQL Cluster creation failed: %s", err)
	}

	if err := createPGClusterHosts(ctx, config, d); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts creation failed: %s", d.Id(), err)
	}

	if err := updatePGClusterAfterCreate(d, meta); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v update params failed: %s", d.Id(), err)
	}

	if err := startPGFailoverIfNeed(d, meta); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts set master failed: %s", d.Id(), err)
	}

	return resourceYandexMDBPostgreSQLClusterRead(d, meta)
}

func resourceYandexMDBPostgreSQLClusterRestore(d *schema.ResourceData, meta interface{}, createClusterRequest *postgresql.CreateClusterRequest, backupID string) error {
	config := meta.(*Config)

	var timeBackup *timestamp.Timestamp = nil
	timeInclusive := false

	if backupTime, ok := d.GetOk("restore.0.time"); ok {
		time, err := parseStringToTime(backupTime.(string))
		if err != nil {
			return fmt.Errorf("Error while parsing restore.0.time to create PostgreSQL Cluster from backup %v, value: %v error: %s", backupID, backupTime, err)
		}
		timeBackup = &timestamp.Timestamp{
			Seconds: time.Unix(),
		}
	}

	if timeInclusiveData, ok := d.GetOk("restore.0.time_inclusive"); ok {
		timeInclusive = timeInclusiveData.(bool)
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()
	request := &postgresql.RestoreClusterRequest{
		BackupId:           backupID,
		Time:               timeBackup,
		TimeInclusive:      timeInclusive,
		Name:               createClusterRequest.Name,
		Description:        createClusterRequest.Description,
		Labels:             createClusterRequest.Labels,
		Environment:        createClusterRequest.Environment,
		ConfigSpec:         createClusterRequest.ConfigSpec,
		HostSpecs:          createClusterRequest.HostSpecs,
		NetworkId:          createClusterRequest.NetworkId,
		FolderId:           createClusterRequest.FolderId,
		SecurityGroupIds:   createClusterRequest.SecurityGroupIds,
		HostGroupIds:       createClusterRequest.HostGroupIds,
		DeletionProtection: createClusterRequest.DeletionProtection,
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster restore request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().Restore(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("Error while requesting API to create PostgreSQL Cluster from backup %v: %s", backupID, err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get PostgreSQL Cluster create from backup %v operation metadata: %s", backupID, err)
	}

	md, ok := protoMetadata.(*postgresql.RestoreClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get PostgreSQL Cluster ID from create from backup %v operation metadata", backupID)
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create PostgreSQL Cluster from backup %v: %s", backupID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("PostgreSQL Cluster creation from backup %v failed: %s", backupID, err)
	}

	if err := createPGClusterHosts(ctx, config, d); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts creation from backup %v failed: %s", d.Id(), backupID, err)
	}

	if err := startPGFailoverIfNeed(d, meta); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts set master failed: %s", d.Id(), err)
	}

	if err := updatePGClusterAfterCreate(d, meta); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v update params failed: %s", d.Id(), err)
	}

	return resourceYandexMDBPostgreSQLClusterRead(d, meta)
}

func prepareCreatePostgreSQLRequest(d *schema.ResourceData, meta *Config) (*postgresql.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on PostgreSQL Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating PostgreSQL Cluster: %s", err)
	}

	hostsFromScheme, err := expandPGHosts(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding host specs on PostgreSQL Cluster create: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parsePostgreSQLEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating PostgreSQL Cluster: %s", err)
	}

	confSpec, _, err := expandPGConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding cluster config on PostgreSQL Cluster create: %s", err)
	}

	confSpec.DiskSizeAutoscaling = nil

	userSpecs, err := expandPGUserSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding user specs on PostgreSQL Cluster create: %s", err)
	}

	databaseSpecs, err := expandPGDatabaseSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding database specs on PostgreSQL Cluster create: %s", err)
	}
	hostSpecs := make([]*postgresql.HostSpec, 0)
	for _, host := range hostsFromScheme {
		if host.HostSpec.ReplicationSource == "" {
			hostSpecs = append(hostSpecs, host.HostSpec)
		}
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))
	hostGroupIds := expandHostGroupIds(d.Get("host_group_ids"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on PostgreSQL Cluster create: %s", err)
	}

	return &postgresql.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Labels:             labels,
		Environment:        env,
		ConfigSpec:         confSpec,
		UserSpecs:          userSpecs,
		DatabaseSpecs:      databaseSpecs,
		HostSpecs:          hostSpecs,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
		HostGroupIds:       hostGroupIds,
	}, nil
}

func updatePGClusterAfterCreate(d *schema.ResourceData, meta interface{}) error {
	maintenanceWindow, err := expandPGMaintenanceWindow(d)
	if err != nil {
		return fmt.Errorf("error expanding maintenance_window while updating PostgreSQL after creation: %s", err)
	}

	paths := make([]string, 0)
	if maintenanceWindow != nil {
		paths = append(paths, "maintenance_window")
	}

	configSpec := &postgresql.ConfigSpec{}
	diskSizeAutoscaling := expandPGDiskSizeAutoscaling(d)
	if diskSizeAutoscaling != nil {
		paths = append(paths, "config_spec.disk_size_autoscaling")
		configSpec.DiskSizeAutoscaling = diskSizeAutoscaling
	}

	if len(paths) < 1 {
		return nil
	}

	request := &postgresql.UpdateClusterRequest{
		ClusterId:         d.Id(),
		MaintenanceWindow: maintenanceWindow,
		ConfigSpec:        configSpec,
		UpdateMask:        &field_mask.FieldMask{Paths: paths},
	}

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster update request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().Update(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to update PostgreSQL Cluster after creation %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to update PostgreSQL Cluster after creation %q: %s", d.Id(), err)
	}

	return nil
}

func resourceYandexMDBPostgreSQLClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if err := setPGFolderID(d, meta); err != nil {
		return err
	}

	if err := updatePGClusterParams(d, meta); err != nil {
		return err
	}

	stateUser := d.Get("user").([]interface{})
	if d.HasChange("user") && len(stateUser) > 0 {
		if err := updatePGClusterUsersAdd(d, meta); err != nil {
			return err
		}
	}

	stateDatabase := d.Get("database").([]interface{})
	if d.HasChange("database") && len(stateDatabase) > 0 {
		if err := updatePGClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") && len(stateUser) > 0 {
		if err := updatePGClusterUsersUpdateAndDrop(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updatePGClusterHosts(d, meta); err != nil {
			return err
		}
	}

	if err := startPGFailoverIfNeed(d, meta); err != nil {
		return err
	}

	d.Partial(false)

	return resourceYandexMDBPostgreSQLClusterRead(d, meta)
}

func updatePGClusterParams(d *schema.ResourceData, meta interface{}) error {
	log.Println("[DEBUG] updatePGClusterParams")
	config := meta.(*Config)
	request, err := prepareUpdatePostgreSQLClusterParamsRequest(d, config)
	if err != nil {
		return err
	}

	if len(request.UpdateMask.Paths) == 0 {
		return nil
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster update request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().Update(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to update PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting for operation to update PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func prepareUpdatePostgreSQLClusterParamsRequest(d *schema.ResourceData, config *Config) (request *postgresql.UpdateClusterRequest, err error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating PostgreSQL Cluster: %s", err)
	}

	configSpec, settingNames, err := expandPGConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding config while updating PostgreSQL Cluster: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))
	if d.HasChange("host_group_ids") {
		return nil, fmt.Errorf("host_group_ids change is not supported yet")
	}

	maintenanceWindow, err := expandPGMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding maintenance_window while updating PostgreSQL cluster: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, config)
	if err != nil {
		return nil, fmt.Errorf("error expanding network_id while updating PostgreSQL cluster: %s", err)
	}

	return &postgresql.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		NetworkId:          networkID,
		ConfigSpec:         configSpec,
		MaintenanceWindow:  maintenanceWindow,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
		UpdateMask:         &field_mask.FieldMask{Paths: expandPGParamsUpdatePath(d, settingNames)},
	}, nil
}

func updatePGClusterDatabases(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currDBs, err := listPGDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandPGDatabaseSpecs(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := pgDatabasesDiff(currDBs, targetDBs)

	for _, dbn := range toDelete {
		err := deletePGDatabase(ctx, config, d, dbn)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createPGDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("database")

	changedDatabases, err := pgChangedDatabases(oldSpecs.([]interface{}), newSpecs.([]interface{}))
	if err != nil {
		return err
	}

	dDatabase := make(map[string]string)
	cnt := d.Get("database.#").(int)
	for i := 0; i < cnt; i++ {
		dDatabase[d.Get(fmt.Sprintf("database.%v.name", i)).(string)] = fmt.Sprintf("database.%v.", i)
	}

	for _, u := range changedDatabases {
		err := updatePGDatabase(ctx, config, d, u, dDatabase[u.Name])
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePGClusterUsersAdd(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currUsers, err := listPGUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	usersForCreate, err := pgUserForCreate(d, currUsers)
	if err != nil {
		return err
	}
	for _, u := range usersForCreate {
		err := createPGUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePGClusterUsersUpdateAndDrop(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currUsers, err := listPGUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	dUser := make(map[string]string)
	cnt := d.Get("user.#").(int)
	for i := 0; i < cnt; i++ {
		dUser[d.Get(fmt.Sprintf("user.%v.name", i)).(string)] = fmt.Sprintf("user.%v.", i)
	}

	deleteNames := make([]string, 0)

	for _, v := range currUsers {
		path, ok := dUser[v.Name]
		if !ok {
			deleteNames = append(deleteNames, v.Name)
		} else {
			err := updatePGUser(ctx, config, d, v, path)
			if err != nil {
				return err
			}
		}
	}

	for _, u := range deleteNames {
		err := deletePGUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePGClusterHosts(d *schema.ResourceData, meta interface{}) error {
	// Ideas:
	// 1. In order to do it safely for clients: firstly add new hosts and only then delete unneeded hosts
	// 2. Batch Add/Update operations are not supported, so we should update hosts one by one
	//    It may produce issues with cascade replicas: we should change replication-source in such way, that
	//    there is no attempts to create replication loop
	//    Solution: update HA-replicas first, then use BFS (using `comparePGHostsInfoResult.hierarchyExists`)

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// Step 1: Add new hosts (as HA-hosts):
	err := createPGClusterHosts(ctx, config, d)
	if err != nil {
		return err
	}

	// Step 2: update hosts:
	currHosts, err := retryListPGHostsWrapper(ctx, config, d.Id())
	if err != nil {
		return err
	}

	compareHostsInfo, err := comparePGHostsInfo(d, currHosts, true)
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
			if len(maskPaths) > 0 {
				if err := updatePGHost(ctx, config, d, &postgresql.UpdateHostSpec{
					HostName:          hostInfo.fqdn,
					ReplicationSource: hostInfo.newReplicationSource,
					AssignPublicIp:    hostInfo.newAssignPublicIP,
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
			if err := deletePGHost(ctx, config, d, hostInfo.fqdn); err != nil {
				return err
			}
		}
	}

	return nil
}

func createPGClusterHosts(ctx context.Context, config *Config, d *schema.ResourceData) error {
	hosts, err := retryListPGHostsWrapper(ctx, config, d.Id())
	if err != nil {
		return err
	}
	compareHostsInfo, err := comparePGHostsInfo(d, hosts, true)
	if err != nil {
		return err
	}

	if compareHostsInfo.hierarchyExists && len(compareHostsInfo.createHostsInfo) == 0 {
		return fmt.Errorf("Create cluster hosts error. Exists host with replication source, which can't be created. Possibly there is a loop")
	}

	for _, newHostInfo := range compareHostsInfo.createHostsInfo {
		host := &postgresql.HostSpec{
			ZoneId:         newHostInfo.zone,
			SubnetId:       newHostInfo.subnetID,
			AssignPublicIp: newHostInfo.newAssignPublicIP,
		}
		if compareHostsInfo.haveHostWithName {
			host.ReplicationSource = newHostInfo.newReplicationSource
		}
		if err := addPGHost(ctx, config, d, host); err != nil {
			return err
		}
	}
	if compareHostsInfo.hierarchyExists {
		return createPGClusterHosts(ctx, config, d)
	}

	return nil
}

func startPGFailoverIfNeed(d *schema.ResourceData, meta interface{}) error {
	rawHostMasterName, ok := d.GetOk("host_master_name")
	if !ok {
		return nil
	}
	hostMasterName := rawHostMasterName.(string)

	log.Printf("[DEBUG] startPGFailoverIfNeed")
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currHosts, err := retryListPGHostsWrapper(ctx, config, d.Id())
	if err != nil {
		return err
	}
	compareHostsInfo, err := comparePGHostsInfo(d, currHosts, true)
	if err != nil {
		return err
	}

	if !compareHostsInfo.haveHostWithName {
		return nil
	}

	log.Printf("[DEBUG] hostMasterName: %+v", hostMasterName)
	for _, hostInfo := range compareHostsInfo.hostsInfo {
		log.Printf("[DEBUG] hostInfox: %+v", hostInfo)
		if hostMasterName == hostInfo.name && hostInfo.role != postgresql.Host_MASTER {
			if err := startPGFailover(ctx, config, d, hostInfo.fqdn); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func resourceYandexMDBPostgreSQLClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting PostgreSQL Cluster %q", d.Id())

	request := &postgresql.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster delete request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().Delete(ctx, request)
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("PostgreSQL Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting PostgreSQL Cluster %q", d.Id())

	return nil
}

func createPGUser(ctx context.Context, config *Config, d *schema.ResourceData, user *postgresql.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().User().Create(ctx, &postgresql.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating user for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func updatePGUser(ctx context.Context, config *Config, d *schema.ResourceData, user *postgresql.User, path string) (err error) {

	us, err := expandPGUser(d, &postgresql.UserSpec{
		Name:        user.Name,
		Permissions: user.Permissions,
		ConnLimit:   &wrappers.Int64Value{Value: user.ConnLimit},
		Settings:    user.Settings,
		Login:       user.Login,
		Grants:      user.Grants,
	}, path)
	if err != nil {
		return err
	}

	changeMask := map[string]string{
		"password":   "password",
		"permission": "permissions",
		"login":      "login",
		"grants":     "grants",
		"conn_limit": "conn_limit",
		"settings":   "settings",
	}

	updatePath := []string{}
	onDone := make([]func(), 0)

	for field, mask := range changeMask {
		if d.HasChange(path + field) {
			updatePath = append(updatePath, mask)
			onDone = append(onDone, func() {

			})
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().User().Update(ctx, &postgresql.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    us.Name,
			Password:    us.Password,
			Permissions: us.Permissions,
			ConnLimit:   us.ConnLimit.GetValue(),
			Login:       us.Login,
			Grants:      us.Grants,
			Settings:    us.Settings,
			UpdateMask:  &field_mask.FieldMask{Paths: updatePath},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating user for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}
	return nil
}

func deletePGUser(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().User().Delete(ctx, &postgresql.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  name,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting user from PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func listPGUsers(ctx context.Context, config *Config, id string) ([]*postgresql.User, error) {
	users := []*postgresql.User{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().PostgreSQL().User().List(ctx, &postgresql.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of users for PostgreSQL Cluster '%q': %s", id, err)
		}

		users = append(users, resp.Users...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return users, nil
}

func createPGDatabase(ctx context.Context, config *Config, d *schema.ResourceData, db *postgresql.DatabaseSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Database().Create(ctx, &postgresql.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &postgresql.DatabaseSpec{
				Name:       db.Name,
				Owner:      db.Owner,
				LcCollate:  db.LcCollate,
				LcCtype:    db.LcCtype,
				TemplateDb: db.TemplateDb,
				Extensions: db.Extensions,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating database for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func updatePGDatabase(ctx context.Context, config *Config, d *schema.ResourceData, db *postgresql.DatabaseSpec, path string) error {
	changeMask := map[string]string{
		"extension": "extensions",
	}

	updatePath := []string{}
	for field, mask := range changeMask {
		if d.HasChange(path + field) {
			updatePath = append(updatePath, mask)
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	// Deletion protection and dbname changing is not supported on purpose
	// User should use separate resources for that
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Database().Update(ctx, &postgresql.UpdateDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: db.Name,
			Extensions:   db.Extensions,
			UpdateMask:   &field_mask.FieldMask{Paths: updatePath},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update database in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating database in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating database for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func deletePGDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Database().Delete(ctx, &postgresql.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting database from PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func listPGDatabases(ctx context.Context, config *Config, id string) ([]*postgresql.Database, error) {
	databases := []*postgresql.Database{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().PostgreSQL().Database().List(ctx, &postgresql.ListDatabasesRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of databases for PostgreSQL Cluster '%q': %s", id, err)
		}

		databases = append(databases, resp.Databases...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return databases, nil
}

func addPGHost(ctx context.Context, config *Config, d *schema.ResourceData, host *postgresql.HostSpec) error {
	request := &postgresql.AddClusterHostsRequest{
		ClusterId: d.Id(),
		HostSpecs: []*postgresql.HostSpec{host},
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster add hosts request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().AddHosts(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to create host for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating host for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating host for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func deletePGHost(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	request := &postgresql.DeleteClusterHostsRequest{
		ClusterId: d.Id(),
		HostNames: []string{name},
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster delete hosts request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().DeleteHosts(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting host from PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func startPGFailover(ctx context.Context, config *Config, d *schema.ResourceData, hostName string) error {
	request := &postgresql.StartClusterFailoverRequest{
		ClusterId: d.Id(),
		HostName:  hostName,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster start failover request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().StartFailover(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to start failover host in PostgreSQL Cluster %q - host %v: %s", d.Id(), hostName, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while start failover host in PostgreSQL Cluster %q - host %v: %s", d.Id(), hostName, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("start failover host in PostgreSQL Cluster %q - host %v failed: %s", d.Id(), hostName, err)
	}

	return nil
}

func updatePGHost(ctx context.Context, config *Config, d *schema.ResourceData, host *postgresql.UpdateHostSpec) error {
	request := &postgresql.UpdateClusterHostsRequest{
		ClusterId:       d.Id(),
		UpdateHostSpecs: []*postgresql.UpdateHostSpec{host},
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL cluster update hosts request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Cluster().UpdateHosts(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to update host for PostgreSQL Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating host for PostgreSQL Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating host for PostgreSQL Cluster %q - host %v failed: %s", d.Id(), host.HostName, err)
	}

	return nil
}

func retryListPGHosts(ctx context.Context, config *Config, id string, attempt int, maxAttempt int, condition func([]*postgresql.Host) bool) ([]*postgresql.Host, error) {
	log.Printf("[DEBUG] Try ListPGHosts, attempt: %d", attempt)
	hosts, err := func(ctx context.Context, config *Config, id string) ([]*postgresql.Host, error) {
		hosts := []*postgresql.Host{}
		pageToken := ""

		for {
			request := &postgresql.ListClusterHostsRequest{
				ClusterId: id,
				PageSize:  defaultMDBPageSize,
				PageToken: pageToken,
			}
			resp, err := config.sdk.MDB().PostgreSQL().Cluster().ListHosts(ctx, request)
			log.Printf("[DEBUG] Sending PostgreSQL cluster list hosts request: %+v", request)
			if err != nil {
				return nil, fmt.Errorf("Error while getting list of hosts for PostgreSQL Cluster '%q': %s", id, err)
			}

			hosts = append(hosts, resp.Hosts...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}

		return hosts, nil
	}(ctx, config, id)
	if condition(hosts) || maxAttempt <= attempt {
		return hosts, err // We tried to do our best
	}

	timeout := int(math.Pow(2, float64(attempt)))
	log.Printf("[DEBUG] Condition failed, waiting %ds before the next attempt", timeout)
	time.Sleep(time.Second * time.Duration(timeout))

	return retryListPGHosts(ctx, config, id, attempt+1, maxAttempt, condition)
}

// retry with 1, 2, 4, 8, 16, 32, 64, 128 seconds if no succeess
// while at least one host is unknown and there is no master
func retryListPGHostsWrapper(ctx context.Context, config *Config, id string) ([]*postgresql.Host, error) {
	attempts := 7
	return retryListPGHosts(ctx, config, id, 0, attempts, func(hosts []*postgresql.Host) bool {
		masterExists := false
		for _, host := range hosts {
			// Check that every host has a role
			if host.Role == postgresql.Host_ROLE_UNKNOWN {
				return false
			}
			// And one of them is master
			if host.Role == postgresql.Host_MASTER {
				masterExists = true
			}
		}
		return masterExists
	})
}

func setPGFolderID(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
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
		request := &postgresql.MoveClusterRequest{
			ClusterId:           d.Id(),
			DestinationFolderId: folderID.(string),
		}
		op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending PostgreSQL cluster move request: %+v", request)
			return config.sdk.MDB().PostgreSQL().Cluster().Move(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("error while requesting API to move PostgreSQL Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while moving PostgreSQL Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("moving PostgreSQL Cluster %q to folder %v failed: %s", d.Id(), folderID, err)
		}

	}

	return nil
}
