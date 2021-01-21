package yandex

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
)

// Sorts list of hosts in accordance with the order in config.
// We need to keep the original order so there's no diff appears on each apply.
// Removes implicit ZooKeeper hosts from the `hosts` slice.
func sortClickHouseHosts(hosts []*clickhouse.Host, specs []*clickhouse.HostSpec) []*clickhouse.Host {
	implicitZk := true
	for _, h := range specs {
		if h.Type == clickhouse.Host_ZOOKEEPER {
			implicitZk = false
			break
		}
	}

	if implicitZk {
		n := 0
		for _, h := range hosts {
			// Filter out implicit ZooKeeper hosts.
			if h.Type == clickhouse.Host_CLICKHOUSE {
				hosts[n] = h
				n++
			}
		}
		hosts = hosts[:n]
	}

	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if h.ZoneId == hosts[j].ZoneId && (h.ShardName == "" || h.ShardName == hosts[j].ShardName) && h.Type == hosts[j].Type {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}
	return hosts
}

func clickHouseUserPermissionHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["database_name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func clickHouseUserHash(v interface{}) int {
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

func clickHouseDatabaseHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

// Takes the current list of dbs and the desirable list of dbs.
// Returns the slice of dbs to delete and the slice of dbs to add.
func clickHouseDatabasesDiff(currDBs []*clickhouse.Database, targetDBs []*clickhouse.DatabaseSpec) ([]string, []string) {
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

// Takes the current list of users and the desirable list of users.
// Returns the slice of usernames to delete and the slice of users to add.
func clickHouseUsersDiff(currUsers []*clickhouse.User, targetUsers []*clickhouse.UserSpec) ([]string, []*clickhouse.UserSpec) {
	m := map[string]bool{}
	toDelete := map[string]bool{}
	toAdd := []*clickhouse.UserSpec{}

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

// Takes the old set of user specs and the new set of user specs.
// Returns the slice of user specs which have changed.
func clickHouseChangedUsers(oldSpecs *schema.Set, newSpecs *schema.Set) []*clickhouse.UserSpec {
	result := []*clickhouse.UserSpec{}
	m := map[string]*clickhouse.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}))
		m[user.Name] = user
	}
	for _, spec := range newSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}))
		if u, ok := m[user.Name]; ok {
			if user.Password != u.Password || fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				result = append(result, user)
			}
		}
	}
	return result
}

// Takes the current list of hosts and the desirable list of hosts.
// Returns the map of hostnames to delete grouped by shard,
// and the map of hosts to add grouped by shard as well.
// All the ZOOKEEPER hosts will reside under the key "zk".
func clickHouseHostsDiff(currHosts []*clickhouse.Host, targetHosts []*clickhouse.HostSpec) (map[string][]string, map[string][]*clickhouse.HostSpec) {
	m := map[string][]*clickhouse.HostSpec{}

	for _, h := range targetHosts {
		shardName := "shard1"
		if h.ShardName != "" {
			shardName = h.ShardName
		}
		if h.Type == clickhouse.Host_ZOOKEEPER {
			shardName = "zk"
		}
		key := h.Type.String() + h.ZoneId + shardName
		m[key] = append(m[key], h)
	}

	toDelete := map[string][]string{}
	for _, h := range currHosts {
		shardName := h.ShardName
		if h.Type == clickhouse.Host_ZOOKEEPER {
			shardName = "zk"
		}
		key := h.Type.String() + h.ZoneId + shardName
		hs, ok := m[key]
		if !ok {
			toDelete[shardName] = append(toDelete[h.ShardName], h.Name)
		}
		if len(hs) > 1 {
			m[key] = hs[1:]
		} else {
			delete(m, key)
		}
	}

	toAdd := map[string][]*clickhouse.HostSpec{}
	for _, hs := range m {
		for _, h := range hs {
			if h.Type == clickhouse.Host_ZOOKEEPER {
				toAdd["zk"] = append(toAdd["zk"], h)
			} else {
				toAdd[h.ShardName] = append(toAdd[h.ShardName], h)
			}
		}
	}

	return toDelete, toAdd
}

type shardGroupDiffInfo struct {
	toDelete []string
	toUpdate []*clickhouse.ShardGroup
	toAdd    []*clickhouse.ShardGroup
}

