package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	// Single topic should be created/updated/deleted much faster but we set these timeouts
	// to larger values to allow for serial modification of multiple topics.
	yandexMDBKafkaTopicCreateTimeout = 10 * time.Minute
	yandexMDBKafkaTopicReadTimeout   = 1 * time.Minute
	yandexMDBKafkaTopicUpdateTimeout = 10 * time.Minute
	yandexMDBKafkaTopicDeleteTimeout = 10 * time.Minute
)

func resourceYandexMDBKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a topic of a Kafka Topic within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).",

		Create: resourceYandexMDBKafkaTopicCreate,
		Read:   resourceYandexMDBKafkaTopicRead,
		Update: resourceYandexMDBKafkaTopicUpdate,
		Delete: resourceYandexMDBKafkaTopicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBKafkaTopicCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBKafkaTopicReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBKafkaTopicUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBKafkaTopicDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Kafka cluster.",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
				ForceNew:    true,
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
				Description: "User-defined settings for the topic. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts/settings-list#topic-settings) and [the Kafka documentation](https://kafka.apache.org/documentation/#topicconfigs).",
				Optional:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterTopicConfig(),
			},
		},
	}
}

func resourceYandexMDBKafkaTopicCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	version, err := getKafkaVersion(ctx, d, config)
	if err != nil {
		return err
	}

	topicSpec, err := buildKafkaTopicSpec(d, "", version)
	if err != nil {
		return err
	}

	req := &kafka.CreateTopicRequest{
		ClusterId: d.Get("cluster_id").(string),
		TopicSpec: topicSpec,
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Creating Kafka topic: %+v", req)
		return config.sdk.MDB().Kafka().Topic().Create(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to create Kafka topic: %s", err)
	}

	topicID := constructResourceId(req.ClusterId, req.TopicSpec.Name)
	d.SetId(topicID)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for Kafka topic create operation: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("kafka topic creation failed: %s", err)
	}
	log.Printf("[DEBUG] Finished creating Kafka topic %q", req.TopicSpec.Name)

	return resourceYandexMDBKafkaTopicRead(d, meta)
}

func getKafkaVersion(ctx context.Context, d *schema.ResourceData, config *Config) (string, error) {
	clusterID := d.Get("cluster_id").(string)
	req := &kafka.GetClusterRequest{ClusterId: clusterID}
	cluster, err := config.sdk.MDB().Kafka().Cluster().Get(ctx, req)
	if err != nil {
		return "", err
	}

	return cluster.GetConfig().GetVersion(), nil
}

func buildKafkaTopicSpec(d *schema.ResourceData, prefixKey string, version string) (*kafka.TopicSpec, error) {
	key := func(key string) string {
		return fmt.Sprintf("%s%s", prefixKey, key)
	}

	topicName := d.Get(key("name")).(string)
	topicSpec := &kafka.TopicSpec{
		Name:              topicName,
		Partitions:        &wrappers.Int64Value{Value: int64(d.Get(key("partitions")).(int))},
		ReplicationFactor: &wrappers.Int64Value{Value: int64(d.Get(key("replication_factor")).(int))},
	}

	if _, ok := d.GetOk(key("topic_config.0")); ok {
		if strings.HasPrefix(version, "3") {
			cfg, err := expandKafkaTopicConfig3x(d, key("topic_config.0."))
			if err != nil {
				return nil, err
			}
			topicSpec.SetTopicConfig_3(cfg)
		} else if version == "2.8" {
			cfg, err := expandKafkaTopicConfig2_8(d, key("topic_config.0."))
			if err != nil {
				return nil, err
			}
			topicSpec.SetTopicConfig_2_8(cfg)
		} else if version == "" {
			return nil, fmt.Errorf("you must specify version of Kafka")
		} else {
			return nil, fmt.Errorf("this version of Kafka not supported by Terraform provider")
		}
	}

	return topicSpec, nil
}

func resourceYandexMDBKafkaTopicRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid topic resource id format: %q", d.Id())
	}

	clusterID := parts[0]
	topicName := parts[1]
	topic, err := config.sdk.MDB().Kafka().Topic().Get(ctx, &kafka.GetTopicRequest{
		ClusterId: clusterID,
		TopicName: topicName,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Topic %q", topicName))
	}
	d.Set("cluster_id", clusterID)
	d.Set("name", topic.Name)
	d.Set("partitions", topic.Partitions.GetValue())
	d.Set("replication_factor", topic.ReplicationFactor.GetValue())

	var cfg map[string]interface{}
	if topic.GetTopicConfig_3() != nil {
		cfg = flattenKafkaTopicConfig3(topic.GetTopicConfig_3())
	}
	if topic.GetTopicConfig_2_8() != nil {
		cfg = flattenKafkaTopicConfig2_8(topic.GetTopicConfig_2_8())
	}
	if len(cfg) != 0 {
		if err := d.Set("topic_config", []map[string]interface{}{cfg}); err != nil {
			return err
		}
	}

	return nil
}

func resourceYandexMDBKafkaTopicUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	version, err := getKafkaVersion(ctx, d, config)
	if err != nil {
		return err
	}

	topicSpec, err := buildKafkaTopicSpec(d, "", version)
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	topicName := d.Get("name").(string)
	request := &kafka.UpdateTopicRequest{
		ClusterId: clusterID,
		TopicName: topicName,
		TopicSpec: topicSpec,
	}

	var updatePath []string
	versionPath := "3"
	if strings.HasPrefix(version, "2") {
		versionPath = strings.Replace(version, ".", "_", -1)
	}
	for field, path := range mdbKafkaTopicUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, strings.Replace(path, "{version}", versionPath, -1))
		}
	}
	request.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	if len(updatePath) == 0 {
		return nil
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending topic update request: %+v", request)
		return config.sdk.MDB().Kafka().Topic().Update(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to update topic %q in Kafka Cluster %q: %s",
			topicName, clusterID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating topic in Kafka Cluster %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating Kafka topic %q", topicName)
	return resourceYandexMDBKafkaTopicRead(d, meta)
}

var mdbKafkaTopicUpdateFieldsMap = map[string]string{
	"partitions":         "topic_spec.partitions",
	"replication_factor": "topic_spec.replication_factor",
}

func init() {
	topicConfigSchema := resourceYandexMDBKafkaClusterTopicConfig().Schema
	for name := range topicConfigSchema {
		key := fmt.Sprintf("topic_config.0.%s", name)
		val := fmt.Sprintf("topic_spec.topic_config_{version}.%s", name)
		mdbKafkaTopicUpdateFieldsMap[key] = val
	}
}

func resourceYandexMDBKafkaTopicDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	topicName := d.Get("name").(string)
	clusterID := d.Get("cluster_id").(string)
	request := &kafka.DeleteTopicRequest{
		ClusterId: clusterID,
		TopicName: topicName,
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Deleting Kafka topic %q", topicName)
		return config.sdk.MDB().Kafka().Topic().Delete(ctx, request)
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kafka topic %q", topicName))
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting topic %q from Kafka Cluster %q: %s", topicName, clusterID, err)
	}

	log.Printf("[DEBUG] Finished deleting Kafka topic %q", topicName)
	return nil
}
