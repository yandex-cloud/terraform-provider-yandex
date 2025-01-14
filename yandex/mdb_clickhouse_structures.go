package yandex

import (
	"bytes"
	"fmt"
	"log"
	"reflect"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
)

var originalClusterResources *clickhouse.Resources

func isEqualResources(clusterResources *clickhouse.Resources, shardResources *clickhouse.Resources) bool {
	if clusterResources.GetDiskSize() != shardResources.GetDiskSize() {
		log.Printf("[DEBUG] resource is different by disk_size: cluster disk_size=%d, shard disk_size=%d\n", clusterResources.GetDiskSize(), shardResources.GetDiskSize())
		return false
	}
	if clusterResources.GetDiskTypeId() != shardResources.GetDiskTypeId() {
		log.Printf("[DEBUG] resource is different by disk_type_id: cluster disk_type_id=%s, shard disk_type_id=%s\n", clusterResources.GetDiskTypeId(), shardResources.GetDiskTypeId())
		return false
	}
	if clusterResources.GetResourcePresetId() != shardResources.GetResourcePresetId() {
		log.Printf("[DEBUG] resource is different by resource_preset_id: cluster resource_preset_id=%s, shard resource_preset_id=%s\n", clusterResources.GetResourcePresetId(), shardResources.GetResourcePresetId())
		return false
	}
	log.Println("[DEBUG] resources are equal")
	return true
}

func backupOriginalClusterResource(d *schema.ResourceData) {
	originalClusterResources = expandClickHouseResources(d, "clickhouse.0.resources.0")
	log.Printf("[DEBUG] update original schema cluster resources=%v\n", originalClusterResources)
}

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
			v, ok := settings.(map[string]interface{})
			if !ok {
				break
			}
			settings := expandClickHouseUserSettings(v)
			p := flattenClickHouseUserSettings(settings)
			buf.WriteString(fmt.Sprintf("%v-", p))
			emptySettings = false
			// TODO: SA4004: the surrounding loop is unconditionally terminated (staticcheck)
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

func clickHouseShardHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})
	if n, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n.(string)))
	}
	if w, ok := m["weight"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", w.(int)))
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
// Returns the slice of user specs which have changed with changed fields of each user
func clickHouseChangedUsers(oldSpecs *schema.Set, newSpecs *schema.Set, d *schema.ResourceData) ([]*clickhouse.UserSpec, [][]string) {
	result := []*clickhouse.UserSpec{}
	updatedPathsOfChangedUsers := [][]string{}
	m := map[string]*clickhouse.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}), nil, 0)
		m[user.Name] = user
	}

	for _, spec := range newSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}), nil, 0)
		if u, ok := m[user.Name]; ok {
			paths := []string{}
			if user.Password != u.Password {
				paths = append(paths, "password")
			}
			if fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				paths = append(paths, "permissions")
			}
			if fmt.Sprintf("%v", user.Settings) != fmt.Sprintf("%v", u.Settings) {
				paths = append(paths, "settings")
			}
			if fmt.Sprintf("%v", user.Quotas) != fmt.Sprintf("%v", u.Quotas) {
				paths = append(paths, "quotas")
			}

			if len(paths) > 0 {
				hash := clickHouseUserHash(spec)
				userWithExistsFields := expandClickHouseUser(spec.(map[string]interface{}), d, hash)
				result = append(result, userWithExistsFields)
				updatedPathsOfChangedUsers = append(updatedPathsOfChangedUsers, paths)
			}
		}
	}
	return result, updatedPathsOfChangedUsers
}

func createKey(hostType clickhouse.Host_Type, zoneId, shardName string) string {
	if hostType == clickhouse.Host_ZOOKEEPER {
		shardName = "zk"
	}
	return hostType.String() + zoneId + shardName
}

func getKey(h *clickhouse.HostSpec) string {
	if h == nil {
		log.Println("[ERROR] host is nil. failed create key")
		return ""
	}
	shardName := "shard1"
	if h.ShardName != "" {
		shardName = h.ShardName
	}
	key := createKey(h.Type, h.ZoneId, shardName)

	return key
}

func getKeyChangesHosts(h *clickhouse.Host) string {
	if h == nil {
		log.Println("[ERROR] host is nil. failed create key")
		return ""
	}

	key := createKey(h.Type, h.ZoneId, h.ShardName)
	return key
}

func getHostsToAdd(keysHosts map[string][]*clickhouse.HostSpec, mKeys []string) map[string][]*clickhouse.HostSpec {
	toAdd := map[string][]*clickhouse.HostSpec{}

	for _, key := range mKeys {
		hs, ok := keysHosts[key]
		// we already proccessed host with such key via update or delete action
		if !ok || len(hs) == 0 {
			continue
		}
		h := hs[0]
		if h.Type == clickhouse.Host_ZOOKEEPER {
			toAdd["zk"] = append(toAdd["zk"], h)
		} else {
			toAdd[h.ShardName] = append(toAdd[h.ShardName], h)
		}
		if len(hs) > 1 {
			keysHosts[key] = hs[1:]
		} else {
			delete(keysHosts, key)
		}
	}

	return toAdd
}

func getChangesHosts(currHosts []*clickhouse.Host, keysHosts map[string][]*clickhouse.HostSpec) (map[string][]string, map[string]*clickhouse.UpdateHostSpec) {
	toDelete := map[string][]string{}
	toUpdate := map[string]*clickhouse.UpdateHostSpec{}

	for _, h := range currHosts {
		key := getKeyChangesHosts(h)

		hs, ok := keysHosts[key]
		if !ok {
			toDelete[h.ShardName] = append(toDelete[h.ShardName], h.Name)
		} else {
			updateRequired := false
			uh := &clickhouse.UpdateHostSpec{HostName: h.Name, UpdateMask: &field_mask.FieldMask{}}
			if hs[0].AssignPublicIp != h.AssignPublicIp {
				updateRequired = true
				uh.AssignPublicIp = &wrapperspb.BoolValue{Value: hs[0].AssignPublicIp}
				uh.UpdateMask.Paths = append(uh.UpdateMask.Paths, "assign_public_ip")
			}
			if updateRequired {
				toUpdate[h.Name] = uh
			}
		}
		if len(hs) > 1 {
			keysHosts[key] = hs[1:]
		} else {
			delete(keysHosts, key)
		}
	}

	return toDelete, toUpdate
}

