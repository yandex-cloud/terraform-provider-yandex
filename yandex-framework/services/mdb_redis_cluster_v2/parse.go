package mdb_redis_cluster_v2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
)

func getEnumValueMapKeys(m map[string]int32) []string {
	return getEnumValueMapKeysExt(m, true)
}

func getEnumValueMapKeysExt(m map[string]int32, skipDefault bool) []string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if v == 0 && skipDefault {
			continue
		}

		keys = append(keys, k)
	}
	return keys
}

func getJoinedKeys(keys []string) string {
	return "`" + strings.Join(keys, "`, `") + "`"
}

func parseRedisMaxmemoryPolicy(s string) (config.RedisConfig_MaxmemoryPolicy, error) {
	v, ok := config.RedisConfig_MaxmemoryPolicy_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'maxmemory_policy' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(config.RedisConfig_MaxmemoryPolicy_value)), s)
	}
	return config.RedisConfig_MaxmemoryPolicy(v), nil
}

func limitToStr(hard, soft, secs *wrappers.Int64Value) string {
	if hard == nil && soft == nil && secs == nil {
		return ""
	}
	vals := []string{
		strconv.FormatInt(hard.GetValue(), 10),
		strconv.FormatInt(soft.GetValue(), 10),
		strconv.FormatInt(secs.GetValue(), 10),
	}
	return strings.Join(vals, " ")
}

func parseRedisEnv(e string) (redis.Cluster_Environment, error) {
	v, ok := redis.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(redis.Cluster_Environment_value)), e)
	}
	return redis.Cluster_Environment(v), nil
}

func parsePersistenceMode(e string) (redis.Cluster_PersistenceMode, error) {
	if e == "" {
		return redis.Cluster_ON, nil
	}

	v, ok := redis.Cluster_PersistenceMode_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'persistence_mode' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeysExt(redis.Cluster_PersistenceMode_value, false)), e)
	}
	return redis.Cluster_PersistenceMode(v), nil
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
