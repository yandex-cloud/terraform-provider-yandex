package yandex

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
)

type TopicCleanupPolicy int32

const (
	Topic_CLEANUP_POLICY_UNSPECIFIED TopicCleanupPolicy = 0
	// this policy discards log segments when either their retention time or log size limit is reached. See also: [KafkaConfig2_1.log_retention_ms] and other similar parameters.
	Topic_CLEANUP_POLICY_DELETE TopicCleanupPolicy = 1
	// this policy compacts messages in log.
	Topic_CLEANUP_POLICY_COMPACT TopicCleanupPolicy = 2
	// this policy use both compaction and deletion for messages and log segments.
	Topic_CLEANUP_POLICY_COMPACT_AND_DELETE TopicCleanupPolicy = 3
)

const kafkaConfigPath = "config.0.kafka.0.kafka_config.0"

// Enum value maps for TopicCleanupPolicy.
var (
	Topic_CleanupPolicy_name = map[int32]string{
		0: "CLEANUP_POLICY_UNSPECIFIED",
		1: "CLEANUP_POLICY_DELETE",
		2: "CLEANUP_POLICY_COMPACT",
		3: "CLEANUP_POLICY_COMPACT_AND_DELETE",
	}
	Topic_CleanupPolicy_value = map[string]int32{
		"CLEANUP_POLICY_UNSPECIFIED":        0,
		"CLEANUP_POLICY_DELETE":             1,
		"CLEANUP_POLICY_COMPACT":            2,
		"CLEANUP_POLICY_COMPACT_AND_DELETE": 3,
	}
)

func parseKafkaEnv(e string) (kafka.Cluster_Environment, error) {
	v, ok := kafka.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(kafka.Cluster_Environment_value)), e)
	}
	return kafka.Cluster_Environment(v), nil
}

func parseKafkaCompression(e string) (kafka.CompressionType, error) {
	v, ok := kafka.CompressionType_value[e]
	if !ok || e == "COMPRESSION_TYPE_UNSPECIFIED" {
		return 0, fmt.Errorf("value for 'compression_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeysExt(kafka.CompressionType_value, true)), e)
	}
	return kafka.CompressionType(v), nil
}

func parseKafkaPermission(e string) (kafka.Permission_AccessRole, error) {
	v, ok := kafka.Permission_AccessRole_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'role' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(kafka.Permission_AccessRole_value)), e)
	}
	return kafka.Permission_AccessRole(v), nil
}

