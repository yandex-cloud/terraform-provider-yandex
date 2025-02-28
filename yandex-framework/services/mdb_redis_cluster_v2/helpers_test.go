package mdb_redis_cluster_v2_test

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_redis_cluster_v2"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const redisResource = "yandex_mdb_redis_cluster_v2.bar"
const defaultTestMDBPageSize = 1000

type Op string

const (
	OpCreate      Op = "Create Valkey cluster"
	OpModify      Op = "Modify Valkey cluster"
	OpRebalance   Op = "Rebalance slot distribution in Valkey cluster"
	OpRestore     Op = "Restore Valkey cluster"
	OpAddHosts    Op = "Add hosts to Valkey cluster"
	OpDeleteHosts Op = "Delete hosts from Valkey cluster"
	OpModifyHosts Op = "Modify hosts in Valkey cluster"
	OpAddShard    Op = "Add shard to Valkey cluster"
	OpDeleteShard Op = "Delete shard from Valkey cluster"
	OpEnableSh    Op = "Enable sharding on Valkey cluster"
	OpDelete      Op = "Delete Valkey cluster"
)

func newPtr[T any](v T) *T {
	return &v
}

func mdbRedisClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"config.password", // not returned
			"hosts",           // todo change after fix import
			"access",
			"maintenance_window",
			"disk_size_autoscaling",
		},
	}
}

func testAccCheckMDBRedisClusterDestroy(s *terraform.State) error {
	config := test.AccProvider.(*provider.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_redis_cluster_v2" {
			continue
		}

		_, err := config.SDK.MDB().Redis().Cluster().Get(context.Background(), &redis.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Redis Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBRedisClusterExists(n string, r *redis.Cluster, hosts int, tlsEnabled, announceHostnames bool,
	persistenceMode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*provider.Provider).GetConfig()

		found, err := config.SDK.MDB().Redis().Cluster().Get(context.Background(), &redis.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Redis Cluster not found")
		}

		if found.TlsEnabled != tlsEnabled {
			return fmt.Errorf("tls mode: found = %t; expected = %t", found.TlsEnabled, tlsEnabled)
		}

		if found.AnnounceHostnames != announceHostnames {
			return fmt.Errorf("announceHostnames mode: found = %t; expected = %t", found.AnnounceHostnames, announceHostnames)
		}

		if found.GetPersistenceMode().String() != persistenceMode {
			return fmt.Errorf("persistence mode: found = %s; expected = %s", found.PersistenceMode, persistenceMode)
		}

		*r = *found

		resp, err := config.SDK.MDB().Redis().Cluster().ListHosts(context.Background(), &redis.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultTestMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != hosts {
			return fmt.Errorf("Expected %d hosts, got %d", hosts, len(resp.Hosts))
		}

		return nil
	}
}

func testAccCheckMDBRedisOperations(n string, ops []Op) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*provider.Provider).GetConfig()

		resp, err := config.SDK.MDB().Redis().Cluster().ListOperations(context.Background(), &redis.ListClusterOperationsRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}
		if len(resp.Operations) != len(ops) {
			return fmt.Errorf("Expected %d operations, got %d", len(ops), len(resp.Operations))
		}

		slices.Reverse(resp.Operations)
		for i, operation := range resp.Operations {
			if operation.Description != string(ops[i]) {
				return fmt.Errorf("Expected operation %s on position %d, got %s", operation.Description, i, string(ops[i]))
			}
		}

		return nil
	}
}

