package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceMDBRedisCluster_byID(t *testing.T) {
	t.Parallel()

	redisName := acctest.RandomWithPrefix("ds-redis-by-id")
	redisDesc := "Redis Cluster Terraform Datasource Test"
	persistenceMode := "OFF"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, nil, nil, persistenceMode,
					"7.2", true),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, nil, nil, persistenceMode),
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, nil, nil, persistenceMode,
					"7.2", true),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, nil, nil, persistenceMode),
			},
		},
	})
}

func TestAccDataSourceMDBRedisCluster_byName(t *testing.T) {
	t.Parallel()

	redisName := acctest.RandomWithPrefix("ds-redis-by-name")
	redisDesc := "Redis Cluster Terraform Datasource Test"
	tlsEnabled := true
	persistenceMode := "ON"
	announceHostnames := true

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, &tlsEnabled, &announceHostnames, persistenceMode,
					"7.2", false),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, &tlsEnabled, &announceHostnames, persistenceMode),
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, &tlsEnabled, &announceHostnames, persistenceMode,
					"7.2", false),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, &tlsEnabled, &announceHostnames, persistenceMode),
			},
		},
	})
}

func testAccDataSourceMDBRedisClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		if ds.Primary.ID != rs.Primary.ID {
			return fmt.Errorf("instance `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		instanceAttrsToTest := []string{
			"name",
			"folder_id",
			"network_id",
			"created_at",
			"description",
			"labels",
			"environment",
			"resources",
			"host",
			"sharded",
			"tls_enabled",
			"persistence_mode",
			"announce_hostnames",
			"config.0.timeout", // Cannot test full config, because API doesn't return password
			"config.0.maxmemory_policy",
			"config.0.notify_keyspace_events",
			"config.0.slowlog_log_slower_than",
			"config.0.slowlog_max_len",
			"config.0.client_output_buffer_limit_normal",
			"config.0.client_output_buffer_limit_pubsub",
			"config.0.use_luajit",
			"config.0.io_threads_allowed",
			"config.0.databases",
			"config.0.maxmemory_percent",
			"config.0.lua_time_limit",
			"config.0.repl_backlog_size_percent",
			"config.0.cluster_require_full_coverage",
			"config.0.cluster_allow_reads_when_down",
			"config.0.cluster_allow_pubsubshard_when_down",
			"config.0.lfu_decay_time",
			"config.0.lfu_log_factor",
			"config.0.turn_before_switchover",
			"config.0.allow_data_loss",
			"config.0.version",
			"security_group_ids",
			"maintenance_window.0.type",
			"maintenance_window.0.day",
			"maintenance_window.0.hour",
			"deletion_protection",
			"disk_size_autoscaling.0.disk_size_limit",
			"disk_size_autoscaling.0.planned_usage_threshold",
			"disk_size_autoscaling.0.emergency_usage_threshold",
		}

		for _, attrToCheck := range instanceAttrsToTest {
			if datasourceAttributes[attrToCheck] != resourceAttributes[attrToCheck] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck,
					datasourceAttributes[attrToCheck],
					resourceAttributes[attrToCheck],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceMDBRedisClusterCheck(datasourceName string, resourceName string, redisName string, desc string,
	tlsEnabled, announceHostnames *bool, persistenceMode string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"
	tlsEnabledStr := "false"
	if tlsEnabled != nil && *tlsEnabled {
		tlsEnabledStr = "true"
	}
	announceHostnamesStr := "false"
	if announceHostnames != nil && *announceHostnames {
		announceHostnamesStr = "true"
	}
	persistenceModeStr := "ON"
	if persistenceMode == "OFF" {
		persistenceModeStr = "OFF"
	}

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBRedisClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", redisName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "sharded", "false"),
		resource.TestCheckResourceAttr(datasourceName, "tls_enabled", tlsEnabledStr),
		resource.TestCheckResourceAttr(datasourceName, "persistence_mode", persistenceModeStr),
		resource.TestCheckResourceAttr(datasourceName, "announce_hostnames", announceHostnamesStr),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "1"),
		resource.TestCheckResourceAttrSet(datasourceName, "host.0.fqdn"),
		resource.TestCheckResourceAttr(datasourceName, "host.0.replica_priority", fmt.Sprintf("%d", defaultReplicaPriority)),
		resource.TestCheckResourceAttr(datasourceName, "host.0.assign_public_ip", "false"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(redisResource, "maintenance_window.0.type", "WEEKLY"),
		resource.TestCheckResourceAttr(redisResource, "maintenance_window.0.day", "FRI"),
		resource.TestCheckResourceAttr(redisResource, "maintenance_window.0.hour", "20"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
		resource.TestCheckResourceAttr(datasourceName, "disk_size_autoscaling.0.disk_size_limit", fmt.Sprintf("%d", mdbRedisDiskSizeGB*2)),
		resource.TestCheckResourceAttr(datasourceName, "disk_size_autoscaling.0.planned_usage_threshold", "70"),
		resource.TestCheckResourceAttr(datasourceName, "disk_size_autoscaling.0.emergency_usage_threshold", "85"),
	)
}

const mdbRedisClusterByIDConfig = `
data "yandex_mdb_redis_cluster" "bar" {
  cluster_id = "${yandex_mdb_redis_cluster.foo.id}"
}
`

const mdbRedisClusterByNameConfig = `
data "yandex_mdb_redis_cluster" "bar" {
  name = "${yandex_mdb_redis_cluster.foo.name}"
}
`

const mdbRedisDiskSizeGB = 16

func testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc string, tlsEnabled, announceHostnames *bool,
	persistenceMode, version string, useDataID bool) string {
	if useDataID {
		return testAccMDBRedisClusterConfigMainWithMW(redisName, redisDesc, "PRESTABLE", false,
			tlsEnabled, announceHostnames, persistenceMode, version, "hm2.nano", mdbRedisDiskSizeGB, "", "", "",
			[]*bool{nil}, []*int{nil}) + mdbRedisClusterByIDConfig
	}

	return testAccMDBRedisClusterConfigMainWithMW(redisName, redisDesc, "PRESTABLE", false,
		tlsEnabled, announceHostnames, persistenceMode, version, "hm2.nano", mdbRedisDiskSizeGB, "", "", "",
		[]*bool{nil}, []*int{nil}) + mdbRedisClusterByNameConfig
}
