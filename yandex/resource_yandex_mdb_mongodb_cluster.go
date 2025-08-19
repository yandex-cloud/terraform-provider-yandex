package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"golang.org/x/exp/maps"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type key int

const (
	ReadModeKey key = iota
)

const (
	yandexMDBMongoDBClusterCreateTimeout = time.Hour
	yandexMDBMongoDBClusterDeleteTimeout = time.Hour
	yandexMDBMongoDBClusterUpdateTimeout = 2 * time.Hour
)

func resourceYandexMDBMongodbCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a MongoDB cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).",

		CreateContext: resourceYandexMDBMongodbClusterCreate,
		ReadContext:   resourceYandexMDBMongodbClusterRead,
		UpdateContext: resourceYandexMDBMongodbClusterUpdate,
		DeleteContext: resourceYandexMDBMongodbClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceYandexMDBMongodbClusterImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBMongoDBClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBMongoDBClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBMongoDBClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
				ForceNew:    true,
			},
			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the MongoDB cluster. Can be either `PRESTABLE` or `PRODUCTION`.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseMongoDBEnv),
			},
			"user": {
				Type:        schema.TypeSet,
				Description: "A user of the MongoDB cluster.",
				Optional:    true,
				Computed:    true,
				Set:         mongodbUserHash,
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
							Set:         mongodbUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:        schema.TypeString,
										Description: "The name of the database that the permission grants access to.",
										Required:    true,
									},
									"roles": {
										Type:        schema.TypeList,
										Description: "The roles of the user in this database. For more information see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts/users-and-roles).",
										Optional:    true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
				Deprecated: useResourceInstead("user", "yandex_mdb_mongodb_user"),
			},
			"database": {
				Type:        schema.TypeSet,
				Description: "A database of the MongoDB cluster.",
				Optional:    true,
				Computed:    true,
				Set:         mongodbDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the database.",
							Required:    true,
						},
					},
				},
				Deprecated: useResourceInstead("database", "yandex_mdb_mongodb_database"),
			},
			"host": {
				Type:        schema.TypeList,
				Description: "A host of the MongoDB cluster.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The fully qualified domain name of the host. Computed on server side.",
							Computed:    true,
						},
						"zone_id": {
							Type:        schema.TypeString,
							Description: common.ResourceDescriptions["zone"],
							Required:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "The role of the cluster (either PRIMARY or SECONDARY).",
							Optional:    true,
							Computed:    true,
						},
						"health": {
							Type:        schema.TypeString,
							Description: "The health of the host.",
							Computed:    true,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
							Required:    true,
						},
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Description: "Should this host have assigned public IP assigned. Can be either `true` or `false`.",
							Optional:    true,
							Default:     false,
						},
						"shard_name": {
							Type:        schema.TypeString,
							Description: "The name of the shard to which the host belongs. Only for sharded cluster.",
							Optional:    true,
							Computed:    true,
						},
						"type": {
							Type:         schema.TypeString,
							Description:  "Type of Mongo daemon which runs on this host (mongod, mongos, mongocfg, mongoinfra). Defaults to `mongod`.",
							Optional:     true,
							Default:      "MONGOD",
							ValidateFunc: validation.StringInSlice([]string{"MONGOS", "MONGOINFRA", "MONGOD", "MONGOCFG"}, true),
							StateFunc:    stateToUpper,
						},
						"host_parameters": {
							Type:        schema.TypeList,
							Description: "The parameters of mongod host in replicaset.",
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hidden": {
										Type:        schema.TypeBool,
										Description: "Should this host be hidden in replicaset. Can be either `true` of `false`. For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.hidden).",
										Optional:    true,
									},
									"priority": {
										Type:        schema.TypeFloat,
										Description: "A floating point number that indicates the relative likelihood of a replica set member to become the primary. For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.priority).",
										Optional:    true,
									},
									"secondary_delay_secs": {
										Type:        schema.TypeInt,
										Description: "The number of seconds `behind` the primary that this replica set member should `lag`. For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.secondaryDelaySecs).",
										Optional:    true,
									},
									"tags": {
										Type:        schema.TypeMap,
										Description: "A set of key/value pairs to assign for the replica set member. For more information see [the official documentation](https://www.mongodb.com/docs/current/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.tags).",
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
									},
								},
							},
						},
					},
				},
			},
			"resources": {
				Type:          schema.TypeList,
				Description:   "(**DEPRECATED**, use `resources_*` instead) Resources allocated to hosts of the MongoDB cluster.",
				Optional:      true,
				MaxItems:      1,
				Deprecated:    useResourceInstead("`resources`", "`resources_mongo*`"),
				ExactlyOneOf:  []string{"resources_mongod"},
				ConflictsWith: []string{"resources_mongod", "resources_mongoinfra", "resources_mongocfg", "resources_mongos"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a MongoDB host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).",
							Required:    true,
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Description: "Volume of the storage available to a MongoDB host, in gigabytes.",
							Required:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of MongoDB hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
							Required:    true,
						},
					},
				},
			},
			"resources_mongod": {
				Type:        schema.TypeList,
				Description: "Resources allocated to `mongod` hosts of the MongoDB cluster.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a MongoDB host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).",
							Required:    true,
						},
						"disk_size": {
							Type:             schema.TypeInt,
							Description:      "Volume of the storage available to a MongoDB host, in gigabytes.",
							Required:         true,
							DiffSuppressFunc: suppressDiskSizeChangeOnAutoscaling("disk_size_autoscaling_mongod.0.disk_size_limit"),
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of MongoDB hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
							Required:    true,
						},
					},
				},
			},
			"resources_mongoinfra": {
				Type:         schema.TypeList,
				Description:  "Resources allocated to `mongoinfra` hosts of the MongoDB cluster.",
				Optional:     true,
				MaxItems:     1,
				RequiredWith: []string{"resources_mongod"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a MongoDB host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).",
							Required:    true,
						},
						"disk_size": {
							Type:             schema.TypeInt,
							Description:      "Volume of the storage available to a MongoDB host, in gigabytes.",
							Required:         true,
							DiffSuppressFunc: suppressDiskSizeChangeOnAutoscaling("disk_size_autoscaling_mongoinfra.0.disk_size_limit"),
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of MongoDB hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
							Required:    true,
						},
					},
				},
			},
			"resources_mongocfg": {
				Type:         schema.TypeList,
				Description:  "Resources allocated to `mongocfg` hosts of the MongoDB cluster.",
				Optional:     true,
				MaxItems:     1,
				RequiredWith: []string{"resources_mongod"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a MongoDB host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).",
							Required:    true,
						},
						"disk_size": {
							Type:             schema.TypeInt,
							Description:      "Volume of the storage available to a MongoDB host, in gigabytes.",
							Required:         true,
							DiffSuppressFunc: suppressDiskSizeChangeOnAutoscaling("disk_size_autoscaling_mongocfg.0.disk_size_limit"),
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of MongoDB hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
							Required:    true,
						},
					},
				},
			},
			"resources_mongos": {
				Type:         schema.TypeList,
				Description:  "Resources allocated to `mongos` hosts of the MongoDB cluster.",
				Optional:     true,
				MaxItems:     1,
				RequiredWith: []string{"resources_mongod", "resources_mongocfg"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a MongoDB host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).",
							Required:    true,
						},
						"disk_size": {
							Type:             schema.TypeInt,
							Description:      "Volume of the storage available to a MongoDB host, in gigabytes.",
							Required:         true,
							DiffSuppressFunc: suppressDiskSizeChangeOnAutoscaling("disk_size_autoscaling_mongos.0.disk_size_limit"),
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of MongoDB hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
							Required:    true,
						},
					},
				},
			},
			"disk_size_autoscaling_mongod": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Disk size autoscaling settings for mongod.",
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
			"disk_size_autoscaling_mongoinfra": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Disk size autoscaling settings for mongoinfra.",
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
			"disk_size_autoscaling_mongocfg": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Disk size autoscaling settings for mongocfg.",
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
			"disk_size_autoscaling_mongos": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Disk size autoscaling settings for mongos.",
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
			"cluster_config": {
				Type:        schema.TypeList,
				Description: "Configuration of the MongoDB subcluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:        schema.TypeString,
							Description: "Version of the MongoDB server software. Can be either `4.2`, `4.4`, `4.4-enterprise`, `5.0`, `5.0-enterprise`, `6.0` and `6.0-enterprise`.",
							Required:    true,
						},
						"feature_compatibility_version": {
							Type:        schema.TypeString,
							Description: "Feature compatibility version of MongoDB. If not provided version is taken. Can be either `6.0`, `5.0`, `4.4` and `4.2`.",
							Optional:    true,
							Computed:    true,
						},
						"backup_retain_period_days": {
							Type:        schema.TypeInt,
							Description: "Retain period of automatically created backup in days.",
							Optional:    true,
							Computed:    true,
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
						"performance_diagnostics": {
							Type:        schema.TypeList,
							Description: "Performance diagnostics to the MongoDB cluster.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Description: "Enable or disable performance diagnostics.",
										Optional:    true,
									},
								},
							},
						},
						"access": {
							Type:        schema.TypeList,
							Description: "Access policy to the MongoDB cluster.",
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_lens": {
										Type:        schema.TypeBool,
										Description: "Allow access for [Yandex DataLens](https://yandex.cloud/services/datalens).",
										Optional:    true,
										Default:     false,
									},
									"data_transfer": {
										Type:        schema.TypeBool,
										Description: "Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer).",
										Optional:    true,
										Default:     false,
									},
									"web_sql": {
										Type:        schema.TypeBool,
										Description: "Allow access for [WebSQL](https://yandex.cloud/ru/docs/websql/).",
										Optional:    true,
										Default:     false,
									},
								},
							},
						},
						"mongod": {
							Type:        schema.TypeList,
							Description: "Configuration of the mongod service.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audit_log": {
										Type:        schema.TypeList,
										Description: "A set of audit log settings (see the [auditLog](https://www.mongodb.com/docs/manual/reference/configuration-options/#auditlog-options) option). Available only in enterprise edition.",
										MaxItems:    1,
										Optional:    true,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"filter": {
													Type:        schema.TypeString,
													Description: "Configuration of the audit log filter in JSON format. For more information see [auditLog.filter](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-auditLog.filter) description in the official documentation. Available only in enterprise edition.",
													Optional:    true,
												},
												"runtime_configuration": {
													Type:        schema.TypeBool,
													Description: "Specifies if a node allows runtime configuration of audit filters and the auditAuthorizationSuccess variable. For more information see [auditLog.runtimeConfiguration](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-auditLog.runtimeConfiguration) description in the official documentation. Available only in enterprise edition.",
													Optional:    true,
												},
											},
										},
									},
									"set_parameter": {
										Type:        schema.TypeList,
										Description: "A set of MongoDB Server Parameters (see the [setParameter](https://www.mongodb.com/docs/manual/reference/configuration-options/#setparameter-option) option).",
										MaxItems:    1,
										Optional:    true,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"audit_authorization_success": {
													Type:        schema.TypeBool,
													Description: "Enables the auditing of authorization successes. Can be either true or false. For more information, see the [auditAuthorizationSuccess](https://www.mongodb.com/docs/manual/reference/parameters/#mongodb-parameter-param.auditAuthorizationSuccess) description in the official documentation. Available only in enterprise edition.",
													Optional:    true,
												},
												"enable_flow_control": {
													Type:        schema.TypeBool,
													Description: "Enables the flow control. Can be either true or false. For more information, see the [enableFlowControl](https://www.mongodb.com/docs/rapid/reference/parameters/#mongodb-parameter-param.enableFlowControl) description in the official documentation.",
													Optional:    true,
												},
												"min_snapshot_history_window_in_seconds": {
													Type:        schema.TypeInt,
													Description: "The minimum time window in seconds for which the storage engine keeps the snapshot history. For more information, see the [minSnapshotHistoryWindowInSeconds](https://www.mongodb.com/docs/manual/reference/parameters/#mongodb-parameter-param.minSnapshotHistoryWindowInSeconds) description in the official documentation.",
													Optional:    true,
												},
											},
										},
									},
									"security": {
										Type:        schema.TypeList,
										Description: "A set of MongoDB Security settings (see the [security](https://www.mongodb.com/docs/manual/reference/configuration-options/#security-options) option). Available only in enterprise edition.",
										MaxItems:    1,
										Optional:    true,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enable_encryption": {
													Type:        schema.TypeBool,
													Description: "Enables the encryption for the WiredTiger storage engine. Can be either true or false. For more information see [security.enableEncryption](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.enableEncryption) description in the official documentation. Available only in enterprise edition.",
													Optional:    true,
												},
												"kmip": {
													Type:        schema.TypeList,
													Description: "Configuration of the third party key management appliance via the Key Management Interoperability Protocol (KMIP) (see [Encryption tutorial](https://www.mongodb.com/docs/rapid/tutorial/configure-encryption) ). Requires `enable_encryption` to be true. The structure is documented below. Available only in enterprise edition.",
													MaxItems:    1,
													Optional:    true,
													Computed:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"server_name": {
																Type:        schema.TypeString,
																Description: "Hostname or IP address of the KMIP server to connect to. For more information see [security.kmip.serverName](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.serverName) description in the official documentation.",
																Optional:    true,
															},
															"port": {
																Type:        schema.TypeInt,
																Description: "Port number to use to communicate with the KMIP server. Default: 5696 For more information see [security.kmip.port](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.port) description in the official documentation.",
																Optional:    true,
															},
															"server_ca": {
																Type:        schema.TypeString,
																Description: "Path to CA File. Used for validating secure client connection to KMIP server. For more information see [security.kmip.serverCAFile](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.serverCAFile) description in the official documentation.",
																Optional:    true,
															},
															"client_certificate": {
																Type:        schema.TypeString,
																Description: "String containing the client certificate used for authenticating MongoDB to the KMIP server. For more information see [security.kmip.clientCertificateFile](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.clientCertificateFile) description in the official documentation.",
																Optional:    true,
															},
															"key_identifier": {
																Type:        schema.TypeString,
																Description: "Unique KMIP identifier for an existing key within the KMIP server. For more information see [security.kmip.keyIdentifier](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-security.kmip.keyIdentifier) description in the official documentation.",
																Optional:    true,
															},
														},
													},
												},
											},
										},
									},
									"operation_profiling": {
										Type:        schema.TypeList,
										Description: "A set of profiling settings (see the [operationProfiling](https://www.mongodb.com/docs/manual/reference/configuration-options/#operationprofiling-options) option).",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mode": {
													Type:         schema.TypeString,
													Description:  "Specifies which operations should be profiled. The following profiler levels are available: off, slow_op, all. For more information, see the [operationProfiling.mode](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.mode) description in the official documentation.",
													Optional:     true,
													StateFunc:    stateToUpper,
													ValidateFunc: validation.StringInSlice([]string{"OFF", "SLOW_OP", "ALL"}, true),
												},
												"slow_op_threshold": {
													Type:        schema.TypeInt,
													Description: "The slow operation time threshold, in milliseconds. Operations that run for longer than this threshold are considered slow. For more information, see the [operationProfiling.slowOpThresholdMs](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.slowOpThresholdMs) description in the official documentation.",
													Optional:    true,
												},
												"slow_op_sample_rate": {
													Type:        schema.TypeFloat,
													Description: "The fraction of slow operations that should be profiled or logged. Accepts values between 0 and 1, inclusive. For more information, see the [operationProfiling.slowOpSampleRate](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.slowOpSampleRate) description in the official documentation.",
													Optional:    true,
												},
											},
										},
									},
									"net": {
										Type:        schema.TypeList,
										Description: "A set of network settings (see the [net](https://www.mongodb.com/docs/manual/reference/configuration-options/#net-options) option).",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_incoming_connections": {
													Type:        schema.TypeInt,
													Description: "The maximum number of simultaneous connections that host will accept. For more information, see the [net.maxIncomingConnections](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-net.maxIncomingConnections) description in the official documentation.",
													Optional:    true,
												},
												"compressors": {
													Type:        schema.TypeList,
													Description: "Specifies the default compressor(s) to use for communication between this mongod or mongos. Accepts array of compressors. Order matters. Available compressors: snappy, zlib, zstd, disabled. To disable network compression, make `disabled` the only value. For more information, see the [net.Compression.Compressors](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-net.compression.compressors) description in the official documentation.",
													MaxItems:    3,
													Optional:    true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														StateFunc:    stateToUpper,
														ValidateFunc: validation.StringInSlice([]string{"ZSTD", "ZLIB", "SNAPPY", "DISABLED"}, true),
													},
												},
											},
										},
									},
									"storage": {
										Type:        schema.TypeList,
										Description: "A set of storage settings (see the [storage](https://www.mongodb.com/docs/manual/reference/configuration-options/#storage-options) option).",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"wired_tiger": {
													Type:        schema.TypeList,
													Description: "The WiredTiger engine settings. (see the [storage.wiredTiger](https://www.mongodb.com/docs/manual/reference/configuration-options/#storage.wiredtiger-options) option). These settings available only on `mongod` hosts.",
													MaxItems:    1,
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"cache_size_gb": {
																Type:        schema.TypeFloat,
																Description: "Defines the maximum size of the internal cache that WiredTiger will use for all data. For more information, see the [storage.wiredTiger.engineConfig.cacheSizeGB](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.wiredTiger.engineConfig.cacheSizeGB) description in the official documentation.",
																Optional:    true,
															},
															"block_compressor": {
																Type:         schema.TypeString,
																Description:  "Specifies the default compression for collection data. You can override this on a per-collection basis when creating collections. Available compressors are: none, snappy, zlib, zstd. This setting available only on `mongod` hosts. For more information, see the [storage.wiredTiger.collectionConfig.blockCompressor](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.wiredTiger.collectionConfig.blockCompressor) description in the official documentation.",
																Optional:     true,
																StateFunc:    stateToUpper,
																ValidateFunc: validation.StringInSlice([]string{"NONE", "ZLIB", "SNAPPY", "ZSTD"}, true),
															},
															"prefix_compression": {
																Type:        schema.TypeBool,
																Description: "Enables or disables prefix compression for index data. Сan be either true or false. For more information, see the [storage.wiredTiger.indexConfig.prefixCompression](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.wiredTiger.indexConfig.prefixCompression) description in the official documentation.",
																Optional:    true,
															},
														},
													},
												},
												"journal": {
													Type:        schema.TypeList,
													Description: "The durability journal to ensure data files remain valid and recoverable.",
													MaxItems:    1,
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"commit_interval": {
																Type:        schema.TypeInt,
																Description: "The maximum amount of time in milliseconds that the mongod process allows between journal operations. For more information, see the [storage.journal.commitIntervalMs](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.journal.commitIntervalMs) description in the official documentation.",
																Optional:    true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"mongocfg": {
							Type:        schema.TypeList,
							Description: "Configuration of the mongocfg service.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"operation_profiling": {
										Type:        schema.TypeList,
										Description: "A set of profiling settings (see the [operationProfiling](https://www.mongodb.com/docs/manual/reference/configuration-options/#operationprofiling-options) option).",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mode": {
													Description:  "Specifies which operations should be profiled. The following profiler levels are available: off, slow_op, all. For more information, see the [operationProfiling.mode](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.mode) description in the official documentation.",
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice([]string{"OFF", "SLOW_OP", "ALL"}, true),
												},
												"slow_op_threshold": {
													Type:        schema.TypeInt,
													Optional:    true,
													Description: "The slow operation time threshold, in milliseconds. Operations that run for longer than this threshold are considered slow. For more information, see the [operationProfiling.slowOpThresholdMs](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-operationProfiling.slowOpThresholdMs) description in the official documentation.",
												},
											},
										},
									},
									"net": {
										Type:        schema.TypeList,
										Description: " A set of network settings (see the [net](https://www.mongodb.com/docs/manual/reference/configuration-options/#net-options) option).",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_incoming_connections": {
													Type:        schema.TypeInt,
													Description: "The maximum number of simultaneous connections that host will accept. For more information, see the [net.maxIncomingConnections](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-net.maxIncomingConnections) description in the official documentation.",
													Optional:    true,
												},
											},
										},
									},
									"storage": {
										Type:        schema.TypeList,
										Description: "A set of storage settings (see the [storage](https://www.mongodb.com/docs/manual/reference/configuration-options/#storage-options) option).",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"wired_tiger": {
													Type:        schema.TypeList,
													MaxItems:    1,
													Optional:    true,
													Description: "The WiredTiger engine settings. (see the [storage.wiredTiger](https://www.mongodb.com/docs/manual/reference/configuration-options/#storage.wiredtiger-options) option).",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"cache_size_gb": {
																Type:        schema.TypeFloat,
																Optional:    true,
																Description: "Defines the maximum size of the internal cache that WiredTiger will use for all data. For more information, see the [storage.wiredTiger.engineConfig.cacheSizeGB](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-storage.wiredTiger.engineConfig.cacheSizeGB) description in the official documentation.",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"mongos": {
							Type:        schema.TypeList,
							Description: "Configuration of the mongos service.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"net": {
										Type:        schema.TypeList,
										Description: "A set of network settings (see the [net](https://www.mongodb.com/docs/manual/reference/configuration-options/#net-options) option).",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_incoming_connections": {
													Type:        schema.TypeInt,
													Description: "The maximum number of simultaneous connections that host will accept. For more information, see the [net.maxIncomingConnections](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-net.maxIncomingConnections) description in the official documentation.",
													Optional:    true,
												},
												"compressors": {
													Type:        schema.TypeList,
													MaxItems:    3,
													Optional:    true,
													Description: "Specifies the default compressor(s) to use for communication between this mongod or mongos. Accepts array of compressors. Order matters. Available compressors: snappy, zlib, zstd, disabled. To disable network compression, make `disabled` the only value. For more information, see the [net.Compression.Compressors](https://www.mongodb.com/docs/manual/reference/configuration-options/#mongodb-setting-net.compression.compressors) description in the official documentation.",
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														StateFunc:    stateToUpper,
														ValidateFunc: validation.StringInSlice([]string{"ZSTD", "ZlIB", "SNAPPY", "DISABLED"}, true),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the cluster.",
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
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"sharded": {
				Type:        schema.TypeBool,
				Description: "MongoDB Cluster mode enabled/disabled.",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-mongodb/api-ref/Cluster/).",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-mongodb/api-ref/Cluster/).",
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
							Description: "Backup ID. The cluster will be created from the specified backup. [How to get a list of PostgreSQL backups](https://yandex.cloud/docs/managed-mongodb/operations/cluster-backups).",
							Required:    true,
							ForceNew:    true,
						},
						"time": {
							Type:         schema.TypeString,
							Description:  "Timestamp of the moment to which the MongoDB cluster should be restored. (Format: `2006-01-02T15:04:05` - UTC). When not set, current time is used.",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: stringToTimeValidateFunc,
						},
					},
				},
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: "Maintenance window settings of the MongoDB cluster.",
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
							Description:  "Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
							ValidateFunc: validateParsableValue(parseMongoDBWeekDay),
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							Description:  "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
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
			"disk_encryption_key_id": {
				Type:        schema.TypeString,
				Description: "ID of the KMS key for cluster disk encryption.",
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
		},
		CustomizeDiff: customdiff.All(
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				stateDbsCnt := d.Get("database").(*schema.Set).Len()
				if stateDbsCnt == 0 {
					if err := d.Clear("database"); err != nil {
						return err
					}
				}

				stateUsersCnt := d.Get("user").(*schema.Set).Len()
				if stateUsersCnt == 0 {
					if err := d.Clear("user"); err != nil {
						return err
					}
				}
				return nil
			},
		),
	}
}

