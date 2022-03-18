package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, nil, persistenceMode,
					"6.2", true),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, nil, persistenceMode),
			},
		},
	})
}

func TestAccDataSourceMDBRedisCluster_byName(t *testing.T) {
	t.Parallel()

	redisName := acctest.RandomWithPrefix("ds-redis-by-name")
	redisDesc := "Redis Cluster Terraform Datasource Test"
	tlsEnabled := false
	persistenceMode := "ON"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, &tlsEnabled, persistenceMode,
					"6.2", false),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, &tlsEnabled, persistenceMode),
			},
		},
	})
}

func TestAccDataSourceMDBRedis6Cluster_byID(t *testing.T) {
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
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, nil, persistenceMode,
					"6.0", true),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, nil, persistenceMode),
			},
		},
	})
}

func TestAccDataSourceMDBRedis6Cluster_byName(t *testing.T) {
	t.Parallel()

	redisName := acctest.RandomWithPrefix("ds-redis-by-name")
	redisDesc := "Redis Cluster Terraform Datasource Test"
	tlsEnabled := true
	persistenceMode := "ON"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc, &tlsEnabled, persistenceMode,
					"6.0", false),
				Check: testAccDataSourceMDBRedisClusterCheck(
					"data.yandex_mdb_redis_cluster.bar",
					"yandex_mdb_redis_cluster.foo", redisName, redisDesc, &tlsEnabled, persistenceMode),
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
			"config.0.timeout", // Cannot test full config, because API doesn't return password
			"config.0.maxmemory_policy",
			"config.0.notify_keyspace_events",
			"config.0.slowlog_log_slower_than",
			"config.0.slowlog_max_len",
			"config.0.databases",
			"config.0.version",
			"security_group_ids",
			"maintenance_window.0.type",
			"maintenance_window.0.day",
			"maintenance_window.0.hour",
			"deletion_protection",
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
	tlsEnabled *bool, persistenceMode string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"
	tlsEnabledStr := "false"
	if tlsEnabled != nil && *tlsEnabled {
		tlsEnabledStr = "true"
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
		resource.TestCheckResourceAttr(datasourceName, "host.#", "1"),
		resource.TestCheckResourceAttrSet(datasourceName, "host.0.fqdn"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.type", "WEEKLY"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.day", "FRI"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.hour", "20"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
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

func testAccDataSourceMDBRedisClusterConfig(redisName, redisDesc string, tlsEnabled *bool, persistenceMode, version string,
	useDataID bool) string {
	if useDataID {
		return testAccMDBRedisClusterConfigMain(redisName, redisDesc, "PRESTABLE", false,
			tlsEnabled, persistenceMode, version, "hm1.nano", 16, "") + mdbRedisClusterByIDConfig
	}

	return testAccMDBRedisClusterConfigMain(redisName, redisDesc, "PRESTABLE", false,
		tlsEnabled, persistenceMode, version, "hm1.nano", 16, "") + mdbRedisClusterByNameConfig
}
