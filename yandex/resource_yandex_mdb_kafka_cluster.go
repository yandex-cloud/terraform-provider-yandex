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
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBKafkaClusterCreateTimeout = 60 * time.Minute
	yandexMDBKafkaClusterReadTimeout   = 5 * time.Minute
	yandexMDBKafkaClusterDeleteTimeout = 60 * time.Minute
	yandexMDBKafkaClusterUpdateTimeout = 60 * time.Minute
)

func resourceYandexMDBKafkaCluster() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterConfig(),
			},
			"environment": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "PRODUCTION",
				ValidateFunc: validateParsableValue(parseKafkaEnv),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"topic": {
				Type:       schema.TypeList,
				Optional:   true,
				Elem:       resourceYandexMDBKafkaClusterTopicBlock(),
				Deprecated: "to manage topics, please switch to using a separate resource type yandex_mdb_kafka_topic",
			},
			"user": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      kafkaUserHash,
				Elem:     resourceYandexMDBKafkaUser(),
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			"host_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"host": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      kafkaHostHash,
				Elem:     resourceYandexMDBKafkaHost(),
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
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"zones": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"kafka": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterKafkaConfig(),
			},
			"brokers_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"assign_public_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"unmanaged_topics": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"schema_registry": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"zookeeper": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterZookeeperConfig(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterResources() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_preset_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"disk_type_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceYandexMDBKafkaZookeeperResources() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_preset_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"disk_type_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterTopicBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"partitions": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"replication_factor": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"topic_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterTopicConfig(),
			},
		},
	}
}

func resourceYandexMDBKafkaUser() *schema.Resource {
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
			"permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      kafkaUserPermissionHash,
				Elem:     resourceYandexMDBKafkaPermission(),
			},
		},
	}
}

func resourceYandexMDBKafkaPermission() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"topic_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterKafkaConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterResources(),
			},
			"kafka_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterKafkaSettings(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterKafkaSettings() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaCompression),
			},
			"log_flush_interval_messages": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_flush_interval_ms": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_flush_scheduler_interval_ms": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_bytes": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_hours": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_minutes": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_retention_ms": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_segment_bytes": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"log_preallocate": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"socket_send_buffer_bytes": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"socket_receive_buffer_bytes": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"auto_create_topics_enable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"num_partitions": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
			"default_replication_factor": {
				Type:         schema.TypeString,
				ValidateFunc: ConvertableToInt(),
				Optional:     true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterTopicConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cleanup_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaTopicCleanupPolicy),
			},
			"compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaCompression),
			},
			"delete_retention_ms": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"file_delete_delay_ms": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"flush_messages": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"flush_ms": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"min_compaction_lag_ms": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"retention_bytes": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"retention_ms": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"max_message_bytes": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"min_insync_replicas": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"segment_bytes": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ConvertableToInt(),
			},
			"preallocate": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterZookeeperConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaZookeeperResources(),
			},
		},
	}
}

