package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexMDBGreenplumClusterDefaultTimeout = 120 * time.Minute
	yandexMDBGreenplumClusterUpdateTimeout  = 120 * time.Minute
	yandexMDBGreenplumClusterExpandTimeout  = 24 * 60 * time.Minute
	yandexMDBGreenplumClusterExpandDuration = 7200 // in seconds
)

func resourceYandexMDBGreenplumCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Greenplum cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-greenplum/).\n\nPlease read [Pricing for Managed Service for Greenplum](https://yandex.cloud/docs/managed-greenplum/) before using Greenplum cluster.\n",

		Create: resourceYandexMDBGreenplumClusterCreate,
		Read:   resourceYandexMDBGreenplumClusterRead,
		Update: resourceYandexMDBGreenplumClusterUpdate,
		Delete: resourceYandexMDBGreenplumClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBGreenplumClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBGreenplumClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBGreenplumClusterDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the Greenplum cluster. (PRODUCTION, PRESTABLE)",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseGreenplumEnv),
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
				ForceNew:    true,
			},
			"zone": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["zone"],
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Description:  "The ID of the subnet, to which the hosts belongs. The subnet must be a part of the network to which the cluster belongs.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"assign_public_ip": {
				Type:        schema.TypeBool,
				Description: "Sets whether the master hosts should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.",
				Required:    true,
			},
			"version": {
				Type:         schema.TypeString,
				Description:  "Version of the Greenplum cluster. (`6.28`)",
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"6.28", "6.25"}, true),
			},
			"master_host_count": {
				Type:        schema.TypeInt,
				Description: "Number of hosts in master subcluster (1 or 2).",
				Required:    true,
			},
			"segment_host_count": {
				Type:        schema.TypeInt,
				Description: "Number of hosts in segment subcluster (from 1 to 32).",
				Required:    true,
			},
			"segment_in_host": {
				Type:        schema.TypeInt,
				Description: "Number of segments on segment host (not more then 1 + RAM/8).",
				Required:    true,
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
			"master_subcluster": {
				Type:        schema.TypeList,
				Description: "Settings for master subcluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:        schema.TypeList,
							Description: "Resources allocated to hosts for master subcluster of the Greenplum cluster.",
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/ru/docs/managed-greenplum/concepts/instance-types).",
									},
									"disk_type_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of the storage of Greenplum hosts - environment default is used if missing.",
									},
									"disk_size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Volume of the storage available to a host, in gigabytes.",
									},
								},
							},
						},
					},
				},
			},
			"segment_subcluster": {
				Type:        schema.TypeList,
				Description: "Settings for segment subcluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:        schema.TypeList,
							Description: "Resources allocated to hosts for segment subcluster of the Greenplum cluster.",
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/ru/docs/managed-greenplum/concepts/instance-types).",
									},
									"disk_type_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of the storage of Greenplum hosts - environment default is used if missing.",
									},
									"disk_size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Volume of the storage available to a host, in gigabytes.",
									},
								},
							},
						},
					},
				},
			},
			"master_hosts": {
				Type:        schema.TypeList,
				Description: "Info about hosts in master subcluster.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Description: "Flag indicating that master hosts should be created with a public IP address.",
							Computed:    true,
						},
						"fqdn": {
							Type:        schema.TypeString,
							Description: "The fully qualified domain name of the host.",
							Computed:    true,
						},
					},
				},
			},
			"segment_hosts": {
				Type:        schema.TypeList,
				Description: "Info about hosts in segment subcluster.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fqdn": {
							Type:        schema.TypeString,
							Description: "The fully qualified domain name of the host.",
							Computed:    true,
						},
					},
				},
			},
			"user_name": {
				Type:        schema.TypeString,
				Description: "Greenplum cluster admin user name.",
				Required:    true,
			},
			"user_password": {
				Type:        schema.TypeString,
				Description: "Greenplum cluster admin password name.",
				Required:    true,
				Sensitive:   true,
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
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: "Maintenance policy of the Greenplum cluster.",
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
							Description:  "Day of the week (in `DDD` format). Allowed values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
							ValidateFunc: mdbMaintenanceWindowSchemaValidateFunc,
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							Description:  "Hour of the day in UTC (in `HH` format). Allowed value is between 0 and 23.",
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
			"service_account_id": {
				Type:        schema.TypeString,
				Description: "ID of service account to use with Yandex Cloud resources (e.g. S3, Cloud Logging).",
				Optional:    true,
			},
			"logging": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Cloud Logging settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Flag that indicates whether log delivery to Cloud Logging is enabled.",
						},
						"log_group_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"logging.0.folder_id"},
							Description:   "Cloud Logging group ID to send logs to.",
						},
						"folder_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"logging.0.log_group_id"},
							Description:   "ID of folder to which deliver logs.",
						},
						"command_center_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Deliver Yandex Command Center's logs to Cloud Logging.",
						},
						"greenplum_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Deliver Greenplum's logs to Cloud Logging.",
						},
						"pooler_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Deliver connection pooler's logs to Cloud Logging.",
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
							Description:  "The minute at which backup will be started (UTC).",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},
			"access": {
				Type:        schema.TypeList,
				Description: "Access policy to the Greenplum cluster.",
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
							Description: "Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer)",
							Optional:    true,
							Default:     false,
						},
						"yandex_query": {
							Type:        schema.TypeBool,
							Description: "Allow access for [Yandex Query](https://yandex.cloud/services/query)",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"pooler_config": {
				Type:        schema.TypeList,
				Description: "Configuration of the connection pooler.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pooling_mode": {
							Type:        schema.TypeString,
							Description: "Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/docs/configuration/rules.md#pool).",
							Optional:    true,
							Default:     "POOL_MODE_UNSPECIFIED",
						},
						"pool_size": {
							Type:        schema.TypeInt,
							Description: "Value for `pool_size` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/docs/configuration/rules.md#pool_size).",
							Optional:    true,
							Default:     nil,
						},
						"pool_client_idle_timeout": {
							Type:        schema.TypeInt,
							Description: "Value for `pool_client_idle_timeout` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/docs/configuration/rules.md#pool_client_idle_timeout).",
							Optional:    true,
							Default:     nil,
						},
						"pool_idle_in_transaction_timeout": {
							Type:        schema.TypeInt,
							Description: "Value for `pool_idle_in_transaction_timeout` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/docs/configuration/rules.md#pool_idle_in_transaction_timeout).",
							Optional:    true,
							Default:     nil,
						},
					},
				},
			},
			"pxf_config": {
				Type:        schema.TypeList,
				Description: "Configuration of the PXF daemon.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_timeout": {
							Type:        schema.TypeInt,
							Description: "The Tomcat server connection timeout for read operations in seconds. Value is between 5 and 600.",
							Optional:    true,
							Default:     nil,
						},
						"upload_timeout": {
							Type:        schema.TypeInt,
							Description: "The Tomcat server connection timeout for write operations in seconds. Value is between 5 and 600.",
							Optional:    true,
							Default:     nil,
						},
						"max_threads": {
							Type:        schema.TypeInt,
							Description: "The maximum number of PXF tomcat threads. Value is between 1 and 1024.",
							Optional:    true,
							Default:     nil,
						},
						"pool_allow_core_thread_timeout": {
							Type:        schema.TypeBool,
							Description: "Identifies whether or not core streaming threads are allowed to time out.",
							Optional:    true,
							Default:     nil,
						},
						"pool_core_size": {
							Type:        schema.TypeInt,
							Description: "The number of core streaming threads. Value is between 1 and 1024.",
							Optional:    true,
							Default:     nil,
						},
						"pool_queue_capacity": {
							Type:        schema.TypeInt,
							Description: "The capacity of the core streaming thread pool queue. Value is positive.",
							Optional:    true,
							Default:     nil,
						},
						"pool_max_size": {
							Type:        schema.TypeInt,
							Description: "The maximum allowed number of core streaming threads. Value is between 1 and 1024.",
							Optional:    true,
							Default:     nil,
						},
						"xmx": {
							Type:        schema.TypeInt,
							Description: "Initial JVM heap size for PXF daemon. Value is between 64 and 16384.",
							Optional:    true,
							Default:     nil,
						},
						"xms": {
							Type:        schema.TypeInt,
							Description: "Maximum JVM heap size for PXF daemon. Value is between 64 and 16384.",
							Optional:    true,
							Default:     nil,
						},
					},
				},
			},
			"greenplum_config": {
				Type:             schema.TypeMap,
				Description:      "Greenplum cluster config. Detail info in `Greenplum cluster settings` block.",
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbGreenplumSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbGreenplumSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cloud_storage": {
				Type:        schema.TypeList,
				Description: "Cloud Storage settings of the Greenplum cluster.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:        schema.TypeBool,
							Description: "Whether to use cloud storage or not.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"master_host_group_ids": {
				Type:        schema.TypeSet,
				Description: "A list of IDs of the host groups to place master subclusters' VMs of the cluster on.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"segment_host_group_ids": {
				Type:        schema.TypeSet,
				Description: "A list of IDs of the host groups to place segment subclusters' VMs of the cluster on.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"background_activities": {
				Type:        schema.TypeList,
				Description: "Background activities settings.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"analyze_and_vacuum": {
							Type:        schema.TypeList,
							Description: "Block to configure 'ANALYZE' and 'VACUUM' daily operations.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_time": {
										Type:        schema.TypeString,
										Description: "Time of day in 'HH:MM' format when scripts should run.",
										Optional:    true,
									},
									"analyze_timeout": {
										Type:        schema.TypeInt,
										Description: "Maximum duration of the `ANALYZE` operation, in seconds. The default value is `36000`. As soon as this period expires, the `ANALYZE` operation will be forced to terminate.",
										Optional:    true,
									},
									"vacuum_timeout": {
										Type:        schema.TypeInt,
										Description: "Maximum duration of the `VACUUM` operation, in seconds. The default value is `36000`. As soon as this period expires, the `VACUUM` operation will be forced to terminate.",
										Optional:    true,
									},
								},
							},
						},
						"query_killer_idle": {
							Type:        schema.TypeList,
							Description: "Block to configure script that kills long running queries that are in `idle` state.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:        schema.TypeBool,
										Description: "Flag that indicates whether script is enabled.",
										Optional:    true,
									},
									"max_age": {
										Type:        schema.TypeInt,
										Description: "Maximum duration for this type of queries (in seconds).",
										Optional:    true,
									},
									"ignore_users": {
										Type:        schema.TypeList,
										Description: "List of users to ignore when considering queries to terminate.",
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"query_killer_idle_in_transaction": {
							Type:        schema.TypeList,
							Description: "Block to configure script that kills long running queries that are in `idle in transaction` state.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:        schema.TypeBool,
										Description: "Flag that indicates whether script is enabled.",
										Optional:    true,
									},
									"max_age": {
										Type:        schema.TypeInt,
										Description: "Maximum duration for this type of queries (in seconds).",
										Optional:    true,
									},
									"ignore_users": {
										Type:        schema.TypeList,
										Description: "List of users to ignore when considering queries to terminate.",
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"query_killer_long_running": {
							Type:        schema.TypeList,
							Description: "Block to configure script that kills long running queries (in any state).",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:        schema.TypeBool,
										Description: "Flag that indicates whether script is enabled.",
										Optional:    true,
									},
									"max_age": {
										Type:        schema.TypeInt,
										Description: "Maximum duration for this type of queries (in seconds).",
										Optional:    true,
									},
									"ignore_users": {
										Type:        schema.TypeList,
										Description: "List of users to ignore when considering queries to terminate.",
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceYandexMDBGreenplumClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := prepareCreateGreenplumClusterRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.MDB().Greenplum().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Greenplum Cluster: %s", err)
	}
	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while get Greenplum create operation metadata: %s", err)
	}
	md, ok := protoMetadata.(*greenplum.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Greenplum Cluster ID from create operation metadata")
	}
	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create Greenplum Cluster: %s", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("failed to create Greenplum Cluster: %s", err)
	}
	return resourceYandexMDBGreenplumClusterRead(d, meta)
}