// Takes the current list of shard groups and the desirable list of shard groups.
// Returns the list of shard group names to delete, to update and to add
func clickHouseShardGroupDiff(currGroups []*clickhouse.ShardGroup, targetGroups []*clickhouse.ShardGroup) shardGroupDiffInfo {
	m := map[string]*clickhouse.ShardGroup{}

	for _, g := range targetGroups {
		m[g.Name] = g
	}

	var toDelete []string
	var toAdd []*clickhouse.ShardGroup
	var toUpdate []*clickhouse.ShardGroup

	for _, currentGroup := range currGroups {
		if targetGroup, ok := m[currentGroup.Name]; ok {
			if currentGroup.Description != targetGroup.Description || !reflect.DeepEqual(currentGroup.ShardNames, targetGroup.ShardNames) {
				toUpdate = append(toUpdate, targetGroup)
			}

			delete(m, currentGroup.Name)
		} else {
			toDelete = append(toDelete, currentGroup.Name)
		}
	}

	for _, sg := range m {
		toAdd = append(toAdd, sg)
	}

	return shardGroupDiffInfo{toDelete, toUpdate, toAdd}
}

type formatSchemaDiffInfo struct {
	toDelete []string
	toUpdate []*clickhouse.FormatSchema
	toAdd    []*clickhouse.FormatSchema
}

// Takes the current list of format schemas and the desirable list of format schemas.
// Returns the list of format schemas names to delete, to update and to add
func clickHouseFormatSchemaDiff(currSchemas []*clickhouse.FormatSchema, targetSchemas []*clickhouse.FormatSchema) formatSchemaDiffInfo {
	m := map[string]*clickhouse.FormatSchema{}

	for _, s := range targetSchemas {
		m[s.Name] = s
	}

	var toDelete []string
	var toAdd []*clickhouse.FormatSchema
	var toUpdate []*clickhouse.FormatSchema

	for _, currentSchema := range currSchemas {
		if targetSchema, ok := m[currentSchema.Name]; ok {
			if currentSchema.Type != targetSchema.Type {
				toDelete = append(toDelete, currentSchema.Name)
			} else {
				if currentSchema.Uri != targetSchema.Uri {
					toUpdate = append(toUpdate, targetSchema)
				}

				delete(m, currentSchema.Name)
			}
		} else {
			toDelete = append(toDelete, currentSchema.Name)
		}
	}

	for _, fs := range m {
		toAdd = append(toAdd, fs)
	}

	return formatSchemaDiffInfo{toDelete, toUpdate, toAdd}
}

type mlModelDiffInfo struct {
	toDelete []string
	toUpdate []*clickhouse.MlModel
	toAdd    []*clickhouse.MlModel
}

// Takes the current list of ml models and the desirable list of ml models.
// Returns the list of ml model names to delete, to update and to add
func clickHouseMlModelDiff(currModels []*clickhouse.MlModel, targetModels []*clickhouse.MlModel) mlModelDiffInfo {
	m := map[string]*clickhouse.MlModel{}

	for _, s := range targetModels {
		m[s.Name] = s
	}

	var toDelete []string
	var toAdd []*clickhouse.MlModel
	var toUpdate []*clickhouse.MlModel

	for _, currentModel := range currModels {
		if targetModel, ok := m[currentModel.Name]; ok {
			if currentModel.Type != targetModel.Type {
				toDelete = append(toDelete, currentModel.Name)
			} else {
				if currentModel.Uri != targetModel.Uri {
					toUpdate = append(toUpdate, targetModel)
				}

				delete(m, currentModel.Name)
			}
		} else {
			toDelete = append(toDelete, currentModel.Name)
		}
	}

	for _, model := range m {
		toAdd = append(toAdd, model)
	}

	return mlModelDiffInfo{toDelete, toUpdate, toAdd}
}

func parseClickHouseEnv(e string) (clickhouse.Cluster_Environment, error) {
	v, ok := clickhouse.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(clickhouse.Cluster_Environment_value)), e)
	}
	return clickhouse.Cluster_Environment(v), nil
}

func parseClickHouseHostType(t string) (clickhouse.Host_Type, error) {
	v, ok := clickhouse.Host_Type_value[t]
	if !ok {
		return 0, fmt.Errorf("value for 'host.type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(clickhouse.Host_Type_value)), t)
	}
	return clickhouse.Host_Type(v), nil
}

func expandClickHouseHosts(d *schema.ResourceData) ([]*clickhouse.HostSpec, error) {
	var result []*clickhouse.HostSpec
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host, err := expandClickHouseHost(config)
		if err != nil {
			return nil, err
		}
		result = append(result, host)
	}

	return result, nil
}

func expandClickHouseHost(config map[string]interface{}) (*clickhouse.HostSpec, error) {
	host := &clickhouse.HostSpec{}
	if v, ok := config["zone"]; ok {
		host.ZoneId = v.(string)
	}

	if v, ok := config["type"]; ok {
		t, err := parseClickHouseHostType(v.(string))
		if err != nil {
			return nil, err
		}
		host.Type = t
	}

	if v, ok := config["subnet_id"]; ok {
		host.SubnetId = v.(string)
	}

	if v, ok := config["shard_name"]; ok {
		host.ShardName = v.(string)
		if host.Type == clickhouse.Host_ZOOKEEPER && host.ShardName != "" {
			return nil, fmt.Errorf("ZooKeeper hosts cannot have a 'shard_name'")
		}
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.AssignPublicIp = v.(bool)
	}

	return host, nil
}