func testAccCheckMDBRedisClusterHasConfig(
	r *redis.Cluster,
	maxmemoryPolicy string,
	timeout int64,
	notifyKeyspaceEvents string,
	slowlogLogSlowerThan, slowlogMaxLen, databases int64,
	version, clientOutputBufferLimitNormal, clientOutputBufferLimitPubsub string,
	maxmemoryPercent int64,
	luaTimeLimit int64,
	replBacklogSizePercent int64,
	clusterRequireFullCoverage bool,
	clusterAllowReadsWhenDown bool,
	clusterAllowPubsubshardWhenDown bool,
	lfuDecayTime int64,
	lfuLogFactor int64,
	turnBeforeSwitchover bool,
	allowDataLoss bool,
	useLuajit bool,
	ioThreadsAllowed bool,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := mdb_redis_cluster_v2.FlattenConfig(r.Config)
		if c.MaxmemoryPolicy.ValueString() != maxmemoryPolicy {
			return fmt.Errorf("expected config.maxmemory_policy '%s', got '%s'", maxmemoryPolicy, c.MaxmemoryPolicy.ValueString())
		}
		if c.Timeout.ValueInt64() != timeout {
			return fmt.Errorf("expected config.timeout '%d', got '%d'", timeout, c.Timeout.ValueInt64())
		}
		if c.NotifyKeyspaceEvents.ValueString() != notifyKeyspaceEvents {
			return fmt.Errorf("expected config.notify_keyspace_events '%s', got '%s'", notifyKeyspaceEvents, c.NotifyKeyspaceEvents.ValueString())
		}
		if c.SlowlogLogSlowerThan.ValueInt64() != slowlogLogSlowerThan {
			return fmt.Errorf("expected config.slowlog_log_slower_than '%d', got '%d'", slowlogLogSlowerThan, c.SlowlogLogSlowerThan.ValueInt64())
		}
		if c.SlowlogMaxLen.ValueInt64() != slowlogMaxLen {
			return fmt.Errorf("expected config.slowlog_max_len '%d', got '%d'", slowlogMaxLen, c.SlowlogMaxLen.ValueInt64())
		}
		if c.Databases.ValueInt64() != databases {
			return fmt.Errorf("expected config.databases '%d', got '%d'", databases, c.Databases.ValueInt64())
		}
		if c.MaxmemoryPercent.ValueInt64() != maxmemoryPercent {
			return fmt.Errorf("expected config.maxmemory_percent '%d', got '%d'", maxmemoryPercent, c.MaxmemoryPercent.ValueInt64())
		}
		if c.Version.ValueString() != version {
			return fmt.Errorf("expected config.version '%s', got '%s'", version, c.Version.ValueString())
		}
		if c.ClientOutputBufferLimitNormal.ValueString() != clientOutputBufferLimitNormal {
			return fmt.Errorf("expected config.clientOutputBufferLimitNormal '%s', got '%s'",
				clientOutputBufferLimitNormal, c.ClientOutputBufferLimitNormal.String())
		}
		if c.ClientOutputBufferLimitPubsub.ValueString() != clientOutputBufferLimitPubsub {
			return fmt.Errorf("expected config.clientOutputBufferLimitPubsub '%s', got '%s'",
				clientOutputBufferLimitPubsub, c.ClientOutputBufferLimitPubsub.ValueString())
		}
		if c.LuaTimeLimit.ValueInt64() != luaTimeLimit {
			return fmt.Errorf("expected config.lua_time_limit '%d', got '%d'", luaTimeLimit, c.LuaTimeLimit.ValueInt64())
		}
		if c.ReplBacklogSizePercent.ValueInt64() != replBacklogSizePercent {
			return fmt.Errorf("expected config.repl_backlog_size_percent '%d', got '%d'", replBacklogSizePercent, c.ReplBacklogSizePercent.ValueInt64())
		}
		if c.ClusterRequireFullCoverage.ValueBool() != clusterRequireFullCoverage {
			return fmt.Errorf("expected config.cluster_require_full_coverage '%t', got '%t'", clusterRequireFullCoverage, c.ClusterRequireFullCoverage.ValueBool())
		}
		if c.ClusterAllowReadsWhenDown.ValueBool() != clusterAllowReadsWhenDown {
			return fmt.Errorf("expected config.cluster_allow_reads_when_down '%t', got '%t'", clusterAllowReadsWhenDown, c.ClusterAllowReadsWhenDown.ValueBool())
		}
		if c.ClusterAllowPubsubshardWhenDown.ValueBool() != clusterAllowPubsubshardWhenDown {
			return fmt.Errorf("expected config.cluster_allow_pubsubshard_when_down '%t', got '%t'", clusterAllowPubsubshardWhenDown, c.ClusterAllowPubsubshardWhenDown.ValueBool())
		}
		if c.LfuDecayTime.ValueInt64() != lfuDecayTime {
			return fmt.Errorf("expected config.lfu_decay_time '%d', got '%d'", lfuDecayTime, c.LfuDecayTime.ValueInt64())
		}
		if c.LfuLogFactor.ValueInt64() != lfuLogFactor {
			return fmt.Errorf("expected config.lfu_log_factor '%d', got '%d'", lfuLogFactor, c.LfuLogFactor.ValueInt64())
		}
		if c.TurnBeforeSwitchover.ValueBool() != turnBeforeSwitchover {
			return fmt.Errorf("expected config.turn_before_switchover '%t', got '%t'", turnBeforeSwitchover, c.TurnBeforeSwitchover.ValueBool())
		}
		if c.AllowDataLoss.ValueBool() != allowDataLoss {
			return fmt.Errorf("expected config.allow_data_loss '%t', got '%t'", allowDataLoss, c.AllowDataLoss.ValueBool())
		}
		if c.UseLuajit.ValueBool() != useLuajit {
			return fmt.Errorf("expected config.use_luajit '%t', got '%t'", useLuajit, c.UseLuajit.ValueBool())
		}
		if c.IoThreadsAllowed.ValueBool() != ioThreadsAllowed {
			return fmt.Errorf("expected config.io_threads_allowed '%t', got '%t'", ioThreadsAllowed, c.IoThreadsAllowed.ValueBool())
		}
		return nil
	}
}

func testAccCheckMDBRedisClusterHasResources(r *redis.Cluster, resourcePresetID string, diskSizeGb int,
	diskTypeId string) resource.TestCheckFunc {
	diskSize := int64(diskSizeGb * int(math.Pow(2, 30)))
	return func(s *terraform.State) error {
		rs := r.Config.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("Expected label with key '%d', got '%d'", diskSize, rs.DiskSize)
		}
		if rs.DiskTypeId != diskTypeId {
			return fmt.Errorf("Expected label with key '%s', got '%s'", diskTypeId, rs.DiskTypeId)
		}
		return nil
	}
}

func testAccCheckMDBRedisClusterHasShards(r *redis.Cluster, shards []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*provider.Provider).GetConfig()

		resp, err := config.SDK.MDB().Redis().Cluster().ListShards(context.Background(), &redis.ListClusterShardsRequest{
			ClusterId: r.Id,
			PageSize:  defaultTestMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Shards) != len(shards) {
			return fmt.Errorf("Expected %d shards, got %d", len(shards), len(resp.Shards))
		}
		for _, s := range shards {
			found := false
			for _, rs := range resp.Shards {
				if s == rs.Name {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("Shard '%s' not found", s)
			}
		}
		return nil
	}
}

func testAccCheckMDBRedisClusterContainsLabel(r *redis.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := r.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}