func prepareCreateGreenplumClusterRequest(d *schema.ResourceData, meta *Config) (*greenplum.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Greenplum Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting folder ID while creating Greenplum Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseGreenplumEnv(e)
	if err != nil {
		return nil, fmt.Errorf("error resolving environment while creating Greenplum Cluster: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error while expanding network id on Greenplum Cluster create: %s", err)
	}

	configSpec, _, err := expandGreenplumConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding config spec on Greenplum Cluster create: %s", err)
	}

	maintenanceWindow, err := expandGreenplumMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding maintenance_window on Greenplum Cluster create: %s", err)
	}

	return &greenplum.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Environment:        env,
		Labels:             labels,
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		DeletionProtection: d.Get("deletion_protection").(bool),
		MaintenanceWindow:  maintenanceWindow,
		ServiceAccountId:   d.Get("service_account_id").(string),
		Logging:            expandGreenplumLogging(d),

		MasterHostCount:  int64(d.Get("master_host_count").(int)),
		SegmentInHost:    int64(d.Get("segment_in_host").(int)),
		SegmentHostCount: int64(d.Get("segment_host_count").(int)),

		Config: &greenplum.GreenplumConfig{
			Version:           d.Get("version").(string),
			BackupWindowStart: expandGreenplumBackupWindowStart(d),
			Access:            expandGreenplumAccess(d),
			ZoneId:            d.Get("zone").(string),
			SubnetId:          d.Get("subnet_id").(string),
			AssignPublicIp:    d.Get("assign_public_ip").(bool),
		},
		MasterConfig: &greenplum.MasterSubclusterConfigSpec{
			Resources: &greenplum.Resources{
				ResourcePresetId: d.Get("master_subcluster.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("master_subcluster.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("master_subcluster.0.resources.0.disk_size").(int)),
			},
		},
		SegmentConfig: &greenplum.SegmentSubclusterConfigSpec{
			Resources: &greenplum.Resources{
				ResourcePresetId: d.Get("segment_subcluster.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("segment_subcluster.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("segment_subcluster.0.resources.0.disk_size").(int)),
			},
		},

		UserName:     d.Get("user_name").(string),
		UserPassword: d.Get("user_password").(string),

		ConfigSpec:          configSpec,
		CloudStorage:        expandGreenplumCloudStorage(d),
		MasterHostGroupIds:  expandHostGroupIds(d.Get("master_host_group_ids")),
		SegmentHostGroupIds: expandHostGroupIds(d.Get("segment_host_group_ids")),
	}, nil
}

func resourceYandexMDBGreenplumClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Greenplum().Cluster().Get(ctx, &greenplum.GetClusterRequest{
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
	d.Set("deletion_protection", cluster.DeletionProtection)

	d.Set("zone", cluster.GetConfig().ZoneId)
	d.Set("subnet_id", cluster.GetConfig().SubnetId)
	d.Set("assign_public_ip", cluster.GetConfig().AssignPublicIp)
	d.Set("version", cluster.GetConfig().Version)
	d.Set("service_account_id", cluster.ServiceAccountId)

	d.Set("master_host_count", cluster.GetMasterHostCount())
	d.Set("segment_host_count", cluster.GetSegmentHostCount())
	d.Set("segment_in_host", cluster.GetSegmentInHost())

	d.Set("user_name", cluster.GetUserName())

	d.Set("master_subcluster", flattenGreenplumMasterSubcluster(cluster.GetMasterConfig().Resources))
	d.Set("segment_subcluster", flattenGreenplumSegmentSubcluster(cluster.GetSegmentConfig().Resources))
	d.Set("logging", flattenGreenplumLogging(cluster.GetLogging()))

	poolConfig, err := flattenGreenplumPoolerConfig(cluster.GetClusterConfig().GetPool())
	if err != nil {
		return err
	}
	if err := d.Set("pooler_config", poolConfig); err != nil {
		return err
	}

	pxfConfig, err := flattenGreenplumPXFConfig(cluster.GetClusterConfig().GetPxfConfig())
	if err != nil {
		return err
	}
	if err := d.Set("pxf_config", pxfConfig); err != nil {
		return err
	}

	gpConfig, err := flattenGreenplumClusterConfig(cluster.ClusterConfig)
	if err != nil {
		return err
	}
	if err := d.Set("greenplum_config", gpConfig); err != nil {
		return err
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	masterHosts, err := listGreenplumMasterHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	segmentHosts, err := listGreenplumSegmentHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	mHost, sHost := flattenGreenplumHosts(masterHosts, segmentHosts)
	if err := d.Set("master_hosts", mHost); err != nil {
		return err
	}
	if err := d.Set("segment_hosts", sHost); err != nil {
		return err
	}

	if err := d.Set("access", flattenGreenplumAccess(cluster.Config)); err != nil {
		return err
	}

	if err := d.Set("cloud_storage", flattenGreenplumCloudStorage(cluster.CloudStorage)); err != nil {
		return err
	}
	d.Set("master_host_group_ids", cluster.GetMasterHostGroupIds())
	d.Set("segment_host_group_ids", cluster.GetSegmentHostGroupIds())

	backgroundActivities, err := flattenGreenplumBackgroundActivities(cluster.ClusterConfig.BackgroundActivities)
	if err != nil {
		return err
	}
	if err := d.Set("background_activities", backgroundActivities); err != nil {
		return err
	}

	maintenanceWindow, err := flattenGreenplumMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}

	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
		return err
	}

	if err := d.Set("backup_window_start", flattenMDBBackupWindowStart(cluster.Config.BackupWindowStart)); err != nil {
		return err
	}

	if err := d.Set("created_at", getTimestamp(cluster.CreatedAt)); err != nil {
		return err
	}

	return nil
}