func flattenClickHouseDatabases(dbs []*clickhouse.Database) *schema.Set {
	result := schema.NewSet(clickHouseDatabaseHash, nil)

	for _, d := range dbs {
		m := make(map[string]interface{})
		m["name"] = d.Name
		result.Add(m)
	}
	return result
}

func flattenClickHouseResources(r *clickhouse.Resources) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = toGigabytes(r.DiskSize)

	return []map[string]interface{}{res}, nil
}

func flattenClickhouseMergeTreeConfig(c *clickhouseConfig.ClickhouseConfig_MergeTree) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["replicated_deduplication_window"] = c.ReplicatedDeduplicationWindow.Value
	res["replicated_deduplication_window_seconds"] = c.ReplicatedDeduplicationWindowSeconds.Value
	res["parts_to_delay_insert"] = c.PartsToDelayInsert.Value
	res["parts_to_throw_insert"] = c.PartsToThrowInsert.Value
	res["max_replicated_merges_in_queue"] = c.MaxReplicatedMergesInQueue.Value
	res["number_of_free_entries_in_pool_to_lower_max_size_of_merge"] = c.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge.Value
	res["max_bytes_to_merge_at_min_space_in_pool"] = c.MaxBytesToMergeAtMinSpaceInPool.Value

	return []map[string]interface{}{res}, nil
}

func flattenClickhouseKafkaSettings(d *schema.ResourceData, keyPath string, c *clickhouseConfig.ClickhouseConfig_Kafka) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["security_protocol"] = c.SecurityProtocol.String()
	res["sasl_mechanism"] = c.SaslMechanism.String()
	res["sasl_username"] = c.SaslUsername
	if v, ok := d.GetOk(keyPath + ".sasl_password"); ok {
		res["sasl_password"] = v.(string)
	}

	return []map[string]interface{}{res}, nil
}

func flattenClickhouseKafkaTopicsSettings(d *schema.ResourceData, c []*clickhouseConfig.ClickhouseConfig_KafkaTopic) ([]interface{}, error) {
	var result []interface{}

	for i, t := range c {
		settings, err := flattenClickhouseKafkaSettings(d, fmt.Sprintf("clickhouse.0.config.0.kafka_topic.%d.settings.0", i), t.Settings)
		if err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"name":     t.Name,
			"settings": settings,
		})
	}

	return result, nil
}

func flattenClickhouseRabbitmqSettings(d *schema.ResourceData, c *clickhouseConfig.ClickhouseConfig_Rabbitmq) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["username"] = c.Username
	if v, ok := d.GetOk("clickhouse.0.config.0.rabbitmq.0.password"); ok {
		res["password"] = v.(string)
	}

	return []map[string]interface{}{res}, nil
}

func flattenClickhouseCompressionSettings(c []*clickhouseConfig.ClickhouseConfig_Compression) ([]interface{}, error) {
	var result []interface{}

	for _, r := range c {
		result = append(result, map[string]interface{}{
			"method":              r.Method.String(),
			"min_part_size":       r.MinPartSize,
			"min_part_size_ratio": r.MinPartSizeRatio,
		})
	}

	return result, nil
}

func flattenClickhouseGraphiteRollupSettings(c []*clickhouseConfig.ClickhouseConfig_GraphiteRollup) ([]interface{}, error) {
	var result []interface{}

	for _, r := range c {
		rollup := map[string]interface{}{
			"name":    r.Name,
			"pattern": []interface{}{},
		}
		for _, p := range r.Patterns {
			pattern := map[string]interface{}{
				"function":  p.Function,
				"regexp":    p.Regexp,
				"retention": []interface{}{},
			}

			for _, r := range p.Retention {
				pattern["retention"] = append(pattern["retention"].([]interface{}), map[string]interface{}{
					"age":       r.Age,
					"precision": r.Precision,
				})
			}

			rollup["pattern"] = append(rollup["pattern"].([]interface{}), pattern)
		}

		result = append(result, rollup)
	}

	return result, nil
}

