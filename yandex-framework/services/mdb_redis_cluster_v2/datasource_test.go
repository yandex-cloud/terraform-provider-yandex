package mdb_redis_cluster_v2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBRedisClusterV2_byID(t *testing.T) {
	t.Parallel()

	redisName := acctest.RandomWithPrefix("ds-redisv2-by-id")
	redisDesc := "Redis Cluster Terraform Datasource Test #1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, "7.2", true),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster_v2.bar",
					"yandex_mdb_redis_cluster_v2.bar", redisName, redisDesc),
			},
		},
	})
}

func TestAccDataSourceMDBRedisClusterV2_byName(t *testing.T) {
	t.Parallel()

	redisName := acctest.RandomWithPrefix("ds-redisv2-by-name")
	redisDesc := "Redis Cluster Terraform Datasource Test #2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, "7.2", false),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster_v2.bar",
					"yandex_mdb_redis_cluster_v2.bar", redisName, redisDesc),
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
			"config.timeout", // Cannot test full config, because API doesn't return password
			"config.maxmemory_policy",
			"config.notify_keyspace_events",
			"config.slowlog_log_slower_than",
			"config.slowlog_max_len",
			"config.client_output_buffer_limit_normal",
			"config.client_output_buffer_limit_pubsub",
			"config.use_luajit",
			"config.io_threads_allowed",
			"config.databases",
			"config.maxmemory_percent",
			"config.lua_time_limit",
			"config.repl_backlog_size_percent",
			"config.cluster_require_full_coverage",
			"config.cluster_allow_reads_when_down",
			"config.cluster_allow_pubsubshard_when_down",
			"config.lfu_decay_time",
			"config.lfu_log_factor",
			"config.turn_before_switchover",
			"config.allow_data_loss",
			"config.version",
			"security_group_ids",
			"maintenance_window.type",
			"maintenance_window.day",
			"maintenance_window.hour",
			"deletion_protection",
			"disk_size_autoscaling.disk_size_limit",
			"disk_size_autoscaling.planned_usage_threshold",
			"disk_size_autoscaling.emergency_usage_threshold",
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

func testAccDataSourceMDBRedisClusterCheck(datasourceName string, resourceName string, redisName string, desc string) resource.TestCheckFunc {
	folderID := test.GetExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeAggregateTestCheckFunc(
		testAccDataSourceMDBRedisClusterAttributesCheck(datasourceName, resourceName),
		test.AccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", redisName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.foo", "bar"),
		resource.TestCheckResourceAttr(datasourceName, "labels.foo2", "bar2"),
		resource.TestCheckResourceAttr(datasourceName, "sharded", "false"),
		resource.TestCheckResourceAttr(datasourceName, "tls_enabled", "true"),
		resource.TestCheckResourceAttr(datasourceName, "persistence_mode", "OFF"),
		resource.TestCheckResourceAttr(datasourceName, "announce_hostnames", "true"),
		resource.TestCheckResourceAttr(datasourceName, "hosts.%", "1"),
		test.AccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(redisResource, "maintenance_window.type", "WEEKLY"),
		resource.TestCheckResourceAttr(redisResource, "maintenance_window.day", "MON"),
		resource.TestCheckResourceAttr(redisResource, "maintenance_window.hour", "1"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
		resource.TestCheckResourceAttr(datasourceName, "disk_size_autoscaling.disk_size_limit", "32"),
		//resource.TestCheckResourceAttr(datasourceName, "disk_size_autoscaling.planned_usage_threshold", "70"),
		resource.TestCheckResourceAttr(datasourceName, "disk_size_autoscaling.emergency_usage_threshold", "83"),
	)
}

const mdbRedisClusterByIDConfig = `
data "yandex_mdb_redis_cluster_v2" "bar" {
  cluster_id = "${yandex_mdb_redis_cluster_v2.bar.id}"
}
`

const mdbRedisClusterByNameConfig = `
data "yandex_mdb_redis_cluster_v2" "bar" {
  name = "${yandex_mdb_redis_cluster_v2.bar.name}"
}
`

func testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, version string, importByID bool) string {
	baseDiskSize := 16
	diskTypeId := "network-ssd"
	baseFlavor := "hm3-c2-m8"
	hosts := map[string]host{
		"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
	}
	conf := makeConfig(nil, testAccAllSettingsConfig(redisName, redisDesc, version, baseDiskSize, diskTypeId, baseFlavor, hosts), &redisConfigTest{DeletionProtection: newPtr(false)})

	if importByID {
		return conf + mdbRedisClusterByIDConfig
	}

	return conf + mdbRedisClusterByNameConfig
}