func parseKafkaTopicCleanupPolicy(e string) (TopicCleanupPolicy, error) {
	v, ok := Topic_CleanupPolicy_value[e]
	if !ok || e == "CLEANUP_POLICY_UNSPECIFIED" {
		return 0, fmt.Errorf("value for 'cleanup_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeysExt(Topic_CleanupPolicy_value, true)), e)
	}
	return TopicCleanupPolicy(v), nil
}

func parseIntKafkaConfigParam(d *schema.ResourceData, paramName string, retErr *error) *wrappers.Int64Value {
	v, ok := d.GetOk(kafkaConfigPath + "." + paramName)
	if !ok {
		return nil
	}

	i, err := strconv.ParseInt(v.(string), 10, 64)
	if err != nil {
		if *retErr != nil {
			*retErr = err
		}
		return nil
	}
	return &wrappers.Int64Value{Value: i}
}

type KafkaConfig struct {
	CompressionType             kafka.CompressionType
	LogFlushIntervalMessages    *wrappers.Int64Value
	LogFlushIntervalMs          *wrappers.Int64Value
	LogFlushSchedulerIntervalMs *wrappers.Int64Value
	LogRetentionBytes           *wrappers.Int64Value
	LogRetentionHours           *wrappers.Int64Value
	LogRetentionMinutes         *wrappers.Int64Value
	LogRetentionMs              *wrappers.Int64Value
	LogSegmentBytes             *wrappers.Int64Value
	LogPreallocate              *wrappers.BoolValue
	SocketSendBufferBytes       *wrappers.Int64Value
	SocketReceiveBufferBytes    *wrappers.Int64Value
	AutoCreateTopicsEnable      *wrappers.BoolValue
	NumPartitions               *wrappers.Int64Value
	DefaultReplicationFactor    *wrappers.Int64Value
}

func parseKafkaConfig(d *schema.ResourceData) (*KafkaConfig, error) {
	res := &KafkaConfig{}

	if v, ok := d.GetOk(kafkaConfigPath + ".compression_type"); ok {
		value, err := parseKafkaCompression(v.(string))
		if err != nil {
			return nil, err
		}
		res.CompressionType = value
	}

	var retErr error

	res.LogFlushIntervalMessages = parseIntKafkaConfigParam(d, "log_flush_interval_messages", &retErr)
	res.LogFlushIntervalMs = parseIntKafkaConfigParam(d, "log_flush_interval_ms", &retErr)
	res.LogFlushSchedulerIntervalMs = parseIntKafkaConfigParam(d, "log_flush_scheduler_interval_ms", &retErr)
	res.LogRetentionBytes = parseIntKafkaConfigParam(d, "log_retention_bytes", &retErr)
	res.LogRetentionHours = parseIntKafkaConfigParam(d, "log_retention_hours", &retErr)
	res.LogRetentionMinutes = parseIntKafkaConfigParam(d, "log_retention_minutes", &retErr)
	res.LogRetentionMs = parseIntKafkaConfigParam(d, "log_retention_ms", &retErr)
	res.LogSegmentBytes = parseIntKafkaConfigParam(d, "log_segment_bytes", &retErr)
	res.SocketSendBufferBytes = parseIntKafkaConfigParam(d, "socket_send_buffer_bytes", &retErr)
	res.SocketReceiveBufferBytes = parseIntKafkaConfigParam(d, "socket_receive_buffer_bytes", &retErr)
	res.NumPartitions = parseIntKafkaConfigParam(d, "num_partitions", &retErr)
	res.DefaultReplicationFactor = parseIntKafkaConfigParam(d, "default_replication_factor", &retErr)

	if v, ok := d.GetOk(kafkaConfigPath + ".log_preallocate"); ok {
		res.LogPreallocate = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(kafkaConfigPath + ".auto_create_topics_enable"); ok {
		res.AutoCreateTopicsEnable = &wrappers.BoolValue{Value: v.(bool)}
	}

	if retErr != nil {
		return nil, retErr
	}

	return res, nil
}

func expandKafkaConfig2_6(d *schema.ResourceData) (*kafka.KafkaConfig2_6, error) {
	kafkaConfig, err := parseKafkaConfig(d)
	if err != nil {
		return nil, err
	}
	return &kafka.KafkaConfig2_6{
		CompressionType:             kafkaConfig.CompressionType,
		LogFlushIntervalMessages:    kafkaConfig.LogFlushIntervalMessages,
		LogFlushIntervalMs:          kafkaConfig.LogFlushIntervalMs,
		LogFlushSchedulerIntervalMs: kafkaConfig.LogFlushSchedulerIntervalMs,
		LogRetentionBytes:           kafkaConfig.LogRetentionBytes,
		LogRetentionHours:           kafkaConfig.LogRetentionHours,
		LogRetentionMinutes:         kafkaConfig.LogRetentionMinutes,
		LogRetentionMs:              kafkaConfig.LogRetentionMs,
		LogSegmentBytes:             kafkaConfig.LogSegmentBytes,
		LogPreallocate:              kafkaConfig.LogPreallocate,
		SocketSendBufferBytes:       kafkaConfig.SocketSendBufferBytes,
		SocketReceiveBufferBytes:    kafkaConfig.SocketReceiveBufferBytes,
		AutoCreateTopicsEnable:      kafkaConfig.AutoCreateTopicsEnable,
		NumPartitions:               kafkaConfig.NumPartitions,
		DefaultReplicationFactor:    kafkaConfig.DefaultReplicationFactor,
	}, nil
}

func expandKafkaConfig2_1(d *schema.ResourceData) (*kafka.KafkaConfig2_1, error) {
	kafkaConfig, err := parseKafkaConfig(d)
	if err != nil {
		return nil, err
	}
	return &kafka.KafkaConfig2_1{
		CompressionType:             kafkaConfig.CompressionType,
		LogFlushIntervalMessages:    kafkaConfig.LogFlushIntervalMessages,
		LogFlushIntervalMs:          kafkaConfig.LogFlushIntervalMs,
		LogFlushSchedulerIntervalMs: kafkaConfig.LogFlushSchedulerIntervalMs,
		LogRetentionBytes:           kafkaConfig.LogRetentionBytes,
		LogRetentionHours:           kafkaConfig.LogRetentionHours,
		LogRetentionMinutes:         kafkaConfig.LogRetentionMinutes,
		LogRetentionMs:              kafkaConfig.LogRetentionMs,
		LogSegmentBytes:             kafkaConfig.LogSegmentBytes,
		LogPreallocate:              kafkaConfig.LogPreallocate,
		SocketSendBufferBytes:       kafkaConfig.SocketSendBufferBytes,
		SocketReceiveBufferBytes:    kafkaConfig.SocketReceiveBufferBytes,
		AutoCreateTopicsEnable:      kafkaConfig.AutoCreateTopicsEnable,
		NumPartitions:               kafkaConfig.NumPartitions,
		DefaultReplicationFactor:    kafkaConfig.DefaultReplicationFactor,
	}, nil
}

func expandKafkaConfig2_8(d *schema.ResourceData) (*kafka.KafkaConfig2_8, error) {
	kafkaConfig, err := parseKafkaConfig(d)
	if err != nil {
		return nil, err
	}
	return &kafka.KafkaConfig2_8{
		CompressionType:             kafkaConfig.CompressionType,
		LogFlushIntervalMessages:    kafkaConfig.LogFlushIntervalMessages,
		LogFlushIntervalMs:          kafkaConfig.LogFlushIntervalMs,
		LogFlushSchedulerIntervalMs: kafkaConfig.LogFlushSchedulerIntervalMs,
		LogRetentionBytes:           kafkaConfig.LogRetentionBytes,
		LogRetentionHours:           kafkaConfig.LogRetentionHours,
		LogRetentionMinutes:         kafkaConfig.LogRetentionMinutes,
		LogRetentionMs:              kafkaConfig.LogRetentionMs,
		LogSegmentBytes:             kafkaConfig.LogSegmentBytes,
		LogPreallocate:              kafkaConfig.LogPreallocate,
		SocketSendBufferBytes:       kafkaConfig.SocketSendBufferBytes,
		SocketReceiveBufferBytes:    kafkaConfig.SocketReceiveBufferBytes,
		AutoCreateTopicsEnable:      kafkaConfig.AutoCreateTopicsEnable,
		NumPartitions:               kafkaConfig.NumPartitions,
		DefaultReplicationFactor:    kafkaConfig.DefaultReplicationFactor,
	}, nil
}

type TopicConfig struct {
	CleanupPolicy      string
	CompressionType    kafka.CompressionType
	DeleteRetentionMs  *wrappers.Int64Value
	FileDeleteDelayMs  *wrappers.Int64Value
	FlushMessages      *wrappers.Int64Value
	FlushMs            *wrappers.Int64Value
	MinCompactionLagMs *wrappers.Int64Value
	RetentionBytes     *wrappers.Int64Value
	RetentionMs        *wrappers.Int64Value
	MaxMessageBytes    *wrappers.Int64Value
	MinInsyncReplicas  *wrappers.Int64Value
	SegmentBytes       *wrappers.Int64Value
	Preallocate        *wrappers.BoolValue
}

func parseIntTopicConfigParam(config map[string]interface{}, paramName string, retErr *error) *wrappers.Int64Value {
	paramValue, ok := config[paramName]
	if !ok {
		return nil
	}
	str := paramValue.(string)
	if str == "" {
		return nil
	}
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		if *retErr != nil {
			*retErr = err
		}
		return nil
	}
	return &wrappers.Int64Value{Value: i}
}

func parseKafkaTopicConfig(config map[string]interface{}) (*TopicConfig, error) {
	res := &TopicConfig{}

	if cleanupPolicy := config["cleanup_policy"].(string); cleanupPolicy != "" {
		_, err := parseKafkaTopicCleanupPolicy(cleanupPolicy)
		if err != nil {
			return nil, err
		}
		res.CleanupPolicy = cleanupPolicy
	}

	if compressionType := config["compression_type"].(string); compressionType != "" {
		value, err := parseKafkaCompression(compressionType)
		if err != nil {
			return nil, err
		}
		res.CompressionType = value
	}

	var retErr error
	res.DeleteRetentionMs = parseIntTopicConfigParam(config, "delete_retention_ms", &retErr)
	res.FileDeleteDelayMs = parseIntTopicConfigParam(config, "file_delete_delay_ms", &retErr)
	res.FlushMessages = parseIntTopicConfigParam(config, "flush_messages", &retErr)
	res.FlushMs = parseIntTopicConfigParam(config, "flush_ms", &retErr)
	res.MinCompactionLagMs = parseIntTopicConfigParam(config, "min_compaction_lag_ms", &retErr)
	res.RetentionBytes = parseIntTopicConfigParam(config, "retention_bytes", &retErr)
	res.RetentionMs = parseIntTopicConfigParam(config, "retention_ms", &retErr)
	res.MaxMessageBytes = parseIntTopicConfigParam(config, "max_message_bytes", &retErr)
	res.MinInsyncReplicas = parseIntTopicConfigParam(config, "min_insync_replicas", &retErr)
	res.SegmentBytes = parseIntTopicConfigParam(config, "segment_bytes", &retErr)

	if v, ok := config["preallocate"]; ok {
		res.Preallocate = &wrappers.BoolValue{Value: v.(bool)}
	}

	if retErr != nil {
		return nil, retErr
	}

	return res, nil
}

func expandKafkaTopicConfig2_6(config map[string]interface{}) (*kafka.TopicConfig2_6, error) {
	topicConfig, err := parseKafkaTopicConfig(config)
	if err != nil {
		return nil, err
	}
	res := &kafka.TopicConfig2_6{
		CleanupPolicy:      kafka.TopicConfig2_6_CleanupPolicy(kafka.TopicConfig2_6_CleanupPolicy_value[topicConfig.CleanupPolicy]),
		CompressionType:    topicConfig.CompressionType,
		DeleteRetentionMs:  topicConfig.DeleteRetentionMs,
		FileDeleteDelayMs:  topicConfig.FileDeleteDelayMs,
		FlushMessages:      topicConfig.FlushMessages,
		FlushMs:            topicConfig.FlushMs,
		MinCompactionLagMs: topicConfig.MinCompactionLagMs,
		RetentionBytes:     topicConfig.RetentionBytes,
		RetentionMs:        topicConfig.RetentionMs,
		MaxMessageBytes:    topicConfig.MaxMessageBytes,
		MinInsyncReplicas:  topicConfig.MinInsyncReplicas,
		SegmentBytes:       topicConfig.SegmentBytes,
		Preallocate:        topicConfig.Preallocate,
	}

	return res, nil
}

func expandKafkaTopicConfig2_1(config map[string]interface{}) (*kafka.TopicConfig2_1, error) {
	topicConfig, err := parseKafkaTopicConfig(config)
	if err != nil {
		return nil, err
	}
	res := &kafka.TopicConfig2_1{
		CleanupPolicy:      kafka.TopicConfig2_1_CleanupPolicy(kafka.TopicConfig2_1_CleanupPolicy_value[topicConfig.CleanupPolicy]),
		CompressionType:    topicConfig.CompressionType,
		DeleteRetentionMs:  topicConfig.DeleteRetentionMs,
		FileDeleteDelayMs:  topicConfig.FileDeleteDelayMs,
		FlushMessages:      topicConfig.FlushMessages,
		FlushMs:            topicConfig.FlushMs,
		MinCompactionLagMs: topicConfig.MinCompactionLagMs,
		RetentionBytes:     topicConfig.RetentionBytes,
		RetentionMs:        topicConfig.RetentionMs,
		MaxMessageBytes:    topicConfig.MaxMessageBytes,
		MinInsyncReplicas:  topicConfig.MinInsyncReplicas,
		SegmentBytes:       topicConfig.SegmentBytes,
		Preallocate:        topicConfig.Preallocate,
	}

	return res, nil
}

func expandKafkaTopicConfig2_8(config map[string]interface{}) (*kafka.TopicConfig2_8, error) {
	topicConfig, err := parseKafkaTopicConfig(config)
	if err != nil {
		return nil, err
	}
	res := &kafka.TopicConfig2_8{
		CleanupPolicy:      kafka.TopicConfig2_8_CleanupPolicy(kafka.TopicConfig2_8_CleanupPolicy_value[topicConfig.CleanupPolicy]),
		CompressionType:    topicConfig.CompressionType,
		DeleteRetentionMs:  topicConfig.DeleteRetentionMs,
		FileDeleteDelayMs:  topicConfig.FileDeleteDelayMs,
		FlushMessages:      topicConfig.FlushMessages,
		FlushMs:            topicConfig.FlushMs,
		MinCompactionLagMs: topicConfig.MinCompactionLagMs,
		RetentionBytes:     topicConfig.RetentionBytes,
		RetentionMs:        topicConfig.RetentionMs,
		MaxMessageBytes:    topicConfig.MaxMessageBytes,
		MinInsyncReplicas:  topicConfig.MinInsyncReplicas,
		SegmentBytes:       topicConfig.SegmentBytes,
		Preallocate:        topicConfig.Preallocate,
	}

	return res, nil
}

func expandKafkaConfigSpec(d *schema.ResourceData) (*kafka.ConfigSpec, error) {
	result := &kafka.ConfigSpec{}

	if v, ok := d.GetOk("config.0.version"); ok {
		result.Version = v.(string)
	}

	if v, ok := d.GetOk("config.0.brokers_count"); ok {
		result.BrokersCount = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOk("config.0.assign_public_ip"); ok {
		result.AssignPublicIp = v.(bool)
	}

	if v, ok := d.GetOk("config.0.unmanaged_topics"); ok {
		result.UnmanagedTopics = v.(bool)
	}

	if v, ok := d.GetOk("config.0.schema_registry"); ok {
		result.SchemaRegistry = v.(bool)
	}

	if v, ok := d.GetOk("config.0.zones"); ok {
		zones := v.([]interface{})
		result.ZoneId = []string{}
		for _, zone := range zones {
			result.ZoneId = append(result.ZoneId, zone.(string))
		}
	}
	result.Kafka = &kafka.ConfigSpec_Kafka{}
	result.Kafka.Resources = expandKafkaResources(d, "config.0.kafka.0.resources.0")

	switch version := result.Version; version {
	case "2.8":
		cfg, err := expandKafkaConfig2_8(d)
		if err != nil {
			return nil, err
		}
		result.Kafka.SetKafkaConfig_2_8(cfg)
	case "2.6":
		cfg, err := expandKafkaConfig2_6(d)
		if err != nil {
			return nil, err
		}
		result.Kafka.SetKafkaConfig_2_6(cfg)
	case "2.1":
		cfg, err := expandKafkaConfig2_1(d)
		if err != nil {
			return nil, err
		}
		result.Kafka.SetKafkaConfig_2_1(cfg)
	default:
		return nil, fmt.Errorf("you must specify version of Kafka")
	}

	if _, ok := d.GetOk("config.0.zookeeper"); ok {
		result.Zookeeper = &kafka.ConfigSpec_Zookeeper{}
		result.Zookeeper.Resources = expandKafkaResources(d, "config.0.zookeeper.0.resources.0")
	}

	return result, nil
}

func expandKafkaTopics(d *schema.ResourceData) ([]*kafka.TopicSpec, error) {
	var result []*kafka.TopicSpec
	version, ok := d.GetOk("config.0.version")
	if !ok {
		return nil, fmt.Errorf("you must specify version of Kafka")
	}
	topics := d.Get("topic").([]interface{})

	for _, topic := range topics {
		topicSpec, err := expandKafkaTopic(topic.(map[string]interface{}), version.(string))
		if err != nil {
			return nil, err
		}
		result = append(result, topicSpec)
	}
	return result, nil
}

func expandKafkaUsers(d *schema.ResourceData) ([]*kafka.UserSpec, error) {
	users := d.Get("user").(*schema.Set)
	result := make([]*kafka.UserSpec, 0, users.Len())

	for _, u := range users.List() {
		user, err := expandKafkaUser(u)
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	return result, nil
}

func expandKafkaUser(u interface{}) (*kafka.UserSpec, error) {
	m := u.(map[string]interface{})
	user := &kafka.UserSpec{}
	if v, ok := m["name"]; ok {
		user.Name = v.(string)
	}
	if v, ok := m["password"]; ok {
		user.Password = v.(string)
	}
	if v, ok := m["permission"]; ok {
		permissions, err := expandKafkaPermissions(v.(*schema.Set))
		if err != nil {
			return nil, err
		}
		user.Permissions = permissions
	}
	return user, nil
}

func expandKafkaPermissions(ps *schema.Set) ([]*kafka.Permission, error) {
	result := []*kafka.Permission{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &kafka.Permission{}
		if v, ok := m["topic_name"]; ok {
			permission.TopicName = v.(string)
		}
		if v, ok := m["role"]; ok {
			role, err := parseKafkaPermission(v.(string))
			if err != nil {
				return nil, err
			}
			permission.Role = role
		}
		result = append(result, permission)
	}
	return result, nil
}

func flattenKafkaConfig(cluster *kafka.Cluster) ([]map[string]interface{}, error) {
	kafkaResources, err := flattenKafkaResources(cluster.Config.Kafka.Resources)
	if err != nil {
		return nil, err
	}

	var kafkaConfig map[string]interface{}
	if cluster.Config.Kafka.GetKafkaConfig_2_6() != nil {
		kafkaConfig, err = flattenKafkaConfig2_6Settings(cluster.Config.Kafka.GetKafkaConfig_2_6())
		if err != nil {
			return nil, err
		}
	}
	if cluster.Config.Kafka.GetKafkaConfig_2_1() != nil {
		kafkaConfig, err = flattenKafkaConfig2_1Settings(cluster.Config.Kafka.GetKafkaConfig_2_1())
		if err != nil {
			return nil, err
		}
	}
	if cluster.Config.Kafka.GetKafkaConfig_2_8() != nil {
		kafkaConfig, err = flattenKafkaConfig2_8Settings(cluster.Config.Kafka.GetKafkaConfig_2_8())
		if err != nil {
			return nil, err
		}
	}

	config := map[string]interface{}{
		"brokers_count":    cluster.Config.BrokersCount.GetValue(),
		"assign_public_ip": cluster.Config.AssignPublicIp,
		"unmanaged_topics": cluster.Config.UnmanagedTopics,
		"schema_registry":  cluster.Config.SchemaRegistry,
		"zones":            cluster.Config.ZoneId,
		"version":          cluster.Config.Version,
		"kafka": []map[string]interface{}{
			{
				"resources":    []map[string]interface{}{kafkaResources},
				"kafka_config": []map[string]interface{}{kafkaConfig},
			},
		},
	}
	if cluster.Config.Zookeeper != nil {
		zkResources, err := flattenKafkaResources(cluster.Config.Zookeeper.Resources)
		if err != nil {
			return nil, err
		}
		config["zookeeper"] = []map[string]interface{}{
			{
				"resources": []map[string]interface{}{zkResources},
			},
		}
	}

	return []map[string]interface{}{config}, nil
}

type KafkaConfigSettings interface {
	GetCompressionType() kafka.CompressionType
	GetLogFlushIntervalMessages() *wrappers.Int64Value
	GetLogFlushIntervalMs() *wrappers.Int64Value
	GetLogFlushSchedulerIntervalMs() *wrappers.Int64Value
	GetLogRetentionBytes() *wrappers.Int64Value
	GetLogRetentionHours() *wrappers.Int64Value
	GetLogRetentionMinutes() *wrappers.Int64Value
	GetLogRetentionMs() *wrappers.Int64Value
	GetLogSegmentBytes() *wrappers.Int64Value
	GetLogPreallocate() *wrappers.BoolValue
	GetSocketSendBufferBytes() *wrappers.Int64Value
	GetSocketReceiveBufferBytes() *wrappers.Int64Value
	GetAutoCreateTopicsEnable() *wrappers.BoolValue
	GetNumPartitions() *wrappers.Int64Value
	GetDefaultReplicationFactor() *wrappers.Int64Value
}

func flattenKafkaConfigSettings(kafkaConfig KafkaConfigSettings) (map[string]interface{}, error) {
	res := map[string]interface{}{}

	if kafkaConfig.GetCompressionType() != kafka.CompressionType_COMPRESSION_TYPE_UNSPECIFIED {
		res["compression_type"] = kafkaConfig.GetCompressionType().String()
	}
	if kafkaConfig.GetLogFlushIntervalMessages() != nil {
		res["log_flush_interval_messages"] = strconv.FormatInt(kafkaConfig.GetLogFlushIntervalMessages().GetValue(), 10)
	}
	if kafkaConfig.GetLogFlushIntervalMs() != nil {
		res["log_flush_interval_ms"] = strconv.FormatInt(kafkaConfig.GetLogFlushIntervalMs().GetValue(), 10)
	}
	if kafkaConfig.GetLogFlushSchedulerIntervalMs() != nil {
		res["log_flush_scheduler_interval_ms"] = strconv.FormatInt(kafkaConfig.GetLogFlushSchedulerIntervalMs().GetValue(), 10)
	}
	if kafkaConfig.GetLogRetentionBytes() != nil {
		res["log_retention_bytes"] = strconv.FormatInt(kafkaConfig.GetLogRetentionBytes().GetValue(), 10)
	}
	if kafkaConfig.GetLogRetentionHours() != nil {
		res["log_retention_hours"] = strconv.FormatInt(kafkaConfig.GetLogRetentionHours().GetValue(), 10)
	}
	if kafkaConfig.GetLogRetentionMinutes() != nil {
		res["log_retention_minutes"] = strconv.FormatInt(kafkaConfig.GetLogRetentionMinutes().GetValue(), 10)
	}
	if kafkaConfig.GetLogRetentionMs() != nil {
		res["log_retention_ms"] = strconv.FormatInt(kafkaConfig.GetLogRetentionMs().GetValue(), 10)
	}
	if kafkaConfig.GetLogSegmentBytes() != nil {
		res["log_segment_bytes"] = strconv.FormatInt(kafkaConfig.GetLogSegmentBytes().GetValue(), 10)
	}
	if kafkaConfig.GetLogPreallocate() != nil {
		res["log_preallocate"] = kafkaConfig.GetLogPreallocate().GetValue()
	}
	if kafkaConfig.GetSocketSendBufferBytes() != nil {
		res["socket_send_buffer_bytes"] = strconv.FormatInt(kafkaConfig.GetSocketSendBufferBytes().GetValue(), 10)
	}
	if kafkaConfig.GetSocketReceiveBufferBytes() != nil {
		res["socket_receive_buffer_bytes"] = strconv.FormatInt(kafkaConfig.GetSocketReceiveBufferBytes().GetValue(), 10)
	}
	if kafkaConfig.GetAutoCreateTopicsEnable() != nil {
		res["auto_create_topics_enable"] = kafkaConfig.GetAutoCreateTopicsEnable().GetValue()
	}
	if kafkaConfig.GetNumPartitions() != nil {
		res["num_partitions"] = strconv.FormatInt(kafkaConfig.GetNumPartitions().GetValue(), 10)
	}
	if kafkaConfig.GetDefaultReplicationFactor() != nil {
		res["default_replication_factor"] = strconv.FormatInt(kafkaConfig.GetDefaultReplicationFactor().GetValue(), 10)
	}

	return res, nil
}

func flattenKafkaConfig2_6Settings(r *kafka.KafkaConfig2_6) (map[string]interface{}, error) {
	return flattenKafkaConfigSettings(r)
}

func flattenKafkaConfig2_1Settings(r *kafka.KafkaConfig2_1) (map[string]interface{}, error) {
	return flattenKafkaConfigSettings(r)
}

func flattenKafkaConfig2_8Settings(r *kafka.KafkaConfig2_8) (map[string]interface{}, error) {
	return flattenKafkaConfigSettings(r)
}

func flattenKafkaResources(r *kafka.Resources) (map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = toGigabytes(r.DiskSize)

	return res, nil
}

func expandKafkaResources(d *schema.ResourceData, rootKey string) *kafka.Resources {
	resources := &kafka.Resources{}

	if v, ok := d.GetOk(rootKey + ".resource_preset_id"); ok {
		resources.ResourcePresetId = v.(string)
	}
	if v, ok := d.GetOk(rootKey + ".disk_size"); ok {
		resources.DiskSize = toBytes(v.(int))
	}
	if v, ok := d.GetOk(rootKey + ".disk_type_id"); ok {
		resources.DiskTypeId = v.(string)
	}
	return resources
}

func kafkaUserHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if n, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n.(string)))
	}
	if p, ok := m["password"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", p.(string)))
	}
	if ps, ok := m["permission"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", ps.(*schema.Set).List()))
	}
	return hashcode.String(buf.String())
}

func kafkaUserPermissionHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if n, ok := m["topic_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n.(string)))
	}
	if r, ok := m["role"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", r))
	}
	return hashcode.String(buf.String())
}

func kafkaHostHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if n, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n.(string)))
	}
	return hashcode.String(buf.String())
}

func flattenKafkaTopics(topics []*kafka.Topic) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, d := range topics {
		m := make(map[string]interface{})
		m["name"] = d.GetName()
		m["partitions"] = d.GetPartitions().GetValue()
		m["replication_factor"] = d.GetReplicationFactor().GetValue()
		var cfg map[string]interface{}
		if d.GetTopicConfig_2_6() != nil {
			cfg = flattenKafkaTopicConfig2_6(d.GetTopicConfig_2_6())
		}
		if d.GetTopicConfig_2_1() != nil {
			cfg = flattenKafkaTopicConfig2_1(d.GetTopicConfig_2_1())
		}
		if d.GetTopicConfig_2_8() != nil {
			cfg = flattenKafkaTopicConfig2_8(d.GetTopicConfig_2_8())
		}
		if len(cfg) != 0 {
			m["topic_config"] = []map[string]interface{}{cfg}
		}
		result = append(result, m)
	}

	return result
}

type TopicConfigSpec interface {
	GetCompressionType() kafka.CompressionType
	GetDeleteRetentionMs() *wrappers.Int64Value
	GetFileDeleteDelayMs() *wrappers.Int64Value
	GetFlushMessages() *wrappers.Int64Value
	GetFlushMs() *wrappers.Int64Value
	GetMinCompactionLagMs() *wrappers.Int64Value
	GetRetentionBytes() *wrappers.Int64Value
	GetRetentionMs() *wrappers.Int64Value
	GetMaxMessageBytes() *wrappers.Int64Value
	GetMinInsyncReplicas() *wrappers.Int64Value
	GetSegmentBytes() *wrappers.Int64Value
	GetPreallocate() *wrappers.BoolValue
}