func flattenClickHouseConfig(d *schema.ResourceData, c *clickhouseConfig.ClickhouseConfigSet) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["log_level"] = c.EffectiveConfig.LogLevel.String()
	res["max_connections"] = c.EffectiveConfig.MaxConnections.Value
	res["max_concurrent_queries"] = c.EffectiveConfig.MaxConcurrentQueries.Value
	res["keep_alive_timeout"] = c.EffectiveConfig.KeepAliveTimeout.Value
	res["uncompressed_cache_size"] = c.EffectiveConfig.UncompressedCacheSize.Value
	res["mark_cache_size"] = c.EffectiveConfig.MarkCacheSize.Value
	res["max_table_size_to_drop"] = c.EffectiveConfig.MaxTableSizeToDrop.Value
	res["max_partition_size_to_drop"] = c.EffectiveConfig.MaxPartitionSizeToDrop.Value
	res["timezone"] = c.EffectiveConfig.Timezone
	res["geobase_uri"] = c.EffectiveConfig.GeobaseUri
	res["query_log_retention_size"] = c.EffectiveConfig.QueryLogRetentionSize.Value
	res["query_log_retention_time"] = c.EffectiveConfig.QueryLogRetentionTime.Value
	res["query_thread_log_enabled"] = c.EffectiveConfig.QueryThreadLogEnabled.Value
	res["query_thread_log_retention_size"] = c.EffectiveConfig.QueryThreadLogRetentionSize.Value
	res["query_thread_log_retention_time"] = c.EffectiveConfig.QueryLogRetentionTime.Value
	res["part_log_retention_size"] = c.EffectiveConfig.PartLogRetentionSize.Value
	res["part_log_retention_time"] = c.EffectiveConfig.PartLogRetentionTime.Value
	res["metric_log_enabled"] = c.EffectiveConfig.MetricLogEnabled.Value
	res["metric_log_retention_size"] = c.EffectiveConfig.MetricLogRetentionSize.Value
	res["metric_log_retention_time"] = c.EffectiveConfig.MetricLogRetentionTime.Value
	res["trace_log_enabled"] = c.EffectiveConfig.TraceLogEnabled.Value
	res["trace_log_retention_size"] = c.EffectiveConfig.TraceLogRetentionSize.Value
	res["trace_log_retention_time"] = c.EffectiveConfig.TraceLogRetentionTime.Value
	res["text_log_enabled"] = c.EffectiveConfig.TextLogEnabled.Value
	res["text_log_retention_size"] = c.EffectiveConfig.TextLogRetentionSize.Value
	res["text_log_retention_time"] = c.EffectiveConfig.TextLogRetentionTime.Value
	res["text_log_level"] = c.EffectiveConfig.TextLogLevel.String()

	if c.EffectiveConfig.BackgroundSchedulePoolSize != nil {
		res["background_pool_size"] = c.EffectiveConfig.BackgroundSchedulePoolSize.Value
	}

	if c.EffectiveConfig.BackgroundSchedulePoolSize != nil {
		res["background_schedule_pool_size"] = c.EffectiveConfig.BackgroundSchedulePoolSize.Value
	}

	mergeTreeSettings, err := flattenClickhouseMergeTreeConfig(c.EffectiveConfig.MergeTree)
	if err != nil {
		return nil, err
	}
	res["merge_tree"] = mergeTreeSettings

	kafkaConfig, err := flattenClickhouseKafkaSettings(d, "clickhouse.0.config.0.kafka.0", c.EffectiveConfig.Kafka)
	if err != nil {
		return nil, err
	}
	res["kafka"] = kafkaConfig

	kafkaTopicsConfig, err := flattenClickhouseKafkaTopicsSettings(d, c.EffectiveConfig.KafkaTopics)
	if err != nil {
		return nil, err
	}
	res["kafka_topic"] = kafkaTopicsConfig

	rabbitmqSettings, err := flattenClickhouseRabbitmqSettings(d, c.EffectiveConfig.Rabbitmq)
	if err != nil {
		return nil, err
	}
	res["rabbitmq"] = rabbitmqSettings

	compressions, err := flattenClickhouseCompressionSettings(c.EffectiveConfig.Compression)
	if err != nil {
		return nil, err
	}
	res["compression"] = compressions

	graphiteRollups, err := flattenClickhouseGraphiteRollupSettings(c.EffectiveConfig.GraphiteRollup)
	if err != nil {
		return nil, err
	}
	res["graphite_rollup"] = graphiteRollups

	return []map[string]interface{}{res}, nil
}

