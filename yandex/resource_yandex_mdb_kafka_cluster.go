package yandex

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexMDBKafkaClusterCreateTimeout = 60 * time.Minute
	yandexMDBKafkaClusterReadTimeout   = 5 * time.Minute
	yandexMDBKafkaClusterDeleteTimeout = 60 * time.Minute
	yandexMDBKafkaClusterUpdateTimeout = 60 * time.Minute
)

func resourceYandexMDBKafkaCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).\n\n~> Historically, `topic` blocks of the `yandex_mdb_kafka_cluster` resource were used to manage topics of the Kafka cluster. However, this approach has a number of disadvantages. In particular, when adding and removing topics from the tf recipe, terraform generates a diff that misleads the user about the planned changes. Also, this approach turned out to be inconvenient when managing topics through the Kafka Admin API. Therefore, topic management through a separate resource type `yandex_mdb_kafka_topic` was implemented and is now recommended.\n",

		Create: resourceYandexMDBKafkaClusterCreate,
		Read:   resourceYandexMDBKafkaClusterRead,
		Update: resourceYandexMDBKafkaClusterUpdate,
		Delete: resourceYandexMDBKafkaClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBKafkaClusterCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBKafkaClusterReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBKafkaClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBKafkaClusterDeleteTimeout),
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
			"config": {
				Type:        schema.TypeList,
				Description: "Configuration of the Kafka cluster.",
				Required:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterConfig(),
			},
			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the Kafka cluster. Can be either `PRESTABLE` or `PRODUCTION`. The default is `PRODUCTION`.",
				Optional:     true,
				ForceNew:     true,
				Default:      "PRODUCTION",
				ValidateFunc: validateParsableValue(parseKafkaEnv),
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
			},
			"subnet_ids": {
				Type:        schema.TypeList,
				Description: common.ResourceDescriptions["subnet_ids"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"topic": {
				Type:        schema.TypeList,
				Description: "To manage topics, please switch to using a separate resource type `yandex_mdb_kafka_topic`.",
				Optional:    true,
				Elem:        resourceYandexMDBKafkaClusterTopicBlock(),
				Deprecated:  useResourceInstead("topic", "yandex_mdb_kafka_topic"),
			},
			"user": {
				Type:        schema.TypeSet,
				Description: "To manage users, please switch to using a separate resource type `yandex_mdb_kafka_user`.",
				Optional:    true,
				Set:         kafkaUserHash,
				Elem:        resourceYandexMDBKafkaClusterUserBlock(),
				Deprecated:  useResourceInstead("user", "yandex_mdb_kafka_user"),
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: "A list of IDs of the host groups to place VMs of the cluster on.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"host": {
				Type:        schema.TypeSet,
				Description: "A host of the Kafka cluster.",
				Computed:    true,
				Set:         kafkaHostHash,
				Elem:        resourceYandexMDBKafkaHost(),
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-kafka/api-ref/Cluster/).",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-kafka/api-ref/Cluster/).",
				Computed:    true,
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: "Maintenance policy of the Kafka cluster.",
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
							ValidateFunc: kafkaMaintenanceWindowSchemaValidateFunc,
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
		},
	}
}

func resourceYandexMDBKafkaClusterConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"version": {
				Type:        schema.TypeString,
				Description: "Version of the Kafka server software.",
				Required:    true,
			},
			"zones": {
				Type:        schema.TypeList,
				Description: "List of availability zones.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
			},
			"kafka": {
				Type:        schema.TypeList,
				Description: "Configuration of the Kafka subcluster.",
				Required:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterKafkaConfig(),
			},
			"brokers_count": {
				Type:        schema.TypeInt,
				Description: "Count of brokers per availability zone. The default is `1`.",
				Optional:    true,
				Default:     1,
			},
			"assign_public_ip": {
				Type:        schema.TypeBool,
				Description: "Determines whether each broker will be assigned a public IP address. The default is `false`.",
				Optional:    true,
				Default:     false,
			},
			"unmanaged_topics": {
				Type:        schema.TypeBool,
				Description: "",
				Optional:    true,
				Default:     false,
				Deprecated:  "The 'unmanaged_topics' field has been deprecated, because feature enabled permanently and can't be disabled.",
			},
			"schema_registry": {
				Type:        schema.TypeBool,
				Description: "Enables managed schema registry on cluster. The default is `false`.",
				Optional:    true,
				Default:     false,
			},
			"zookeeper": {
				Type:        schema.TypeList,
				Description: "Configuration of the ZooKeeper subcluster.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterZookeeperConfig(),
			},
			"kraft": {
				Type:        schema.TypeList,
				Description: "Configuration of the KRaft-controller subcluster.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterKRaftControllerConfig(),
			},
			"disk_size_autoscaling": {
				Type:        schema.TypeList,
				Description: "Disk autoscaling settings of the Kafka cluster.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size_limit": {
							Type:        schema.TypeInt,
							Description: "Maximum possible size of disk in bytes.",
							Required:    true,
						},
						"planned_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Percent of disk utilization. During maintenance disk will autoscale, if this threshold reached. Value is between 0 and 100. Default value is 0 (autoscaling disabled).",
							Optional:    true,
						},
						"emergency_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Percent of disk utilization. Disk will autoscale immediately, if this threshold reached. Value is between 0 and 100. Default value is 0 (autoscaling disabled). Must be not less then 'planned_usage_threshold' value.",
							Optional:    true,
						},
					},
				},
			},
			"access": {
				Type:        schema.TypeList,
				Description: "Access policy to the Kafka cluster.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_transfer": {
							Type:        schema.TypeBool,
							Description: "Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer).",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"rest_api": {
				Type:        schema.TypeList,
				Description: "REST API settings of the Kafka cluster.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enables REST API on cluster. The default is `false`.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
		},
	}
}

