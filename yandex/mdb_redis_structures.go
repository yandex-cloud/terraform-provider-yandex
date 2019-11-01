package yandex

import (
	"fmt"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
)

type redisConfig struct {
	timeout         int64
	maxmemoryPolicy string
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
	rc := (cc.RedisConfig).(*redis.ClusterConfig_RedisConfig_5_0)
	c := rc.RedisConfig_5_0.EffectiveConfig

	res := redisConfig{}
	res.timeout = c.GetTimeout().GetValue()
	res.maxmemoryPolicy = c.GetMaxmemoryPolicy().String()
	return res
}

func expandRedisConfig(d *schema.ResourceData) (*redis.ConfigSpec_RedisConfig_5_0, error) {
	cs := &redis.ConfigSpec_RedisConfig_5_0{}
	c := &config.RedisConfig5_0{}

	if v, ok := d.GetOk("config.0.password"); ok {
		c.Password = v.(string)
	}
	if v, ok := d.GetOk("config.0.timeout"); ok {
		c.Timeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk("config.0.maxmemory_policy"); ok {
		mp, err := parseRedisMaxmemoryPolicy(v.(string))
		if err != nil {
			return nil, err
		}
		c.MaxmemoryPolicy = mp
	}
	cs.RedisConfig_5_0 = c

	return cs, nil
}

func flattenRedisResources(r *redis.Resources) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_size"] = toGigabytes(r.DiskSize)

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

	return rs, nil
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

func parseRedisMaxmemoryPolicy(s string) (config.RedisConfig5_0_MaxmemoryPolicy, error) {
	v, ok := config.RedisConfig5_0_MaxmemoryPolicy_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'maxmemory_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(config.RedisConfig5_0_MaxmemoryPolicy_value)), s)
	}
	return config.RedisConfig5_0_MaxmemoryPolicy(v), nil
}