func expandClickHouseResources(d *schema.ResourceData, rootKey string) *clickhouse.Resources {
	resources := &clickhouse.Resources{}

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

func expandClickhouseMergeTreeConfig(d *schema.ResourceData, rootKey string) (*clickhouseConfig.ClickhouseConfig_MergeTree, error) {
	config := &clickhouseConfig.ClickhouseConfig_MergeTree{}
	if v, ok := d.GetOk(rootKey + ".replicated_deduplication_window"); ok {
		config.ReplicatedDeduplicationWindow = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".replicated_deduplication_window_seconds"); ok {
		config.ReplicatedDeduplicationWindowSeconds = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".parts_to_delay_insert"); ok {
		config.PartsToDelayInsert = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".parts_to_throw_insert"); ok {
		config.PartsToThrowInsert = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".max_replicated_merges_in_queue"); ok {
		config.MaxReplicatedMergesInQueue = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".number_of_free_entries_in_pool_to_lower_max_size_of_merge"); ok {
		config.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".max_bytes_to_merge_at_min_space_in_pool"); ok {
		config.MaxBytesToMergeAtMinSpaceInPool = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return config, nil
}

func expandEnum(keyName string, value string, enumValues map[string]int32) (*int32, error) {
	if val, ok := enumValues[value]; ok {
		return &val, nil
	} else {
		return nil, fmt.Errorf("value for '%s' must be one of %s, not `%s`",
			keyName, getJoinedKeys(getEnumValueMapKeys(enumValues)), value)
	}
}

func expandClickhouseKafkaSettings(d *schema.ResourceData, rootKey string) (*clickhouseConfig.ClickhouseConfig_Kafka, error) {
	config := &clickhouseConfig.ClickhouseConfig_Kafka{}

	if v, ok := d.GetOk(rootKey + ".security_protocol"); ok {
		if val, err := expandEnum("security_protocol", v.(string), clickhouseConfig.ClickhouseConfig_Kafka_SecurityProtocol_value); val != nil && err == nil {
			config.SecurityProtocol = clickhouseConfig.ClickhouseConfig_Kafka_SecurityProtocol(*val)
		} else {
			return nil, err
		}
	}
	if v, ok := d.GetOk(rootKey + ".sasl_mechanism"); ok {
		if val, err := expandEnum("sasl_mechanism", v.(string), clickhouseConfig.ClickhouseConfig_Kafka_SaslMechanism_value); val != nil && err == nil {
			config.SaslMechanism = clickhouseConfig.ClickhouseConfig_Kafka_SaslMechanism(*val)
		} else {
			return nil, err
		}
	}
	if v, ok := d.GetOk(rootKey + ".sasl_username"); ok {
		config.SaslUsername = v.(string)
	}
	if v, ok := d.GetOk(rootKey + ".sasl_password"); ok {
		config.SaslPassword = v.(string)
	}

	return config, nil
}

func expandClickhouseKafkaTopicsSettings(d *schema.ResourceData, rootKey string) ([]*clickhouseConfig.ClickhouseConfig_KafkaTopic, error) {
	var result []*clickhouseConfig.ClickhouseConfig_KafkaTopic
	topics := d.Get(rootKey).([]interface{})

	for i := range topics {
		settings, err := expandClickhouseKafkaSettings(d, rootKey+fmt.Sprintf(".%d.settings.0", i))
		if err != nil {
			return nil, err
		}

		result = append(result, &clickhouseConfig.ClickhouseConfig_KafkaTopic{
			Name:     d.Get(rootKey + fmt.Sprintf(".%d.name", i)).(string),
			Settings: settings,
		})
	}
	return result, nil
}

func expandClickhouseRabbitmqSettings(d *schema.ResourceData, rootKey string) (*clickhouseConfig.ClickhouseConfig_Rabbitmq, error) {
	config := &clickhouseConfig.ClickhouseConfig_Rabbitmq{}

	if v, ok := d.GetOk(rootKey + ".username"); ok {
		config.Username = v.(string)
	}
	if v, ok := d.GetOk(rootKey + ".password"); ok {
		config.Password = v.(string)
	}

	return config, nil
}

func expandClickhouseCompressionSettings(d *schema.ResourceData, rootKey string) ([]*clickhouseConfig.ClickhouseConfig_Compression, error) {
	var result []*clickhouseConfig.ClickhouseConfig_Compression
	compressions := d.Get(rootKey).([]interface{})

	for i := range compressions {
		keyPrefix := rootKey + fmt.Sprintf(".%d", i)
		compression := &clickhouseConfig.ClickhouseConfig_Compression{}

		if v, ok := d.GetOk(keyPrefix + ".method"); ok {
			if val, err := expandEnum("method", v.(string), clickhouseConfig.ClickhouseConfig_Compression_Method_value); val != nil && err == nil {
				compression.Method = clickhouseConfig.ClickhouseConfig_Compression_Method(*val)
			} else {
				return nil, err
			}
		}
		if v, ok := d.GetOk(keyPrefix + ".min_part_size"); ok {
			compression.MinPartSize = int64(v.(int))
		}
		if v, ok := d.GetOk(keyPrefix + ".min_part_size_ratio"); ok {
			compression.MinPartSizeRatio = v.(float64)
		}

		result = append(result, compression)
	}
	return result, nil
}

func expandClickhouseGraphiteRollupSettings(d *schema.ResourceData, rootKey string) ([]*clickhouseConfig.ClickhouseConfig_GraphiteRollup, error) {
	var result []*clickhouseConfig.ClickhouseConfig_GraphiteRollup

	for r := range d.Get(rootKey).([]interface{}) {
		rollupKey := rootKey + fmt.Sprintf(".%d", r)
		rollup := &clickhouseConfig.ClickhouseConfig_GraphiteRollup{Name: d.Get(rollupKey + ".name").(string)}

		for p := range d.Get(rollupKey + ".pattern").([]interface{}) {
			patternKey := rollupKey + fmt.Sprintf(".pattern.%d", p)

			pattern := &clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern{
				Function: d.Get(patternKey + ".function").(string),
			}

			if v, ok := d.GetOk(patternKey + ".regexp"); ok {
				pattern.Regexp = v.(string)
			}

			for r := range d.Get(patternKey + ".retention").([]interface{}) {
				retentionKey := patternKey + fmt.Sprintf(".retention.%d", r)
				retention := &clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
					Age:       int64(d.Get(retentionKey + ".age").(int)),
					Precision: int64(d.Get(retentionKey + ".precision").(int)),
				}
				pattern.Retention = append(pattern.Retention, retention)
			}

			rollup.Patterns = append(rollup.Patterns, pattern)
		}

		result = append(result, rollup)
	}
	return result, nil
}

func expandClickHouseConfig(d *schema.ResourceData, rootKey string) (*clickhouseConfig.ClickhouseConfig, error) {
	config := &clickhouseConfig.ClickhouseConfig{}

	if v, ok := d.GetOk(rootKey + ".log_level"); ok {
		if val, err := expandEnum("log_level", v.(string), clickhouseConfig.ClickhouseConfig_LogLevel_value); val != nil && err == nil {
			config.LogLevel = clickhouseConfig.ClickhouseConfig_LogLevel(*val)
		} else {
			return nil, err
		}
	}
	if v, ok := d.GetOk(rootKey + ".max_connections"); ok {
		config.MaxConnections = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".max_concurrent_queries"); ok {
		config.MaxConcurrentQueries = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".keep_alive_timeout"); ok {
		config.KeepAliveTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".uncompressed_cache_size"); ok {
		config.UncompressedCacheSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".mark_cache_size"); ok {
		config.MarkCacheSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".max_table_size_to_drop"); ok {
		config.MaxTableSizeToDrop = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".max_partition_size_to_drop"); ok {
		config.MaxPartitionSizeToDrop = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".timezone"); ok {
		config.Timezone = v.(string)
	}
	if v, ok := d.GetOk(rootKey + ".geobase_uri"); ok {
		config.GeobaseUri = v.(string)
	}
	if v, ok := d.GetOk(rootKey + ".query_log_retention_size"); ok {
		config.QueryLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".query_log_retention_time"); ok {
		config.QueryLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".query_thread_log_enabled"); ok {
		config.QueryThreadLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".query_thread_log_retention_size"); ok {
		config.QueryThreadLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".query_thread_log_retention_time"); ok {
		config.QueryThreadLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".part_log_retention_size"); ok {
		config.PartLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".part_log_retention_time"); ok {
		config.PartLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".metric_log_enabled"); ok {
		config.MetricLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".metric_log_retention_size"); ok {
		config.MetricLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".metric_log_retention_time"); ok {
		config.MetricLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".trace_log_enabled"); ok {
		config.TraceLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".trace_log_retention_size"); ok {
		config.TraceLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".trace_log_retention_time"); ok {
		config.TraceLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".text_log_enabled"); ok {
		config.TextLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".text_log_retention_size"); ok {
		config.TextLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".text_log_retention_time"); ok {
		config.TextLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".text_log_level"); ok {
		if val, err := expandEnum("text_log_level", v.(string), clickhouseConfig.ClickhouseConfig_LogLevel_value); val != nil && err == nil {
			config.TextLogLevel = clickhouseConfig.ClickhouseConfig_LogLevel(*val)
		} else {
			return nil, err
		}
	}
	if v, ok := d.GetOk(rootKey + ".background_pool_size"); ok {
		config.BackgroundPoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".background_schedule_pool_size"); ok {
		config.BackgroundSchedulePoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	mergeTreeSettings, err := expandClickhouseMergeTreeConfig(d, rootKey+".merge_tree.0")
	if err != nil {
		return nil, err
	}
	config.MergeTree = mergeTreeSettings

	kafkaConfig, err := expandClickhouseKafkaSettings(d, rootKey+".kafka.0")
	if err != nil {
		return nil, err
	}
	config.Kafka = kafkaConfig

	kafkaTopicsConfig, err := expandClickhouseKafkaTopicsSettings(d, rootKey+".kafka_topic")
	if err != nil {
		return nil, err
	}
	config.KafkaTopics = kafkaTopicsConfig

	rabbitmqSettings, err := expandClickhouseRabbitmqSettings(d, rootKey+".rabbitmq.0")
	if err != nil {
		return nil, err
	}
	config.Rabbitmq = rabbitmqSettings

	compressions, err := expandClickhouseCompressionSettings(d, rootKey+".compression")
	if err != nil {
		return nil, err
	}
	config.Compression = compressions

	graphiteRollups, err := expandClickhouseGraphiteRollupSettings(d, rootKey+".graphite_rollup")
	if err != nil {
		return nil, err
	}
	config.GraphiteRollup = graphiteRollups

	return config, nil
}

func expandClickHouseZookeeperSpec(d *schema.ResourceData) *clickhouse.ConfigSpec_Zookeeper {
	result := &clickhouse.ConfigSpec_Zookeeper{}
	result.Resources = expandClickHouseResources(d, "zookeeper.0.resources.0")
	return result
}

func expandClickHouseSpec(d *schema.ResourceData) (*clickhouse.ConfigSpec_Clickhouse, error) {
	result := &clickhouse.ConfigSpec_Clickhouse{}
	result.Resources = expandClickHouseResources(d, "clickhouse.0.resources.0")
	config, err := expandClickHouseConfig(d, "clickhouse.0.config.0")
	if err != nil {
		return nil, err
	}

	result.Config = config

	return result, nil
}

func flattenClickHouseBackupWindowStart(t *timeofday.TimeOfDay) []map[string]interface{} {
	res := map[string]interface{}{}

	res["hours"] = int(t.Hours)
	res["minutes"] = int(t.Minutes)

	return []map[string]interface{}{res}
}

func expandClickHouseBackupWindowStart(d *schema.ResourceData) *timeofday.TimeOfDay {
	result := &timeofday.TimeOfDay{}

	if v, ok := d.GetOk("backup_window_start.0.hours"); ok {
		result.Hours = int32(v.(int))
	}
	if v, ok := d.GetOk("backup_window_start.0.minutes"); ok {
		result.Minutes = int32(v.(int))
	}
	return result
}

func flattenClickHouseAccess(a *clickhouse.Access) []map[string]interface{} {
	res := map[string]interface{}{}

	res["web_sql"] = a.WebSql
	res["data_lens"] = a.DataLens
	res["metrika"] = a.Metrika
	res["serverless"] = a.Serverless

	return []map[string]interface{}{res}
}

func expandClickHouseAccess(d *schema.ResourceData) *clickhouse.Access {
	result := &clickhouse.Access{}

	if v, ok := d.GetOk("access.0.web_sql"); ok {
		result.WebSql = v.(bool)
	}
	if v, ok := d.GetOk("access.0.data_lens"); ok {
		result.DataLens = v.(bool)
	}
	if v, ok := d.GetOk("access.0.metrika"); ok {
		result.Metrika = v.(bool)
	}
	if v, ok := d.GetOk("access.0.serverless"); ok {
		result.Serverless = v.(bool)
	}
	return result
}

func expandClickHouseUserPermissions(ps *schema.Set) []*clickhouse.Permission {
	result := []*clickhouse.Permission{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &clickhouse.Permission{}
		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}
		result = append(result, permission)
	}
	return result
}

func flattenClickHouseUsers(users []*clickhouse.User, passwords map[string]string) *schema.Set {
	result := schema.NewSet(clickHouseUserHash, nil)

	for _, user := range users {
		u := map[string]interface{}{}
		u["name"] = user.Name

		perms := schema.NewSet(clickHouseUserPermissionHash, nil)
		for _, perm := range user.Permissions {
			p := map[string]interface{}{}
			p["database_name"] = perm.DatabaseName
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

func expandClickHouseUser(u map[string]interface{}) *clickhouse.UserSpec {
	user := &clickhouse.UserSpec{}

	if v, ok := u["name"]; ok {
		user.Name = v.(string)
	}

	if v, ok := u["password"]; ok {
		user.Password = v.(string)
	}

	if v, ok := u["permission"]; ok {
		user.Permissions = expandClickHouseUserPermissions(v.(*schema.Set))
	}

	return user
}

func expandClickHouseUserSpecs(d *schema.ResourceData) ([]*clickhouse.UserSpec, error) {
	result := []*clickhouse.UserSpec{}
	users := d.Get("user").(*schema.Set)

	for _, u := range users.List() {
		m := u.(map[string]interface{})

		result = append(result, expandClickHouseUser(m))
	}

	return result, nil
}

func clickHouseUsersPasswords(users []*clickhouse.UserSpec) map[string]string {
	result := map[string]string{}
	for _, u := range users {
		result[u.Name] = u.Password
	}
	return result
}

func expandClickHouseDatabases(d *schema.ResourceData) ([]*clickhouse.DatabaseSpec, error) {
	var result []*clickhouse.DatabaseSpec
	dbs := d.Get("database").(*schema.Set).List()

	for _, d := range dbs {
		m := d.(map[string]interface{})
		db := &clickhouse.DatabaseSpec{}

		if v, ok := m["name"]; ok {
			db.Name = v.(string)
		}

		result = append(result, db)
	}
	return result, nil
}

func flattenClickHouseHosts(hs []*clickhouse.Host) ([]map[string]interface{}, error) {
	res := []map[string]interface{}{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["type"] = h.GetType().String()
		m["zone"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["shard_name"] = h.ShardName
		m["assign_public_ip"] = h.AssignPublicIp
		m["fqdn"] = h.Name
		res = append(res, m)
	}

	return res, nil
}

func expandClickHouseShardGroups(d *schema.ResourceData) ([]*clickhouse.ShardGroup, error) {
	var result []*clickhouse.ShardGroup
	groups := d.Get("shard_group").([]interface{})

	for _, g := range groups {
		result = append(result, expandClickHouseShardGroup(g.(map[string]interface{})))
	}
	return result, nil
}

func expandClickHouseShardGroup(g map[string]interface{}) *clickhouse.ShardGroup {
	group := &clickhouse.ShardGroup{}

	if v, ok := g["name"]; ok {
		group.Name = v.(string)
	}

	if v, ok := g["description"]; ok {
		group.Description = v.(string)
	}

	if v, ok := g["shard_names"]; ok {
		for _, shard := range v.([]interface{}) {
			group.ShardNames = append(group.ShardNames, shard.(string))
		}
	}

	return group
}

func flattenClickHouseShardGroups(sg []*clickhouse.ShardGroup) ([]map[string]interface{}, error) {
	var res []map[string]interface{}

	for _, g := range sg {
		m := map[string]interface{}{}
		m["name"] = g.Name
		m["description"] = g.Description
		m["shard_names"] = g.ShardNames
		res = append(res, m)
	}

	return res, nil
}

func expandClickHouseFormatSchemas(d *schema.ResourceData) ([]*clickhouse.FormatSchema, error) {
	var result []*clickhouse.FormatSchema
	schemas := d.Get("format_schema").(*schema.Set).List()

	for _, g := range schemas {
		formatSchema, err := expandClickHouseFormatSchema(g.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		result = append(result, formatSchema)
	}
	return result, nil
}

func expandClickHouseFormatSchema(g map[string]interface{}) (*clickhouse.FormatSchema, error) {
	formatSchema := &clickhouse.FormatSchema{}

	if v, ok := g["name"]; ok {
		formatSchema.Name = v.(string)
	}
	if v, ok := g["type"]; ok {
		if val, err := expandEnum("type", v.(string), clickhouse.FormatSchemaType_value); val != nil && err == nil {
			formatSchema.Type = clickhouse.FormatSchemaType(*val)
		} else {
			return nil, err
		}
	}
	if v, ok := g["uri"]; ok {
		formatSchema.Uri = v.(string)
	}

	return formatSchema, nil
}

func flattenClickHouseFormatSchemas(schemas []*clickhouse.FormatSchema) ([]map[string]interface{}, error) {
	var res []map[string]interface{}

	for _, s := range schemas {
		m := map[string]interface{}{}
		m["name"] = s.Name
		m["type"] = s.Type.String()
		m["uri"] = s.Uri
		res = append(res, m)
	}

	return res, nil
}

func expandClickHouseMlModels(d *schema.ResourceData) ([]*clickhouse.MlModel, error) {
	var result []*clickhouse.MlModel
	schemas := d.Get("ml_model").(*schema.Set).List()

	for _, g := range schemas {
		mlModel, err := expandClickHouseMlModel(g.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		result = append(result, mlModel)
	}
	return result, nil
}

func expandClickHouseMlModel(g map[string]interface{}) (*clickhouse.MlModel, error) {
	mlModel := &clickhouse.MlModel{}

	if v, ok := g["name"]; ok {
		mlModel.Name = v.(string)
	}
	if v, ok := g["type"]; ok {
		if val, err := expandEnum("type", v.(string), clickhouse.MlModelType_value); val != nil && err == nil {
			mlModel.Type = clickhouse.MlModelType(*val)
		} else {
			return nil, err
		}
	}
	if v, ok := g["uri"]; ok {
		mlModel.Uri = v.(string)
	}

	return mlModel, nil
}

func flattenClickHouseMlModels(schemas []*clickhouse.MlModel) ([]map[string]interface{}, error) {
	var res []map[string]interface{}

	for _, s := range schemas {
		m := map[string]interface{}{}
		m["name"] = s.Name
		m["type"] = s.Type.String()
		m["uri"] = s.Uri
		res = append(res, m)
	}

	return res, nil
}
