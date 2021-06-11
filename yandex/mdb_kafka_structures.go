package yandex

import (
	"bytes"
	"fmt"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
	if !ok {
		return 0, fmt.Errorf("value for 'compression_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(kafka.CompressionType_value)), e)
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
	if !ok {
		return 0, fmt.Errorf("value for 'cleanup_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(Topic_CleanupPolicy_value)), e)
	}
	return TopicCleanupPolicy(v), nil
}

func expandKafkaConfig2_6(d *schema.ResourceData, rootKey string) (*kafka.KafkaConfig2_6, error) {
	res := &kafka.KafkaConfig2_6{}

	if v, ok := d.GetOk(rootKey + ".compression_type"); ok {
		value, err := parseKafkaCompression(v.(string))
		if err != nil {
			return nil, err
		}
		res.CompressionType = value
	}
	if v, ok := d.GetOk(rootKey + ".log_flush_interval_messages"); ok {
		res.LogFlushIntervalMessages = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_flush_interval_ms"); ok {
		res.LogFlushIntervalMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_flush_scheduler_interval_ms"); ok {
		res.LogFlushSchedulerIntervalMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_bytes"); ok {
		res.LogRetentionBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_hours"); ok {
		res.LogRetentionHours = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_minutes"); ok {
		res.LogRetentionMinutes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_ms"); ok {
		res.LogRetentionMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_segment_bytes"); ok {
		res.LogSegmentBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".log_preallocate"); ok {
		res.LogPreallocate = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".socket_send_buffer_bytes"); ok {
		res.SocketSendBufferBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".socket_receive_buffer_bytes"); ok {
		res.SocketReceiveBufferBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".auto_create_topics_enable"); ok {
		res.AutoCreateTopicsEnable = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".num_partitions"); ok {
		res.NumPartitions = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".default_replication_factor"); ok {
		res.DefaultReplicationFactor = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return res, nil
}

func expandKafkaConfig2_1(d *schema.ResourceData, rootKey string) (*kafka.KafkaConfig2_1, error) {
	res := &kafka.KafkaConfig2_1{}

	if v, ok := d.GetOk(rootKey + ".compression_type"); ok {
		value, err := parseKafkaCompression(v.(string))
		if err != nil {
			return nil, err
		}
		res.CompressionType = value
	}
	if v, ok := d.GetOk(rootKey + ".log_flush_interval_messages"); ok {
		res.LogFlushIntervalMessages = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_flush_interval_ms"); ok {
		res.LogFlushIntervalMs = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_flush_scheduler_interval_ms"); ok {
		res.LogFlushSchedulerIntervalMs = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_bytes"); ok {
		res.LogRetentionBytes = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_hours"); ok {
		res.LogRetentionHours = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_minutes"); ok {
		res.LogRetentionMinutes = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_retention_ms"); ok {
		res.LogRetentionMs = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_segment_bytes"); ok {
		res.LogSegmentBytes = v.(*wrappers.Int64Value)
	}
	if v, ok := d.GetOk(rootKey + ".log_preallocate"); ok {
		res.LogPreallocate = v.(*wrappers.BoolValue)
	}
	if v, ok := d.GetOk(rootKey + ".socket_send_buffer_bytes"); ok {
		res.SocketSendBufferBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".socket_receive_buffer_bytes"); ok {
		res.SocketReceiveBufferBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".auto_create_topics_enable"); ok {
		res.AutoCreateTopicsEnable = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".num_partitions"); ok {
		res.NumPartitions = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".default_replication_factor"); ok {
		res.DefaultReplicationFactor = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return res, nil
}

func expandKafkaTopicConfig2_6(config map[string]interface{}) (*kafka.TopicConfig2_6, error) {
	res := &kafka.TopicConfig2_6{}

	if v, ok := config["cleanup_policy"]; ok {
		_, err := parseKafkaTopicCleanupPolicy(v.(string))
		if err == nil {
			res.CleanupPolicy = kafka.TopicConfig2_6_CleanupPolicy(kafka.TopicConfig2_6_CleanupPolicy_value[v.(string)])
		}
	}
	if v, ok := config["compression_type"]; ok {
		value, err := parseKafkaCompression(v.(string))
		if err == nil {
			res.CompressionType = value
		}
	}
	if v, ok := config["delete_retention_ms"]; ok {
		res.DeleteRetentionMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["file_delete_delay_ms"]; ok {
		res.FileDeleteDelayMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["flush_messages"]; ok {
		res.FlushMessages = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["flush_ms"]; ok {
		res.FlushMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["min_compaction_lag_ms"]; ok {
		res.MinCompactionLagMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["retention_bytes"]; ok {
		res.RetentionBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["retention_ms"]; ok {
		res.RetentionMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["max_message_bytes"]; ok {
		res.MaxMessageBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["min_insync_replicas"]; ok {
		res.MinInsyncReplicas = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["segment_bytes"]; ok {
		res.SegmentBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["preallocate"]; ok {
		res.Preallocate = &wrappers.BoolValue{Value: v.(bool)}
	}
	return res, nil
}

func expandKafkaTopicConfig2_1(config map[string]interface{}) (*kafka.TopicConfig2_1, error) {
	res := &kafka.TopicConfig2_1{}

	if v, ok := config["cleanup_policy"]; ok {
		_, err := parseKafkaTopicCleanupPolicy(v.(string))
		if err == nil {
			res.CleanupPolicy = kafka.TopicConfig2_1_CleanupPolicy(kafka.TopicConfig2_1_CleanupPolicy_value[v.(string)])
		}
	}
	if v, ok := config["compression_type"]; ok {
		value, err := parseKafkaCompression(v.(string))
		if err == nil {
			res.CompressionType = value
		}
	}
	if v, ok := config["delete_retention_ms"]; ok {
		res.DeleteRetentionMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["file_delete_delay_ms"]; ok {
		res.FileDeleteDelayMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["flush_messages"]; ok {
		res.FlushMessages = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["flush_ms"]; ok {
		res.FlushMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["min_compaction_lag_ms"]; ok {
		res.MinCompactionLagMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["retention_bytes"]; ok {
		res.RetentionBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["retention_ms"]; ok {
		res.RetentionMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["max_message_bytes"]; ok {
		res.MaxMessageBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["min_insync_replicas"]; ok {
		res.MinInsyncReplicas = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["segment_bytes"]; ok {
		res.SegmentBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := config["preallocate"]; ok {
		res.Preallocate = &wrappers.BoolValue{Value: v.(bool)}
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
	case "2.6":
		cfg, err := expandKafkaConfig2_6(d, "config.0.kafka.0.kafka_config.0")
		if err != nil {
			return nil, err
		}
		result.Kafka.SetKafkaConfig_2_6(cfg)
	case "2.1":
		cfg, err := expandKafkaConfig2_1(d, "config.0.kafka.0.kafka_config.0")
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

	config := map[string]interface{}{
		"brokers_count":    cluster.Config.BrokersCount.GetValue(),
		"assign_public_ip": cluster.Config.AssignPublicIp,
		"unmanaged_topics": cluster.Config.UnmanagedTopics,
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

func flattenKafkaConfig2_6Settings(r *kafka.KafkaConfig2_6) (map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["compression_type"] = r.GetCompressionType().String()
	res["log_flush_interval_messages"] = r.GetLogFlushIntervalMessages().GetValue()
	res["log_flush_interval_ms"] = r.GetLogFlushIntervalMs().GetValue()
	res["log_flush_scheduler_interval_ms"] = r.GetLogFlushSchedulerIntervalMs().GetValue()
	res["log_retention_bytes"] = r.GetLogRetentionBytes().GetValue()
	res["log_retention_hours"] = r.GetLogRetentionHours().GetValue()
	res["log_retention_minutes"] = r.GetLogRetentionMinutes().GetValue()
	res["log_retention_ms"] = r.GetLogRetentionMs().GetValue()
	res["log_segment_bytes"] = r.GetLogSegmentBytes().GetValue()
	res["log_preallocate"] = r.GetLogPreallocate().GetValue()
	res["socket_send_buffer_bytes"] = r.GetSocketSendBufferBytes().GetValue()
	res["socket_receive_buffer_bytes"] = r.GetSocketReceiveBufferBytes().GetValue()
	res["auto_create_topics_enable"] = r.GetAutoCreateTopicsEnable().GetValue()
	res["num_partitions"] = r.GetNumPartitions().GetValue()
	res["default_replication_factor"] = r.GetDefaultReplicationFactor().GetValue()

	return res, nil
}

func flattenKafkaConfig2_1Settings(r *kafka.KafkaConfig2_1) (map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["compression_type"] = r.GetCompressionType().String()
	res["log_flush_interval_messages"] = r.GetLogFlushIntervalMessages()
	res["log_flush_interval_ms"] = r.GetLogFlushIntervalMs()
	res["log_flush_scheduler_interval_ms"] = r.GetLogFlushSchedulerIntervalMs()
	res["log_retention_bytes"] = r.GetLogRetentionBytes()
	res["log_retention_hours"] = r.GetLogRetentionHours()
	res["log_retention_minutes"] = r.GetLogRetentionMinutes()
	res["log_retention_ms"] = r.GetLogRetentionMs()
	res["log_segment_bytes"] = r.GetLogSegmentBytes()
	res["log_preallocate"] = r.GetLogPreallocate()
	res["socket_send_buffer_bytes"] = r.GetSocketSendBufferBytes().GetValue()
	res["socket_receive_buffer_bytes"] = r.GetSocketReceiveBufferBytes().GetValue()
	res["auto_create_topics_enable"] = r.GetAutoCreateTopicsEnable().GetValue()
	res["num_partitions"] = r.GetNumPartitions().GetValue()
	res["default_replication_factor"] = r.GetDefaultReplicationFactor().GetValue()

	return res, nil
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
		if len(cfg) != 0 {
			m["topic_config"] = []map[string]interface{}{cfg}
		}
		result = append(result, m)
	}

	return result
}

func flattenKafkaTopicConfig2_6(topicConfig *kafka.TopicConfig2_6) map[string]interface{} {
	result := make(map[string]interface{})

	if topicConfig.GetCleanupPolicy() != kafka.TopicConfig2_6_CLEANUP_POLICY_UNSPECIFIED {
		result["cleanup_policy"] = topicConfig.GetCleanupPolicy().String()
	}
	if topicConfig.GetCompressionType() != kafka.CompressionType_COMPRESSION_TYPE_UNSPECIFIED {
		result["compression_type"] = topicConfig.GetCompressionType().String()
	}
	if topicConfig.GetDeleteRetentionMs() != nil {
		result["delete_retention_ms"] = topicConfig.GetDeleteRetentionMs().GetValue()
	}
	if topicConfig.GetFileDeleteDelayMs() != nil {
		result["file_delete_delay_ms"] = topicConfig.GetFileDeleteDelayMs().GetValue()
	}
	if topicConfig.GetFlushMessages() != nil {
		result["flush_messages"] = topicConfig.GetFlushMessages().GetValue()
	}
	if topicConfig.GetFlushMs() != nil {
		result["flush_ms"] = topicConfig.GetFlushMs().GetValue()
	}
	if topicConfig.GetMinCompactionLagMs() != nil {
		result["min_compaction_lag_ms"] = topicConfig.GetMinCompactionLagMs().GetValue()
	}
	if topicConfig.GetRetentionBytes() != nil {
		result["retention_bytes"] = topicConfig.GetRetentionBytes().GetValue()
	}
	if topicConfig.GetRetentionMs() != nil {
		result["retention_ms"] = topicConfig.GetRetentionMs().GetValue()
	}
	if topicConfig.GetMaxMessageBytes() != nil {
		result["max_message_bytes"] = topicConfig.GetMaxMessageBytes().GetValue()
	}
	if topicConfig.GetMinInsyncReplicas() != nil {
		result["min_insync_replicas"] = topicConfig.GetMinInsyncReplicas().GetValue()
	}
	if topicConfig.GetSegmentBytes() != nil {
		result["segment_bytes"] = topicConfig.GetSegmentBytes().GetValue()
	}
	if topicConfig.GetPreallocate() != nil {
		result["preallocate"] = topicConfig.GetPreallocate().GetValue()
	}
	return result
}

func flattenKafkaTopicConfig2_1(topicConfig *kafka.TopicConfig2_1) map[string]interface{} {
	result := make(map[string]interface{})

	if topicConfig.GetCleanupPolicy() != kafka.TopicConfig2_1_CLEANUP_POLICY_UNSPECIFIED {
		result["cleanup_policy"] = topicConfig.GetCleanupPolicy().String()
	}
	if topicConfig.GetCompressionType() != kafka.CompressionType_COMPRESSION_TYPE_UNSPECIFIED {
		result["compression_type"] = topicConfig.GetCompressionType().String()
	}
	if topicConfig.GetDeleteRetentionMs() != nil {
		result["delete_retention_ms"] = topicConfig.GetDeleteRetentionMs().GetValue()
	}
	if topicConfig.GetFileDeleteDelayMs() != nil {
		result["file_delete_delay_ms"] = topicConfig.GetFileDeleteDelayMs().GetValue()
	}
	if topicConfig.GetFlushMessages() != nil {
		result["flush_messages"] = topicConfig.GetFlushMessages().GetValue()
	}
	if topicConfig.GetFlushMs() != nil {
		result["flush_ms"] = topicConfig.GetFlushMs().GetValue()
	}
	if topicConfig.GetMinCompactionLagMs() != nil {
		result["min_compaction_lag_ms"] = topicConfig.GetMinCompactionLagMs().GetValue()
	}
	if topicConfig.GetRetentionBytes() != nil {
		result["retention_bytes"] = topicConfig.GetRetentionBytes().GetValue()
	}
	if topicConfig.GetRetentionMs() != nil {
		result["retention_ms"] = topicConfig.GetRetentionMs().GetValue()
	}
	if topicConfig.GetMaxMessageBytes() != nil {
		result["max_message_bytes"] = topicConfig.GetMaxMessageBytes().GetValue()
	}
	if topicConfig.GetMinInsyncReplicas() != nil {
		result["min_insync_replicas"] = topicConfig.GetMinInsyncReplicas().GetValue()
	}
	if topicConfig.GetSegmentBytes() != nil {
		result["segment_bytes"] = topicConfig.GetSegmentBytes().GetValue()
	}
	if topicConfig.GetPreallocate() != nil {
		result["preallocate"] = topicConfig.GetPreallocate().GetValue()
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
		switch version {
		case "2.6":
			configList := v.([]interface{})
			if len(configList) > 0 {
				cfg, err := expandKafkaTopicConfig2_6(configList[0].(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				topic.SetTopicConfig_2_6(cfg)
			}
		case "2.1":
			configList := v.([]interface{})
			if len(configList) > 0 {
				cfg, err := expandKafkaTopicConfig2_1(configList[0].(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				topic.SetTopicConfig_2_1(cfg)
			}
		default:
			return nil, fmt.Errorf("specified version %v of Kafka is not supported", version)
		}
	}
	return topic, nil
}
