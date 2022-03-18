package yandex

import (
	"fmt"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
)

type redisConfig struct {
	timeout              int64
	maxmemoryPolicy      string
	notifyKeyspaceEvents string
	slowlogLogSlowerThan int64
	slowlogMaxLen        int64
	databases            int64
	version              string
}

// Sorts list of hosts in accordance with the order in config.
// We need to keep the original order so there's no diff appears on each apply.
func sortRedisHosts(hosts []*redis.Host, specs []*redis.HostSpec) {
	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if h.ZoneId == hosts[j].ZoneId && (h.ShardName == "" || h.ShardName == hosts[j].ShardName) {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}
}

// Takes the current list of hosts and the desirable list of hosts.
// Returns the map of hostnames to delete grouped by shard,
// and the map of hosts to add grouped by shard as well.
func redisHostsDiff(currHosts []*redis.Host, targetHosts []*redis.HostSpec) (map[string][]string, map[string][]*redis.HostSpec) {
	m := map[string][]*redis.HostSpec{}

	for _, h := range targetHosts {
		key := h.ZoneId + h.ShardName
		m[key] = append(m[key], h)
	}

	toDelete := map[string][]string{}
	for _, h := range currHosts {
		key := h.ZoneId + h.ShardName
		hs, ok := m[key]
		if !ok {
			toDelete[h.ShardName] = append(toDelete[h.ShardName], h.Name)
		}
		if len(hs) > 1 {
			m[key] = hs[1:]
		} else {
			delete(m, key)
		}
	}

	toAdd := map[string][]*redis.HostSpec{}
	for _, hs := range m {
		for _, h := range hs {
			toAdd[h.ShardName] = append(toAdd[h.ShardName], h)
		}
	}

	return toDelete, toAdd
}

func extractRedisConfig(cc *redis.ClusterConfig) redisConfig {
	res := redisConfig{
		version: cc.Version,
	}
	switch rc := cc.RedisConfig.(type) {
	case *redis.ClusterConfig_RedisConfig_5_0:
		c := rc.RedisConfig_5_0.EffectiveConfig
		res.maxmemoryPolicy = c.GetMaxmemoryPolicy().String()
		res.timeout = c.GetTimeout().GetValue()
		res.notifyKeyspaceEvents = c.GetNotifyKeyspaceEvents()
		res.slowlogLogSlowerThan = c.GetSlowlogLogSlowerThan().GetValue()
		res.slowlogMaxLen = c.GetSlowlogMaxLen().GetValue()
		res.databases = c.GetDatabases().GetValue()
	case *redis.ClusterConfig_RedisConfig_6_0:
		c := rc.RedisConfig_6_0.EffectiveConfig
		res.maxmemoryPolicy = c.GetMaxmemoryPolicy().String()
		res.timeout = c.GetTimeout().GetValue()
		res.notifyKeyspaceEvents = c.GetNotifyKeyspaceEvents()
		res.slowlogLogSlowerThan = c.GetSlowlogLogSlowerThan().GetValue()
		res.slowlogMaxLen = c.GetSlowlogMaxLen().GetValue()
		res.databases = c.GetDatabases().GetValue()
	case *redis.ClusterConfig_RedisConfig_6_2:
		c := rc.RedisConfig_6_2.EffectiveConfig
		res.maxmemoryPolicy = c.GetMaxmemoryPolicy().String()
		res.timeout = c.GetTimeout().GetValue()
		res.notifyKeyspaceEvents = c.GetNotifyKeyspaceEvents()
		res.slowlogLogSlowerThan = c.GetSlowlogLogSlowerThan().GetValue()
		res.slowlogMaxLen = c.GetSlowlogMaxLen().GetValue()
		res.databases = c.GetDatabases().GetValue()
	}

	return res
}