// Takes the current list of hosts and the desirable list of hosts.
// Returns the map of hostnames to delete grouped by shard,
// and the map of hosts to add grouped by shard as well.
// All the ZOOKEEPER hosts will reside under the key "zk".
func clickHouseHostsDiff(currHosts []*clickhouse.Host, targetHosts []*clickhouse.HostSpec) (map[string][]string, map[string][]*clickhouse.HostSpec, map[string]*clickhouse.UpdateHostSpec) {
	keysHosts := map[string][]*clickhouse.HostSpec{}
	var mKeys []string

	for _, h := range targetHosts {
		key := getKey(h)
		keysHosts[key] = append(keysHosts[key], h)
		mKeys = append(mKeys, key)
	}

	toDelete, toUpdate := getChangesHosts(currHosts, keysHosts)
	toAdd := getHostsToAdd(keysHosts, mKeys)

	return toDelete, toAdd, toUpdate
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

	if c.ReplicatedDeduplicationWindow != nil {
		res["replicated_deduplication_window"] = c.ReplicatedDeduplicationWindow.Value
	}
	if c.ReplicatedDeduplicationWindowSeconds != nil {
		res["replicated_deduplication_window_seconds"] = c.ReplicatedDeduplicationWindowSeconds.Value
	}
	if c.PartsToDelayInsert != nil {
		res["parts_to_delay_insert"] = c.PartsToDelayInsert.Value
	}
	if c.PartsToThrowInsert != nil {
		res["parts_to_throw_insert"] = c.PartsToThrowInsert.Value
	}
	if c.InactivePartsToDelayInsert != nil {
		res["inactive_parts_to_delay_insert"] = c.InactivePartsToDelayInsert.Value
	}
	if c.InactivePartsToThrowInsert != nil {
		res["inactive_parts_to_throw_insert"] = c.InactivePartsToThrowInsert.Value
	}
	if c.MaxReplicatedMergesInQueue != nil {
		res["max_replicated_merges_in_queue"] = c.MaxReplicatedMergesInQueue.Value
	}
	if c.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge != nil {
		res["number_of_free_entries_in_pool_to_lower_max_size_of_merge"] = c.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge.Value
	}
	if c.MaxBytesToMergeAtMinSpaceInPool != nil {
		res["max_bytes_to_merge_at_min_space_in_pool"] = c.MaxBytesToMergeAtMinSpaceInPool.Value
	}
	if c.MaxBytesToMergeAtMaxSpaceInPool != nil {
		res["max_bytes_to_merge_at_max_space_in_pool"] = c.MaxBytesToMergeAtMaxSpaceInPool.Value
	}
	if c.MinBytesForWidePart != nil {
		res["min_bytes_for_wide_part"] = c.MinBytesForWidePart.Value
	}
	if c.MinRowsForWidePart != nil {
		res["min_rows_for_wide_part"] = c.MinRowsForWidePart.Value
	}
	if c.TtlOnlyDropParts != nil {
		res["ttl_only_drop_parts"] = c.TtlOnlyDropParts.Value
	}
	if c.AllowRemoteFsZeroCopyReplication != nil {
		res["allow_remote_fs_zero_copy_replication"] = c.AllowRemoteFsZeroCopyReplication.Value
	}
	if c.MergeWithTtlTimeout != nil {
		res["merge_with_ttl_timeout"] = c.MergeWithTtlTimeout.Value
	}
	if c.MergeWithRecompressionTtlTimeout != nil {
		res["merge_with_recompression_ttl_timeout"] = c.MergeWithRecompressionTtlTimeout.Value
	}
	if c.MaxPartsInTotal != nil {
		res["max_parts_in_total"] = c.MaxPartsInTotal.Value
	}
	if c.MaxNumberOfMergesWithTtlInPool != nil {
		res["max_number_of_merges_with_ttl_in_pool"] = c.MaxNumberOfMergesWithTtlInPool.Value
	}
	if c.CleanupDelayPeriod != nil {
		res["cleanup_delay_period"] = c.CleanupDelayPeriod.Value
	}
	if c.NumberOfFreeEntriesInPoolToExecuteMutation != nil {
		res["number_of_free_entries_in_pool_to_execute_mutation"] = c.NumberOfFreeEntriesInPoolToExecuteMutation.Value
	}
	if c.MaxAvgPartSizeForTooManyParts != nil {
		res["max_avg_part_size_for_too_many_parts"] = c.MaxAvgPartSizeForTooManyParts.Value
	}
	if c.MinAgeToForceMergeSeconds != nil {
		res["min_age_to_force_merge_seconds"] = c.MinAgeToForceMergeSeconds.Value
	}
	if c.MinAgeToForceMergeOnPartitionOnly != nil {
		res["min_age_to_force_merge_on_partition_only"] = c.MinAgeToForceMergeOnPartitionOnly.Value
	}
	if c.MergeSelectingSleepMs != nil {
		res["merge_selecting_sleep_ms"] = c.MergeSelectingSleepMs.Value
	}
	if c.MergeMaxBlockSize != nil {
		res["merge_max_block_size"] = c.MergeMaxBlockSize.Value
	}
	if c.CheckSampleColumnIsCorrect != nil {
		res["check_sample_column_is_correct"] = c.CheckSampleColumnIsCorrect.Value
	}
	if c.MaxMergeSelectingSleepMs != nil {
		res["max_merge_selecting_sleep_ms"] = c.MaxMergeSelectingSleepMs.Value
	}
	if c.MaxCleanupDelayPeriod != nil {
		res["max_cleanup_delay_period"] = c.MaxCleanupDelayPeriod.Value
	}

	return []map[string]interface{}{res}, nil
}

func flattenClickhouseKafkaSettings(d *schema.ResourceData, keyPath string, c *clickhouseConfig.ClickhouseConfig_Kafka) ([]map[string]interface{}, error) {
	if c == nil {
		return []map[string]interface{}{}, nil
	}

	res := map[string]interface{}{}

	res["security_protocol"] = c.SecurityProtocol.String()
	res["sasl_mechanism"] = c.SaslMechanism.String()
	res["sasl_username"] = c.SaslUsername
	if v, ok := d.GetOk(keyPath + ".sasl_password"); ok {
		res["sasl_password"] = v.(string)
	}
	if c.EnableSslCertificateVerification != nil {
		res["enable_ssl_certificate_verification"] = c.EnableSslCertificateVerification.Value
	}
	if c.MaxPollIntervalMs != nil {
		res["max_poll_interval_ms"] = c.MaxPollIntervalMs.Value
	}
	if c.SessionTimeoutMs != nil {
		res["session_timeout_ms"] = c.SessionTimeoutMs.Value
	}
	res["debug"] = c.Debug.String()
	res["auto_offset_reset"] = c.AutoOffsetReset.String()

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
	if c == nil {
		return []map[string]interface{}{}, nil
	}

	res := map[string]interface{}{}

	res["username"] = c.Username
	if v, ok := d.GetOk("clickhouse.0.config.0.rabbitmq.0.password"); ok {
		res["password"] = v.(string)
	}
	if v, ok := d.GetOk("clickhouse.0.config.0.rabbitmq.0.vhost"); ok {
		res["vhost"] = v.(string)
	}

	return []map[string]interface{}{res}, nil
}

func flattenClickhouseCompressionSettings(c []*clickhouseConfig.ClickhouseConfig_Compression) ([]interface{}, error) {
	var result []interface{}

	for _, r := range c {
		compressionSettings := map[string]interface{}{
			"method":              r.Method.String(),
			"min_part_size":       r.MinPartSize,
			"min_part_size_ratio": r.MinPartSizeRatio,
		}
		if r.Level != nil && r.Level.GetValue() > 0 {
			compressionSettings["level"] = r.Level.Value
		}
		result = append(result, compressionSettings)
	}

	return result, nil
}