func flattenKafkaTopicConfig(topicConfig TopicConfigSpec) map[string]interface{} {
	result := make(map[string]interface{})

	if topicConfig.GetCompressionType() != kafka.CompressionType_COMPRESSION_TYPE_UNSPECIFIED {
		result["compression_type"] = topicConfig.GetCompressionType().String()
	}
	if topicConfig.GetDeleteRetentionMs() != nil {
		result["delete_retention_ms"] = strconv.FormatInt(topicConfig.GetDeleteRetentionMs().GetValue(), 10)
	}
	if topicConfig.GetFileDeleteDelayMs() != nil {
		result["file_delete_delay_ms"] = strconv.FormatInt(topicConfig.GetFileDeleteDelayMs().GetValue(), 10)
	}
	if topicConfig.GetFlushMessages() != nil {
		result["flush_messages"] = strconv.FormatInt(topicConfig.GetFlushMessages().GetValue(), 10)
	}
	if topicConfig.GetFlushMs() != nil {
		result["flush_ms"] = strconv.FormatInt(topicConfig.GetFlushMs().GetValue(), 10)
	}
	if topicConfig.GetMinCompactionLagMs() != nil {
		result["min_compaction_lag_ms"] = strconv.FormatInt(topicConfig.GetMinCompactionLagMs().GetValue(), 10)
	}
	if topicConfig.GetRetentionBytes() != nil {
		result["retention_bytes"] = strconv.FormatInt(topicConfig.GetRetentionBytes().GetValue(), 10)
	}
	if topicConfig.GetRetentionMs() != nil {
		result["retention_ms"] = strconv.FormatInt(topicConfig.GetRetentionMs().GetValue(), 10)
	}
	if topicConfig.GetMaxMessageBytes() != nil {
		result["max_message_bytes"] = strconv.FormatInt(topicConfig.GetMaxMessageBytes().GetValue(), 10)
	}
	if topicConfig.GetMinInsyncReplicas() != nil {
		result["min_insync_replicas"] = strconv.FormatInt(topicConfig.GetMinInsyncReplicas().GetValue(), 10)
	}
	if topicConfig.GetSegmentBytes() != nil {
		result["segment_bytes"] = strconv.FormatInt(topicConfig.GetSegmentBytes().GetValue(), 10)
	}
	if topicConfig.GetPreallocate() != nil {
		result["preallocate"] = topicConfig.GetPreallocate().GetValue()
	}
	return result
}

