package yandex

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func clickHouseUserQuotaHash(v interface{}) int {
	var buf bytes.Buffer

	if v != nil {
		m := v.(map[string]interface{})
		if n, ok := m["interval_duration"]; ok {
			buf.WriteString(fmt.Sprintf("%v", n))
		} else {
			buf.WriteString("")
		}
	}

	return hashcode.String(buf.String())
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
	emptySettings := true
	if s, ok := m["settings"]; ok {
		for _, settings := range s.([]interface{}) {
			settings := expandClickHouseUserSettings(settings.(map[string]interface{}))
			p := flattenClickHouseUserSettings(settings)
			buf.WriteString(fmt.Sprintf("%v-", p))
			emptySettings = false
			break
		}
	}
	if emptySettings {
		settings := &clickhouse.UserSettings{}
		p := flattenClickHouseUserSettings(settings)
		buf.WriteString(fmt.Sprintf("%v-", p))
	}
	if q, ok := m["quota"]; ok {
		quotaSet := q.(*schema.Set)
		if len(quotaSet.List()) > 0 {
			quotas := expandClickHouseUserQuotas(quotaSet)
			for _, quota := range quotas {
				p := flattenClickHouseUserQuota(quota)
				buf.WriteString(fmt.Sprintf(" %v", p))
			}
		}
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
func clickHouseChangedUsers(oldSpecs *schema.Set, newSpecs *schema.Set, d *schema.ResourceData) []*clickhouse.UserSpec {
	result := []*clickhouse.UserSpec{}
	m := map[string]*clickhouse.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}), nil, 0)
		m[user.Name] = user
	}
	for _, spec := range newSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}), nil, 0)
		if u, ok := m[user.Name]; ok {
			if user.Password != u.Password || fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) || fmt.Sprintf("%v", user.Settings) != fmt.Sprintf("%v", u.Settings) || fmt.Sprintf("%v", user.Quotas) != fmt.Sprintf("%v", u.Quotas) {
				hash := clickHouseUserHash(spec)
				userWithExistsFields := expandClickHouseUser(spec.(map[string]interface{}), d, hash)
				result = append(result, userWithExistsFields)
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
	res["query_thread_log_retention_time"] = c.EffectiveConfig.QueryThreadLogRetentionTime.Value
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
	if v, ok := d.GetOkExists(rootKey + ".replicated_deduplication_window"); ok {
		config.ReplicatedDeduplicationWindow = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".replicated_deduplication_window_seconds"); ok {
		config.ReplicatedDeduplicationWindowSeconds = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".parts_to_delay_insert"); ok {
		config.PartsToDelayInsert = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".parts_to_throw_insert"); ok {
		config.PartsToThrowInsert = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_replicated_merges_in_queue"); ok {
		config.MaxReplicatedMergesInQueue = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".number_of_free_entries_in_pool_to_lower_max_size_of_merge"); ok {
		config.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_bytes_to_merge_at_min_space_in_pool"); ok {
		config.MaxBytesToMergeAtMinSpaceInPool = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return config, nil
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
	if v, ok := d.GetOkExists(rootKey + ".sasl_username"); ok {
		config.SaslUsername = v.(string)
	}
	if v, ok := d.GetOkExists(rootKey + ".sasl_password"); ok {
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

	if v, ok := d.GetOkExists(rootKey + ".username"); ok {
		config.Username = v.(string)
	}
	if v, ok := d.GetOkExists(rootKey + ".password"); ok {
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
		if v, ok := d.GetOkExists(keyPrefix + ".min_part_size"); ok {
			compression.MinPartSize = int64(v.(int))
		}
		if v, ok := d.GetOkExists(keyPrefix + ".min_part_size_ratio"); ok {
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

			if v, ok := d.GetOkExists(patternKey + ".regexp"); ok {
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
	if v, ok := d.GetOkExists(rootKey + ".max_connections"); ok {
		config.MaxConnections = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_concurrent_queries"); ok {
		config.MaxConcurrentQueries = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".keep_alive_timeout"); ok {
		config.KeepAliveTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".uncompressed_cache_size"); ok {
		config.UncompressedCacheSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".mark_cache_size"); ok {
		config.MarkCacheSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_table_size_to_drop"); ok {
		config.MaxTableSizeToDrop = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_partition_size_to_drop"); ok {
		config.MaxPartitionSizeToDrop = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".timezone"); ok {
		config.Timezone = v.(string)
	}
	if v, ok := d.GetOkExists(rootKey + ".geobase_uri"); ok {
		config.GeobaseUri = v.(string)
	}
	if v, ok := d.GetOkExists(rootKey + ".query_log_retention_size"); ok {
		config.QueryLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".query_log_retention_time"); ok {
		config.QueryLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".query_thread_log_enabled"); ok {
		config.QueryThreadLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".query_thread_log_retention_size"); ok {
		config.QueryThreadLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".query_thread_log_retention_time"); ok {
		config.QueryThreadLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".part_log_retention_size"); ok {
		config.PartLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".part_log_retention_time"); ok {
		config.PartLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".metric_log_enabled"); ok {
		config.MetricLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".metric_log_retention_size"); ok {
		config.MetricLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".metric_log_retention_time"); ok {
		config.MetricLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".trace_log_enabled"); ok {
		config.TraceLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".trace_log_retention_size"); ok {
		config.TraceLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".trace_log_retention_time"); ok {
		config.TraceLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".text_log_enabled"); ok {
		config.TextLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".text_log_retention_size"); ok {
		config.TextLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".text_log_retention_time"); ok {
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

func makeReversedMap(m map[int32]string, addMap map[string]int32) map[string]int32 {
	r := addMap
	for k, v := range m {
		r[v] = k
	}
	return r
}

var (
	UserSettings_OverflowMode_name = map[int32]string{
		0: "unspecified",
		1: "throw",
		2: "break",
	}
	UserSettings_OverflowMode_value       = makeReversedMap(UserSettings_OverflowMode_name, clickhouse.UserSettings_OverflowMode_value)
	UserSettings_GroupByOverflowMode_name = map[int32]string{
		0: "unspecified",
		1: "throw",
		2: "break",
		3: "any",
	}
	UserSettings_GroupByOverflowMode_value   = makeReversedMap(UserSettings_GroupByOverflowMode_name, clickhouse.UserSettings_GroupByOverflowMode_value)
	UserSettings_DistributedProductMode_name = map[int32]string{
		0: "unspecified",
		1: "deny",
		2: "local",
		3: "global",
		4: "allow",
	}
	UserSettings_DistributedProductMode_value     = makeReversedMap(UserSettings_DistributedProductMode_name, clickhouse.UserSettings_DistributedProductMode_value)
	UserSettings_CountDistinctImplementation_name = map[int32]string{
		0: "unspecified",
		1: "uniq",
		2: "uniq_combined",
		3: "uniq_combined_64",
		4: "uniq_hll_12",
		5: "uniq_exact",
	}
	UserSettings_CountDistinctImplementation_value = makeReversedMap(UserSettings_CountDistinctImplementation_name, clickhouse.UserSettings_CountDistinctImplementation_value)
	UserSettings_QuotaMode_name                    = map[int32]string{
		0: "unspecified",
		1: "default",
		2: "keyed",
		3: "keyed_by_ip",
	}
	UserSettings_QuotaMode_value = makeReversedMap(UserSettings_QuotaMode_name, clickhouse.UserSettings_QuotaMode_value)
)

func getOverflowModeName(value clickhouse.UserSettings_OverflowMode) string {
	if name, ok := UserSettings_OverflowMode_name[int32(value)]; ok {
		return name
	}
	return UserSettings_OverflowMode_name[0]
}

func getOverflowModeValue(name string) clickhouse.UserSettings_OverflowMode {
	if value, ok := UserSettings_OverflowMode_value[name]; ok {
		return clickhouse.UserSettings_OverflowMode(value)
	}
	return 0
}

func getGroupByOverflowModeName(value clickhouse.UserSettings_GroupByOverflowMode) string {
	if name, ok := UserSettings_GroupByOverflowMode_name[int32(value)]; ok {
		return name
	}
	return UserSettings_GroupByOverflowMode_name[0]
}

func getGroupByOverflowModeValue(name string) clickhouse.UserSettings_GroupByOverflowMode {
	if value, ok := UserSettings_GroupByOverflowMode_value[name]; ok {
		return clickhouse.UserSettings_GroupByOverflowMode(value)
	}
	return 0
}

func getDistributedProductModeName(value clickhouse.UserSettings_DistributedProductMode) string {
	if name, ok := UserSettings_DistributedProductMode_name[int32(value)]; ok {
		return name
	}
	return UserSettings_DistributedProductMode_name[0]
}

func getDistributedProductModeValue(name string) clickhouse.UserSettings_DistributedProductMode {
	if value, ok := UserSettings_DistributedProductMode_value[name]; ok {
		return clickhouse.UserSettings_DistributedProductMode(value)
	}
	return 0
}

func getCountDistinctImplementationName(value clickhouse.UserSettings_CountDistinctImplementation) string {
	if name, ok := UserSettings_CountDistinctImplementation_name[int32(value)]; ok {
		return name
	}
	return UserSettings_CountDistinctImplementation_name[0]
}

func getCountDistinctImplementationValue(name string) clickhouse.UserSettings_CountDistinctImplementation {
	if value, ok := UserSettings_CountDistinctImplementation_value[name]; ok {
		return clickhouse.UserSettings_CountDistinctImplementation(value)
	}
	return 0
}

func getQuotaModeName(value clickhouse.UserSettings_QuotaMode) string {
	if name, ok := UserSettings_QuotaMode_name[int32(value)]; ok {
		return name
	}
	return UserSettings_QuotaMode_name[0]
}

func getQuotaModeValue(name string) clickhouse.UserSettings_QuotaMode {
	if value, ok := UserSettings_QuotaMode_value[name]; ok {
		return clickhouse.UserSettings_QuotaMode(value)
	}
	return 0
}

func setSettingFromMapInt64(us map[string]interface{}, key string, setting **wrappers.Int64Value) {
	if v, ok := us[key]; ok {
		switch vt := v.(type) {
		case int:
			if vt > 0 {
				*setting = &wrappers.Int64Value{Value: int64(vt)}
			}
		case int64:
			if vt > 0 {
				*setting = &wrappers.Int64Value{Value: vt}
			}
		}
	}
}

func setSettingFromDataInt64(d *schema.ResourceData, fullKey string, setting **wrappers.Int64Value) {
	if v, ok := d.GetOk(fullKey); ok {
		if v.(int) > 0 {
			*setting = &wrappers.Int64Value{Value: int64(v.(int))}
		}
	}
}

func setSettingFromMapBool(us map[string]interface{}, key string, setting **wrappers.BoolValue) {
	if v, ok := us[key]; ok {
		*setting = &wrappers.BoolValue{Value: v.(bool)}
	}
}

func setSettingFromDataBool(d *schema.ResourceData, fullKey string, setting **wrappers.BoolValue) {
	if v, ok := d.GetOk(fullKey); ok {
		*setting = &wrappers.BoolValue{Value: v.(bool)}
	}
}

func expandClickHouseUserSettings(us map[string]interface{}) *clickhouse.UserSettings {
	result := &clickhouse.UserSettings{}

	setSettingFromMapInt64(us, "readonly", &result.Readonly)
	setSettingFromMapBool(us, "allow_ddl", &result.AllowDdl)
	setSettingFromMapInt64(us, "insert_quorum", &result.InsertQuorum)
	setSettingFromMapInt64(us, "connect_timeout", &result.ConnectTimeout)
	setSettingFromMapInt64(us, "receive_timeout", &result.ReceiveTimeout)
	setSettingFromMapInt64(us, "send_timeout", &result.SendTimeout)
	setSettingFromMapInt64(us, "insert_quorum_timeout", &result.InsertQuorumTimeout)
	setSettingFromMapBool(us, "select_sequential_consistency", &result.SelectSequentialConsistency)
	setSettingFromMapInt64(us, "max_replica_delay_for_distributed_queries", &result.MaxReplicaDelayForDistributedQueries)
	setSettingFromMapBool(us, "fallback_to_stale_replicas_for_distributed_queries", &result.FallbackToStaleReplicasForDistributedQueries)
	setSettingFromMapInt64(us, "replication_alter_partitions_sync", &result.ReplicationAlterPartitionsSync)

	if v, ok := us["distributed_product_mode"]; ok {
		result.DistributedProductMode = getDistributedProductModeValue(v.(string))
	}

	setSettingFromMapBool(us, "distributed_aggregation_memory_efficient", &result.DistributedAggregationMemoryEfficient)
	setSettingFromMapInt64(us, "distributed_ddl_task_timeout", &result.DistributedDdlTaskTimeout)
	setSettingFromMapBool(us, "skip_unavailable_shards", &result.SkipUnavailableShards)
	setSettingFromMapBool(us, "compile", &result.Compile)
	setSettingFromMapInt64(us, "min_count_to_compile", &result.MinCountToCompile)
	setSettingFromMapBool(us, "compile_expressions", &result.CompileExpressions)
	setSettingFromMapInt64(us, "min_count_to_compile_expression", &result.MinCountToCompileExpression)
	setSettingFromMapInt64(us, "max_block_size", &result.MaxBlockSize)
	setSettingFromMapInt64(us, "min_insert_block_size_rows", &result.MinInsertBlockSizeRows)
	setSettingFromMapInt64(us, "min_insert_block_size_bytes", &result.MinInsertBlockSizeBytes)
	setSettingFromMapInt64(us, "max_insert_block_size", &result.MaxInsertBlockSize)
	setSettingFromMapInt64(us, "min_bytes_to_use_direct_io", &result.MinBytesToUseDirectIo)
	setSettingFromMapBool(us, "use_uncompressed_cache", &result.UseUncompressedCache)
	setSettingFromMapInt64(us, "merge_tree_max_rows_to_use_cache", &result.MergeTreeMaxRowsToUseCache)
	setSettingFromMapInt64(us, "merge_tree_max_bytes_to_use_cache", &result.MergeTreeMaxBytesToUseCache)
	setSettingFromMapInt64(us, "merge_tree_min_rows_for_concurrent_read", &result.MergeTreeMinRowsForConcurrentRead)
	setSettingFromMapInt64(us, "merge_tree_min_bytes_for_concurrent_read", &result.MergeTreeMinBytesForConcurrentRead)
	setSettingFromMapInt64(us, "max_bytes_before_external_group_by", &result.MaxBytesBeforeExternalGroupBy)
	setSettingFromMapInt64(us, "max_bytes_before_external_sort", &result.MaxBytesBeforeExternalSort)
	setSettingFromMapInt64(us, "group_by_two_level_threshold", &result.GroupByTwoLevelThreshold)
	setSettingFromMapInt64(us, "group_by_two_level_threshold_bytes", &result.GroupByTwoLevelThresholdBytes)
	setSettingFromMapInt64(us, "priority", &result.Priority)
	setSettingFromMapInt64(us, "max_threads", &result.MaxThreads)
	setSettingFromMapInt64(us, "max_memory_usage", &result.MaxMemoryUsage)
	setSettingFromMapInt64(us, "max_memory_usage_for_user", &result.MaxMemoryUsageForUser)
	setSettingFromMapInt64(us, "max_network_bandwidth", &result.MaxNetworkBandwidth)
	setSettingFromMapInt64(us, "max_network_bandwidth_for_user", &result.MaxNetworkBandwidthForUser)
	setSettingFromMapBool(us, "force_index_by_date", &result.ForceIndexByDate)
	setSettingFromMapBool(us, "force_primary_key", &result.ForcePrimaryKey)
	setSettingFromMapInt64(us, "max_rows_to_read", &result.MaxRowsToRead)
	setSettingFromMapInt64(us, "max_bytes_to_read", &result.MaxBytesToRead)

	if v, ok := us["read_overflow_mode"]; ok {
		result.ReadOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_rows_to_group_by", &result.MaxRowsToGroupBy)

	if v, ok := us["group_by_overflow_mode"]; ok {
		result.GroupByOverflowMode = getGroupByOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_rows_to_sort", &result.MaxRowsToSort)
	setSettingFromMapInt64(us, "max_bytes_to_sort", &result.MaxBytesToSort)

	if v, ok := us["sort_overflow_mode"]; ok {
		result.SortOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_result_rows", &result.MaxResultRows)
	setSettingFromMapInt64(us, "max_result_bytes", &result.MaxResultBytes)

	if v, ok := us["result_overflow_mode"]; ok {
		result.ResultOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_rows_in_distinct", &result.MaxRowsInDistinct)
	setSettingFromMapInt64(us, "max_bytes_in_distinct", &result.MaxBytesInDistinct)

	if v, ok := us["distinct_overflow_mode"]; ok {
		result.DistinctOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_rows_to_transfer", &result.MaxRowsToTransfer)
	setSettingFromMapInt64(us, "max_bytes_to_transfer", &result.MaxBytesToTransfer)

	if v, ok := us["transfer_overflow_mode"]; ok {
		result.TransferOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_execution_time", &result.MaxExecutionTime)

	if v, ok := us["timeout_overflow_mode"]; ok {
		result.TimeoutOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_rows_in_set", &result.MaxRowsInSet)
	setSettingFromMapInt64(us, "max_bytes_in_set", &result.MaxBytesInSet)

	if v, ok := us["set_overflow_mode"]; ok {
		result.SetOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_rows_in_join", &result.MaxRowsInJoin)
	setSettingFromMapInt64(us, "max_bytes_in_join", &result.MaxBytesInJoin)

	if v, ok := us["join_overflow_mode"]; ok {
		result.JoinOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromMapInt64(us, "max_columns_to_read", &result.MaxColumnsToRead)
	setSettingFromMapInt64(us, "max_temporary_columns", &result.MaxTemporaryColumns)
	setSettingFromMapInt64(us, "max_temporary_non_const_columns", &result.MaxTemporaryNonConstColumns)
	setSettingFromMapInt64(us, "max_query_size", &result.MaxQuerySize)
	setSettingFromMapInt64(us, "max_ast_depth", &result.MaxAstDepth)
	setSettingFromMapInt64(us, "max_ast_elements", &result.MaxAstElements)
	setSettingFromMapInt64(us, "max_expanded_ast_elements", &result.MaxExpandedAstElements)
	setSettingFromMapInt64(us, "min_execution_speed", &result.MinExecutionSpeed)
	setSettingFromMapInt64(us, "min_execution_speed_bytes", &result.MinExecutionSpeedBytes)

	if v, ok := us["count_distinct_implementation"]; ok {
		result.CountDistinctImplementation = getCountDistinctImplementationValue(v.(string))
	}

	setSettingFromMapBool(us, "input_format_values_interpret_expressions", &result.InputFormatValuesInterpretExpressions)
	setSettingFromMapBool(us, "input_format_defaults_for_omitted_fields", &result.InputFormatDefaultsForOmittedFields)
	setSettingFromMapBool(us, "output_format_json_quote_64bit_integers", &result.OutputFormatJsonQuote_64BitIntegers)
	setSettingFromMapBool(us, "output_format_json_quote_denormals", &result.OutputFormatJsonQuoteDenormals)
	setSettingFromMapBool(us, "low_cardinality_allow_in_native_format", &result.LowCardinalityAllowInNativeFormat)
	setSettingFromMapBool(us, "empty_result_for_aggregation_by_empty_set", &result.EmptyResultForAggregationByEmptySet)
	setSettingFromMapBool(us, "joined_subquery_requires_alias", &result.JoinedSubqueryRequiresAlias)
	setSettingFromMapBool(us, "join_use_nulls", &result.JoinUseNulls)
	setSettingFromMapBool(us, "transform_null_in", &result.TransformNullIn)

	setSettingFromMapInt64(us, "http_connection_timeout", &result.HttpConnectionTimeout)
	setSettingFromMapInt64(us, "http_receive_timeout", &result.HttpReceiveTimeout)
	setSettingFromMapInt64(us, "http_send_timeout", &result.HttpSendTimeout)
	setSettingFromMapBool(us, "enable_http_compression", &result.EnableHttpCompression)
	setSettingFromMapBool(us, "send_progress_in_http_headers", &result.SendProgressInHttpHeaders)
	setSettingFromMapInt64(us, "http_headers_progress_interval", &result.HttpHeadersProgressInterval)
	setSettingFromMapBool(us, "add_http_cors_header", &result.AddHttpCorsHeader)

	if v, ok := us["quota_mode"]; ok {
		result.QuotaMode = getQuotaModeValue(v.(string))
	}

	return result
}

func expandClickHouseUserSettingsExists(d *schema.ResourceData, hash int) *clickhouse.UserSettings {
	result := &clickhouse.UserSettings{}

	rootKey := fmt.Sprintf("user.%d.settings.0", hash)

	setSettingFromDataInt64(d, rootKey+".readonly", &result.Readonly)
	setSettingFromDataBool(d, rootKey+".allow_ddl", &result.AllowDdl)
	setSettingFromDataInt64(d, rootKey+".insert_quorum", &result.InsertQuorum)
	setSettingFromDataInt64(d, rootKey+".connect_timeout", &result.ConnectTimeout)
	setSettingFromDataInt64(d, rootKey+".receive_timeout", &result.ReceiveTimeout)
	setSettingFromDataInt64(d, rootKey+".send_timeout", &result.SendTimeout)
	setSettingFromDataInt64(d, rootKey+".insert_quorum_timeout", &result.InsertQuorumTimeout)
	setSettingFromDataBool(d, rootKey+".select_sequential_consistency", &result.SelectSequentialConsistency)
	setSettingFromDataInt64(d, rootKey+".max_replica_delay_for_distributed_queries", &result.MaxReplicaDelayForDistributedQueries)
	setSettingFromDataBool(d, rootKey+".fallback_to_stale_replicas_for_distributed_queries", &result.FallbackToStaleReplicasForDistributedQueries)
	setSettingFromDataInt64(d, rootKey+".replication_alter_partitions_sync", &result.ReplicationAlterPartitionsSync)

	if v, ok := d.GetOk(rootKey + ".distributed_product_mode"); ok {
		result.DistributedProductMode = getDistributedProductModeValue(v.(string))
	}

	setSettingFromDataBool(d, rootKey+".distributed_aggregation_memory_efficient", &result.DistributedAggregationMemoryEfficient)
	setSettingFromDataInt64(d, rootKey+".distributed_ddl_task_timeout", &result.DistributedDdlTaskTimeout)
	setSettingFromDataBool(d, rootKey+".skip_unavailable_shards", &result.SkipUnavailableShards)
	setSettingFromDataBool(d, rootKey+".compile", &result.Compile)
	setSettingFromDataInt64(d, rootKey+".min_count_to_compile", &result.MinCountToCompile)
	setSettingFromDataBool(d, rootKey+".compile_expressions", &result.CompileExpressions)
	setSettingFromDataInt64(d, rootKey+".min_count_to_compile_expression", &result.MinCountToCompileExpression)
	setSettingFromDataInt64(d, rootKey+".max_block_size", &result.MaxBlockSize)
	setSettingFromDataInt64(d, rootKey+".min_insert_block_size_rows", &result.MinInsertBlockSizeRows)
	setSettingFromDataInt64(d, rootKey+".min_insert_block_size_bytes", &result.MinInsertBlockSizeBytes)
	setSettingFromDataInt64(d, rootKey+".max_insert_block_size", &result.MaxInsertBlockSize)
	setSettingFromDataInt64(d, rootKey+".min_bytes_to_use_direct_io", &result.MinBytesToUseDirectIo)
	setSettingFromDataBool(d, rootKey+".use_uncompressed_cache", &result.UseUncompressedCache)
	setSettingFromDataInt64(d, rootKey+".merge_tree_max_rows_to_use_cache", &result.MergeTreeMaxRowsToUseCache)
	setSettingFromDataInt64(d, rootKey+".merge_tree_max_bytes_to_use_cache", &result.MergeTreeMaxBytesToUseCache)
	setSettingFromDataInt64(d, rootKey+".merge_tree_min_rows_for_concurrent_read", &result.MergeTreeMinRowsForConcurrentRead)
	setSettingFromDataInt64(d, rootKey+".merge_tree_min_bytes_for_concurrent_read", &result.MergeTreeMinBytesForConcurrentRead)
	setSettingFromDataInt64(d, rootKey+".max_bytes_before_external_group_by", &result.MaxBytesBeforeExternalGroupBy)
	setSettingFromDataInt64(d, rootKey+".max_bytes_before_external_sort", &result.MaxBytesBeforeExternalSort)
	setSettingFromDataInt64(d, rootKey+".group_by_two_level_threshold", &result.GroupByTwoLevelThreshold)
	setSettingFromDataInt64(d, rootKey+".group_by_two_level_threshold_bytes", &result.GroupByTwoLevelThresholdBytes)
	setSettingFromDataInt64(d, rootKey+".priority", &result.Priority)
	setSettingFromDataInt64(d, rootKey+".max_threads", &result.MaxThreads)
	setSettingFromDataInt64(d, rootKey+".max_memory_usage", &result.MaxMemoryUsage)
	setSettingFromDataInt64(d, rootKey+".max_memory_usage_for_user", &result.MaxMemoryUsageForUser)
	setSettingFromDataInt64(d, rootKey+".max_network_bandwidth", &result.MaxNetworkBandwidth)
	setSettingFromDataInt64(d, rootKey+".max_network_bandwidth_for_user", &result.MaxNetworkBandwidthForUser)
	setSettingFromDataBool(d, rootKey+".force_index_by_date", &result.ForceIndexByDate)
	setSettingFromDataBool(d, rootKey+".force_primary_key", &result.ForcePrimaryKey)
	setSettingFromDataInt64(d, rootKey+".max_rows_to_read", &result.MaxRowsToRead)
	setSettingFromDataInt64(d, rootKey+".max_bytes_to_read", &result.MaxBytesToRead)

	if v, ok := d.GetOk(rootKey + ".read_overflow_mode"); ok {
		result.ReadOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_rows_to_group_by", &result.MaxRowsToGroupBy)

	if v, ok := d.GetOk(rootKey + ".group_by_overflow_mode"); ok {
		result.GroupByOverflowMode = getGroupByOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_rows_to_sort", &result.MaxRowsToSort)
	setSettingFromDataInt64(d, rootKey+".max_bytes_to_sort", &result.MaxBytesToSort)

	if v, ok := d.GetOk(rootKey + ".sort_overflow_mode"); ok {
		result.SortOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_result_rows", &result.MaxResultRows)
	setSettingFromDataInt64(d, rootKey+".max_result_bytes", &result.MaxResultBytes)

	if v, ok := d.GetOk(rootKey + ".result_overflow_mode"); ok {
		result.ResultOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_rows_in_distinct", &result.MaxRowsInDistinct)
	setSettingFromDataInt64(d, rootKey+".max_bytes_in_distinct", &result.MaxBytesInDistinct)

	if v, ok := d.GetOk(rootKey + ".distinct_overflow_mode"); ok {
		result.DistinctOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_rows_to_transfer", &result.MaxRowsToTransfer)
	setSettingFromDataInt64(d, rootKey+".max_bytes_to_transfer", &result.MaxBytesToTransfer)

	if v, ok := d.GetOk(rootKey + ".transfer_overflow_mode"); ok {
		result.TransferOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_execution_time", &result.MaxExecutionTime)

	if v, ok := d.GetOk(rootKey + ".timeout_overflow_mode"); ok {
		result.TimeoutOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_rows_in_set", &result.MaxRowsInSet)
	setSettingFromDataInt64(d, rootKey+".max_bytes_in_set", &result.MaxBytesInSet)

	if v, ok := d.GetOk(rootKey + ".set_overflow_mode"); ok {
		result.SetOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_rows_in_join", &result.MaxRowsInJoin)
	setSettingFromDataInt64(d, rootKey+".max_bytes_in_join", &result.MaxBytesInJoin)

	if v, ok := d.GetOk(rootKey + ".join_overflow_mode"); ok {
		result.JoinOverflowMode = getOverflowModeValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".max_columns_to_read", &result.MaxColumnsToRead)
	setSettingFromDataInt64(d, rootKey+".max_temporary_columns", &result.MaxTemporaryColumns)
	setSettingFromDataInt64(d, rootKey+".max_temporary_non_const_columns", &result.MaxTemporaryNonConstColumns)
	setSettingFromDataInt64(d, rootKey+".max_query_size", &result.MaxQuerySize)
	setSettingFromDataInt64(d, rootKey+".max_ast_depth", &result.MaxAstDepth)
	setSettingFromDataInt64(d, rootKey+".max_ast_elements", &result.MaxAstElements)
	setSettingFromDataInt64(d, rootKey+".max_expanded_ast_elements", &result.MaxExpandedAstElements)
	setSettingFromDataInt64(d, rootKey+".min_execution_speed", &result.MinExecutionSpeed)
	setSettingFromDataInt64(d, rootKey+".min_execution_speed_bytes", &result.MinExecutionSpeedBytes)

	if v, ok := d.GetOk(rootKey + ".count_distinct_implementation"); ok {
		result.CountDistinctImplementation = getCountDistinctImplementationValue(v.(string))
	}

	setSettingFromDataBool(d, rootKey+".input_format_values_interpret_expressions", &result.InputFormatValuesInterpretExpressions)
	setSettingFromDataBool(d, rootKey+".input_format_defaults_for_omitted_fields", &result.InputFormatDefaultsForOmittedFields)
	setSettingFromDataBool(d, rootKey+".output_format_json_quote_64bit_integers", &result.OutputFormatJsonQuote_64BitIntegers)
	setSettingFromDataBool(d, rootKey+".output_format_json_quote_denormals", &result.OutputFormatJsonQuoteDenormals)
	setSettingFromDataBool(d, rootKey+".low_cardinality_allow_in_native_format", &result.LowCardinalityAllowInNativeFormat)
	setSettingFromDataBool(d, rootKey+".empty_result_for_aggregation_by_empty_set", &result.EmptyResultForAggregationByEmptySet)
	setSettingFromDataBool(d, rootKey+".joined_subquery_requires_alias", &result.JoinedSubqueryRequiresAlias)
	setSettingFromDataBool(d, rootKey+".join_use_nulls", &result.JoinUseNulls)
	setSettingFromDataBool(d, rootKey+".transform_null_in", &result.TransformNullIn)

	setSettingFromDataInt64(d, rootKey+".http_connection_timeout", &result.HttpConnectionTimeout)
	setSettingFromDataInt64(d, rootKey+".http_receive_timeout", &result.HttpReceiveTimeout)
	setSettingFromDataInt64(d, rootKey+".http_send_timeout", &result.HttpSendTimeout)
	setSettingFromDataBool(d, rootKey+".enable_http_compression", &result.EnableHttpCompression)
	setSettingFromDataBool(d, rootKey+".send_progress_in_http_headers", &result.SendProgressInHttpHeaders)
	setSettingFromDataInt64(d, rootKey+".http_headers_progress_interval", &result.HttpHeadersProgressInterval)
	setSettingFromDataBool(d, rootKey+".add_http_cors_header", &result.AddHttpCorsHeader)

	if v, ok := d.GetOk(rootKey + ".quota_mode"); ok {
		result.QuotaMode = getQuotaModeValue(v.(string))
	}

	return result
}

func flattenClickHouseUserQuota(quota *clickhouse.UserQuota) map[string]interface{} {
	p := map[string]interface{}{}
	if quota.IntervalDuration != nil {
		p["interval_duration"] = quota.IntervalDuration.Value
	}
	if quota.Queries != nil {
		p["queries"] = quota.Queries.Value
	}
	if quota.Errors != nil {
		p["errors"] = quota.Errors.Value
	}
	if quota.ResultRows != nil {
		p["result_rows"] = quota.ResultRows.Value
	}
	if quota.ReadRows != nil {
		p["read_rows"] = quota.ReadRows.Value
	}
	if quota.ExecutionTime != nil {
		p["execution_time"] = quota.ExecutionTime.Value
	}
	return p
}

func falseOnNil(param *wrappers.BoolValue) bool {
	if param != nil {
		return param.Value
	}
	return false
}

func flattenClickHouseUserSettings(settings *clickhouse.UserSettings) map[string]interface{} {
	result := map[string]interface{}{}

	if settings.Readonly != nil {
		result["readonly"] = settings.Readonly.Value
	}
	result["allow_ddl"] = falseOnNil(settings.AllowDdl)
	if settings.InsertQuorum != nil {
		result["insert_quorum"] = settings.InsertQuorum.Value
	}
	if settings.ConnectTimeout != nil {
		result["connect_timeout"] = settings.ConnectTimeout.Value
	}
	if settings.ReceiveTimeout != nil {
		result["receive_timeout"] = settings.ReceiveTimeout.Value
	}
	if settings.SendTimeout != nil {
		result["send_timeout"] = settings.SendTimeout.Value
	}
	if settings.InsertQuorumTimeout != nil {
		result["insert_quorum_timeout"] = settings.InsertQuorumTimeout.Value
	}
	result["select_sequential_consistency"] = falseOnNil(settings.SelectSequentialConsistency)
	if settings.MaxReplicaDelayForDistributedQueries != nil {
		result["max_replica_delay_for_distributed_queries"] = settings.MaxReplicaDelayForDistributedQueries.Value
	}
	result["fallback_to_stale_replicas_for_distributed_queries"] = falseOnNil(settings.FallbackToStaleReplicasForDistributedQueries)
	if settings.ReplicationAlterPartitionsSync != nil {
		result["replication_alter_partitions_sync"] = settings.ReplicationAlterPartitionsSync.Value
	}
	result["distributed_product_mode"] = getDistributedProductModeName(settings.DistributedProductMode)
	result["distributed_aggregation_memory_efficient"] = falseOnNil(settings.DistributedAggregationMemoryEfficient)
	if settings.DistributedDdlTaskTimeout != nil {
		result["distributed_ddl_task_timeout"] = settings.DistributedDdlTaskTimeout.Value
	}
	result["skip_unavailable_shards"] = falseOnNil(settings.SkipUnavailableShards)
	result["compile"] = falseOnNil(settings.Compile)
	if settings.MinCountToCompile != nil {
		result["min_count_to_compile"] = settings.MinCountToCompile.Value
	}
	result["compile_expressions"] = falseOnNil(settings.CompileExpressions)
	if settings.MinCountToCompileExpression != nil {
		result["min_count_to_compile_expression"] = settings.MinCountToCompileExpression.Value
	}
	if settings.MaxBlockSize != nil {
		result["max_block_size"] = settings.MaxBlockSize.Value
	}
	if settings.MinInsertBlockSizeRows != nil {
		result["min_insert_block_size_rows"] = settings.MinInsertBlockSizeRows.Value
	}
	if settings.MinInsertBlockSizeBytes != nil {
		result["min_insert_block_size_bytes"] = settings.MinInsertBlockSizeBytes.Value
	}
	if settings.MaxInsertBlockSize != nil {
		result["max_insert_block_size"] = settings.MaxInsertBlockSize.Value
	}
	if settings.MinBytesToUseDirectIo != nil {
		result["min_bytes_to_use_direct_io"] = settings.MinBytesToUseDirectIo.Value
	}
	result["use_uncompressed_cache"] = falseOnNil(settings.UseUncompressedCache)
	if settings.MergeTreeMaxRowsToUseCache != nil {
		result["merge_tree_max_rows_to_use_cache"] = settings.MergeTreeMaxRowsToUseCache.Value
	}
	if settings.MergeTreeMaxBytesToUseCache != nil {
		result["merge_tree_max_bytes_to_use_cache"] = settings.MergeTreeMaxBytesToUseCache.Value
	}
	if settings.MergeTreeMinRowsForConcurrentRead != nil {
		result["merge_tree_min_rows_for_concurrent_read"] = settings.MergeTreeMinRowsForConcurrentRead.Value
	}
	if settings.MergeTreeMinBytesForConcurrentRead != nil {
		result["merge_tree_min_bytes_for_concurrent_read"] = settings.MergeTreeMinBytesForConcurrentRead.Value
	}
	if settings.MaxBytesBeforeExternalGroupBy != nil {
		result["max_bytes_before_external_group_by"] = settings.MaxBytesBeforeExternalGroupBy.Value
	}
	if settings.MaxBytesBeforeExternalSort != nil {
		result["max_bytes_before_external_sort"] = settings.MaxBytesBeforeExternalSort.Value
	}
	if settings.GroupByTwoLevelThreshold != nil {
		result["group_by_two_level_threshold"] = settings.GroupByTwoLevelThreshold.Value
	}
	if settings.GroupByTwoLevelThresholdBytes != nil {
		result["group_by_two_level_threshold_bytes"] = settings.GroupByTwoLevelThresholdBytes.Value
	}
	if settings.Priority != nil {
		result["priority"] = settings.Priority.Value
	}
	if settings.MaxThreads != nil {
		result["max_threads"] = settings.MaxThreads.Value
	}
	if settings.MaxMemoryUsage != nil {
		result["max_memory_usage"] = settings.MaxMemoryUsage.Value
	}
	if settings.MaxMemoryUsageForUser != nil {
		result["max_memory_usage_for_user"] = settings.MaxMemoryUsageForUser.Value
	}
	if settings.MaxNetworkBandwidth != nil {
		result["max_network_bandwidth"] = settings.MaxNetworkBandwidth.Value
	}
	if settings.MaxNetworkBandwidthForUser != nil {
		result["max_network_bandwidth_for_user"] = settings.MaxNetworkBandwidthForUser.Value
	}
	result["force_index_by_date"] = falseOnNil(settings.ForceIndexByDate)
	result["force_primary_key"] = falseOnNil(settings.ForcePrimaryKey)
	if settings.MaxRowsToRead != nil {
		result["max_rows_to_read"] = settings.MaxRowsToRead.Value
	}
	if settings.MaxBytesToRead != nil {
		result["max_bytes_to_read"] = settings.MaxBytesToRead.Value
	}
	result["read_overflow_mode"] = getOverflowModeName(settings.ReadOverflowMode)
	if settings.MaxRowsToGroupBy != nil {
		result["max_rows_to_group_by"] = settings.MaxRowsToGroupBy.Value
	}
	result["group_by_overflow_mode"] = getGroupByOverflowModeName(settings.GroupByOverflowMode)
	if settings.MaxRowsToSort != nil {
		result["max_rows_to_sort"] = settings.MaxRowsToSort.Value
	}
	if settings.MaxBytesToSort != nil {
		result["max_bytes_to_sort"] = settings.MaxBytesToSort.Value
	}
	result["sort_overflow_mode"] = getOverflowModeName(settings.SortOverflowMode)
	if settings.MaxResultRows != nil {
		result["max_result_rows"] = settings.MaxResultRows.Value
	}
	if settings.MaxResultBytes != nil {
		result["max_result_bytes"] = settings.MaxResultBytes.Value
	}
	result["result_overflow_mode"] = getOverflowModeName(settings.ResultOverflowMode)
	if settings.MaxRowsInDistinct != nil {
		result["max_rows_in_distinct"] = settings.MaxRowsInDistinct.Value
	}
	if settings.MaxBytesInDistinct != nil {
		result["max_bytes_in_distinct"] = settings.MaxBytesInDistinct.Value
	}
	result["distinct_overflow_mode"] = getOverflowModeName(settings.DistinctOverflowMode)
	if settings.MaxRowsToTransfer != nil {
		result["max_rows_to_transfer"] = settings.MaxRowsToTransfer.Value
	}
	if settings.MaxBytesToTransfer != nil {
		result["max_bytes_to_transfer"] = settings.MaxBytesToTransfer.Value
	}
	result["transfer_overflow_mode"] = getOverflowModeName(settings.TransferOverflowMode)
	if settings.MaxExecutionTime != nil {
		result["max_execution_time"] = settings.MaxExecutionTime.Value
	}
	result["timeout_overflow_mode"] = getOverflowModeName(settings.TimeoutOverflowMode)
	if settings.MaxRowsInSet != nil {
		result["max_rows_in_set"] = settings.MaxRowsInSet.Value
	}
	if settings.MaxBytesInSet != nil {
		result["max_bytes_in_set"] = settings.MaxBytesInSet.Value
	}
	result["set_overflow_mode"] = getOverflowModeName(settings.SetOverflowMode)
	if settings.MaxRowsInJoin != nil {
		result["max_rows_in_join"] = settings.MaxRowsInJoin.Value
	}
	if settings.MaxBytesInJoin != nil {
		result["max_bytes_in_join"] = settings.MaxBytesInJoin.Value
	}
	result["join_overflow_mode"] = getOverflowModeName(settings.JoinOverflowMode)
	if settings.MaxColumnsToRead != nil {
		result["max_columns_to_read"] = settings.MaxColumnsToRead.Value
	}
	if settings.MaxTemporaryColumns != nil {
		result["max_temporary_columns"] = settings.MaxTemporaryColumns.Value
	}
	if settings.MaxTemporaryNonConstColumns != nil {
		result["max_temporary_non_const_columns"] = settings.MaxTemporaryNonConstColumns.Value
	}
	if settings.MaxQuerySize != nil {
		result["max_query_size"] = settings.MaxQuerySize.Value
	}
	if settings.MaxAstDepth != nil {
		result["max_ast_depth"] = settings.MaxAstDepth.Value
	}
	if settings.MaxAstElements != nil {
		result["max_ast_elements"] = settings.MaxAstElements.Value
	}
	if settings.MaxExpandedAstElements != nil {
		result["max_expanded_ast_elements"] = settings.MaxExpandedAstElements.Value
	}
	if settings.MinExecutionSpeed != nil {
		result["min_execution_speed"] = settings.MinExecutionSpeed.Value
	}
	if settings.MinExecutionSpeedBytes != nil {
		result["min_execution_speed_bytes"] = settings.MinExecutionSpeedBytes.Value
	}
	result["count_distinct_implementation"] = getCountDistinctImplementationName(settings.CountDistinctImplementation)
	result["input_format_values_interpret_expressions"] = falseOnNil(settings.InputFormatValuesInterpretExpressions)
	result["input_format_defaults_for_omitted_fields"] = falseOnNil(settings.InputFormatDefaultsForOmittedFields)
	result["output_format_json_quote_64bit_integers"] = falseOnNil(settings.OutputFormatJsonQuote_64BitIntegers)
	result["output_format_json_quote_denormals"] = falseOnNil(settings.OutputFormatJsonQuoteDenormals)
	result["low_cardinality_allow_in_native_format"] = falseOnNil(settings.LowCardinalityAllowInNativeFormat)
	result["empty_result_for_aggregation_by_empty_set"] = falseOnNil(settings.EmptyResultForAggregationByEmptySet)
	result["joined_subquery_requires_alias"] = falseOnNil(settings.JoinedSubqueryRequiresAlias)
	result["join_use_nulls"] = falseOnNil(settings.JoinUseNulls)
	result["transform_null_in"] = falseOnNil(settings.TransformNullIn)
	if settings.HttpConnectionTimeout != nil {
		result["http_connection_timeout"] = settings.HttpConnectionTimeout.Value
	}
	if settings.HttpReceiveTimeout != nil {
		result["http_receive_timeout"] = settings.HttpReceiveTimeout.Value
	}
	if settings.HttpSendTimeout != nil {
		result["http_send_timeout"] = settings.HttpSendTimeout.Value
	}
	result["enable_http_compression"] = falseOnNil(settings.EnableHttpCompression)
	result["send_progress_in_http_headers"] = falseOnNil(settings.SendProgressInHttpHeaders)
	if settings.HttpHeadersProgressInterval != nil {
		result["http_headers_progress_interval"] = settings.HttpHeadersProgressInterval.Value
	}
	result["add_http_cors_header"] = falseOnNil(settings.AddHttpCorsHeader)
	result["quota_mode"] = getQuotaModeName(settings.QuotaMode)

	return result
}

func expandClickHouseUserQuotas(ps *schema.Set) []*clickhouse.UserQuota {
	result := []*clickhouse.UserQuota{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		quota := &clickhouse.UserQuota{}

		setSettingFromMapInt64(m, "interval_duration", &quota.IntervalDuration)
		setSettingFromMapInt64(m, "queries", &quota.Queries)
		setSettingFromMapInt64(m, "errors", &quota.Errors)
		setSettingFromMapInt64(m, "result_rows", &quota.ResultRows)
		setSettingFromMapInt64(m, "ead_rows", &quota.ReadRows)
		setSettingFromMapInt64(m, "execution_time", &quota.ExecutionTime)

		result = append(result, quota)
	}
	return result
}

func expandClickHouseUserQuotasExists(d *schema.ResourceData, hash int) []*clickhouse.UserQuota {
	result := []*clickhouse.UserQuota{}

	rootKey := fmt.Sprintf("user.%d.quota", hash)

	quotas := d.Get(rootKey).(*schema.Set)

	for _, q := range quotas.List() {
		quotaHash := clickHouseUserQuotaHash(q)
		quota := &clickhouse.UserQuota{}
		quotaKey := fmt.Sprintf("user.%d.quota.%d", hash, quotaHash)

		setSettingFromDataInt64(d, quotaKey+".interval_duration", &quota.IntervalDuration)
		setSettingFromDataInt64(d, quotaKey+".queries", &quota.Queries)
		setSettingFromDataInt64(d, quotaKey+".errors", &quota.Errors)
		setSettingFromDataInt64(d, quotaKey+".result_rows", &quota.ResultRows)
		setSettingFromDataInt64(d, quotaKey+".read_rows", &quota.ReadRows)
		setSettingFromDataInt64(d, quotaKey+".execution_time", &quota.ExecutionTime)

		result = append(result, quota)
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

		u["settings"] = []interface{}{flattenClickHouseUserSettings(user.Settings)}

		if len(user.Quotas) > 0 {
			quotas := schema.NewSet(clickHouseUserQuotaHash, nil)
			for _, quota := range user.Quotas {
				p := flattenClickHouseUserQuota(quota)
				quotas.Add(p)
			}
			u["quota"] = quotas
		}

		result.Add(u)
	}
	return result
}

func expandClickHouseUser(u map[string]interface{}, d *schema.ResourceData, hash int) *clickhouse.UserSpec {
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

	if v, ok := u["settings"]; ok {
		if d != nil {
			user.Settings = expandClickHouseUserSettingsExists(d, hash)
		} else {
			// for compare, when we have old Set without ResourceData
			for _, settings := range v.([]interface{}) {
				user.Settings = expandClickHouseUserSettings(settings.(map[string]interface{}))
			}
		}
	}

	if v, ok := u["quota"]; ok {
		if d != nil {
			user.Quotas = expandClickHouseUserQuotasExists(d, hash)
		} else {
			user.Quotas = expandClickHouseUserQuotas(v.(*schema.Set))
		}
	}

	return user
}

func expandClickHouseUserSpecs(d *schema.ResourceData) ([]*clickhouse.UserSpec, error) {
	result := []*clickhouse.UserSpec{}
	users := d.Get("user").(*schema.Set)

	for _, u := range users.List() {
		m := u.(map[string]interface{})
		hash := clickHouseUserHash(u)
		result = append(result, expandClickHouseUser(m, d, hash))
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

func expandClickHouseCloudStorage(d *schema.ResourceData) *clickhouse.CloudStorage {
	result := &clickhouse.CloudStorage{}
	cloudStorage := d.Get("cloud_storage").([]interface{})

	for _, g := range cloudStorage {
		cloudStorageSpec := g.(map[string]interface{})
		if val, ok := cloudStorageSpec["enabled"]; ok {
			result.SetEnabled(val.(bool))
		}
	}

	return result
}

func flattenClickHouseCloudStorage(cs *clickhouse.CloudStorage) []map[string]interface{} {
	result := []map[string]interface{}{}

	if cs != nil && cs.GetEnabled() {
		m := map[string]interface{}{}
		m["enabled"] = true
		result = append(result, m)
	}

	return result
}

func parseClickHouseWeekDay(wd string) (clickhouse.WeeklyMaintenanceWindow_WeekDay, error) {
	val, ok := clickhouse.WeeklyMaintenanceWindow_WeekDay_value[wd]
	// do not allow WEEK_DAY_UNSPECIFIED
	if !ok || val == 0 {
		return clickhouse.WeeklyMaintenanceWindow_WEEK_DAY_UNSPECIFIED,
			fmt.Errorf("value for 'day' should be one of %s, not `%s`",
				getJoinedKeys(getEnumValueMapKeysExt(clickhouse.WeeklyMaintenanceWindow_WeekDay_value, true)), wd)
	}

	return clickhouse.WeeklyMaintenanceWindow_WeekDay(val), nil
}

func expandClickHouseMaintenanceWindow(d *schema.ResourceData) (*clickhouse.MaintenanceWindow, error) {
	mwType, ok := d.GetOk("maintenance_window.0.type")
	if !ok {
		return nil, nil
	}

	result := &clickhouse.MaintenanceWindow{}

	switch mwType {
	case "ANYTIME":
		timeSet := false
		if _, ok := d.GetOk("maintenance_window.0.day"); ok {
			timeSet = true
		}
		if _, ok := d.GetOk("maintenance_window.0.hour"); ok {
			timeSet = true
		}
		if timeSet {
			return nil, fmt.Errorf("with ANYTIME type of maintenance window both DAY and HOUR should be omitted")
		}
		result.SetAnytime(&clickhouse.AnytimeMaintenanceWindow{})

	case "WEEKLY":
		weekly := &clickhouse.WeeklyMaintenanceWindow{}
		if val, ok := d.GetOk("maintenance_window.0.day"); ok {
			var err error
			weekly.Day, err = parseClickHouseWeekDay(val.(string))
			if err != nil {
				return nil, err
			}
		}
		if v, ok := d.GetOk("maintenance_window.0.hour"); ok {
			weekly.Hour = int64(v.(int))
		}

		result.SetWeeklyMaintenanceWindow(weekly)
	}

	return result, nil
}

func flattenClickHouseMaintenanceWindow(mw *clickhouse.MaintenanceWindow) []map[string]interface{} {
	result := map[string]interface{}{}

	if val := mw.GetAnytime(); val != nil {
		result["type"] = "ANYTIME"
	}

	if val := mw.GetWeeklyMaintenanceWindow(); val != nil {
		result["type"] = "WEEKLY"
		result["day"] = val.Day.String()
		result["hour"] = val.Hour
	}

	return []map[string]interface{}{result}
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
