package yandex

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
)

type redisConfig struct {
	timeout                         int64
	maxmemoryPolicy                 string
	notifyKeyspaceEvents            string
	slowlogLogSlowerThan            int64
	slowlogMaxLen                   int64
	databases                       int64
	version                         string
	clientOutputBufferLimitNormal   string
	clientOutputBufferLimitPubsub   string
	maxmemoryPercent                int64
	luaTimeLimit                    int64
	replBacklogSizePercent          int64
	clusterRequireFullCoverage      bool
	clusterAllowReadsWhenDown       bool
	clusterAllowPubsubshardWhenDown bool
	lfuDecayTime                    int64
	lfuLogFactor                    int64
	turnBeforeSwitchover            bool
	allowDataLoss                   bool
	useLuajit                       bool
	ioThreadsAllowed                bool
}

const defaultReplicaPriority = 100

func weightFunc(zone, shard, subnet string, priority *wrappers.Int64Value, ipFlag bool) int {
	weight := 0
	if zone != "" {
		weight += 10000
	}
	if shard != "" {
		weight += 1000
	}
	if subnet != "" {
		weight += 100
	}
	if priority != nil {
		weight += 10
	}
	if ipFlag {
		weight += 1
	}
	return weight
}

func getHostWeight(spec *redis.Host) int {
	return weightFunc(spec.ZoneId, spec.ShardName, spec.SubnetId, spec.ReplicaPriority, spec.AssignPublicIp)
}

// Sorts list of hosts in accordance with the order in config.
// We need to keep the original order so there's no diff appears on each apply.
func sortRedisHosts(sharded bool, hosts []*redis.Host, specs []*redis.HostSpec) {
	for i, hs := range specs {
		switched := false
		for j := i; j < len(hosts); j++ {
			if (hs.ZoneId == hosts[j].ZoneId) &&
				(hs.ShardName == "" || hs.ShardName == hosts[j].ShardName) &&
				(hs.SubnetId == "" || hs.SubnetId == hosts[j].SubnetId) &&
				(sharded || hosts[j].ReplicaPriority != nil && (hs.ReplicaPriority == nil && hosts[j].ReplicaPriority.GetValue() == defaultReplicaPriority ||
					hs.ReplicaPriority.GetValue() == hosts[j].ReplicaPriority.GetValue())) &&
				(hs.AssignPublicIp == hosts[j].AssignPublicIp) {
				if !switched || getHostWeight(hosts[j]) > getHostWeight(hosts[i]) {
					hosts[i], hosts[j] = hosts[j], hosts[i]
					switched = true
				}
			}
		}
	}
}

func keyFunc(zone, shard, subnet string) string {
	return fmt.Sprintf("zone:%s;shard:%s;subnet:%s",
		zone, shard, subnet,
	)
}

func getHostSpecBaseKey(h *redis.HostSpec) string {
	return keyFunc(h.ZoneId, h.ShardName, h.SubnetId)
}

func getHostBaseKey(h *redis.Host) string {
	return keyFunc(h.ZoneId, h.ShardName, h.SubnetId)
}

func getHostSpecWeight(spec *redis.HostSpec) int {
	return weightFunc(spec.ZoneId, spec.ShardName, spec.SubnetId, spec.ReplicaPriority, spec.AssignPublicIp)
}

// used to detect specs to update, add, delete
func sortHostSpecs(targetHosts []*redis.HostSpec) []*redis.HostSpec {
	weightedHosts := make(map[int][]*redis.HostSpec)
	for _, spec := range targetHosts {
		weight := getHostSpecWeight(spec)
		weightedHosts[weight] = append(weightedHosts[weight], spec)
	}

	keys := make([]int, 0, len(weightedHosts))
	for k := range weightedHosts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	res := []*redis.HostSpec{}
	for _, k := range keys {
		res = append(res, weightedHosts[k]...)
	}

	return res
}