func expandRedisConfig(d *schema.ResourceData) (*redis.ConfigSpec_RedisSpec, string, error) {
	var cs redis.ConfigSpec_RedisSpec

	var password string
	if v, ok := d.GetOk("config.0.password"); ok {
		password = v.(string)
	}

	var timeout *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.timeout"); ok {
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
	if v, ok := d.GetOk("config.0.slowlog_max_len"); ok {
		slowlogMaxLen = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var databases *wrappers.Int64Value
	if v, ok := d.GetOk("config.0.databases"); ok {
		databases = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	var version string
	if v, ok := d.GetOk("config.0.version"); ok {
		version = v.(string)
	}
	switch version {
	case "5.0":
		c := config.RedisConfig5_0{
			Password:             password,
			Timeout:              timeout,
			NotifyKeyspaceEvents: notifyKeyspaceEvents,
			SlowlogLogSlowerThan: slowlogLogSlowerThan,
			SlowlogMaxLen:        slowlogMaxLen,
			Databases:            databases,
		}
		err := setMaxMemory5_0(&c, d)
		if err != nil {
			return nil, version, err
		}
		cs = &redis.ConfigSpec_RedisConfig_5_0{
			RedisConfig_5_0: &c,
		}
	case "6.0":
		c := config.RedisConfig6_0{
			Password:             password,
			Timeout:              timeout,
			NotifyKeyspaceEvents: notifyKeyspaceEvents,
			SlowlogLogSlowerThan: slowlogLogSlowerThan,
			SlowlogMaxLen:        slowlogMaxLen,
			Databases:            databases,
		}
		err := setMaxMemory6_0(&c, d)
		if err != nil {
			return nil, version, err
		}
		cs = &redis.ConfigSpec_RedisConfig_6_0{
			RedisConfig_6_0: &c,
		}
	case "6.2":
		c := config.RedisConfig6_2{
			Password:             password,
			Timeout:              timeout,
			NotifyKeyspaceEvents: notifyKeyspaceEvents,
			SlowlogLogSlowerThan: slowlogLogSlowerThan,
			SlowlogMaxLen:        slowlogMaxLen,
			Databases:            databases,
		}
		err := setMaxMemory6_2(&c, d)
		if err != nil {
			return nil, version, err
		}
		cs = &redis.ConfigSpec_RedisConfig_6_2{
			RedisConfig_6_2: &c,
		}
	}

	return &cs, version, nil
}

func setMaxMemory5_0(c *config.RedisConfig5_0, d *schema.ResourceData) error {
	if v, ok := d.GetOk("config.0.maxmemory_policy"); ok {
		mp, err := parseRedisMaxmemoryPolicy5_0(v.(string))
		if err != nil {
			return err
		}
		c.MaxmemoryPolicy = mp
	}
	return nil
}

func setMaxMemory6_0(c *config.RedisConfig6_0, d *schema.ResourceData) error {
	if v, ok := d.GetOk("config.0.maxmemory_policy"); ok {
		mp, err := parseRedisMaxmemoryPolicy6_0(v.(string))
		if err != nil {
			return err
		}
		c.MaxmemoryPolicy = mp
	}
	return nil
}

func setMaxMemory6_2(c *config.RedisConfig6_2, d *schema.ResourceData) error {
	if v, ok := d.GetOk("config.0.maxmemory_policy"); ok {
		mp, err := parseRedisMaxmemoryPolicy6_2(v.(string))
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

func flattenRedisHosts(hs []*redis.Host) ([]map[string]interface{}, error) {
	res := []map[string]interface{}{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["zone"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["shard_name"] = h.ShardName
		m["fqdn"] = h.Name
		res = append(res, m)
	}

	return res, nil
}

func expandRedisHosts(d *schema.ResourceData) ([]*redis.HostSpec, error) {
	var result []*redis.HostSpec
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host := expandRedisHost(config)
		result = append(result, host)
	}

	return result, nil
}

func expandRedisHost(config map[string]interface{}) *redis.HostSpec {
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

func parseRedisMaxmemoryPolicy5_0(s string) (config.RedisConfig5_0_MaxmemoryPolicy, error) {
	v, ok := config.RedisConfig5_0_MaxmemoryPolicy_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'maxmemory_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(config.RedisConfig5_0_MaxmemoryPolicy_value)), s)
	}
	return config.RedisConfig5_0_MaxmemoryPolicy(v), nil
}

func parseRedisMaxmemoryPolicy6_0(s string) (config.RedisConfig6_0_MaxmemoryPolicy, error) {
	v, ok := config.RedisConfig6_0_MaxmemoryPolicy_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'maxmemory_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(config.RedisConfig6_0_MaxmemoryPolicy_value)), s)
	}
	return config.RedisConfig6_0_MaxmemoryPolicy(v), nil
}

func parseRedisMaxmemoryPolicy6_2(s string) (config.RedisConfig6_2_MaxmemoryPolicy, error) {
	v, ok := config.RedisConfig6_2_MaxmemoryPolicy_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'maxmemory_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(config.RedisConfig6_2_MaxmemoryPolicy_value)), s)
	}
	return config.RedisConfig6_2_MaxmemoryPolicy(v), nil
}