func resourceYandexMDBKafkaClusterResources() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_preset_id": {
				Type:        schema.TypeString,
				Description: "The ID of the preset for computational resources available to a Kafka host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).",
				Required:    true,
			},
			"disk_size": {
				Type:        schema.TypeInt,
				Description: "Volume of the storage available to a Kafka host, in gigabytes.",
				Required:    true,
			},
			"disk_type_id": {
				Type:        schema.TypeString,
				Description: "Type of the storage of Kafka hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/storage).",
				Required:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaZookeeperResources() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_preset_id": {
				Type:        schema.TypeString,
				Description: "The ID of the preset for computational resources available to a ZooKeeper host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).",
				Optional:    true,
				Computed:    true,
			},
			"disk_size": {
				Type:        schema.TypeInt,
				Description: "Volume of the storage available to a ZooKeeper host, in gigabytes.",
				Optional:    true,
				Computed:    true,
			},
			"disk_type_id": {
				Type:        schema.TypeString,
				Description: "Type of the storage of ZooKeeper hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/storage).",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaKRaftControllerResources() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_preset_id": {
				Type:        schema.TypeString,
				Description: "The ID of the preset for computational resources available to a KRaft-controller host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).",
				Optional:    true,
				Computed:    true,
			},
			"disk_size": {
				Type:        schema.TypeInt,
				Description: "Volume of the storage available to a KRaft-controller host, in gigabytes.",
				Optional:    true,
				Computed:    true,
			},
			"disk_type_id": {
				Type:        schema.TypeString,
				Description: "Type of the storage of KRaft-controller hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/storage).",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterTopicBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the topic.",
				Required:    true,
			},
			"partitions": {
				Type:        schema.TypeInt,
				Description: "The number of the topic's partitions.",
				Required:    true,
			},
			"replication_factor": {
				Type:        schema.TypeInt,
				Description: "Amount of data copies (replicas) for the topic in the cluster.",
				Required:    true,
			},
			"topic_config": {
				Type:        schema.TypeList,
				Description: "User-defined settings for the topic. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/operations/cluster-topics#update-topic) and [the Kafka documentation](https://kafka.apache.org/documentation/#configuration).",
				Optional:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterTopicConfig(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterUserBlock() *schema.Resource {
	return &schema.Resource{
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
				Set:         kafkaUserPermissionHash,
				Elem:        resourceYandexMDBKafkaPermission(),
			},
		},
	}
}