func resourceYandexMDBKafkaHost() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_public_ip": {
				Type:     schema.TypeBool,
				Computed: true,
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

	cfg, err := flattenKafkaConfig(cluster)
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

	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBKafkaClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Kafka Cluster %q", d.Id())

	d.Partial(true)

	if err := updateKafkaClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("topic") {
		if err := updateKafkaClusterTopics(d, meta); err != nil {
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
	return ret, nil
}

var mdbKafkaUpdateFieldsMap = map[string]string{
	"name":                   "name",
	"description":            "description",
	"labels":                 "labels",
	"security_group_ids":     "security_group_ids",
	"deletion_protection":    "deletion_protection",
	"config.0.zones":         "config_spec.zone_id",
	"config.0.version":       "config_spec.version",
	"config.0.brokers_count": "config_spec.brokers_count",
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
	"config.0.kafka.0.kafka_config.0.log_preallocate":                 "config_spec.kafka.kafka_config_{version}.log_preallocate",
	"config.0.kafka.0.kafka_config.0.socket_send_buffer_bytes":        "config_spec.kafka.kafka_config_{version}.socket_send_buffer_bytes",
	"config.0.kafka.0.kafka_config.0.socket_receive_buffer_bytes":     "config_spec.kafka.kafka_config_{version}.socket_receive_buffer_bytes",
	"config.0.kafka.0.kafka_config.0.auto_create_topics_enable":       "config_spec.kafka.kafka_config_{version}.auto_create_topics_enable",
	"config.0.kafka.0.kafka_config.0.num_partitions":                  "config_spec.kafka.kafka_config_{version}.num_partitions",
	"config.0.kafka.0.kafka_config.0.default_replication_factor":      "config_spec.kafka.kafka_config_{version}.default_replication_factor",
	"config.0.unmanaged_topics":                                       "config_spec.unmanaged_topics",
	"config.0.zookeeper.0.resources.0.resource_preset_id":             "config_spec.zookeeper.resources.resource_preset_id",
	"config.0.zookeeper.0.resources.0.disk_type_id":                   "config_spec.zookeeper.resources.disk_type_id",
	"config.0.zookeeper.0.resources.0.disk_size":                      "config_spec.zookeeper.resources.disk_size",
}

func kafkaClusterUpdateRequest(d *schema.ResourceData) (*kafka.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating Kafka cluster: %s", err)
	}

	configSpec, err := expandKafkaConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding configSpec while updating Kafka cluster: %s", err)
	}

	req := &kafka.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		ConfigSpec:         configSpec,
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		DeletionProtection: d.Get("deletion_protection").(bool),
	}
	return req, nil
}

func kafkaClusterUpdateRequestWithMask(d *schema.ResourceData) (*kafka.UpdateClusterRequest, error) {
	req, err := kafkaClusterUpdateRequest(d)
	if err != nil {
		return nil, err
	}

	updatePath := []string{}
	for field, path := range mdbKafkaUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, strings.Replace(path, "{version}", getSuffixVerion(d), -1))
		}
	}

	if len(updatePath) == 0 {
		return nil, nil
	}

	sort.Strings(updatePath)

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	return req, nil
}

func getSuffixVerion(d *schema.ResourceData) string {
	return strings.Replace(d.Get("config.0.version").(string), ".", "_", -1)
}

func updateKafkaClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := kafkaClusterUpdateRequestWithMask(d)
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

func updateKafkaClusterTopics(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	versionI, ok := d.GetOk("config.0.version")
	if !ok {
		return fmt.Errorf("you must specify version of Kafka")
	}
	version := versionI.(string)

	diffByTopicName := diffByEntityKey(d, "topic", "name")
	for topicName, topicDiff := range diffByTopicName {
		oldTopic := topicDiff[0]
		newTopic := topicDiff[1]

		if oldTopic == nil {
			topicSpec, err := expandKafkaTopic(newTopic, version)
			if err != nil {
				return err
			}
			log.Printf("[DEBUG] Creating topic %+v", topicSpec)
			if err := createKafkaTopic(ctx, config, d, topicSpec); err != nil {
				return err
			}
			continue
		}

		if newTopic == nil {
			log.Printf("[DEBUG] Topic %s is to be deleted", topicName)
			if err := deleteKafkaTopic(ctx, config, d, topicName); err != nil {
				return err
			}
			continue
		}

		if !reflect.DeepEqual(oldTopic, newTopic) {
			topicSpec, err := expandKafkaTopic(newTopic, version)
			if err != nil {
				return err
			}
			paths := kafkaTopicUpdateMask(oldTopic, newTopic, getSuffixVerion(d))
			if err := updateKafkaTopic(ctx, config, d, topicSpec, paths); err != nil {
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
		return fmt.Errorf("error while requesting API to create user in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding user to Kafka Cluster %q: %s", d.Id(), err)
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
				err = updateKafkaUser(ctx, config, d, req)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func updateKafkaUser(ctx context.Context, config *Config, d *schema.ResourceData, req *kafka.UpdateUserRequest) error {
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Updating Kafka user %q: %+v", req.UserName, req)
		return config.sdk.MDB().Kafka().User().Update(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in Kafka Cluster %q: %s", d.Id(), err)
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