func separateHostsToUpdateAndDelete(sharded bool, sortedHosts []*redis.HostSpec, currHosts []*redis.Host) (
	[]*redis.HostSpec, map[string][]*HostUpdateInfo, map[string][]string, error) {
	targetHostsBaseMap := map[string][]*redis.HostSpec{}
	for _, h := range sortedHosts {
		key := getHostSpecBaseKey(h)
		targetHostsBaseMap[key] = append(targetHostsBaseMap[key], h)
	}

	toDelete := map[string][]string{}
	toUpdate := map[string][]*HostUpdateInfo{}
	for _, h := range currHosts {
		key := getHostBaseKey(h)
		hs, ok := targetHostsBaseMap[key]
		if ok {
			newSpec := hs[0]
			hostInfo, err := getHostUpdateInfo(sharded, h.Name, h.ReplicaPriority, h.AssignPublicIp,
				newSpec.ReplicaPriority, newSpec.AssignPublicIp)
			if err != nil {
				return nil, nil, nil, err
			}
			if hostInfo != nil {
				toUpdate[h.ShardName] = append(toUpdate[h.ShardName], hostInfo)
			}
			if len(hs) > 1 {
				targetHostsBaseMap[key] = hs[1:]
			} else {
				delete(targetHostsBaseMap, key)
			}
		} else {
			toDelete[h.ShardName] = append(toDelete[h.ShardName], h.Name)
		}
	}

	hostsLeft := []*redis.HostSpec{}
	for _, specs := range targetHostsBaseMap {
		hostsLeft = append(hostsLeft, specs...)
	}

	return hostsLeft, toUpdate, toDelete, nil
}

// Takes the current list of hosts and the desirable list of hosts.
// Returns the map of hostnames:
// to delete grouped by shard,
// to update grouped by shard,
// to add grouped by shard as well.
func redisHostsDiff(sharded bool, currHosts []*redis.Host, targetHosts []*redis.HostSpec) (map[string][]string,
	map[string][]*HostUpdateInfo, map[string][]*redis.HostSpec, error) {
	sortedHosts := sortHostSpecs(targetHosts)
	hostsLeft, toUpdate, toDelete, err := separateHostsToUpdateAndDelete(sharded, sortedHosts, currHosts)
	if err != nil {
		return nil, nil, nil, err
	}

	toAdd := map[string][]*redis.HostSpec{}
	for _, h := range hostsLeft {
		toAdd[h.ShardName] = append(toAdd[h.ShardName], h)
	}

	return toDelete, toUpdate, toAdd, nil
}

func limitToStr(hard, soft, secs *wrappers.Int64Value) string {
	vals := []string{
		strconv.FormatInt(hard.GetValue(), 10),
		strconv.FormatInt(soft.GetValue(), 10),
		strconv.FormatInt(secs.GetValue(), 10),
	}
	return strings.Join(vals, " ")
}

func extractRedisConfig(cc *redis.ClusterConfig) redisConfig {
	res := redisConfig{
		version: cc.Version,
	}
	c := cc.Redis.EffectiveConfig
	res.maxmemoryPolicy = c.GetMaxmemoryPolicy().String()
	res.timeout = c.GetTimeout().GetValue()
	res.notifyKeyspaceEvents = c.GetNotifyKeyspaceEvents()
	res.slowlogLogSlowerThan = c.GetSlowlogLogSlowerThan().GetValue()
	res.slowlogMaxLen = c.GetSlowlogMaxLen().GetValue()
	res.databases = c.GetDatabases().GetValue()
	res.maxmemoryPercent = c.GetMaxmemoryPercent().GetValue()
	res.clientOutputBufferLimitNormal = limitToStr(
		c.ClientOutputBufferLimitNormal.HardLimit,
		c.ClientOutputBufferLimitNormal.SoftLimit,
		c.ClientOutputBufferLimitNormal.SoftSeconds,
	)
	res.clientOutputBufferLimitPubsub = limitToStr(
		c.ClientOutputBufferLimitPubsub.HardLimit,
		c.ClientOutputBufferLimitPubsub.SoftLimit,
		c.ClientOutputBufferLimitPubsub.SoftSeconds,
	)
	res.luaTimeLimit = c.GetLuaTimeLimit().GetValue()
	res.replBacklogSizePercent = c.GetReplBacklogSizePercent().GetValue()
	res.clusterRequireFullCoverage = c.GetClusterRequireFullCoverage().GetValue()
	res.clusterAllowReadsWhenDown = c.GetClusterAllowReadsWhenDown().GetValue()
	res.clusterAllowPubsubshardWhenDown = c.GetClusterAllowPubsubshardWhenDown().GetValue()
	res.lfuDecayTime = c.GetLfuDecayTime().GetValue()
	res.lfuLogFactor = c.GetLfuLogFactor().GetValue()
	res.turnBeforeSwitchover = c.GetTurnBeforeSwitchover().GetValue()
	res.allowDataLoss = c.GetAllowDataLoss().GetValue()
	res.useLuajit = c.GetUseLuajit().GetValue()
	res.ioThreadsAllowed = c.GetIoThreadsAllowed().GetValue()
	return res
}