func resourceYandexMDBKafkaPermission() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"topic_name": {
				Type:        schema.TypeString,
				Description: "The name of the topic that the permission grants access to.",
				Required:    true,
			},
			"role": {
				Type:        schema.TypeString,
				Description: "The role type to grant to the topic.",
				Required:    true,
			},
			"allow_hosts": {
				Type:        schema.TypeSet,
				Description: "Set of hosts, to which this permission grants access to. Only ip-addresses allowed as value of single host.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterKafkaConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the Kafka subcluster.",
				Required:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterResources(),
			},
			"kafka_config": {
				Type:        schema.TypeList,
				Description: "User-defined settings for the Kafka cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/operations/cluster-update) and [the Kafka documentation](https://kafka.apache.org/documentation/#configuration).",
				Optional:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterKafkaSettings(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterKafkaSettings() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compression_type": {
				Type:         schema.TypeString,
				Description:  "Compression type of kafka topics.",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaCompression),
			},
			"log_flush_interval_messages": {
				Type:         schema.TypeString,
				Description:  "The number of messages accumulated on a log partition before messages are flushed to disk.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_flush_interval_ms": {
				Type:         schema.TypeString,
				Description:  "The maximum time in ms that a message in any topic is kept in memory before flushed to disk. If not set, the value in log.flush.scheduler.interval.ms is used.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_flush_scheduler_interval_ms": {
				Type:         schema.TypeString,
				Description:  "The frequency in ms that the log flusher checks whether any log needs to be flushed to disk.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_bytes": {
				Type:         schema.TypeString,
				Description:  "The maximum size of the log before deleting it.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_hours": {
				Type:         schema.TypeString,
				Description:  "The number of hours to keep a log file before deleting it (in hours), tertiary to log.retention.ms property.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_minutes": {
				Type:         schema.TypeString,
				Description:  "The number of minutes to keep a log file before deleting it (in minutes), secondary to log.retention.ms property. If not set, the value in log.retention.hours is used.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_ms": {
				Type:         schema.TypeString,
				Description:  "The number of milliseconds to keep a log file before deleting it (in milliseconds), If not set, the value in log.retention.minutes is used. If set to -1, no time limit is applied.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_segment_bytes": {
				Type:         schema.TypeString,
				Description:  "The maximum size of a single log file.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_preallocate": {
				Type:        schema.TypeBool,
				Description: "Should pre allocate file when create new segment?",
				Optional:    true,
				Deprecated:  "The 'log_preallocate' field has been deprecated, because feature not useful for Yandex Cloud.",
			},
			"socket_send_buffer_bytes": {
				Type:         schema.TypeString,
				Description:  "The SO_SNDBUF buffer of the socket server sockets. If the value is -1, the OS default will be used.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"socket_receive_buffer_bytes": {
				Type:         schema.TypeString,
				Description:  "The SO_RCVBUF buffer of the socket server sockets. If the value is -1, the OS default will be used.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"auto_create_topics_enable": {
				Type:        schema.TypeBool,
				Description: "Enable auto creation of topic on the server.",
				Optional:    true,
			},
			"num_partitions": {
				Type:         schema.TypeString,
				Description:  "The default number of log partitions per topic.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"default_replication_factor": {
				Type:         schema.TypeString,
				Description:  "The replication factor for automatically created topics, and for topics created with -1 as the replication factor.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"message_max_bytes": {
				Type:         schema.TypeString,
				Description:  "The largest record batch size allowed by Kafka (after compression if compression is enabled).",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"replica_fetch_max_bytes": {
				Type:         schema.TypeString,
				Description:  "The number of bytes of messages to attempt to fetch for each partition.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"ssl_cipher_suites": {
				Type:        schema.TypeSet,
				Description: "A list of cipher suites.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
			},
			"offsets_retention_minutes": {
				Type:         schema.TypeString,
				Description:  "For subscribed consumers, committed offset of a specific partition will be expired and discarded after this period of time.",
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"sasl_enabled_mechanisms": {
				Type:        schema.TypeSet,
				Description: "The list of SASL mechanisms enabled in the Kafka server.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateParsableValue(parseKafkaSaslMechanism),
				},
				Set:      schema.HashString,
				Optional: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterTopicConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cleanup_policy": {
				Type:         schema.TypeString,
				Description:  "Retention policy to use on log segments.",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaTopicCleanupPolicy),
			},
			"compression_type": {
				Type:         schema.TypeString,
				Description:  "Compression type of kafka topic.",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaCompression),
			},
			"delete_retention_ms": {
				Type:         schema.TypeString,
				Description:  "The amount of time to retain delete tombstone markers for log compacted topics.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"file_delete_delay_ms": {
				Type:         schema.TypeString,
				Description:  "The time to wait before deleting a file from the filesystem.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"flush_messages": {
				Type:         schema.TypeString,
				Description:  "This setting allows specifying an interval at which we will force an fsync of data written to the log.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"flush_ms": {
				Type:         schema.TypeString,
				Description:  "This setting allows specifying a time interval at which we will force an fsync of data written to the log.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"min_compaction_lag_ms": {
				Type:         schema.TypeString,
				Description:  "The minimum time a message will remain uncompacted in the log. Only applicable for logs that are being compacted.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"retention_bytes": {
				Type:         schema.TypeString,
				Description:  "This configuration controls the maximum size a partition (which consists of log segments) can grow to before we will discard old log segments to free up space if we are using the \"delete\" retention policy.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"retention_ms": {
				Type:         schema.TypeString,
				Description:  "This configuration controls the maximum time we will retain a log before we will discard old log segments to free up space if we are using the \"delete\" retention policy.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"max_message_bytes": {
				Type:         schema.TypeString,
				Description:  "The largest record batch size allowed by Kafka (after compression if compression is enabled).",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"min_insync_replicas": {
				Type:         schema.TypeString,
				Description:  "When a producer sets acks to \"all\" (or \"-1\"), this configuration specifies the minimum number of replicas that must acknowledge a write for the write to be considered successful. ",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"segment_bytes": {
				Type:         schema.TypeString,
				Description:  "This configuration controls the segment file size for the log.",
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"preallocate": {
				Type:        schema.TypeBool,
				Description: "True if we should preallocate the file on disk when creating a new log segment.",
				Optional:    true,
				Deprecated:  "The 'preallocate' field has been deprecated, because feature not useful for Yandex Cloud.",
			},
		},
	}
}

func resourceYandexMDBKafkaClusterZookeeperConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the ZooKeeper subcluster.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaZookeeperResources(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterKRaftControllerConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the KRaft-controller subcluster.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaKRaftControllerResources(),
			},
		},
	}
}

func resourceYandexMDBKafkaHost() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The fully qualified domain name of the host.",
				Computed:    true,
			},
			"zone_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
			},
			"role": {
				Type:        schema.TypeString,
				Description: "Role of the host in the cluster.",
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: "Health of the host.",
				Computed:    true,
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Description: "The ID of the subnet, to which the host belongs.",
				Computed:    true,
			},
			"assign_public_ip": {
				Type:        schema.TypeBool,
				Description: "The flag that defines whether a public IP address is assigned to the node.",
				Computed:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareKafkaCreateRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	log.Printf("[DEBUG] Creating Kafka cluster: %+v", req)

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Kafka().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Kafka Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while getting Kafka create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*kafka.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create Kafka Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("kafka cluster creation failed: %s", err)
	}
	log.Printf("[DEBUG] Finished creating Kafka cluster %q", md.ClusterId)

	return resourceYandexMDBKafkaClusterRead(d, meta)
}

// Returns request for creating the Cluster.
func prepareKafkaCreateRequest(d *schema.ResourceData, meta *Config) (*kafka.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Kafka Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting folder ID while creating Kafka Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseKafkaEnv(e)
	if err != nil {
		return nil, fmt.Errorf("error resolving environment while creating Kafka Cluster: %s", err)
	}

	configSpec, err := expandKafkaConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding configuration on Kafka Cluster create: %s", err)
	}

	subnets := []string{}
	if v, ok := d.GetOk("subnet_ids"); ok {
		for _, subnet := range v.([]interface{}) {
			subnets = append(subnets, subnet.(string))
		}
	}

	topicSpecs, err := expandKafkaTopics(d)
	if err != nil {
		return nil, err
	}

	userSpecs, err := expandKafkaUsers(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding users on Kafka Cluster create: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))
	hostGroupIds := expandHostGroupIds(d.Get("host_group_ids"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on Kafka Cluster create: %s", err)
	}

	maintenanceWindow, err := expandKafkaMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding maintenance window settings on Kafka Cluster create: %s", err)
	}

	req := kafka.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Environment:        env,
		ConfigSpec:         configSpec,
		Labels:             labels,
		SubnetId:           subnets,
		TopicSpecs:         topicSpecs,
		UserSpecs:          userSpecs,
		SecurityGroupIds:   securityGroupIds,
		HostGroupIds:       hostGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
		MaintenanceWindow:  maintenanceWindow,
	}
	return &req, nil
}

func resourceYandexMDBKafkaClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Kafka().Cluster().Get(ctx, &kafka.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)

	cfg, err := flattenKafkaConfig(d, cluster)
	if err != nil {
		return err
	}
	if err := d.Set("config", cfg); err != nil {
		return err
	}

	stateTopics := d.Get("topic").([]interface{})
	if len(stateTopics) == 0 {
		if err := d.Set("topic", []map[string]interface{}{}); err != nil {
			return err
		}
	} else {
		topics, err := listKafkaTopics(ctx, config, d.Id())
		if err != nil {
			return err
		}

		topicSpecs, err := expandKafkaTopics(d)
		if err != nil {
			return err
		}
		sortKafkaTopics(topics, topicSpecs)

		if err := d.Set("topic", flattenKafkaTopics(topics)); err != nil {
			return err
		}
	}

	if len(d.Get("user").(*schema.Set).List()) == 0 {
		if err := d.Set("user", schema.NewSet(kafkaUserHash, nil)); err != nil {
			return err
		}
	} else {
		dUsers, err := expandKafkaUsers(d)
		if err != nil {
			return err
		}
		passwords := kafkaUsersPasswords(dUsers)

		users, err := listKafkaUsers(ctx, config, d.Id())
		if err != nil {
			return err
		}
		if err := d.Set("user", flattenKafkaUsers(users, passwords)); err != nil {
			return err
		}
	}

	hosts, err := listKafkaHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	if err := d.Set("host", flattenKafkaHosts(hosts)); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	if err := d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	maintenanceWindow, err := flattenKafkaMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}
	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
		return err
	}

	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBKafkaClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Kafka Cluster %q", d.Id())

	d.Partial(true)

	if err := setKafkaFolderID(d, meta); err != nil {
		return err
	}

	if err := updateKafkaClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("topic") {
		topicModifier := NewKafkaTopicManager(meta.(*Config))
		if err := updateKafkaClusterTopics(d, topicModifier); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updateKafkaClusterUsers(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)

	log.Printf("[DEBUG] Finished updating Kafka Cluster %q", d.Id())
	return resourceYandexMDBKafkaClusterRead(d, meta)
}

func resourceYandexMDBKafkaClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Kafka Cluster %q", d.Id())

	req := &kafka.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Kafka().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kafka Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Kafka Cluster %q", d.Id())
	return nil
}

func listKafkaTopics(ctx context.Context, config *Config, id string) ([]*kafka.Topic, error) {
	ret := []*kafka.Topic{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Kafka().Topic().List(ctx, &kafka.ListTopicsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of topics for '%s': %s", id, err)
		}
		ret = append(ret, resp.Topics...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return ret, nil
}

func listKafkaUsers(ctx context.Context, config *Config, id string) ([]*kafka.User, error) {
	ret := []*kafka.User{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Kafka().User().List(ctx, &kafka.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of users for '%s': %s", id, err)
		}
		ret = append(ret, resp.Users...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return ret, nil
}

func listKafkaHosts(ctx context.Context, config *Config, id string) ([]*kafka.Host, error) {
	ret := []*kafka.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Kafka().Cluster().ListHosts(ctx, &kafka.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of hosts for '%s': %s", id, err)
		}
		ret = append(ret, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret, nil
}

var mdbKafkaUpdateFieldsMap = map[string]string{
	"name":                      "name",
	"description":               "description",
	"labels":                    "labels",
	"network_id":                "network_id",
	"security_group_ids":        "security_group_ids",
	"deletion_protection":       "deletion_protection",
	"maintenance_window":        "maintenance_window",
	"subnet_ids":                "subnet_ids",
	"config.0.zones":            "config_spec.zone_id",
	"config.0.version":          "config_spec.version",
	"config.0.brokers_count":    "config_spec.brokers_count",
	"config.0.assign_public_ip": "config_spec.assign_public_ip",
	"config.0.schema_registry":  "config_spec.schema_registry",
	"config.0.access":           "config_spec.access",
	"config.0.disk_size_autoscaling.0.disk_size_limit":                "config_spec.disk_size_autoscaling.disk_size_limit",
	"config.0.disk_size_autoscaling.0.planned_usage_threshold":        "config_spec.disk_size_autoscaling.planned_usage_threshold",
	"config.0.disk_size_autoscaling.0.emergency_usage_threshold":      "config_spec.disk_size_autoscaling.emergency_usage_threshold",
	"config.0.kafka.0.resources.0.resource_preset_id":                 "config_spec.kafka.resources.resource_preset_id",
	"config.0.kafka.0.resources.0.disk_type_id":                       "config_spec.kafka.resources.disk_type_id",
	"config.0.kafka.0.resources.0.disk_size":                          "config_spec.kafka.resources.disk_size",
	"config.0.kafka.0.kafka_config.0.compression_type":                "config_spec.kafka.kafka_config_{version}.compression_type",
	"config.0.kafka.0.kafka_config.0.log_flush_interval_messages":     "config_spec.kafka.kafka_config_{version}.log_flush_interval_messages",
	"config.0.kafka.0.kafka_config.0.log_flush_interval_ms":           "config_spec.kafka.kafka_config_{version}.log_flush_interval_ms",
	"config.0.kafka.0.kafka_config.0.log_flush_scheduler_interval_ms": "config_spec.kafka.kafka_config_{version}.log_flush_scheduler_interval_ms",
	"config.0.kafka.0.kafka_config.0.log_retention_bytes":             "config_spec.kafka.kafka_config_{version}.log_retention_bytes",
	"config.0.kafka.0.kafka_config.0.log_retention_hours":             "config_spec.kafka.kafka_config_{version}.log_retention_hours",
	"config.0.kafka.0.kafka_config.0.log_retention_minutes":           "config_spec.kafka.kafka_config_{version}.log_retention_minutes",
	"config.0.kafka.0.kafka_config.0.log_retention_ms":                "config_spec.kafka.kafka_config_{version}.log_retention_ms",
	"config.0.kafka.0.kafka_config.0.log_segment_bytes":               "config_spec.kafka.kafka_config_{version}.log_segment_bytes",
	"config.0.kafka.0.kafka_config.0.socket_send_buffer_bytes":        "config_spec.kafka.kafka_config_{version}.socket_send_buffer_bytes",
	"config.0.kafka.0.kafka_config.0.socket_receive_buffer_bytes":     "config_spec.kafka.kafka_config_{version}.socket_receive_buffer_bytes",
	"config.0.kafka.0.kafka_config.0.auto_create_topics_enable":       "config_spec.kafka.kafka_config_{version}.auto_create_topics_enable",
	"config.0.kafka.0.kafka_config.0.num_partitions":                  "config_spec.kafka.kafka_config_{version}.num_partitions",
	"config.0.kafka.0.kafka_config.0.default_replication_factor":      "config_spec.kafka.kafka_config_{version}.default_replication_factor",
	"config.0.kafka.0.kafka_config.0.message_max_bytes":               "config_spec.kafka.kafka_config_{version}.message_max_bytes",
	"config.0.kafka.0.kafka_config.0.replica_fetch_max_bytes":         "config_spec.kafka.kafka_config_{version}.replica_fetch_max_bytes",
	"config.0.kafka.0.kafka_config.0.ssl_cipher_suites":               "config_spec.kafka.kafka_config_{version}.ssl_cipher_suites",
	"config.0.kafka.0.kafka_config.0.offsets_retention_minutes":       "config_spec.kafka.kafka_config_{version}.offsets_retention_minutes",
	"config.0.kafka.0.kafka_config.0.sasl_enabled_mechanisms":         "config_spec.kafka.kafka_config_{version}.sasl_enabled_mechanisms",
	"config.0.zookeeper.0.resources.0.resource_preset_id":             "config_spec.zookeeper.resources.resource_preset_id",
	"config.0.zookeeper.0.resources.0.disk_type_id":                   "config_spec.zookeeper.resources.disk_type_id",
	"config.0.zookeeper.0.resources.0.disk_size":                      "config_spec.zookeeper.resources.disk_size",
	"config.0.kraft.0.resources.0.resource_preset_id":                 "config_spec.kraft.resources.resource_preset_id",
	"config.0.kraft.0.resources.0.disk_type_id":                       "config_spec.kraft.resources.disk_type_id",
	"config.0.kraft.0.resources.0.disk_size":                          "config_spec.kraft.resources.disk_size",
	"config.0.rest_api":                                               "config_spec.rest_api_config.enabled",
}

func kafkaClusterUpdateRequest(d *schema.ResourceData, config *Config) (*kafka.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating Kafka cluster: %s", err)
	}

	configSpec, err := expandKafkaConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding configSpec while updating Kafka cluster: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, config)
	if err != nil {
		return nil, fmt.Errorf("error expanding network_id settings while updating Kafka cluster: %s", err)
	}

	maintenanceWindow, err := expandKafkaMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding maintenance window settings while updating Kafka cluster: %s", err)
	}

	var subnets []string
	if v, ok := d.GetOk("subnet_ids"); ok {
		for _, subnet := range v.([]interface{}) {
			subnets = append(subnets, subnet.(string))
		}
	}

	req := &kafka.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		NetworkId:          networkID,
		ConfigSpec:         configSpec,
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		DeletionProtection: d.Get("deletion_protection").(bool),
		SubnetIds:          subnets,
		MaintenanceWindow:  maintenanceWindow,
	}
	return req, nil
}

func kafkaClusterUpdateRequestWithMask(d *schema.ResourceData, config *Config) (*kafka.UpdateClusterRequest, error) {
	req, err := kafkaClusterUpdateRequest(d, config)
	if err != nil {
		return nil, err
	}

	updatePath := []string{}
	for field, path := range mdbKafkaUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, strings.Replace(path, "{version}", getSuffixVersion(d), -1))
		}
	}

	if len(updatePath) == 0 {
		return nil, nil
	}

	sort.Strings(updatePath)

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	return req, nil
}

