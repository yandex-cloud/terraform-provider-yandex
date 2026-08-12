//go:build tf1_12

package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var redisPasswordWoPairError = regexp.MustCompile(
	`(?s)all of.*config\.0\.password_wo,config\.0\.password_wo_version.*must\s+be\s+specified`,
)

func TestAccMDBRedisClusterPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-redis-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBRedisClusterConfigPasswordWo(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					// SDKv2 materializes an omitted optional string inside a TypeList as its zero value.
					resource.TestCheckResourceAttr(redisResourceFoo, "config.0.password", ""),
					resource.TestCheckNoResourceAttr(redisResourceFoo, "config.0.password_wo"),
					resource.TestCheckResourceAttr(redisResourceFoo, "config.0.password_wo_version", "1"),
				),
			},
			{
				Config:             testAccMDBRedisClusterConfigPasswordWo(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBRedisClusterConfigPasswordFields(clusterName, `password = "legacyP@ssw0rd"
    password_wo = "writeOnlyP@ssw0rd"
    password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`only one of .config.password. or .config.password_wo. can be specified`),
			},
			{
				Config:      testAccMDBRedisClusterConfigPasswordFields(clusterName, `password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: redisPasswordWoPairError,
			},
			{
				Config:      testAccMDBRedisClusterConfigPasswordFields(clusterName, `password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: redisPasswordWoPairError,
			},
			{
				Config: testAccMDBRedisClusterConfigPasswordWo(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(redisResourceFoo, "config.0.password", ""),
					resource.TestCheckNoResourceAttr(redisResourceFoo, "config.0.password_wo"),
					resource.TestCheckResourceAttr(redisResourceFoo, "config.0.password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBRedisClusterConfigPasswordWo(name, password string, version int) string {
	return testAccMDBRedisClusterConfigPasswordFields(name, fmt.Sprintf("password_wo = %q\n    password_wo_version = %d", password, version))
}

func testAccMDBRedisClusterConfigPasswordFields(name, passwordFields string) string {
	return fmt.Sprintf(redisVPCDependencies+`
resource "yandex_mdb_redis_cluster" "foo" {
  name        = %q
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    %s
    version = "8.1-valkey"
  }

  resources {
    resource_preset_id = "hm3-c2-m8"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}
`, name, passwordFields)
}