func flattenKafkaTopicConfig2_6(topicConfig *kafka.TopicConfig2_6) map[string]interface{} {
	result := flattenKafkaTopicConfig(topicConfig)

	if topicConfig.GetCleanupPolicy() != kafka.TopicConfig2_6_CLEANUP_POLICY_UNSPECIFIED {
		result["cleanup_policy"] = topicConfig.GetCleanupPolicy().String()
	}

	return result
}

func flattenKafkaTopicConfig2_1(topicConfig *kafka.TopicConfig2_1) map[string]interface{} {
	result := flattenKafkaTopicConfig(topicConfig)

	if topicConfig.GetCleanupPolicy() != kafka.TopicConfig2_1_CLEANUP_POLICY_UNSPECIFIED {
		result["cleanup_policy"] = topicConfig.GetCleanupPolicy().String()
	}

	return result
}

func flattenKafkaTopicConfig2_8(topicConfig *kafka.TopicConfig2_8) map[string]interface{} {
	result := flattenKafkaTopicConfig(topicConfig)

	if topicConfig.GetCleanupPolicy() != kafka.TopicConfig2_8_CLEANUP_POLICY_UNSPECIFIED {
		result["cleanup_policy"] = topicConfig.GetCleanupPolicy().String()
	}

	return result
}