func getSuffixVersion(d *schema.ResourceData) string {
	version := d.Get("config.0.version").(string)
	result := "3"
	if strings.HasPrefix(version, "2") {
		result = strings.Replace(version, ".", "_", -1)
	}
	return result
}

func updateKafkaClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := kafkaClusterUpdateRequestWithMask(d, config)
	if err != nil {
		return err
	}
	if req == nil {
		return nil
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending Kafka cluster update request: %+v", req)
		return config.sdk.MDB().Kafka().Cluster().Update(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to update Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating Kafka Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func updateKafkaClusterTopics(d *schema.ResourceData, topicModifier KafkaTopicModifier) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	versionI, ok := d.GetOk("config.0.version")
	if !ok {
		return fmt.Errorf("you must specify version of Kafka")
	}
	version := versionI.(string)

	diffByTopicName := diffByEntityKey(d, "topic", "name")
	for topicName, topicDiff := range diffByTopicName {
		if topicDiff.OldEntity == nil {
			topicSpec, err := buildKafkaTopicSpec(d, fmt.Sprintf("%s.", topicDiff.NewEntityKey), version)
			if err != nil {
				return err
			}
			log.Printf("[DEBUG] Creating topic %+v", topicSpec)
			if err := topicModifier.CreateKafkaTopic(ctx, d, topicSpec); err != nil {
				return err
			}
			continue
		}

		if topicDiff.NewEntity == nil {
			log.Printf("[DEBUG] Topic %s is to be deleted", topicName)
			if err := topicModifier.DeleteKafkaTopic(ctx, d, topicName); err != nil {
				return err
			}
			continue
		}

		if !reflect.DeepEqual(topicDiff.OldEntity, topicDiff.NewEntity) {
			topicSpec, err := buildKafkaTopicSpec(d, fmt.Sprintf("%s.", topicDiff.NewEntityKey), version)
			if err != nil {
				return err
			}
			paths := kafkaTopicUpdateMask(topicDiff.OldEntity, topicDiff.NewEntity, getSuffixVersion(d))
			if err := topicModifier.UpdateKafkaTopic(ctx, d, topicSpec, paths); err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteKafkaTopic(ctx context.Context, config *Config, d *schema.ResourceData, topicName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().Topic().Delete(ctx, &kafka.DeleteTopicRequest{
			ClusterId: d.Id(),
			TopicName: topicName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete topic from Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting topic from Kafka Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createKafkaTopic(ctx context.Context, config *Config, d *schema.ResourceData, topicSpec *kafka.TopicSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().Topic().Create(ctx, &kafka.CreateTopicRequest{
			ClusterId: d.Id(),
			TopicSpec: topicSpec,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create topic in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding topic to Kafka Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateKafkaTopic(ctx context.Context, config *Config, d *schema.ResourceData, topicSpec *kafka.TopicSpec, paths []string) error {
	request := &kafka.UpdateTopicRequest{
		ClusterId:  d.Id(),
		TopicName:  topicSpec.Name,
		TopicSpec:  topicSpec,
		UpdateMask: &field_mask.FieldMask{Paths: paths},
	}

	log.Printf("[DEBUG] Sending topic update request: %+v", request)

	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().Topic().Update(ctx, request),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update topic in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating topic in Kafka Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func updateKafkaClusterUsers(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currUsers, err := listKafkaUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetUsers, err := expandKafkaUsers(d)
	if err != nil {
		return err
	}
	toDelete, toAdd := kafkaUsersDiff(currUsers, targetUsers)

	for _, user := range toDelete {
		err := deleteKafkaUser(ctx, config, d, user)
		if err != nil {
			return err
		}
	}
	for _, user := range toAdd {
		err := createKafkaUser(ctx, config, d, user)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")
	err = updateKafkaUsers(ctx, config, d, oldSpecs.(*schema.Set), newSpecs.(*schema.Set))
	if err != nil {
		return err
	}

	return nil
}

func deleteKafkaUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	req := &kafka.DeleteUserRequest{
		ClusterId: d.Id(),
		UserName:  userName,
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Deleting Kafka user %q within cluster %q", userName, d.Id())
		return config.sdk.MDB().Kafka().User().Delete(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from Kafka Cluster %q: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Finished deleting Kafka user %q", userName)
	return nil
}

func createKafkaUser(ctx context.Context, config *Config, d *schema.ResourceData, userSpec *kafka.UserSpec) error {
	req := &kafka.CreateUserRequest{
		ClusterId: d.Id(),
		UserSpec:  userSpec,
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Creating Kafka user %q: %+v", userSpec.Name, req)
		return config.sdk.MDB().Kafka().User().Create(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to create user %q in Kafka Cluster %q: %s", userSpec.Name, d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for Kafka user %q in cluster %q create operation: %s", userSpec.Name, d.Id(), err)
	}
	if _, err = op.Response(); err != nil {
		return fmt.Errorf("kafka user %q creation failed in cluster %q: %s", userSpec.Name, d.Id(), err)
	}
	log.Printf("[DEBUG] Finished creating Kafka user %q", userSpec.Name)
	return nil
}

func updateKafkaUsers(ctx context.Context, config *Config, d *schema.ResourceData, oldSpecs *schema.Set, newSpecs *schema.Set) error {
	m := map[string]*kafka.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user, err := expandKafkaUser(spec.(map[string]interface{}))
		if err != nil {
			return err
		}
		m[user.Name] = user
	}
	for _, spec := range newSpecs.List() {
		user, err := expandKafkaUser(spec.(map[string]interface{}))
		if err != nil {
			return err
		}
		if u, ok := m[user.Name]; ok {
			updatePaths := make([]string, 0, 2)

			if user.Password != u.Password {
				updatePaths = append(updatePaths, "password")
			}

			if fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				updatePaths = append(updatePaths, "permissions")
			}

			if len(updatePaths) > 0 {
				req := &kafka.UpdateUserRequest{
					ClusterId:   d.Id(),
					UserName:    user.Name,
					Password:    user.Password,
					Permissions: user.Permissions,
					UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
				}
				err = updateKafkaUser(ctx, config, req)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func updateKafkaUser(ctx context.Context, config *Config, req *kafka.UpdateUserRequest) error {
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Updating Kafka user %q: %+v", req.UserName, req)
		return config.sdk.MDB().Kafka().User().Update(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to update user %q in Kafka Cluster %q: %s", req.UserName, req.ClusterId, err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user %q in Kafka Cluster %q: %s", req.UserName, req.ClusterId, err)
	}
	log.Printf("[DEBUG] Finished updating Kafka user %q", req.UserName)
	return nil
}

func kafkaTopicUpdateMask(oldTopic, newTopic map[string]interface{}, version string) []string {
	var paths []string
	attrs := []string{"partitions", "replication_factor"}
	for _, attr := range attrs {
		val1 := oldTopic[attr]
		val2 := newTopic[attr]
		if !reflect.DeepEqual(val1, val2) {
			paths = append(paths, fmt.Sprintf("topic_spec.%s", attr))
		}
	}

	oldTopicConfig := map[string]interface{}{}
	topicConfigList, ok := oldTopic["topic_config"].([]interface{})
	if ok && len(topicConfigList) > 0 {
		oldTopicConfig = topicConfigList[0].(map[string]interface{})
	}

	newTopicConfig := map[string]interface{}{}
	topicConfigList, ok = newTopic["topic_config"].([]interface{})
	if ok && len(topicConfigList) > 0 {
		newTopicConfig = topicConfigList[0].(map[string]interface{})
	}

	keys := map[string]struct{}{}
	for key := range oldTopicConfig {
		keys[key] = struct{}{}
	}
	for key := range newTopicConfig {
		keys[key] = struct{}{}
	}

	for key := range keys {
		val1 := oldTopicConfig[key]
		val2 := newTopicConfig[key]
		if !reflect.DeepEqual(val1, val2) {
			paths = append(paths, fmt.Sprintf("topic_spec.topic_config_%s.%s", version, key))
		}
	}

	return paths
}

func sortKafkaTopics(topics []*kafka.Topic, specs []*kafka.TopicSpec) {
	for i, spec := range specs {
		for j := i + 1; j < len(topics); j++ {
			if spec.Name == topics[j].Name {
				topics[i], topics[j] = topics[j], topics[i]
				break
			}
		}
	}
}

func setKafkaFolderID(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Kafka().Cluster().Get(ctx, &kafka.GetClusterRequest{
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
		request := &kafka.MoveClusterRequest{
			ClusterId:           d.Id(),
			DestinationFolderId: folderID.(string),
		}
		op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending Kafka cluster move request: %+v", request)
			return config.sdk.MDB().Kafka().Cluster().Move(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("error while requesting API to move Kafka Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while moving Kafka Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("moving Kafka Cluster %q to folder %v failed: %s", d.Id(), folderID, err)
		}

	}

	return nil
}