func expandLimit(limit string) ([]*wrappers.Int64Value, error) {
	vals := strings.Split(limit, " ")
	if len(vals) != 3 {
		return nil, fmt.Errorf("%s should be space-separated 3-values string", limit)
	}
	res := []*wrappers.Int64Value{}
	for _, val := range vals {
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		res = append(res, &wrappers.Int64Value{Value: parsed})
	}
	return res, nil
}

func expandRedisConfig(d *schema.ResourceData) (*config.RedisConfig, string, error) {
	var password string
	if v, ok := d.GetOk("config.0.password"); ok {
		password = v.(string)
	}

	var timeout *wrappers.Int64Value
	if v, ok := d.GetOkExists("config.0.timeout"); ok {
		timeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var notifyKeyspaceEvents string
	if v, ok := d.GetOk("config.0.notify_keyspace_events"); ok {
		notifyKeyspaceEvents = v.(string)
	}

	var slowlogLogSlowerThan *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.slowlog_log_slower_than"); ok {
		slowlogLogSlowerThan = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var slowlogMaxLen *wrappers.Int64Value
	if v, ok := d.GetOkExists("config.0.slowlog_max_len"); ok {
		slowlogMaxLen = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var databases *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.databases"); ok {
		databases = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var maxmemoryPercent *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.maxmemory_percent"); ok {
		maxmemoryPercent = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var version string
	if v, ok := d.GetOk("config.0.version"); ok {
		version = v.(string)
	}

	var luaTimeLimit *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.lua_time_limit"); ok {
		luaTimeLimit = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var replBacklogSizePercent *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.repl_backlog_size_percent"); ok {
		replBacklogSizePercent = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var clusterRequireFullCoverage *wrappers.BoolValue
	if v := d.Get("config.0.cluster_require_full_coverage"); v != nil {
		clusterRequireFullCoverage = wrapperspb.Bool(v.(bool))
	}

	var clusterAllowReadsWhenDown *wrappers.BoolValue
	if v := d.Get("config.0.cluster_allow_reads_when_down"); v != nil {
		clusterAllowReadsWhenDown = wrapperspb.Bool(v.(bool))
	}

	var clusterAllowPubsubshardWhenDown *wrappers.BoolValue
	if v := d.Get("config.0.cluster_allow_pubsubshard_when_down"); v != nil {
		clusterAllowPubsubshardWhenDown = wrapperspb.Bool(v.(bool))
	}

	var lfuDecayTime *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.lfu_decay_time"); ok {
		lfuDecayTime = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var lfuLogFactor *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.lfu_log_factor"); ok {
		lfuLogFactor = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var turnBeforeSwitchover *wrappers.BoolValue
	if v := d.Get("config.0.turn_before_switchover"); v != nil {
		turnBeforeSwitchover = wrapperspb.Bool(v.(bool))
	}

	var allowDataLoss *wrappers.BoolValue
	if v := d.Get("config.0.allow_data_loss"); v != nil {
		allowDataLoss = wrapperspb.Bool(v.(bool))
	}

	var err error
	var expandedNormal []*wrappers.Int64Value
	if v, ok := d.GetOk("config.0.client_output_buffer_limit_normal"); ok {
		expandedNormal, err = expandLimit(v.(string))
		if err != nil {
			return nil, "", err
		}
	}
	var expandedPubsub []*wrappers.Int64Value
	if v, ok := d.GetOk("config.0.client_output_buffer_limit_pubsub"); ok {
		expandedPubsub, err = expandLimit(v.(string))
		if err != nil {
			return nil, "", err
		}
	}

	var useLuajit *wrappers.BoolValue
	if v := d.Get("config.0.use_luajit"); v != nil {
		useLuajit = &wrappers.BoolValue{Value: v.(bool)}
	}

	var ioThreadsAllowed *wrappers.BoolValue
	if v := d.Get("config.0.io_threads_allowed"); v != nil {
		ioThreadsAllowed = &wrappers.BoolValue{Value: v.(bool)}
	}

	c := config.RedisConfig{
		Password:                        password,
		Timeout:                         timeout,
		NotifyKeyspaceEvents:            notifyKeyspaceEvents,
		SlowlogLogSlowerThan:            slowlogLogSlowerThan,
		SlowlogMaxLen:                   slowlogMaxLen,
		Databases:                       databases,
		MaxmemoryPercent:                maxmemoryPercent,
		LuaTimeLimit:                    luaTimeLimit,
		ReplBacklogSizePercent:          replBacklogSizePercent,
		ClusterRequireFullCoverage:      clusterRequireFullCoverage,
		ClusterAllowReadsWhenDown:       clusterAllowReadsWhenDown,
		ClusterAllowPubsubshardWhenDown: clusterAllowPubsubshardWhenDown,
		LfuDecayTime:                    lfuDecayTime,
		LfuLogFactor:                    lfuLogFactor,
		TurnBeforeSwitchover:            turnBeforeSwitchover,
		AllowDataLoss:                   allowDataLoss,
		UseLuajit:                       useLuajit,
		IoThreadsAllowed:                ioThreadsAllowed,
	}

	if len(expandedNormal) != 0 {
		normalLimit := &config.RedisConfig_ClientOutputBufferLimit{
			HardLimit:   expandedNormal[0],
			SoftLimit:   expandedNormal[1],
			SoftSeconds: expandedNormal[2],
		}
		c.SetClientOutputBufferLimitNormal(normalLimit)
	}

	if len(expandedPubsub) != 0 {
		pubsubLimit := &config.RedisConfig_ClientOutputBufferLimit{
			HardLimit:   expandedPubsub[0],
			SoftLimit:   expandedPubsub[1],
			SoftSeconds: expandedPubsub[2],
		}
		c.SetClientOutputBufferLimitPubsub(pubsubLimit)
	}

	if err := setMaxMemory(&c, d); err != nil {
		return nil, version, err
	}

	return &c, version, nil
}

func setMaxMemory(c *config.RedisConfig, d *schema.ResourceData) error {
	if v, ok := d.GetOk("config.0.maxmemory_policy"); ok {
		mp, err := parseRedisMaxmemoryPolicy(v.(string))
		if err != nil {
			return err
		}
		c.MaxmemoryPolicy = mp
	}
	return nil
}

func flattenRedisResources(r *redis.Resources) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_size"] = toGigabytes(r.DiskSize)
	res["disk_type_id"] = r.DiskTypeId

	return []map[string]interface{}{res}, nil
}

func flattenRedisDiskSizeAutoscaling(r *redis.DiskSizeAutoscaling) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["disk_size_limit"] = toGigabytes(r.GetDiskSizeLimit().GetValue())
	res["planned_usage_threshold"] = int(r.GetPlannedUsageThreshold().GetValue())
	res["emergency_usage_threshold"] = int(r.GetEmergencyUsageThreshold().GetValue())

	return []map[string]interface{}{res}, nil
}