func flattenClickhouseGraphiteRollupSettings(c []*clickhouseConfig.ClickhouseConfig_GraphiteRollup) ([]interface{}, error) {
	var result []interface{}

	for _, r := range c {
		rollup := map[string]interface{}{
			"name":                r.Name,
			"pattern":             []interface{}{},
			"path_column_name":    r.PathColumnName,
			"time_column_name":    r.TimeColumnName,
			"value_column_name":   r.ValueColumnName,
			"version_column_name": r.VersionColumnName,
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

func flattenClickhouseQueryMaskingRulesSettings(c []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule) ([]interface{}, error) {
	var result []interface{}

	for _, r := range c {
		queryMaskingRuleSettings := map[string]interface{}{
			"name":    r.Name,
			"regexp":  r.Regexp,
			"replace": r.Replace,
		}
		result = append(result, queryMaskingRuleSettings)
	}

	return result, nil
}

func flattenClickhouseQueryCacheSettings(c *clickhouseConfig.ClickhouseConfig_QueryCache) ([]map[string]interface{}, error) {
	if c == nil {
		return []map[string]interface{}{}, nil
	}

	res := map[string]interface{}{}

	if c.MaxSizeInBytes != nil {
		res["max_size_in_bytes"] = c.MaxSizeInBytes.Value
	}
	if c.MaxEntries != nil {
		res["max_entries"] = c.MaxEntries.Value
	}
	if c.MaxEntrySizeInBytes != nil {
		res["max_entry_size_in_bytes"] = c.MaxEntrySizeInBytes.Value
	}
	if c.MaxEntrySizeInRows != nil {
		res["max_entry_size_in_rows"] = c.MaxEntrySizeInRows.Value
	}

	return []map[string]interface{}{res}, nil
}

func flattenClickHouseConfig(d *schema.ResourceData, c *clickhouseConfig.ClickhouseConfigSet) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["log_level"] = c.EffectiveConfig.LogLevel.String()

	if c.EffectiveConfig.MaxConnections != nil {
		res["max_connections"] = c.EffectiveConfig.MaxConnections.Value
	}
	if c.EffectiveConfig.MaxConcurrentQueries != nil {
		res["max_concurrent_queries"] = c.EffectiveConfig.MaxConcurrentQueries.Value
	}
	if c.EffectiveConfig.KeepAliveTimeout != nil {
		res["keep_alive_timeout"] = c.EffectiveConfig.KeepAliveTimeout.Value
	}
	if c.EffectiveConfig.UncompressedCacheSize != nil {
		res["uncompressed_cache_size"] = c.EffectiveConfig.UncompressedCacheSize.Value
	}
	if c.EffectiveConfig.MarkCacheSize != nil {
		res["mark_cache_size"] = c.EffectiveConfig.MarkCacheSize.Value
	}
	if c.EffectiveConfig.MaxTableSizeToDrop != nil {
		res["max_table_size_to_drop"] = c.EffectiveConfig.MaxTableSizeToDrop.Value
	}
	if c.EffectiveConfig.MaxPartitionSizeToDrop != nil {
		res["max_partition_size_to_drop"] = c.EffectiveConfig.MaxPartitionSizeToDrop.Value
	}
	res["timezone"] = c.EffectiveConfig.Timezone
	res["geobase_uri"] = c.EffectiveConfig.GeobaseUri
	if c.EffectiveConfig.GeobaseEnabled != nil {
		res["geobase_enabled"] = c.EffectiveConfig.GeobaseEnabled.Value
	}

	if c.EffectiveConfig.QueryLogRetentionSize != nil {
		res["query_log_retention_size"] = c.EffectiveConfig.QueryLogRetentionSize.Value
	}

	if c.EffectiveConfig.QueryLogRetentionTime != nil {
		res["query_log_retention_time"] = c.EffectiveConfig.QueryLogRetentionTime.Value
	}
	if c.EffectiveConfig.QueryThreadLogEnabled != nil {
		res["query_thread_log_enabled"] = c.EffectiveConfig.QueryThreadLogEnabled.Value
	}
	if c.EffectiveConfig.QueryThreadLogRetentionSize != nil {
		res["query_thread_log_retention_size"] = c.EffectiveConfig.QueryThreadLogRetentionSize.Value
	}
	if c.EffectiveConfig.QueryThreadLogRetentionTime != nil {
		res["query_thread_log_retention_time"] = c.EffectiveConfig.QueryThreadLogRetentionTime.Value
	}
	if c.EffectiveConfig.PartLogRetentionSize != nil {
		res["part_log_retention_size"] = c.EffectiveConfig.PartLogRetentionSize.Value
	}
	if c.EffectiveConfig.PartLogRetentionTime != nil {
		res["part_log_retention_time"] = c.EffectiveConfig.PartLogRetentionTime.Value
	}
	if c.EffectiveConfig.MetricLogEnabled != nil {
		res["metric_log_enabled"] = c.EffectiveConfig.MetricLogEnabled.Value
	}
	if c.EffectiveConfig.MetricLogRetentionSize != nil {
		res["metric_log_retention_size"] = c.EffectiveConfig.MetricLogRetentionSize.Value
	}
	if c.EffectiveConfig.MetricLogRetentionTime != nil {
		res["metric_log_retention_time"] = c.EffectiveConfig.MetricLogRetentionTime.Value
	}
	if c.EffectiveConfig.TraceLogEnabled != nil {
		res["trace_log_enabled"] = c.EffectiveConfig.TraceLogEnabled.Value
	}
	if c.EffectiveConfig.TraceLogRetentionSize != nil {
		res["trace_log_retention_size"] = c.EffectiveConfig.TraceLogRetentionSize.Value
	}
	if c.EffectiveConfig.TraceLogRetentionTime != nil {
		res["trace_log_retention_time"] = c.EffectiveConfig.TraceLogRetentionTime.Value
	}
	if c.EffectiveConfig.TextLogEnabled != nil {
		res["text_log_enabled"] = c.EffectiveConfig.TextLogEnabled.Value
	}
	if c.EffectiveConfig.TextLogRetentionSize != nil {
		res["text_log_retention_size"] = c.EffectiveConfig.TextLogRetentionSize.Value
	}
	if c.EffectiveConfig.TextLogRetentionTime != nil {
		res["text_log_retention_time"] = c.EffectiveConfig.TextLogRetentionTime.Value
	}
	if c.EffectiveConfig.OpentelemetrySpanLogEnabled != nil {
		res["opentelemetry_span_log_enabled"] = c.EffectiveConfig.OpentelemetrySpanLogEnabled.Value
	}
	if c.EffectiveConfig.OpentelemetrySpanLogRetentionSize != nil {
		res["opentelemetry_span_log_retention_size"] = c.EffectiveConfig.OpentelemetrySpanLogRetentionSize.Value
	}
	if c.EffectiveConfig.OpentelemetrySpanLogRetentionTime != nil {
		res["opentelemetry_span_log_retention_time"] = c.EffectiveConfig.OpentelemetrySpanLogRetentionTime.Value
	}
	if c.EffectiveConfig.QueryViewsLogEnabled != nil {
		res["query_views_log_enabled"] = c.EffectiveConfig.QueryViewsLogEnabled.Value
	}
	if c.EffectiveConfig.QueryViewsLogRetentionSize != nil {
		res["query_views_log_retention_size"] = c.EffectiveConfig.QueryViewsLogRetentionSize.Value
	}
	if c.EffectiveConfig.QueryViewsLogRetentionTime != nil {
		res["query_views_log_retention_time"] = c.EffectiveConfig.QueryViewsLogRetentionTime.Value
	}
	if c.EffectiveConfig.AsynchronousMetricLogEnabled != nil {
		res["asynchronous_metric_log_enabled"] = c.EffectiveConfig.AsynchronousMetricLogEnabled.Value
	}
	if c.EffectiveConfig.AsynchronousMetricLogRetentionSize != nil {
		res["asynchronous_metric_log_retention_size"] = c.EffectiveConfig.AsynchronousMetricLogRetentionSize.Value
	}
	if c.EffectiveConfig.AsynchronousMetricLogRetentionTime != nil {
		res["asynchronous_metric_log_retention_time"] = c.EffectiveConfig.AsynchronousMetricLogRetentionTime.Value
	}
	if c.EffectiveConfig.SessionLogEnabled != nil {
		res["session_log_enabled"] = c.EffectiveConfig.SessionLogEnabled.Value
	}
	if c.EffectiveConfig.SessionLogRetentionSize != nil {
		res["session_log_retention_size"] = c.EffectiveConfig.SessionLogRetentionSize.Value
	}
	if c.EffectiveConfig.SessionLogRetentionTime != nil {
		res["session_log_retention_time"] = c.EffectiveConfig.SessionLogRetentionTime.Value
	}
	if c.EffectiveConfig.ZookeeperLogEnabled != nil {
		res["zookeeper_log_enabled"] = c.EffectiveConfig.ZookeeperLogEnabled.Value
	}
	if c.EffectiveConfig.ZookeeperLogRetentionSize != nil {
		res["zookeeper_log_retention_size"] = c.EffectiveConfig.ZookeeperLogRetentionSize.Value
	}
	if c.EffectiveConfig.ZookeeperLogRetentionTime != nil {
		res["zookeeper_log_retention_time"] = c.EffectiveConfig.ZookeeperLogRetentionTime.Value
	}
	if c.EffectiveConfig.AsynchronousInsertLogEnabled != nil {
		res["asynchronous_insert_log_enabled"] = c.EffectiveConfig.AsynchronousInsertLogEnabled.Value
	}
	if c.EffectiveConfig.AsynchronousInsertLogRetentionSize != nil {
		res["asynchronous_insert_log_retention_size"] = c.EffectiveConfig.AsynchronousInsertLogRetentionSize.Value
	}
	if c.EffectiveConfig.AsynchronousInsertLogRetentionTime != nil {
		res["asynchronous_insert_log_retention_time"] = c.EffectiveConfig.AsynchronousInsertLogRetentionTime.Value
	}

	res["text_log_level"] = c.EffectiveConfig.TextLogLevel.String()

	if c.EffectiveConfig.BackgroundPoolSize != nil {
		res["background_pool_size"] = c.EffectiveConfig.BackgroundPoolSize.Value
	}

	if c.EffectiveConfig.BackgroundSchedulePoolSize != nil {
		res["background_schedule_pool_size"] = c.EffectiveConfig.BackgroundSchedulePoolSize.Value
	}

	if c.EffectiveConfig.BackgroundFetchesPoolSize != nil {
		res["background_fetches_pool_size"] = c.EffectiveConfig.BackgroundFetchesPoolSize.Value
	}
	if c.EffectiveConfig.BackgroundMovePoolSize != nil {
		res["background_move_pool_size"] = c.EffectiveConfig.BackgroundMovePoolSize.Value
	}
	if c.EffectiveConfig.BackgroundDistributedSchedulePoolSize != nil {
		res["background_distributed_schedule_pool_size"] = c.EffectiveConfig.BackgroundDistributedSchedulePoolSize.Value
	}
	if c.EffectiveConfig.BackgroundBufferFlushSchedulePoolSize != nil {
		res["background_buffer_flush_schedule_pool_size"] = c.EffectiveConfig.BackgroundBufferFlushSchedulePoolSize.Value
	}
	if c.EffectiveConfig.BackgroundMessageBrokerSchedulePoolSize != nil {
		res["background_message_broker_schedule_pool_size"] = c.EffectiveConfig.BackgroundMessageBrokerSchedulePoolSize.Value
	}
	if c.EffectiveConfig.BackgroundCommonPoolSize != nil {
		res["background_common_pool_size"] = c.EffectiveConfig.BackgroundCommonPoolSize.Value
	}
	if c.EffectiveConfig.BackgroundMergesMutationsConcurrencyRatio != nil {
		res["background_merges_mutations_concurrency_ratio"] = c.EffectiveConfig.BackgroundMergesMutationsConcurrencyRatio.Value
	}
	if c.EffectiveConfig.DefaultDatabase != nil && len(c.EffectiveConfig.DefaultDatabase.Value) != 0 {
		res["default_database"] = c.EffectiveConfig.DefaultDatabase.Value
	}
	if c.EffectiveConfig.TotalMemoryProfilerStep != nil {
		res["total_memory_profiler_step"] = c.EffectiveConfig.TotalMemoryProfilerStep.Value
	}
	if c.EffectiveConfig.DictionariesLazyLoad != nil {
		res["dictionaries_lazy_load"] = c.EffectiveConfig.DictionariesLazyLoad.Value
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

	queryMaskingRules, err := flattenClickhouseQueryMaskingRulesSettings(c.EffectiveConfig.QueryMaskingRules)
	if err != nil {
		return nil, err
	}
	res["query_masking_rules"] = queryMaskingRules

	queryCache, err := flattenClickhouseQueryCacheSettings(c.EffectiveConfig.QueryCache)
	if err != nil {
		return nil, err
	}
	res["query_cache"] = queryCache

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
	// TODO: SA1019: d.GetOkExists is deprecated: usage is discouraged due to undefined behaviors and may be removed in a future version of the SDK (staticcheck)
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
	if v, ok := d.GetOkExists(rootKey + ".inactive_parts_to_delay_insert"); ok {
		config.InactivePartsToDelayInsert = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".inactive_parts_to_throw_insert"); ok {
		config.InactivePartsToThrowInsert = &wrappers.Int64Value{Value: int64(v.(int))}
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
	if v, ok := d.GetOkExists(rootKey + ".max_bytes_to_merge_at_max_space_in_pool"); ok {
		config.MaxBytesToMergeAtMaxSpaceInPool = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".min_bytes_for_wide_part"); ok {
		config.MinBytesForWidePart = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".min_rows_for_wide_part"); ok {
		config.MinRowsForWidePart = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".ttl_only_drop_parts"); ok {
		config.TtlOnlyDropParts = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".allow_remote_fs_zero_copy_replication"); ok {
		config.AllowRemoteFsZeroCopyReplication = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".merge_with_ttl_timeout"); ok {
		config.MergeWithTtlTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".merge_with_recompression_ttl_timeout"); ok {
		config.MergeWithRecompressionTtlTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_parts_in_total"); ok {
		config.MaxPartsInTotal = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_number_of_merges_with_ttl_in_pool"); ok {
		config.MaxNumberOfMergesWithTtlInPool = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".cleanup_delay_period"); ok {
		config.CleanupDelayPeriod = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".number_of_free_entries_in_pool_to_execute_mutation"); ok {
		config.NumberOfFreeEntriesInPoolToExecuteMutation = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_avg_part_size_for_too_many_parts"); ok {
		config.MaxAvgPartSizeForTooManyParts = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".min_age_to_force_merge_seconds"); ok {
		config.MinAgeToForceMergeSeconds = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".min_age_to_force_merge_on_partition_only"); ok {
		config.MinAgeToForceMergeOnPartitionOnly = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".merge_selecting_sleep_ms"); ok {
		config.MergeSelectingSleepMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".merge_max_block_size"); ok {
		config.MergeMaxBlockSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".check_sample_column_is_correct"); ok {
		config.CheckSampleColumnIsCorrect = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_merge_selecting_sleep_ms"); ok {
		config.MaxMergeSelectingSleepMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_cleanup_delay_period"); ok {
		config.MaxCleanupDelayPeriod = &wrappers.Int64Value{Value: int64(v.(int))}
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
	if v, ok := d.GetOkExists(rootKey + ".enable_ssl_certificate_verification"); ok {
		config.EnableSslCertificateVerification = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(rootKey + ".max_poll_interval_ms"); ok {
		config.MaxPollIntervalMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".session_timeout_ms"); ok {
		config.SessionTimeoutMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".debug"); ok {
		if val, err := expandEnum("debug", v.(string), clickhouseConfig.ClickhouseConfig_Kafka_Debug_value); val != nil && err == nil {
			config.Debug = clickhouseConfig.ClickhouseConfig_Kafka_Debug(*val)
		} else {
			return nil, err
		}
	}
	if v, ok := d.GetOk(rootKey + ".auto_offset_reset"); ok {
		if val, err := expandEnum("auto_offset_reset", v.(string), clickhouseConfig.ClickhouseConfig_Kafka_AutoOffsetReset_value); val != nil && err == nil {
			config.AutoOffsetReset = clickhouseConfig.ClickhouseConfig_Kafka_AutoOffsetReset(*val)
		} else {
			return nil, err
		}
	}

	return config, nil
}

func expandClickhouseKafkaTopicsSettings(d *schema.ResourceData, rootKey string) ([]*clickhouseConfig.ClickhouseConfig_KafkaTopic, error) {
	var result []*clickhouseConfig.ClickhouseConfig_KafkaTopic
	topics := d.Get(rootKey).([]interface{})

	for i := range topics {
		var topicKey = rootKey + fmt.Sprintf(".%d.settings.0", i)
		settings, err := expandClickhouseKafkaSettings(d, topicKey)
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
	if v, ok := d.GetOkExists(rootKey + ".vhost"); ok {
		config.Vhost = v.(string)
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

		if v, ok := d.GetOk(keyPrefix + ".level"); ok {
			compression.Level = &wrapperspb.Int64Value{Value: int64(v.(int))}
		}

		result = append(result, compression)
	}
	return result, nil
}

func expandClickhouseGraphiteRollupSettings(d *schema.ResourceData, rootKey string) ([]*clickhouseConfig.ClickhouseConfig_GraphiteRollup, error) {
	var result []*clickhouseConfig.ClickhouseConfig_GraphiteRollup

	for r := range d.Get(rootKey).([]interface{}) {
		rollupKey := rootKey + fmt.Sprintf(".%d", r)
		rollup := &clickhouseConfig.ClickhouseConfig_GraphiteRollup{
			Name:              d.Get(rollupKey + ".name").(string),
			PathColumnName:    d.Get(rollupKey + ".path_column_name").(string),
			TimeColumnName:    d.Get(rollupKey + ".time_column_name").(string),
			ValueColumnName:   d.Get(rollupKey + ".value_column_name").(string),
			VersionColumnName: d.Get(rollupKey + ".version_column_name").(string),
		}

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

func expandClickhouseQueryMaskingRulesSettings(d *schema.ResourceData, rootKey string) ([]*clickhouseConfig.ClickhouseConfig_QueryMaskingRule, error) {
	var result []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule
	queryMaskingRules := d.Get(rootKey).([]interface{})

	for i := range queryMaskingRules {
		keyPrefix := rootKey + fmt.Sprintf(".%d", i)
		queryMaskingRule := &clickhouseConfig.ClickhouseConfig_QueryMaskingRule{}

		if v, ok := d.GetOk(keyPrefix + ".name"); ok {
			queryMaskingRule.Name = v.(string)
		}

		queryMaskingRule.Regexp = d.Get(keyPrefix + ".regexp").(string)

		if v, ok := d.GetOk(keyPrefix + ".replace"); ok {
			queryMaskingRule.Replace = v.(string)
		}

		result = append(result, queryMaskingRule)
	}
	return result, nil
}

func expandClickhouseQueryCacheConfig(d *schema.ResourceData, rootKey string) (*clickhouseConfig.ClickhouseConfig_QueryCache, error) {
	config := &clickhouseConfig.ClickhouseConfig_QueryCache{}

	if v, ok := d.GetOkExists(rootKey + ".max_size_in_bytes"); ok {
		config.MaxSizeInBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_entries"); ok {
		config.MaxEntries = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_entry_size_in_bytes"); ok {
		config.MaxEntrySizeInBytes = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".max_entry_size_in_rows"); ok {
		config.MaxEntrySizeInRows = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return config, nil
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
	if v, ok := d.GetOkExists(rootKey + ".geobase_enabled"); ok {
		config.GeobaseEnabled = &wrappers.BoolValue{Value: v.(bool)}
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
	if v, ok := d.GetOkExists(rootKey + ".opentelemetry_span_log_enabled"); ok {
		config.OpentelemetrySpanLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".opentelemetry_span_log_retention_size"); ok {
		config.OpentelemetrySpanLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".opentelemetry_span_log_retention_time"); ok {
		config.OpentelemetrySpanLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".query_views_log_enabled"); ok {
		config.QueryViewsLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".query_views_log_retention_size"); ok {
		config.QueryViewsLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".query_views_log_retention_time"); ok {
		config.QueryViewsLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".asynchronous_metric_log_enabled"); ok {
		config.AsynchronousMetricLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".asynchronous_metric_log_retention_size"); ok {
		config.AsynchronousMetricLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".asynchronous_metric_log_retention_time"); ok {
		config.AsynchronousMetricLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".session_log_enabled"); ok {
		config.SessionLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".session_log_retention_size"); ok {
		config.SessionLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".session_log_retention_time"); ok {
		config.SessionLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".zookeeper_log_enabled"); ok {
		config.ZookeeperLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".zookeeper_log_retention_size"); ok {
		config.ZookeeperLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".zookeeper_log_retention_time"); ok {
		config.ZookeeperLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".asynchronous_insert_log_enabled"); ok {
		config.AsynchronousInsertLogEnabled = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOkExists(rootKey + ".asynchronous_insert_log_retention_size"); ok {
		config.AsynchronousInsertLogRetentionSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".asynchronous_insert_log_retention_time"); ok {
		config.AsynchronousInsertLogRetentionTime = &wrappers.Int64Value{Value: int64(v.(int))}
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

	if v, ok := d.GetOk(rootKey + ".background_fetches_pool_size"); ok {
		config.BackgroundFetchesPoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".background_move_pool_size"); ok {
		config.BackgroundMovePoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".background_distributed_schedule_pool_size"); ok {
		config.BackgroundDistributedSchedulePoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".background_buffer_flush_schedule_pool_size"); ok {
		config.BackgroundBufferFlushSchedulePoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".background_common_pool_size"); ok {
		config.BackgroundCommonPoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".background_message_broker_schedule_pool_size"); ok {
		config.BackgroundMessageBrokerSchedulePoolSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(rootKey + ".background_merges_mutations_concurrency_ratio"); ok {
		config.BackgroundMergesMutationsConcurrencyRatio = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".default_database"); ok {
		defaultDatabase, ok := v.(string)
		if ok && len(defaultDatabase) != 0 {
			config.DefaultDatabase = &wrappers.StringValue{Value: defaultDatabase}
		}
	}
	if v, ok := d.GetOk(rootKey + ".total_memory_profiler_step"); ok {
		config.TotalMemoryProfilerStep = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOkExists(rootKey + ".dictionaries_lazy_load"); ok {
		config.DictionariesLazyLoad = &wrappers.BoolValue{Value: v.(bool)}
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

	queryMaskingRules, err := expandClickhouseQueryMaskingRulesSettings(d, rootKey+".query_masking_rules")
	if err != nil {
		return nil, err
	}
	config.QueryMaskingRules = queryMaskingRules

	queryCacheSettings, err := expandClickhouseQueryCacheConfig(d, rootKey+".query_cache.0")
	if err != nil {
		return nil, err
	}
	config.QueryCache = queryCacheSettings

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
	if a != nil {
		res["web_sql"] = a.WebSql
		res["data_lens"] = a.DataLens
		res["metrika"] = a.Metrika
		res["serverless"] = a.Serverless
		res["data_transfer"] = a.DataTransfer
		res["yandex_query"] = a.YandexQuery
	}
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
	if v, ok := d.GetOk("access.0.data_transfer"); ok {
		result.DataTransfer = v.(bool)
	}
	if v, ok := d.GetOk("access.0.yandex_query"); ok {
		result.YandexQuery = v.(bool)
	}
	return result
}

func expandClickhouseBackupRetainPeriodDays(d *schema.ResourceData) *wrappers.Int64Value {
	if v, ok := d.GetOk("backup_retain_period_days"); ok {
		return &wrappers.Int64Value{
			Value: int64(v.(int)),
		}
	}
	return nil
}

func expandClickHouseUserSettingsJoinAlgorithm(jas []interface{}) []clickhouse.UserSettings_JoinAlgorithm {
	result := []clickhouse.UserSettings_JoinAlgorithm{}

	for _, ja := range jas {
		result = append(result, getJoinAlgorithmValue(ja.(string)))
	}

	return result
}

func flattenClickHouseUserSettingsJoinAlgorithm(jas []clickhouse.UserSettings_JoinAlgorithm) []interface{} {
	var result []interface{}

	for _, ja := range jas {
		result = append(result, getJoinAlgorithmName(ja))
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
	UserSettings_QuotaMode_value                = makeReversedMap(UserSettings_QuotaMode_name, clickhouse.UserSettings_QuotaMode_value)
	UserSettings_LocalFilesystemReadMethod_name = map[int32]string{
		0: "unspecified",
		1: "read",
		2: "pread_threadpool",
		3: "pread",
		4: "nmap",
	}
	UserSettings_LocalFilesystemReadMethod_value = makeReversedMap(UserSettings_LocalFilesystemReadMethod_name, clickhouse.UserSettings_LocalFilesystemReadMethod_value)
	UserSettings_RemoteFilesystemReadMethod_name = map[int32]string{
		0: "unspecified",
		1: "read",
		2: "threadpool",
	}
	UserSettings_RemoteFilesystemReadMethod_value = makeReversedMap(UserSettings_RemoteFilesystemReadMethod_name, clickhouse.UserSettings_RemoteFilesystemReadMethod_value)
	UserSettings_LoadBalancing_name               = map[int32]string{
		0: "unspecified",
		1: "random",
		2: "nearest_hostname",
		3: "in_order",
		4: "first_or_random",
		5: "round_robin",
	}
	UserSettings_LoadBalancing_value      = makeReversedMap(UserSettings_LoadBalancing_name, clickhouse.UserSettings_LoadBalancing_value)
	UserSettings_DateTimeInputFormat_name = map[int32]string{
		0: "unspecified",
		1: "best_effort",
		2: "basic",
		3: "best_effort_us",
	}
	UserSettings_DateTimeInputFormat_value = makeReversedMap(UserSettings_DateTimeInputFormat_name, clickhouse.UserSettings_DateTimeInputFormat_value)
	UserSettings_DateTimeOutputFormat_name = map[int32]string{
		0: "unspecified",
		1: "simple",
		2: "iso",
		3: "unix_timestamp",
	}
	UserSettings_DateTimeOutputFormat_value = makeReversedMap(UserSettings_DateTimeOutputFormat_name, clickhouse.UserSettings_DateTimeOutputFormat_value)
	UserSettings_JoinAlgorithm_name         = map[int32]string{
		0: "unspecified",
		1: "hash",
		2: "parallel_hash",
		3: "partial_merge",
		4: "direct",
		5: "auto",
		6: "full_sorting_merge",
		7: "prefer_partial_merge",
	}
	UserSettings_JoinAlgorithm_value = makeReversedMap(UserSettings_JoinAlgorithm_name, clickhouse.UserSettings_JoinAlgorithm_value)
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

func getLocalFilesystemReadMethodName(value clickhouse.UserSettings_LocalFilesystemReadMethod) string {
	if name, ok := UserSettings_LocalFilesystemReadMethod_name[int32(value)]; ok {
		return name
	}
	return UserSettings_LocalFilesystemReadMethod_name[0]
}

func getLocalFilesystemReadMethodValue(name string) clickhouse.UserSettings_LocalFilesystemReadMethod {
	if value, ok := UserSettings_LocalFilesystemReadMethod_value[name]; ok {
		return clickhouse.UserSettings_LocalFilesystemReadMethod(value)
	}
	return 0
}

func getRemoteFilesystemReadMethodName(value clickhouse.UserSettings_RemoteFilesystemReadMethod) string {
	if name, ok := UserSettings_RemoteFilesystemReadMethod_name[int32(value)]; ok {
		return name
	}
	return UserSettings_RemoteFilesystemReadMethod_name[0]
}

func getRemoteFilesystemReadMethodValue(name string) clickhouse.UserSettings_RemoteFilesystemReadMethod {
	if value, ok := UserSettings_RemoteFilesystemReadMethod_value[name]; ok {
		return clickhouse.UserSettings_RemoteFilesystemReadMethod(value)
	}
	return 0
}

func getLoadBalancingName(value clickhouse.UserSettings_LoadBalancing) string {
	if name, ok := UserSettings_LoadBalancing_name[int32(value)]; ok {
		return name
	}
	return UserSettings_LoadBalancing_name[0]
}

func getLoadBalancingValue(name string) clickhouse.UserSettings_LoadBalancing {
	if value, ok := UserSettings_LoadBalancing_value[name]; ok {
		return clickhouse.UserSettings_LoadBalancing(value)
	}
	return 0
}

func getDateTimeInputFormatName(value clickhouse.UserSettings_DateTimeInputFormat) string {
	if name, ok := UserSettings_DateTimeInputFormat_name[int32(value)]; ok {
		return name
	}
	return UserSettings_DateTimeInputFormat_name[0]
}

func getDateTimeInputFormatValue(name string) clickhouse.UserSettings_DateTimeInputFormat {
	if value, ok := UserSettings_DateTimeInputFormat_value[name]; ok {
		return clickhouse.UserSettings_DateTimeInputFormat(value)
	}
	return 0
}

func getDateTimeOutputFormatName(value clickhouse.UserSettings_DateTimeOutputFormat) string {
	if name, ok := UserSettings_DateTimeOutputFormat_name[int32(value)]; ok {
		return name
	}
	return UserSettings_DateTimeOutputFormat_name[0]
}

func getDateTimeOutputFormatValue(name string) clickhouse.UserSettings_DateTimeOutputFormat {
	if value, ok := UserSettings_DateTimeOutputFormat_value[name]; ok {
		return clickhouse.UserSettings_DateTimeOutputFormat(value)
	}
	return 0
}

func getJoinAlgorithmName(value clickhouse.UserSettings_JoinAlgorithm) string {
	if name, ok := UserSettings_JoinAlgorithm_name[int32(value)]; ok {
		return name
	}
	return UserSettings_JoinAlgorithm_name[0]
}

func getJoinAlgorithmValue(name string) clickhouse.UserSettings_JoinAlgorithm {
	if value, ok := UserSettings_JoinAlgorithm_value[name]; ok {
		return clickhouse.UserSettings_JoinAlgorithm(value)
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

func setSettingFromMapDouble(us map[string]interface{}, key string, setting **wrappers.DoubleValue) {
	if v, ok := us[key]; ok {
		if v.(float64) > 0 {
			*setting = &wrappers.DoubleValue{Value: v.(float64)}
		}
	}
}

func setSettingFromDataDouble(d *schema.ResourceData, fullKey string, setting **wrappers.DoubleValue) {
	if v, ok := d.GetOk(fullKey); ok {
		if v.(float64) > 0 {
			*setting = &wrappers.DoubleValue{Value: v.(float64)}
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
	setSettingFromMapBool(us, "insert_quorum_parallel", &result.InsertQuorumParallel)
	setSettingFromMapBool(us, "select_sequential_consistency", &result.SelectSequentialConsistency)
	setSettingFromMapBool(us, "deduplicate_blocks_in_dependent_materialized_views", &result.DeduplicateBlocksInDependentMaterializedViews)
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

	if v, ok := us["join_algorithm"]; ok {
		result.JoinAlgorithm = expandClickHouseUserSettingsJoinAlgorithm(v.([]interface{}))
	}

	setSettingFromMapBool(us, "any_join_distinct_right_table_keys", &result.AnyJoinDistinctRightTableKeys)
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
	setSettingFromMapBool(us, "input_format_null_as_default", &result.InputFormatNullAsDefault)
	setSettingFromMapBool(us, "input_format_with_names_use_header", &result.InputFormatWithNamesUseHeader)
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
	setSettingFromMapInt64(us, "max_concurrent_queries_for_user", &result.MaxConcurrentQueriesForUser)
	setSettingFromMapInt64(us, "memory_profiler_step", &result.MemoryProfilerStep)
	setSettingFromMapDouble(us, "memory_profiler_sample_probability", &result.MemoryProfilerSampleProbability)
	setSettingFromMapBool(us, "insert_null_as_default", &result.InsertNullAsDefault)
	setSettingFromMapBool(us, "allow_suspicious_low_cardinality_types", &result.AllowSuspiciousLowCardinalityTypes)
	setSettingFromMapInt64(us, "connect_timeout_with_failover", &result.ConnectTimeoutWithFailover)
	setSettingFromMapBool(us, "allow_introspection_functions", &result.AllowIntrospectionFunctions)
	setSettingFromMapBool(us, "async_insert", &result.AsyncInsert)
	setSettingFromMapInt64(us, "async_insert_threads", &result.AsyncInsertThreads)
	setSettingFromMapBool(us, "wait_for_async_insert", &result.WaitForAsyncInsert)
	setSettingFromMapInt64(us, "wait_for_async_insert_timeout", &result.WaitForAsyncInsertTimeout)
	setSettingFromMapInt64(us, "async_insert_max_data_size", &result.AsyncInsertMaxDataSize)
	setSettingFromMapInt64(us, "async_insert_busy_timeout", &result.AsyncInsertBusyTimeout)
	setSettingFromMapInt64(us, "async_insert_stale_timeout", &result.AsyncInsertStaleTimeout)

	setSettingFromMapInt64(us, "timeout_before_checking_execution_speed", &result.TimeoutBeforeCheckingExecutionSpeed)
	setSettingFromMapBool(us, "cancel_http_readonly_queries_on_client_close", &result.CancelHttpReadonlyQueriesOnClientClose)
	setSettingFromMapBool(us, "flatten_nested", &result.FlattenNested)

	if v, ok := us["format_regexp"]; ok {
		formatRegexp, ok := v.(string)
		if ok && len(formatRegexp) != 0 {
			result.FormatRegexp = formatRegexp
		}
	}

	setSettingFromMapBool(us, "format_regexp_skip_unmatched", &result.FormatRegexpSkipUnmatched)
	setSettingFromMapInt64(us, "max_http_get_redirects", &result.MaxHttpGetRedirects)

	if v, ok := us["quota_mode"]; ok {
		result.QuotaMode = getQuotaModeValue(v.(string))
	}

	setSettingFromMapBool(us, "input_format_import_nested_json", &result.InputFormatImportNestedJson)
	setSettingFromMapBool(us, "input_format_parallel_parsing", &result.InputFormatParallelParsing)
	setSettingFromMapInt64(us, "max_final_threads", &result.MaxFinalThreads)
	setSettingFromMapInt64(us, "max_read_buffer_size", &result.MaxReadBufferSize)

	if v, ok := us["local_filesystem_read_method"]; ok {
		result.LocalFilesystemReadMethod = getLocalFilesystemReadMethodValue(v.(string))
	}
	if v, ok := us["remote_filesystem_read_method"]; ok {
		result.RemoteFilesystemReadMethod = getRemoteFilesystemReadMethodValue(v.(string))
	}

	setSettingFromMapInt64(us, "insert_keeper_max_retries", &result.InsertKeeperMaxRetries)
	setSettingFromMapInt64(us, "max_temporary_data_on_disk_size_for_user", &result.MaxTemporaryDataOnDiskSizeForUser)
	setSettingFromMapInt64(us, "max_temporary_data_on_disk_size_for_query", &result.MaxTemporaryDataOnDiskSizeForQuery)
	setSettingFromMapInt64(us, "max_parser_depth", &result.MaxParserDepth)
	setSettingFromMapInt64(us, "memory_overcommit_ratio_denominator", &result.MemoryOvercommitRatioDenominator)
	setSettingFromMapInt64(us, "memory_overcommit_ratio_denominator_for_user", &result.MemoryOvercommitRatioDenominatorForUser)
	setSettingFromMapInt64(us, "memory_usage_overcommit_max_wait_microseconds", &result.MemoryUsageOvercommitMaxWaitMicroseconds)
	setSettingFromMapBool(us, "log_query_threads", &result.LogQueryThreads)
	setSettingFromMapInt64(us, "max_insert_threads", &result.MaxInsertThreads)
	setSettingFromMapBool(us, "use_hedged_requests", &result.UseHedgedRequests)
	setSettingFromMapInt64(us, "idle_connection_timeout", &result.IdleConnectionTimeout)
	setSettingFromMapInt64(us, "hedged_connection_timeout_ms", &result.HedgedConnectionTimeoutMs)

	if v, ok := us["load_balancing"]; ok {
		result.LoadBalancing = getLoadBalancingValue(v.(string))
	}

	setSettingFromMapBool(us, "prefer_localhost_replica", &result.PreferLocalhostReplica)

	if v, ok := us["date_time_input_format"]; ok {
		result.DateTimeInputFormat = getDateTimeInputFormatValue(v.(string))
	}

	if v, ok := us["date_time_output_format"]; ok {
		result.DateTimeOutputFormat = getDateTimeOutputFormatValue(v.(string))
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
	setSettingFromDataBool(d, rootKey+".insert_quorum_parallel", &result.InsertQuorumParallel)
	setSettingFromDataBool(d, rootKey+".select_sequential_consistency", &result.SelectSequentialConsistency)
	setSettingFromDataBool(d, rootKey+".deduplicate_blocks_in_dependent_materialized_views", &result.DeduplicateBlocksInDependentMaterializedViews)
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

	if v, ok := d.GetOk(rootKey + ".join_algorithm"); ok {
		result.JoinAlgorithm = expandClickHouseUserSettingsJoinAlgorithm(v.([]interface{}))
	}

	setSettingFromDataBool(d, rootKey+".any_join_distinct_right_table_keys", &result.AnyJoinDistinctRightTableKeys)
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
	setSettingFromDataBool(d, rootKey+".input_format_null_as_default", &result.InputFormatNullAsDefault)
	setSettingFromDataBool(d, rootKey+".input_format_with_names_use_header", &result.InputFormatWithNamesUseHeader)
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
	setSettingFromDataInt64(d, rootKey+".max_concurrent_queries_for_user", &result.MaxConcurrentQueriesForUser)
	setSettingFromDataInt64(d, rootKey+".memory_profiler_step", &result.MemoryProfilerStep)
	setSettingFromDataDouble(d, rootKey+".memory_profiler_sample_probability", &result.MemoryProfilerSampleProbability)
	setSettingFromDataBool(d, rootKey+".insert_null_as_default", &result.InsertNullAsDefault)
	setSettingFromDataBool(d, rootKey+".allow_suspicious_low_cardinality_types", &result.AllowSuspiciousLowCardinalityTypes)
	setSettingFromDataInt64(d, rootKey+".connect_timeout_with_failover", &result.ConnectTimeoutWithFailover)
	setSettingFromDataBool(d, rootKey+".allow_introspection_functions", &result.AllowIntrospectionFunctions)
	setSettingFromDataBool(d, rootKey+".async_insert", &result.AsyncInsert)
	setSettingFromDataInt64(d, rootKey+".async_insert_threads", &result.AsyncInsertThreads)
	setSettingFromDataBool(d, rootKey+".wait_for_async_insert", &result.WaitForAsyncInsert)
	setSettingFromDataInt64(d, rootKey+".wait_for_async_insert_timeout", &result.WaitForAsyncInsertTimeout)
	setSettingFromDataInt64(d, rootKey+".async_insert_max_data_size", &result.AsyncInsertMaxDataSize)
	setSettingFromDataInt64(d, rootKey+".async_insert_busy_timeout", &result.AsyncInsertBusyTimeout)
	setSettingFromDataInt64(d, rootKey+".async_insert_stale_timeout", &result.AsyncInsertStaleTimeout)
	setSettingFromDataInt64(d, rootKey+".timeout_before_checking_execution_speed", &result.TimeoutBeforeCheckingExecutionSpeed)
	setSettingFromDataBool(d, rootKey+".cancel_http_readonly_queries_on_client_close", &result.CancelHttpReadonlyQueriesOnClientClose)
	setSettingFromDataBool(d, rootKey+".flatten_nested", &result.FlattenNested)

	if v, ok := d.GetOkExists(rootKey + ".format_regexp"); ok {
		formatRegexp, ok := v.(string)
		if ok && len(formatRegexp) != 0 {
			result.FormatRegexp = formatRegexp
		}
	}

	setSettingFromDataBool(d, rootKey+".format_regexp_skip_unmatched", &result.FormatRegexpSkipUnmatched)
	setSettingFromDataInt64(d, rootKey+".max_http_get_redirects", &result.MaxHttpGetRedirects)

	if v, ok := d.GetOk(rootKey + ".quota_mode"); ok {
		result.QuotaMode = getQuotaModeValue(v.(string))
	}

	setSettingFromDataBool(d, rootKey+".input_format_import_nested_json", &result.InputFormatImportNestedJson)
	setSettingFromDataBool(d, rootKey+".input_format_parallel_parsing", &result.InputFormatParallelParsing)
	setSettingFromDataInt64(d, rootKey+".max_final_threads", &result.MaxFinalThreads)
	setSettingFromDataInt64(d, rootKey+".max_read_buffer_size", &result.MaxReadBufferSize)

	if v, ok := d.GetOk(rootKey + ".local_filesystem_read_method"); ok {
		result.LocalFilesystemReadMethod = getLocalFilesystemReadMethodValue(v.(string))
	}
	if v, ok := d.GetOk(rootKey + ".remote_filesystem_read_method"); ok {
		result.RemoteFilesystemReadMethod = getRemoteFilesystemReadMethodValue(v.(string))
	}

	setSettingFromDataInt64(d, rootKey+".insert_keeper_max_retries", &result.InsertKeeperMaxRetries)
	setSettingFromDataInt64(d, rootKey+".max_temporary_data_on_disk_size_for_user", &result.MaxTemporaryDataOnDiskSizeForUser)
	setSettingFromDataInt64(d, rootKey+".max_temporary_data_on_disk_size_for_query", &result.MaxTemporaryDataOnDiskSizeForQuery)
	setSettingFromDataInt64(d, rootKey+".max_parser_depth", &result.MaxParserDepth)
	setSettingFromDataInt64(d, rootKey+".memory_overcommit_ratio_denominator", &result.MemoryOvercommitRatioDenominator)
	setSettingFromDataInt64(d, rootKey+".memory_overcommit_ratio_denominator_for_user", &result.MemoryOvercommitRatioDenominatorForUser)
	setSettingFromDataInt64(d, rootKey+".memory_usage_overcommit_max_wait_microseconds", &result.MemoryUsageOvercommitMaxWaitMicroseconds)
	setSettingFromDataBool(d, rootKey+".log_query_threads", &result.LogQueryThreads)
	setSettingFromDataInt64(d, rootKey+".max_insert_threads", &result.MaxInsertThreads)
	setSettingFromDataBool(d, rootKey+".use_hedged_requests", &result.UseHedgedRequests)
	setSettingFromDataInt64(d, rootKey+".idle_connection_timeout", &result.IdleConnectionTimeout)
	setSettingFromDataInt64(d, rootKey+".hedged_connection_timeout_ms", &result.HedgedConnectionTimeoutMs)

	if v, ok := d.GetOk(rootKey + ".load_balancing"); ok {
		result.LoadBalancing = getLoadBalancingValue(v.(string))
	}

	setSettingFromDataBool(d, rootKey+".prefer_localhost_replica", &result.PreferLocalhostReplica)

	if v, ok := d.GetOk(rootKey + ".date_time_input_format"); ok {
		result.DateTimeInputFormat = getDateTimeInputFormatValue(v.(string))
	}

	if v, ok := d.GetOk(rootKey + ".date_time_output_format"); ok {
		result.DateTimeOutputFormat = getDateTimeOutputFormatValue(v.(string))
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
	result["insert_quorum_parallel"] = falseOnNil(settings.InsertQuorumParallel)
	result["select_sequential_consistency"] = falseOnNil(settings.SelectSequentialConsistency)
	result["deduplicate_blocks_in_dependent_materialized_views"] = falseOnNil(settings.DeduplicateBlocksInDependentMaterializedViews)
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
	result["join_algorithm"] = flattenClickHouseUserSettingsJoinAlgorithm(settings.JoinAlgorithm)
	result["any_join_distinct_right_table_keys"] = falseOnNil(settings.AnyJoinDistinctRightTableKeys)
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
	result["input_format_null_as_default"] = falseOnNil(settings.InputFormatNullAsDefault)
	result["input_format_with_names_use_header"] = falseOnNil(settings.InputFormatWithNamesUseHeader)
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
	if settings.MaxConcurrentQueriesForUser != nil {
		result["max_concurrent_queries_for_user"] = settings.MaxConcurrentQueriesForUser.Value
	}
	if settings.MemoryProfilerStep != nil {
		result["memory_profiler_step"] = settings.MemoryProfilerStep.Value
	}
	if settings.MemoryProfilerSampleProbability != nil {
		result["memory_profiler_sample_probability"] = settings.MemoryProfilerSampleProbability.Value
	}
	result["insert_null_as_default"] = falseOnNil(settings.InsertNullAsDefault)
	result["allow_suspicious_low_cardinality_types"] = falseOnNil(settings.AllowSuspiciousLowCardinalityTypes)
	if settings.ConnectTimeoutWithFailover != nil {
		result["connect_timeout_with_failover"] = settings.ConnectTimeoutWithFailover.Value
	}
	result["allow_introspection_functions"] = falseOnNil(settings.AllowIntrospectionFunctions)
	result["async_insert"] = falseOnNil(settings.AsyncInsert)
	if settings.AsyncInsertThreads != nil {
		result["async_insert_threads"] = settings.AsyncInsertThreads.Value
	}
	result["wait_for_async_insert"] = falseOnNil(settings.WaitForAsyncInsert)
	if settings.WaitForAsyncInsertTimeout != nil {
		result["wait_for_async_insert_timeout"] = settings.WaitForAsyncInsertTimeout.Value
	}
	if settings.AsyncInsertMaxDataSize != nil {
		result["async_insert_max_data_size"] = settings.AsyncInsertMaxDataSize.Value
	}
	if settings.AsyncInsertBusyTimeout != nil {
		result["async_insert_busy_timeout"] = settings.AsyncInsertBusyTimeout.Value
	}
	if settings.AsyncInsertStaleTimeout != nil {
		result["async_insert_stale_timeout"] = settings.AsyncInsertStaleTimeout.Value
	}
	if settings.TimeoutBeforeCheckingExecutionSpeed != nil {
		result["timeout_before_checking_execution_speed"] = settings.TimeoutBeforeCheckingExecutionSpeed.Value
	}
	result["cancel_http_readonly_queries_on_client_close"] = falseOnNil(settings.CancelHttpReadonlyQueriesOnClientClose)
	result["flatten_nested"] = falseOnNil(settings.FlattenNested)
	if len(settings.FormatRegexp) != 0 {
		result["format_regexp"] = settings.FormatRegexp
	}
	result["format_regexp_skip_unmatched"] = falseOnNil(settings.FormatRegexpSkipUnmatched)
	if settings.MaxHttpGetRedirects != nil {
		result["max_http_get_redirects"] = settings.MaxHttpGetRedirects.Value
	}

	result["quota_mode"] = getQuotaModeName(settings.QuotaMode)
	if settings.MaxFinalThreads != nil {
		result["max_final_threads"] = settings.MaxFinalThreads.Value
	}
	if settings.MaxReadBufferSize != nil {
		result["max_read_buffer_size"] = settings.MaxReadBufferSize.Value
	}
	result["input_format_import_nested_json"] = falseOnNil(settings.InputFormatImportNestedJson)
	result["input_format_parallel_parsing"] = falseOnNil(settings.InputFormatParallelParsing)

	result["local_filesystem_read_method"] = getLocalFilesystemReadMethodName(settings.LocalFilesystemReadMethod)
	result["remote_filesystem_read_method"] = getRemoteFilesystemReadMethodName(settings.RemoteFilesystemReadMethod)

	if settings.InsertKeeperMaxRetries != nil {
		result["insert_keeper_max_retries"] = settings.InsertKeeperMaxRetries.Value
	}
	if settings.MaxTemporaryDataOnDiskSizeForUser != nil {
		result["max_temporary_data_on_disk_size_for_user"] = settings.MaxTemporaryDataOnDiskSizeForUser.Value
	}
	if settings.MaxTemporaryDataOnDiskSizeForQuery != nil {
		result["max_temporary_data_on_disk_size_for_query"] = settings.MaxTemporaryDataOnDiskSizeForQuery.Value
	}
	if settings.MaxParserDepth != nil {
		result["max_parser_depth"] = settings.MaxParserDepth.Value
	}
	if settings.MemoryOvercommitRatioDenominator != nil {
		result["memory_overcommit_ratio_denominator"] = settings.MemoryOvercommitRatioDenominator.Value
	}
	if settings.MemoryOvercommitRatioDenominatorForUser != nil {
		result["memory_overcommit_ratio_denominator_for_user"] = settings.MemoryOvercommitRatioDenominatorForUser.Value
	}
	if settings.MemoryUsageOvercommitMaxWaitMicroseconds != nil {
		result["memory_usage_overcommit_max_wait_microseconds"] = settings.MemoryUsageOvercommitMaxWaitMicroseconds.Value
	}

	result["log_query_threads"] = falseOnNil(settings.LogQueryThreads)

	if settings.MaxInsertThreads != nil {
		result["max_insert_threads"] = settings.MaxInsertThreads.Value
	}

	result["use_hedged_requests"] = falseOnNil(settings.UseHedgedRequests)

	if settings.IdleConnectionTimeout != nil {
		result["idle_connection_timeout"] = settings.IdleConnectionTimeout.Value
	}
	if settings.HedgedConnectionTimeoutMs != nil {
		result["hedged_connection_timeout_ms"] = settings.HedgedConnectionTimeoutMs.Value
	}

	result["load_balancing"] = getLoadBalancingName(settings.LoadBalancing)

	result["prefer_localhost_replica"] = falseOnNil(settings.PreferLocalhostReplica)

	result["date_time_input_format"] = getDateTimeInputFormatName(settings.DateTimeInputFormat)

	result["date_time_output_format"] = getDateTimeOutputFormatName(settings.DateTimeOutputFormat)

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

func expandClickhouseShard(s map[string]interface{}, _ *schema.ResourceData, hash int) *clickhouse.ShardConfigSpec {
	shardFromSpec := &clickhouse.ShardConfigSpec{
		Clickhouse: &clickhouse.ShardConfigSpec_Clickhouse{},
	}

	if v, ok := s["weight"]; ok {
		shardFromSpec.Clickhouse.Weight = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	shardFromSpec.Clickhouse.Resources = expandClickhouseShardResources(s)

	return shardFromSpec
}

func expandClickhouseShardSpecs(d *schema.ResourceData) (map[string]*clickhouse.ShardConfigSpec, error) {
	rawShardsFromSpec := d.Get("shard").(*schema.Set)
	return expandClickhouseShardSpecsFromSchema(rawShardsFromSpec)
}

func expandClickhouseShardSpecsFromSchema(rawShardsFromSpec *schema.Set) (map[string]*clickhouse.ShardConfigSpec, error) {
	resultShardsFromSpec := map[string]*clickhouse.ShardConfigSpec{}

	log.Printf("[DEBUG] shards config from spec = %v\n", rawShardsFromSpec.List())
	for _, shard := range rawShardsFromSpec.List() {
		m := shard.(map[string]interface{})
		hash := clickHouseShardHash(shard)
		if v, ok := m["name"]; ok {
			resultShardsFromSpec[v.(string)] = expandClickhouseShard(m, nil, hash)
		}
	}
	return resultShardsFromSpec, nil
}

func flattenClickHouseShards(shards []*clickhouse.Shard) ([]map[string]interface{}, error) {
	var res []map[string]interface{}

	for _, shard := range shards {
		m := map[string]interface{}{}
		m["name"] = shard.Name
		if shard.Config.Clickhouse.Weight != nil {
			m["weight"] = shard.Config.Clickhouse.Weight.Value
		}

		if shard.Config.Clickhouse.Resources != nil {
			resources, err := flattenClickHouseResources(shard.Config.Clickhouse.Resources)
			if err != nil {
				return nil, fmt.Errorf("failed parse shard resources from cluster")
			}
			m["resources"] = resources
			log.Printf("[DEBUG] read shard from cluster: shard=%s, resources=%v\n", shard.Name, resources)
		}

		res = append(res, m)
	}

	return res, nil
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

func expandClickHouseCloudStorage(d *schema.ResourceData) (*clickhouse.CloudStorage, error) {
	result := &clickhouse.CloudStorage{}
	cloudStorage := d.Get("cloud_storage").([]interface{})

	for _, g := range cloudStorage {
		cloudStorageSpec := g.(map[string]interface{})
		if val, ok := cloudStorageSpec["enabled"]; ok {
			result.SetEnabled(val.(bool))
			if result.GetEnabled() {
				if moveFactor, ok := cloudStorageSpec["move_factor"]; ok {
					result.SetMoveFactor(&wrapperspb.DoubleValue{Value: moveFactor.(float64)})
				}

				var dataCacheEnabled *wrapperspb.BoolValue
				if cacheEnabled, ok := cloudStorageSpec["data_cache_enabled"]; ok {
					dataCacheEnabled = &wrapperspb.BoolValue{Value: cacheEnabled.(bool)}
				}
				var dataCacheMaxSize *wrapperspb.Int64Value
				if data, ok := cloudStorageSpec["data_cache_max_size"]; ok {
					cacheMaxSize := int64(data.(int))
					if cacheMaxSize > 0 {
						dataCacheMaxSize = &wrapperspb.Int64Value{Value: cacheMaxSize}
					}
				}
				if dataCacheMaxSize != nil && (dataCacheEnabled == nil || !dataCacheEnabled.Value) {
					return nil, fmt.Errorf("setting data_cache_enabled should be enabled to use data_cache_max_size")
				}
				result.SetDataCacheEnabled(dataCacheEnabled)
				result.SetDataCacheMaxSize(dataCacheMaxSize)

				if preferNotToMerge, ok := cloudStorageSpec["prefer_not_to_merge"]; ok {
					result.SetPreferNotToMerge(&wrapperspb.BoolValue{Value: preferNotToMerge.(bool)})
				}
			}
		}
	}

	return result, nil
}

func flattenClickHouseCloudStorage(cs *clickhouse.CloudStorage) []map[string]interface{} {
	var result []map[string]interface{}

	m := map[string]interface{}{}
	if cs != nil {
		m["enabled"] = cs.GetEnabled()
		if cs.GetMoveFactor() != nil {
			m["move_factor"] = cs.GetMoveFactor().Value
		}
		if cs.GetDataCacheEnabled() != nil {
			if cs.GetDataCacheEnabled().Value {
				m["data_cache_enabled"] = cs.GetDataCacheEnabled().Value
				if cs.GetDataCacheMaxSize() != nil {
					m["data_cache_max_size"] = cs.GetDataCacheMaxSize().Value
				}
			}
		}
		if cs.GetPreferNotToMerge() != nil {
			m["prefer_not_to_merge"] = cs.GetPreferNotToMerge().Value
		}
	}

	result = append(result, m)

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

func expandClickhouseShardResources(s map[string]interface{}) *clickhouse.Resources {
	valueResources, ok := s["resources"]
	if !ok {
		log.Println("[TRACE] shard has empty resource.")
		return nil
	}
	log.Printf("[DEBUG] parse shard resources: %v\n", valueResources)

	resourcesShardSpec := valueResources.([]interface{})

	if len(resourcesShardSpec) == 0 {
		log.Println("[DEBUG] shard has resource but it is empty.")
		return nil
	}
	hasResource := false
	resources := clickhouse.Resources{}

	resourceSpec := resourcesShardSpec[0].(map[string]interface{})

	if diskSize, ok := resourceSpec["disk_size"]; ok {
		log.Printf("[DEBUG] shard has resource: disk_size=%d\n", resources.GetDiskSize())
		hasResource = true
		resources.SetDiskSize(toBytes(diskSize.(int)))
	}
	if resourcePresetId, ok := resourceSpec["resource_preset_id"]; ok {
		log.Printf("[DEBUG] shard has resource: resource_preset_id=%s\n", resources.GetResourcePresetId())
		hasResource = true
		resources.SetResourcePresetId(resourcePresetId.(string))
	}
	if diskTypeId, ok := resourceSpec["disk_type_id"]; ok {
		log.Printf("[DEBUG] shard has resource: disk_type_id=%s\n", resources.GetDiskTypeId())
		hasResource = true
		resources.SetDiskTypeId(diskTypeId.(string))
	}

	if hasResource {
		return &resources
	}
	return nil
}