func flattenKafkaUsers(users []*kafka.User, passwords map[string]string) *schema.Set {
	result := schema.NewSet(kafkaUserHash, nil)

	for _, user := range users {
		u := map[string]interface{}{}
		u["name"] = user.Name

		perms := schema.NewSet(kafkaUserPermissionHash, nil)
		for _, perm := range user.Permissions {
			p := map[string]interface{}{}
			p["topic_name"] = perm.TopicName
			p["role"] = perm.Role.String()
			perms.Add(p)
		}
		u["permission"] = perms

		if p, ok := passwords[user.Name]; ok {
			u["password"] = p
		}
		result.Add(u)
	}
	return result
}

func flattenKafkaHosts(hosts []*kafka.Host) *schema.Set {
	result := schema.NewSet(kafkaHostHash, nil)

	for _, host := range hosts {
		h := map[string]interface{}{}
		h["name"] = host.Name
		h["zone_id"] = host.ZoneId
		h["role"] = host.Role.String()
		h["health"] = host.Health.String()
		h["subnet_id"] = host.SubnetId
		h["assign_public_ip"] = host.AssignPublicIp

		result.Add(h)
	}
	return result
}

func kafkaUsersPasswords(users []*kafka.UserSpec) map[string]string {
	result := map[string]string{}
	for _, u := range users {
		result[u.Name] = u.Password
	}
	return result
}