func expandRedisResources(d *schema.ResourceData) (*redis.Resources, error) {
	rs := &redis.Resources{}

	if v, ok := d.GetOk("resources.0.resource_preset_id"); ok {
		rs.ResourcePresetId = v.(string)
	}

	if v, ok := d.GetOk("resources.0.disk_size"); ok {
		rs.DiskSize = toBytes(v.(int))
	}

	if v, ok := d.GetOk("resources.0.disk_type_id"); ok {
		rs.DiskTypeId = v.(string)
	}

	return rs, nil
}

func expandRedisDiskSizeAutoscaling(d *schema.ResourceData) (*redis.DiskSizeAutoscaling, error) {
	dsa := &redis.DiskSizeAutoscaling{}

	if v := d.Get("disk_size_autoscaling.0.disk_size_limit"); v != nil {
		dsa.DiskSizeLimit = &wrappers.Int64Value{Value: toBytes(v.(int))}
	}

	if v, ok := d.GetOk("disk_size_autoscaling.0.planned_usage_threshold"); ok {
		dsa.PlannedUsageThreshold = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOk("disk_size_autoscaling.0.emergency_usage_threshold"); ok {
		dsa.EmergencyUsageThreshold = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return dsa, nil
}

func parseRedisWeekDay(wd string) (redis.WeeklyMaintenanceWindow_WeekDay, error) {
	val, ok := redis.WeeklyMaintenanceWindow_WeekDay_value[wd]
	// do not allow WEEK_DAY_UNSPECIFIED
	if !ok || val == 0 {
		return redis.WeeklyMaintenanceWindow_WEEK_DAY_UNSPECIFIED,
			fmt.Errorf("value for 'day' should be one of %s, not `%s`",
				getJoinedKeys(getEnumValueMapKeysExt(redis.WeeklyMaintenanceWindow_WeekDay_value, true)), wd)
	}

	return redis.WeeklyMaintenanceWindow_WeekDay(val), nil
}

func expandRedisMaintenanceWindow(d *schema.ResourceData) (*redis.MaintenanceWindow, error) {
	mwType, ok := d.GetOk("maintenance_window.0.type")
	if !ok {
		return nil, nil
	}

	result := &redis.MaintenanceWindow{}

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
		result.SetAnytime(&redis.AnytimeMaintenanceWindow{})

	case "WEEKLY":
		weekly := &redis.WeeklyMaintenanceWindow{}
		if val, ok := d.GetOk("maintenance_window.0.day"); ok {
			var err error
			weekly.Day, err = parseRedisWeekDay(val.(string))
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

func flattenRedisMaintenanceWindow(mw *redis.MaintenanceWindow) []map[string]interface{} {
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

func flattenRedisHosts(sharded bool, hs []*redis.Host) ([]map[string]interface{}, error) {
	res := []map[string]interface{}{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["zone"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["shard_name"] = h.ShardName
		m["fqdn"] = h.Name
		if sharded {
			m["replica_priority"] = defaultReplicaPriority
		} else {
			m["replica_priority"] = h.ReplicaPriority.GetValue()
		}
		m["assign_public_ip"] = h.AssignPublicIp
		res = append(res, m)
	}

	return res, nil
}

func expandRedisHosts(d *schema.ResourceData) ([]*redis.HostSpec, error) {
	var result []*redis.HostSpec
	hosts := d.Get("host").([]interface{})
	sharded := d.Get("sharded").(bool)

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host := expandRedisHost(sharded, config)
		result = append(result, host)
	}

	return result, nil
}

func expandRedisHost(sharded bool, config map[string]interface{}) *redis.HostSpec {
	host := &redis.HostSpec{}
	if v, ok := config["zone"]; ok {
		host.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		host.SubnetId = v.(string)
	}

	if v, ok := config["shard_name"]; ok {
		host.ShardName = v.(string)
	}

	if v, ok := config["replica_priority"]; ok && !sharded {
		priority := v.(int)
		host.ReplicaPriority = &wrappers.Int64Value{Value: int64(priority)}
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.AssignPublicIp = v.(bool)
	}
	return host
}

func parseRedisEnv(e string) (redis.Cluster_Environment, error) {
	v, ok := redis.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(redis.Cluster_Environment_value)), e)
	}
	return redis.Cluster_Environment(v), nil
}

func parsePersistenceMode(p interface{}) (redis.Cluster_PersistenceMode, error) {
	e := p.(string)
	if e == "" {
		return redis.Cluster_ON, nil
	}

	v, ok := redis.Cluster_PersistenceMode_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'persistence_mode' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(redis.Cluster_PersistenceMode_value)), e)
	}
	return redis.Cluster_PersistenceMode(v), nil
}

func parseRedisMaxmemoryPolicy(s string) (config.RedisConfig_MaxmemoryPolicy, error) {
	v, ok := config.RedisConfig_MaxmemoryPolicy_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'maxmemory_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(config.RedisConfig_MaxmemoryPolicy_value)), s)
	}
	return config.RedisConfig_MaxmemoryPolicy(v), nil
}

func flattenRedisAccess(ac *redis.Access) []map[string]interface{} {
	res := map[string]interface{}{}
	if ac != nil {
		res["web_sql"] = ac.WebSql
		res["data_lens"] = ac.DataLens
	}
	return []map[string]interface{}{res}
}

func expandRedisAccess(d *schema.ResourceData) *redis.Access {
	result := &redis.Access{}

	if v, ok := d.GetOk("access.0.web_sql"); ok {
		result.WebSql = v.(bool)
	}
	if v, ok := d.GetOk("access.0.data_lens"); ok {
		result.DataLens = v.(bool)
	}

	return result
}