func listGreenplumMasterHosts(ctx context.Context, config *Config, id string) ([]*greenplum.Host, error) {
	hosts := []*greenplum.Host{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().Greenplum().Cluster().ListMasterHosts(ctx, &greenplum.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", id, err)
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts, nil
}

func listGreenplumSegmentHosts(ctx context.Context, config *Config, id string) ([]*greenplum.Host, error) {
	hosts := []*greenplum.Host{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().Greenplum().Cluster().ListSegmentHosts(ctx, &greenplum.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", id, err)
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts, nil
}

func resourceYandexMDBGreenplumClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	config := meta.(*Config)

	reqExpand, err := prepareExpandGreenplumClusterRequest(d)
	if err != nil {
		return err
	}

	req, err := prepareUpdateGreenplumClusterRequest(d, config)
	if err != nil {
		return err
	}

	if len(req.UpdateMask.Paths) != 0 {

		ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
		defer cancel()

		op, err := config.sdk.WrapOperation(config.sdk.MDB().Greenplum().Cluster().Update(ctx, req))
		if err != nil {
			return fmt.Errorf("error while requesting API to update Greenplum Cluster %q: %s", d.Id(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while updating Greenplum Cluster %q: %s", d.Id(), err)
		}
	}

	if reqExpand.AddSegmentsPerHostCount != 0 || reqExpand.SegmentHostCount != 0 {

		ctx, cancelExp := config.ContextWithTimeout(yandexMDBGreenplumClusterExpandTimeout)
		defer cancelExp()

		op, err := config.sdk.WrapOperation(config.sdk.MDB().Greenplum().Cluster().Expand(ctx, reqExpand))
		if err != nil {
			return fmt.Errorf("error while requesting API to expand Greenplum Cluster %q: %s", d.Id(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while expanding Greenplum Cluster %q: %s", d.Id(), err)
		}
	}

	d.Partial(false)

	return resourceYandexMDBGreenplumClusterRead(d, meta)
}

func prepareExpandGreenplumClusterRequest(d *schema.ResourceData) (*greenplum.ExpandRequest, error) {

	var segHostCount int64
	var addSegCount int64

	if d.HasChange("segment_host_count") {
		segHostOld, segHostNew := d.GetChange("segment_host_count")
		segHostCount = int64(segHostNew.(int) - segHostOld.(int))
		// set key value to old as if no changes nas been made - need for correct update
		d.Set("segment_host_count", segHostOld.(int))
	}

	if d.HasChange("segment_in_host") {
		inHostOld, inHostNew := d.GetChange("segment_in_host")
		addSegCount = int64(inHostNew.(int) - inHostOld.(int))
		// set key value to old as if no changes nas been made - need for correct update
		d.Set("segment_in_host", inHostOld.(int))
	}

	return &greenplum.ExpandRequest{
		ClusterId:               d.Id(),
		SegmentHostCount:        segHostCount,
		AddSegmentsPerHostCount: addSegCount,
		Duration:                yandexMDBGreenplumClusterExpandDuration,
	}, nil

}

func prepareUpdateGreenplumClusterRequest(d *schema.ResourceData, config *Config) (*greenplum.UpdateClusterRequest, error) {
	if d.HasChange("security_group_ids") {
		return nil, fmt.Errorf("changing of 'security_group_ids' is not implemented yet")
	}
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Greenplum cluster update: %s", err)
	}

	configSpec, configMask, err := expandGreenplumConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding config spec on Greenplum Cluster update: %s", err)
	}

	maintenanceWindow, err := expandGreenplumMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding maintenance_window on Greenplum Cluster update: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, config)
	if err != nil {
		return nil, fmt.Errorf("error while expanding network id on Greenplum Cluster update: %s", err)
	}

	return &greenplum.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Name:               d.Get("name").(string),
		UserPassword:       d.Get("user_password").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		NetworkId:          networkID,
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		DeletionProtection: d.Get("deletion_protection").(bool),
		MaintenanceWindow:  maintenanceWindow,
		ServiceAccountId:   d.Get("service_account_id").(string),
		Logging:            expandGreenplumLogging(d),

		Config: &greenplum.GreenplumConfig{
			Version:           d.Get("version").(string),
			BackupWindowStart: expandGreenplumBackupWindowStart(d),
			Access:            expandGreenplumAccess(d),
		},
		MasterConfig: &greenplum.MasterSubclusterConfigSpec{
			Resources: &greenplum.Resources{
				ResourcePresetId: d.Get("master_subcluster.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("master_subcluster.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("master_subcluster.0.resources.0.disk_size").(int)),
			},
		},
		SegmentConfig: &greenplum.SegmentSubclusterConfigSpec{
			Resources: &greenplum.Resources{
				ResourcePresetId: d.Get("segment_subcluster.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("segment_subcluster.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("segment_subcluster.0.resources.0.disk_size").(int)),
			},
		},

		UpdateMask: fieldmaskpb.Union(
			expandGreenplumUpdateMask(d),
			configMask,
		),
		ConfigSpec:   configSpec,
		CloudStorage: expandGreenplumCloudStorage(d),
	}, nil
}

func resourceYandexMDBGreenplumClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Greenplum Cluster %q", d.Id())

	req := &greenplum.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Greenplum().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Greenplum Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Greenplum Cluster %q", d.Id())
	return nil
}