func kafkaUsersDiff(currUsers []*kafka.User, targetUsers []*kafka.UserSpec) ([]string, []*kafka.UserSpec) {
	m := map[string]bool{}
	toAdd := []*kafka.UserSpec{}
	toDelete := map[string]bool{}
	for _, user := range currUsers {
		toDelete[user.Name] = true
		m[user.Name] = true
	}

	for _, user := range targetUsers {
		delete(toDelete, user.Name)
		_, ok := m[user.Name]
		if !ok {
			toAdd = append(toAdd, user)
			continue
		}
	}

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func diffByEntityKey(d *schema.ResourceData, path, indexKey string) map[string][]map[string]interface{} {
	result := map[string][]map[string]interface{}{}
	for i := 0; ; i++ {
		oldEntityI, newEntityI := d.GetChange(fmt.Sprintf("%s.%d", path, i))
		empty := true

		oldEntity := oldEntityI.(map[string]interface{})
		oldEntityKey, ok := oldEntity[indexKey].(string)
		if ok {
			empty = false
			pair, ok := result[oldEntityKey]
			if !ok {
				pair = make([]map[string]interface{}, 2)
				result[oldEntityKey] = pair
			}
			pair[0] = oldEntity
		}

		newEntity := newEntityI.(map[string]interface{})
		newEntityKey, ok := newEntity[indexKey].(string)
		if ok {
			empty = false
			if newEntityKey != "" {
				pair, ok := result[newEntityKey]
				if !ok {
					pair = make([]map[string]interface{}, 2)
					result[newEntityKey] = pair
				}
				pair[1] = newEntity
			}
		}

		if empty {
			break
		}
	}
	return result
}

func expandKafkaTopic(spec map[string]interface{}, version string) (*kafka.TopicSpec, error) {
	topic := &kafka.TopicSpec{}

	if v, ok := spec["name"]; ok {
		topic.Name = v.(string)
	}
	if v, ok := spec["partitions"]; ok {
		topic.Partitions = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := spec["replication_factor"]; ok {
		topic.ReplicationFactor = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := spec["topic_config"]; ok {
		configList := v.([]interface{})
		if len(configList) > 0 && configList[0] != nil {
			topicConfig := configList[0].(map[string]interface{})
			switch version {
			case "2.6":
				cfg, err := expandKafkaTopicConfig2_6(topicConfig)
				if err != nil {
					return nil, err
				}
				topic.SetTopicConfig_2_6(cfg)
			case "2.1":
				cfg, err := expandKafkaTopicConfig2_1(topicConfig)
				if err != nil {
					return nil, err
				}
				topic.SetTopicConfig_2_1(cfg)
			case "2.8":
				cfg, err := expandKafkaTopicConfig2_8(topicConfig)
				if err != nil {
					return nil, err
				}
				topic.SetTopicConfig_2_8(cfg)
			default:
				return nil, fmt.Errorf("specified version %v of Kafka is not supported", version)
			}
		}
	}
	return topic, nil
}