func stateToUpper(val interface{}) string {
	return strings.ToUpper(val.(string))
}

func prepareCreateMongodbRequest(d *schema.ResourceData, meta *Config) (*mongodb.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Mongodb Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting folder ID while creating Mongodb Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseMongoDBEnv(e)
	if err != nil {
		return nil, fmt.Errorf("error resolving environment while creating Mongodb Cluster: %s", err)
	}

	version := extractVersion(d)
	configSpec := &mongodb.ConfigSpec{Version: version, FeatureCompatibilityVersion: version}
	if cfgCompVer := d.Get("cluster_config.0.feature_compatibility_version"); cfgCompVer != nil {
		configSpec.FeatureCompatibilityVersion = cfgCompVer.(string)
	}

	if backupStart := d.Get("cluster_config.0.backup_window_start"); backupStart != nil {
		configSpec.BackupWindowStart = expandMongoDBBackupWindowStart(d)
	}

	configSpec.BackupRetainPeriodDays = expandMongoDBBackupRetainPeriod(d)

	if access := d.Get("cluster_config.0.access"); access != nil {
		configSpec.Access = &mongodb.Access{
			DataLens:     d.Get("cluster_config.0.access.0.data_lens").(bool),
			DataTransfer: d.Get("cluster_config.0.access.0.data_transfer").(bool),
			WebSql:       d.Get("cluster_config.0.access.0.web_sql").(bool),
		}
	}

	if pd := d.Get("cluster_config.0.performance_diagnostics"); pd != nil {
		configSpec.PerformanceDiagnostics = &mongodb.PerformanceDiagnosticsConfig{
			ProfilingEnabled: d.Get("cluster_config.0.performance_diagnostics.0.enabled").(bool),
		}
	}

	mongodbSpecHelper := GetMongodbSpecHelper()
	configSpec.Mongodb = mongodbSpecHelper.Expand(d)

	hosts, err := expandMongoDBHosts(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding hosts on MongoDB Cluster create: %s", err)
	}

	dbSpecs, err := expandMongoDBDatabases(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding databases on MongoDB Cluster create: %s", err)
	}

	users, err := expandMongoDBUserSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding user specs on MongoDB Cluster create: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on MongoDB Cluster create: %s", err)
	}

	mw, err := expandMongoDBMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding maintenance window on MongoDB Cluster create: %s", err)
	}

	var diskEncryptionKeyId *wrapperspb.StringValue
	if val, ok := d.GetOk("disk_encryption_key_id"); ok {
		diskEncryptionKeyId = &wrapperspb.StringValue{
			Value: val.(string),
		}
	}

	req := mongodb.CreateClusterRequest{
		FolderId:            folderID,
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		NetworkId:           networkID,
		Environment:         env,
		ConfigSpec:          configSpec,
		HostSpecs:           hosts,
		UserSpecs:           users,
		DatabaseSpecs:       dbSpecs,
		Labels:              labels,
		SecurityGroupIds:    securityGroupIds,
		DeletionProtection:  d.Get("deletion_protection").(bool),
		MaintenanceWindow:   mw,
		DiskEncryptionKeyId: diskEncryptionKeyId,
	}
	return &req, nil
}

func resourceYandexMDBMongodbClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	req, err := prepareCreateMongodbRequest(d, config)

	if err != nil {
		return diag.FromErr(err)
	}

	if backupID, ok := d.GetOk("restore.0.backup_id"); ok && backupID != "" {
		return resourceYandexMDBMongodbClusterRestore(d, meta, req, backupID.(string))
	}

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Create(ctx, req))
	if err != nil {
		return diag.Errorf("error while requesting API to create Mongodb Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while get Mongodb create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*mongodb.CreateClusterMetadata)
	if !ok {
		return diag.Errorf("could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting for operation to create Mongodb Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("Mongodb Cluster creation failed: %s", err)
	}

	return resourceYandexMDBMongodbClusterRead(ctx, d, meta)
}

func resourceYandexMDBMongodbClusterRestore(d *schema.ResourceData, meta interface{}, createClusterRequest *mongodb.CreateClusterRequest, backupID string) diag.Diagnostics {
	config := meta.(*Config)

	var timeBackup *mongodb.RestoreClusterRequest_RecoveryTargetSpec = nil
	if backupTime, ok := d.GetOk("restore.0.time"); ok {
		time, err := parseStringToTime(backupTime.(string))
		if err != nil {
			return diag.Errorf("Error while parsing restore.0.time to create MongoDB Clsuter from backup %v, value: %v error: %s", backupID, backupTime, err)
		}
		timeBackup = &mongodb.RestoreClusterRequest_RecoveryTargetSpec{
			Timestamp: time.Unix(),
		}
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	request := &mongodb.RestoreClusterRequest{
		BackupId:            backupID,
		RecoveryTargetSpec:  timeBackup,
		Name:                createClusterRequest.Name,
		Description:         createClusterRequest.Description,
		Labels:              createClusterRequest.Labels,
		Environment:         createClusterRequest.Environment,
		ConfigSpec:          createClusterRequest.ConfigSpec,
		HostSpecs:           createClusterRequest.HostSpecs,
		NetworkId:           createClusterRequest.NetworkId,
		FolderId:            createClusterRequest.FolderId,
		SecurityGroupIds:    createClusterRequest.SecurityGroupIds,
		DeletionProtection:  createClusterRequest.DeletionProtection,
		DiskEncryptionKeyId: createClusterRequest.DiskEncryptionKeyId,
	}

	// Empty string will remove encryption when restoring
	if request.DiskEncryptionKeyId == nil {
		log.Printf("[WARN] Disk encryption key ID is not set. Encryption will be disabled if present in source cluster.")
		request.DiskEncryptionKeyId = &wrapperspb.StringValue{Value: ""}
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending MongoDB cluste restore request: %+v", request)
		return config.sdk.MDB().MongoDB().Cluster().Restore(ctx, request)
	})

	if err != nil {
		return diag.Errorf("Error while requesting API to create MongoDB Cluster from backup %v: %s", backupID, err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get MongoDB Cluster create from backup %v operation metadata: %s", backupID, err)
	}

	md, ok := protoMetadata.(*mongodb.RestoreClusterMetadata)
	if !ok {
		return diag.Errorf("Could not get MongoDB Cluster ID from create from backup %v operation metadata", backupID)
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while waiting for operation to create MongoDB Cluster from backup %v: %s", backupID, err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("MongoDB Cluster creationg from backup %v failed: %s", backupID, err)
	}

	return resourceYandexMDBMongodbClusterRead(ctx, d, meta)
}

//nolint:unused
func updateMongoDBMaintenanceWindow(ctx context.Context, config *Config, d *schema.ResourceData, mw *mongodb.MaintenanceWindow) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().Update(ctx, &mongodb.UpdateClusterRequest{
			ClusterId:         d.Id(),
			MaintenanceWindow: mw,
			UpdateMask:        &field_mask.FieldMask{Paths: []string{"maintenance_window"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update maintenance window in MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating maintenance window in MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func listMongodbHosts(ctx context.Context, config *Config, d *schema.ResourceData) ([]*mongodb.Host, error) {
	var hosts []*mongodb.Host
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MongoDB().Cluster().ListHosts(ctx, &mongodb.ListClusterHostsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", d.Id(), err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}

func listMongodbShards(ctx context.Context, config *Config, d *schema.ResourceData) ([]*mongodb.Shard, error) {
	var shards []*mongodb.Shard
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MongoDB().Cluster().ListShards(ctx, &mongodb.ListClusterShardsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of shards for '%s': %s", d.Id(), err)
		}
		shards = append(shards, resp.Shards...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return shards, nil
}

func resourceYandexMDBMongodbClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	cluster, err := config.sdk.MDB().MongoDB().Cluster().Get(ctx, &mongodb.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string))))
	}

	if err := d.Set("created_at", getTimestamp(cluster.CreatedAt)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", cluster.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("folder_id", cluster.FolderId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("network_id", cluster.NetworkId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment", cluster.GetEnvironment().String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("health", cluster.GetHealth().String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", cluster.GetStatus().String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", cluster.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sharded", cluster.Sharded); err != nil {
		return diag.FromErr(err)
	}

	mongodbSpecHelper := GetMongodbSpecHelper()
	flattenResources, err := mongodbSpecHelper.FlattenResources(cluster.Config, d)
	if err != nil {
		return diag.FromErr(err)
	}
	for k, v := range flattenResources {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}

	flattenDsa, err := mongodbSpecHelper.FlattenDiskSizeAutoscaling(cluster.Config, d)
	if err != nil {
		return diag.FromErr(err)
	}
	for k, v := range flattenDsa {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}

	expandUsers, err := expandMongoDBUserSpecs(d)
	if err != nil {
		return diag.FromErr(err)
	}
	passwords := mongodbUsersPasswords(expandUsers)

	stateUsersCnt := d.Get("user").(*schema.Set).Len()
	isReadMode, ok := ctx.Value(ReadModeKey).(bool)
	if !(ok && isReadMode) && stateUsersCnt == 0 {
		if err := d.Set("user", nil); err != nil {
			return diag.FromErr(err)
		}
	} else {
		clusterUsers, err := listMongodbUsers(ctx, config, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		flattenUsers := flattenMongoDBUsers(clusterUsers, passwords)
		if err := d.Set("user", flattenUsers); err != nil {
			return diag.FromErr(err)
		}
	}

	stateDbsCnt := d.Get("database").(*schema.Set).Len()
	if !(ok && isReadMode) && stateDbsCnt == 0 {
		if err := d.Set("database", nil); err != nil {
			return diag.FromErr(err)
		}
	} else {
		clusterDatabases, err := listMongodbDatabases(ctx, config, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		flattenDatabases := flattenMongoDBDatabases(clusterDatabases)

		if err := d.Set("database", flattenDatabases); err != nil {
			return diag.FromErr(err)
		}
	}

	flattenClusterConfig, err := flattenMongoDBClusterConfig(cluster.Config, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_config", flattenClusterConfig); err != nil {
		return diag.FromErr(err)
	}

	clusterHosts, err := listMongodbHosts(ctx, config, d)
	if err != nil {
		return diag.FromErr(err)
	}

	expandHosts, err := expandMongoDBHosts(d)
	if err != nil {
		return diag.FromErr(err)
	}

	hosts := sortMongoDBHosts(clusterHosts, expandHosts)

	flattenHosts, err := flattenMongoDBHosts(hosts)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("host", flattenHosts); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return diag.FromErr(err)
	}

	mw := flattenMongoDBMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("deletion_protection", cluster.DeletionProtection); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return diag.FromErr(err)
	}

	if cluster.DiskEncryptionKeyId != nil {
		if err = d.Set("disk_encryption_key_id", cluster.DiskEncryptionKeyId.GetValue()); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func sortMongoDBHosts(hosts []*mongodb.Host, specs []*mongodb.HostSpec) []*mongodb.Host {
	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if matchHosts(h, hosts[j]) {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}

	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if !matchHosts(h, hosts[j]) {
				break
			}

			if matchHostParameters(h, hosts[j]) {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}

	return hosts
}

func matchHosts(hostSpec *mongodb.HostSpec, host *mongodb.Host) bool {
	return hostSpec.ZoneId == host.ZoneId && (hostSpec.ShardName == "" || hostSpec.ShardName == host.ShardName) && hostSpec.Type == host.Type && hostSpec.SubnetId == host.SubnetId
}

func matchHostParameters(hostSpec *mongodb.HostSpec, host *mongodb.Host) bool {
	if host.AssignPublicIp != hostSpec.AssignPublicIp {
		return false
	}

	hostParameters := host.HostParameters
	if hostParameters == nil {
		// Default values
		hostParameters = &mongodb.Host_HostParameters{
			Hidden:             false,
			SecondaryDelaySecs: 0,
			Priority:           1.0,
		}
	}

	if hostParameters.Hidden != getOrDefault(hostSpec.Hidden, false) {
		return false
	}

	if hostParameters.SecondaryDelaySecs != getOrDefault(hostSpec.SecondaryDelaySecs, int64(0)) {
		return false
	}

	if hostParameters.Priority != getOrDefault(hostSpec.Priority, 1.0) {
		return false
	}

	if len(hostParameters.Tags) != len(hostSpec.Tags) {
		return false
	}

	return maps.Equal(hostSpec.Tags, hostParameters.Tags)
}

func resourceYandexMDBMongodbClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.Partial(true)

	if err := setMongoDBFolderID(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	if err := updateMongodbClusterParams(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	stateDbsCnt := d.Get("database").(*schema.Set).Len()
	if d.HasChange("database") && stateDbsCnt > 0 {
		if err := updateMongodbClusterDatabases(ctx, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	stateUsersCnt := d.Get("user").(*schema.Set).Len()
	if d.HasChange("user") && stateUsersCnt > 0 {
		if err := updateMongodbClusterUsers(ctx, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("host") {
		if err := updateMongoDBClusterShards(ctx, d, meta); err != nil {
			return diag.FromErr(err)
		}
		if err := updateMongoDBClusterHosts(ctx, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)
	return resourceYandexMDBMongodbClusterRead(ctx, d, meta)
}

func resourceYandexMDBMongodbClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Mongodb Cluster %q", d.Id())

	req := &mongodb.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Delete(ctx, req))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Mongodb Cluster %q", d.Get("name").(string))))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting Mongodb Cluster %q", d.Id())
	return nil
}

func listMongodbUsers(ctx context.Context, config *Config, id string) ([]*mongodb.User, error) {
	var users []*mongodb.User
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MongoDB().User().List(ctx, &mongodb.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of users for '%s': %s", id, err)
		}
		users = append(users, resp.Users...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return users, nil
}

func listMongodbDatabases(ctx context.Context, config *Config, id string) ([]*mongodb.Database, error) {
	var dbs []*mongodb.Database
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MongoDB().Database().List(ctx, &mongodb.ListDatabasesRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of databases for '%s': %s", id, err)
		}
		dbs = append(dbs, resp.Databases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return dbs, nil
}

func getMongoDBClusterUpdateRequest(d *schema.ResourceData) (*mongodb.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating MongoDB cluster: %s", err)
	}

	version := extractVersion(d)
	mongodbSpecHelper := GetMongodbSpecHelper()
	req := &mongodb.UpdateClusterRequest{
		ClusterId:   d.Id(),
		Description: d.Get("description").(string),
		Labels:      labels,
		Name:        d.Get("name").(string),
		ConfigSpec: &mongodb.ConfigSpec{
			Version:                     version,
			Mongodb:                     mongodbSpecHelper.Expand(d),
			BackupWindowStart:           expandMongoDBBackupWindowStart(d),
			BackupRetainPeriodDays:      expandMongoDBBackupRetainPeriod(d),
			FeatureCompatibilityVersion: d.Get("cluster_config.0.feature_compatibility_version").(string),
			PerformanceDiagnostics: &mongodb.PerformanceDiagnosticsConfig{
				ProfilingEnabled: d.Get("cluster_config.0.performance_diagnostics.0.enabled").(bool),
			},
			Access: &mongodb.Access{
				DataLens:     d.Get("cluster_config.0.access.0.data_lens").(bool),
				DataTransfer: d.Get("cluster_config.0.access.0.data_transfer").(bool),
				WebSql:       d.Get("cluster_config.0.access.0.web_sql").(bool),
			},
		},
		SecurityGroupIds: expandSecurityGroupIds(d.Get("security_group_ids")),
	}
	return req, nil
}

var mdbMongodbUpdateFieldsMap = map[string]string{
	"name":                                           "name",
	"labels":                                         "labels",
	"description":                                    "description",
	"cluster_config.0.access":                        "config_spec.access",
	"security_group_ids":                             "security_group_ids",
	"cluster_config.0.version":                       "config_spec.version",
	"deletion_protection":                            "deletion_protection",
	"cluster_config.0.backup_window_start":           "config_spec.backup_window_start",
	"cluster_config.0.performance_diagnostics":       "config_spec.performance_diagnostics",
	"cluster_config.0.backup_retain_period_days":     "config_spec.backup_retain_period_days",
	"cluster_config.0.feature_compatibility_version": "config_spec.feature_compatibility_version",
}

func updateMongodbClusterParams(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := getMongoDBClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	var updatePath []string
	for field, path := range mdbMongodbUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	types := getSetOfHostTypes(d)
	var sharded bool
	if sharded = d.Get("sharded").(bool); !sharded && mapContainsOneOfKeys(types, []string{
		mongodb.Host_MONGOINFRA.String(),
		mongodb.Host_MONGOS.String(),
		mongodb.Host_MONGOCFG.String(),
	}) {
		err := enableShardingMongoDB(ctx, config, d)
		if err != nil {
			return err
		}
	}

	if sharded && d.HasChange("resources_mongoinfra") {
		if compareResources(d, "resources", "resources_mongoinfra") {
			resourcesSpecPath := "config_spec.mongodb.mongoinfra.resources"
			updatePath = append(updatePath, resourcesSpecPath)
		}
	}

	if sharded && d.HasChange("resources_mongocfg") {
		if compareResources(d, "resources", "resources_mongocfg") {
			resourcesSpecPath := "config_spec.mongodb.mongocfg.resources"
			updatePath = append(updatePath, resourcesSpecPath)
		}
	}

	if d.HasChange("resources_mongod") || d.HasChange("resources") {
		if compareResources(d, "resources", "resources_mongod") {
			resourcesSpecPath := "config_spec.mongodb.mongod.resources"
			updatePath = append(updatePath, resourcesSpecPath)
		}
	}

	if sharded && d.HasChange("resources_mongos") {
		if compareResources(d, "resources", "resources_mongos") {
			resourcesSpecPath := "config_spec.mongodb.mongos.resources"
			updatePath = append(updatePath, resourcesSpecPath)
		}
	}

	if d.HasChange("disk_size_autoscaling_mongod") {
		dsaSpecPath := "config_spec.mongodb.mongod.disk_size_autoscaling"
		updatePath = append(updatePath, dsaSpecPath)
	}

	if sharded && d.HasChange("disk_size_autoscaling_mongoinfra") {
		dsaSpecPath := "config_spec.mongodb.mongoinfra.disk_size_autoscaling"
		updatePath = append(updatePath, dsaSpecPath)
	}
	if sharded && d.HasChange("disk_size_autoscaling_mongocfg") {
		dsaSpecPath := "config_spec.mongodb.mongocfg.disk_size_autoscaling"
		updatePath = append(updatePath, dsaSpecPath)
	}
	if sharded && d.HasChange("disk_size_autoscaling_mongos") {
		dsaSpecPath := "config_spec.mongodb.mongos.disk_size_autoscaling"
		updatePath = append(updatePath, dsaSpecPath)
	}

	if d.HasChange("cluster_config.0.mongod") {
		configSpecPath := "config_spec.mongodb.mongod.config"
		updatePath = append(updatePath, configSpecPath)
	}

	hostTypes := getSetOfHostTypes(d)
	_, hasMongoInfraHosts := hostTypes["MONGOINFRA"]

	if d.HasChange("cluster_config.0.mongos") {
		var configSpecPath string
		if hasMongoInfraHosts {
			configSpecPath = "config_spec.mongodb.mongoinfra.config_mongos"
		} else {
			configSpecPath = "config_spec.mongodb.mongos.config"
		}
		updatePath = append(updatePath, configSpecPath)
	}

	if d.HasChange("cluster_config.0.mongocfg") {
		var configSpecPath string
		if hasMongoInfraHosts {
			configSpecPath = "config_spec.mongodb.mongoinfra.config_mongocfg"
		} else {
			configSpecPath = "config_spec.mongodb.mongocfg.config"
		}
		updatePath = append(updatePath, configSpecPath)
	}

	if d.HasChange("maintenance_window") {
		mw, err := expandMongoDBMaintenanceWindow(d)
		if err != nil {
			return err
		}
		req.MaintenanceWindow = mw
		updatePath = append(updatePath, "maintenance_window")
	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = d.Get("deletion_protection").(bool)
		updatePath = append(updatePath, "deletion_protection")
	}
	if d.HasChange("config_spec.access") {
		if d.HasChange("config_spec.access.0.web_sql") {
			updatePath = append(updatePath, "config_spec.access.web_sql")
		}
		if d.HasChange("config_spec.access.0.data_lens") {
			updatePath = append(updatePath, "config_spec.access.data_lens")
		}
		if d.HasChange("config_spec.access.0.data_transfer") {
			updatePath = append(updatePath, "config_spec.access.data_transfer")
		}
	}
	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating MongoDB Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func updateMongodbClusterDatabases(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	currDBs, err := listMongodbDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandMongoDBDatabases(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mongodbDatabasesDiff(currDBs, targetDBs)

	for _, db := range toDelete {
		err := deleteMongoDBDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createMongoDBDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateMongodbClusterUsers(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	currUsers, err := listMongodbUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetUsers, err := expandMongoDBUserSpecs(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mongodbUsersDiff(currUsers, targetUsers)
	for _, u := range toDelete {
		err := deleteMongoDBUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}
	for _, u := range toAdd {
		err := createMongoDBUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	requests := mongodbCreateChangedUsersRequests(d)
	for _, u := range requests {
		err = updateMongoDBUser(ctx, config, u, d.Id())
		if err != nil {
			return err
		}
	}

	return nil
}

func updateMongoDBClusterHosts(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	currHosts, err := listMongodbHosts(ctx, config, d)
	if err != nil {
		return err
	}
	targetHosts, err := expandMongoDBHosts(d)
	if err != nil {
		return err
	}

	currHosts = sortMongoDBHosts(currHosts, targetHosts)

	toDelete, toAdd, toUpdate := mongodbHostsDiff(currHosts, targetHosts)

	for _, hs := range toAdd {
		for _, h := range hs {
			err = createMongoDBHost(ctx, config, d, h)
			if err != nil {
				return err
			}
		}

	}

	for _, hs := range toDelete {
		for _, h := range hs {
			err = deleteMongoDBHost(ctx, config, d, h)
			if err != nil {
				return err
			}
		}
	}

	for _, hs := range toUpdate {
		for _, h := range hs {
			err = updateMongoDBHost(ctx, config, d, h)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func updateMongoDBClusterShards(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	var currShards = make(map[string]struct{})
	var targetShards = make(map[string][]*mongodb.HostSpec)

	shards, err := listMongodbShards(ctx, config, d)
	if err != nil {
		return err
	}
	for _, shard := range shards {
		currShards[shard.Name] = struct{}{}
	}
	targetHosts, err := expandMongoDBHosts(d)
	if err != nil {
		return err
	}
	for _, h := range targetHosts {
		if h.Type == mongodb.Host_MONGOD {
			targetShards[h.ShardName] = append(targetShards[h.ShardName], h)
		}
	}

	if _, hasHostWithEmptyShard := targetShards[""]; len(targetShards) <= 2 && hasHostWithEmptyShard {
		targetShards[shards[0].Name] = append(targetShards[shards[0].Name], targetShards[""]...)
		delete(targetShards, "")
	}

	toDelete, toAdd := mongodbShardsDiff(currShards, targetShards)

	for _, shName := range toAdd {
		err = createMongoDBShard(ctx, config, d, shName, targetShards[shName])
		if err != nil {
			return err
		}
	}

	for _, shName := range toDelete {
		err = deleteMongoDBShard(ctx, config, d, shName)
		if err != nil {
			return err
		}
	}

	return nil
}

func createMongoDBDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Database().Create(ctx, &mongodb.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &mongodb.DatabaseSpec{
				Name: dbName,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteMongoDBDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Database().Delete(ctx, &mongodb.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMongoDBUser(ctx context.Context, config *Config, d *schema.ResourceData, user *mongodb.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().User().Create(ctx, &mongodb.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteMongoDBUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().User().Delete(ctx, &mongodb.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateMongoDBUser(ctx context.Context, config *Config, req *mongodb.UpdateUserRequest, cid string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().User().Update(ctx, req),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in MongoDB Cluster %q: %s", cid, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in MongoDB Cluster %q: %s", cid, err)
	}
	return nil
}

func createMongoDBHost(ctx context.Context, config *Config, d *schema.ResourceData, spec *mongodb.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().AddHosts(ctx, &mongodb.AddClusterHostsRequest{
			ClusterId: d.Id(),
			HostSpecs: []*mongodb.HostSpec{spec},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add host to MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding host to MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteMongoDBHost(ctx context.Context, config *Config, d *schema.ResourceData, fqdn string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().DeleteHosts(ctx, &mongodb.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: []string{fqdn},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateMongoDBHost(ctx context.Context, config *Config, d *schema.ResourceData, spec *mongodb.UpdateHostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().UpdateHosts(ctx, &mongodb.UpdateClusterHostsRequest{
			ClusterId:       d.Id(),
			UpdateHostSpecs: []*mongodb.UpdateHostSpec{spec},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update host in MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating host to MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMongoDBShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string, hosts []*mongodb.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().AddShard(ctx, &mongodb.AddClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: shardName,
			HostSpecs: hosts,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add shard to MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding shard to MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteMongoDBShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().DeleteShard(ctx, &mongodb.DeleteClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: shardName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete shard from MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting shard from MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func enableShardingMongoDB(ctx context.Context, config *Config, d *schema.ResourceData) error {
	_, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra := getResources(d)
	hosts, err := expandMongoDBHosts(d)
	if err != nil {
		return fmt.Errorf("Error while expanding hosts on enable sharding MongoDB Cluster : %s", err)
	}
	var hostsWithoutD []*mongodb.HostSpec
	for _, host := range hosts {
		if host.Type != mongodb.Host_MONGOD {
			hostsWithoutD = append(hostsWithoutD, host)
		}
	}
	var mongoinfra *mongodb.EnableClusterShardingRequest_MongoInfra
	var mongocfg *mongodb.EnableClusterShardingRequest_MongoCfg
	var mongos *mongodb.EnableClusterShardingRequest_Mongos
	if resourcesMongoInfra != nil {
		mongoinfra = &mongodb.EnableClusterShardingRequest_MongoInfra{
			Resources: resourcesMongoInfra,
		}
	}
	if resourcesMongos != nil {
		mongos = &mongodb.EnableClusterShardingRequest_Mongos{
			Resources: resourcesMongos,
		}
	}
	if resourcesMongoCfg != nil {
		mongocfg = &mongodb.EnableClusterShardingRequest_MongoCfg{
			Resources: resourcesMongoCfg,
		}
	}
	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().EnableSharding(ctx, &mongodb.EnableClusterShardingRequest{
		ClusterId:  d.Id(),
		Mongoinfra: mongoinfra,
		Mongos:     mongos,
		Mongocfg:   mongocfg,
		HostSpecs:  hostsWithoutD,
	}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to enable sharding MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while enabling sharding MongoDB Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func mongodbHostsDiff(currHosts []*mongodb.Host, targetHosts []*mongodb.HostSpec) (map[string][]string, map[string][]*mongodb.HostSpec, map[string][]*mongodb.UpdateHostSpec) {
	m := map[string][]*mongodb.HostSpec{}
	toUpdate := map[string][]*mongodb.UpdateHostSpec{}

	for _, h := range targetHosts {
		key := h.Type.String() + h.ZoneId + h.SubnetId
		if h.Type == mongodb.Host_MONGOD {
			key = key + h.ShardName
		}
		m[key] = append(m[key], h)
	}

	toDelete := map[string][]string{}
	for _, h := range currHosts {
		key := h.Type.String() + h.ZoneId + h.SubnetId
		if h.Type == mongodb.Host_MONGOD {
			key = key + h.ShardName
		}
		hs, ok := m[key]
		if !ok {
			toDelete[h.ShardName] = append(toDelete[h.ShardName], h.Name)
		} else if !matchHostParameters(hs[0], h) {
			updateSpec := getHostUpdateSpec(hs[0], h)
			if len(updateSpec.UpdateMask.Paths) > 0 {
				toUpdate[h.ShardName] = append(toUpdate[h.ShardName], updateSpec)
			}
		}

		if len(hs) > 1 {
			m[key] = hs[1:]
		} else {
			delete(m, key)
		}
	}

	toAdd := map[string][]*mongodb.HostSpec{}
	for _, hs := range m {
		for _, h := range hs {
			toAdd[h.ShardName] = append(toAdd[h.ShardName], h)
		}
	}

	return toDelete, toAdd, toUpdate
}

func getHostUpdateSpec(hostSpec *mongodb.HostSpec, host *mongodb.Host) *mongodb.UpdateHostSpec {
	result := &mongodb.UpdateHostSpec{
		HostName: host.Name,
	}
	updatePaths := make([]string, 0)

	if (host.HostParameters == nil && getOrDefault(hostSpec.Hidden, false)) || host.HostParameters.Hidden != getOrDefault(hostSpec.Hidden, false) {
		updatePaths = append(updatePaths, "hidden")
		result.Hidden = &wrapperspb.BoolValue{Value: getOrDefault(hostSpec.Hidden, false)}
	}

	if (host.HostParameters == nil && getOrDefault(hostSpec.Priority, 1.0) != 1.0) || host.HostParameters.Priority != getOrDefault(hostSpec.Priority, 1.0) {
		updatePaths = append(updatePaths, "priority")
		result.Priority = &wrapperspb.DoubleValue{Value: getOrDefault(hostSpec.Priority, 1.0)}
	}

	if (host.HostParameters == nil && getOrDefault(hostSpec.SecondaryDelaySecs, int64(0)) != 0) || host.HostParameters.SecondaryDelaySecs != getOrDefault(hostSpec.SecondaryDelaySecs, int64(0)) {
		updatePaths = append(updatePaths, "secondary_delay_secs")
		result.SecondaryDelaySecs = &wrapperspb.Int64Value{Value: getOrDefault(hostSpec.SecondaryDelaySecs, int64(0))}
	}

	if (host.HostParameters == nil && len(hostSpec.Tags) != 0) || !maps.Equal(host.HostParameters.Tags, hostSpec.Tags) {
		updatePaths = append(updatePaths, "tags")
		result.Tags = hostSpec.Tags
	}

	if hostSpec.AssignPublicIp != host.AssignPublicIp {
		updatePaths = append(updatePaths, "assign_public_ip")
		result.AssignPublicIp = hostSpec.AssignPublicIp
	}

	result.UpdateMask = &field_mask.FieldMask{Paths: updatePaths}

	return result
}

func mongodbShardsDiff(currShards map[string]struct{}, targetShards map[string][]*mongodb.HostSpec) ([]string, []string) {
	var toDelete []string
	for shardName := range currShards {
		if _, ok := targetShards[shardName]; !ok {
			toDelete = append(toDelete, shardName)
		}
	}
	var toAdd []string
	for shardName := range targetShards {
		if _, ok := currShards[shardName]; !ok {
			toAdd = append(toAdd, shardName)
		}
	}

	return toDelete, toAdd
}

func mongodbUsersDiff(currUsers []*mongodb.User, targetUsers []*mongodb.UserSpec) ([]string, []*mongodb.UserSpec) {
	m := map[string]bool{}
	toDelete := map[string]bool{}
	var toAdd []*mongodb.UserSpec

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

	var toDel []string
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func mongodbCreateChangedUsersRequests(d *schema.ResourceData) []*mongodb.UpdateUserRequest {
	oldSpecs, newSpecs := d.GetChange("user")

	var result []*mongodb.UpdateUserRequest
	m := map[string]*mongodb.UserSpec{}
	for _, spec := range oldSpecs.(*schema.Set).List() {
		user := expandMongoDBUser(spec.(map[string]interface{}))
		m[user.Name] = user
	}
	for _, spec := range newSpecs.(*schema.Set).List() {
		user := expandMongoDBUser(spec.(map[string]interface{}))
		if u, ok := m[user.Name]; ok {
			var updatePaths []string
			if user.Password != u.Password {
				updatePaths = append(updatePaths, "password")
			}
			if fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				updatePaths = append(updatePaths, "permissions")
			}

			if len(updatePaths) == 0 {
				continue
			}
			req := &mongodb.UpdateUserRequest{
				ClusterId:   d.Id(),
				UserName:    user.Name,
				Password:    user.Password,
				Permissions: user.Permissions,
				UpdateMask:  &fieldmaskpb.FieldMask{Paths: updatePaths},
			}
			result = append(result, req)
		}
	}
	return result
}

func mapContainsOneOfKeys(set map[string]struct{}, keys []string) bool {
	for _, key := range keys {
		if _, ok := set[key]; ok {
			return true
		}
	}
	return false
}

// func to compare resources for migrate from resources to resources_*
func compareResources(d *schema.ResourceData, cfg1, cfg2 string) bool {
	if d.HasChange(cfg1) != d.HasChange(cfg2) {
		return true
	}
	cfg1Delete, cfg1Add := d.GetChange(cfg1)
	cfg2Delete, cfg2Add := d.GetChange(cfg2)

	var resourcesOld, resourcesNew map[string]interface{}
	if len(cfg1Delete.([]interface{})) == 1 && len(cfg2Add.([]interface{})) == 1 {
		resourcesOld = cfg1Delete.([]interface{})[0].(map[string]interface{})
		resourcesNew = cfg2Add.([]interface{})[0].(map[string]interface{})
	} else if len(cfg2Delete.([]interface{})) == 1 && len(cfg1Add.([]interface{})) == 1 {
		resourcesOld = cfg2Delete.([]interface{})[0].(map[string]interface{})
		resourcesNew = cfg1Add.([]interface{})[0].(map[string]interface{})
	} else {
		return true
	}

	if len(resourcesOld) != len(resourcesNew) {
		return true
	}

	for k, vOld := range resourcesOld {
		vNew := resourcesNew[k]
		if vNew != vOld {
			return true
		}
	}
	return false
}

func setMongoDBFolderID(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	cluster, err := config.sdk.MDB().MongoDB().Cluster().Get(ctx, &mongodb.GetClusterRequest{
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
		request := &mongodb.MoveClusterRequest{
			ClusterId:           d.Id(),
			DestinationFolderId: folderID.(string),
		}
		op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending MongoDB cluster move request: %+v", request)
			return config.sdk.MDB().MongoDB().Cluster().Move(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("error while requesting API to move MongoDB Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while moving MongoDB Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("moving MongoDB Cluster %q to folder %v failed: %s", d.Id(), folderID, err)
		}

	}

	return nil
}

// to import users. maybe change tests for this
func resourceYandexMDBMongodbClusterImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx = context.WithValue(ctx, ReadModeKey, true)
	_ = resourceYandexMDBMongodbClusterRead(ctx, d, m)
	return []*schema.ResourceData{d}, nil
}
