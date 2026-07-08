package mdb_redis_backup_retention_policy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	redis "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	redisv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/redis/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	redisBRPResource   = "yandex_mdb_redis_backup_retention_policy.test"
	redisBRPTestPrefix = "tf-redis-brp"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccMDBRedisBackupRetentionPolicy(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix(redisBRPTestPrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create policy, verify attributes
			{
				Config: testAccMDBRedisBRPConfig(clusterName, "keep-weekly", "Weekly backups for 30 days", 30, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisBRPExists(redisBRPResource),
					resource.TestCheckResourceAttrSet(redisBRPResource, "id"),
					resource.TestCheckResourceAttrSet(redisBRPResource, "policy_id"),
					resource.TestCheckResourceAttrSet(redisBRPResource, "cluster_id"),
					resource.TestCheckResourceAttr(redisBRPResource, "policy_name", "keep-weekly"),
					resource.TestCheckResourceAttr(redisBRPResource, "description", "Weekly backups for 30 days"),
					resource.TestCheckResourceAttr(redisBRPResource, "retain_for_days", "30"),
					resource.TestCheckResourceAttr(redisBRPResource, "cron.day_of_week", "1"),
					resource.TestCheckResourceAttr(redisBRPResource, "cron.day_of_month", "*"),
					resource.TestCheckResourceAttr(redisBRPResource, "cron.month", "*"),
				),
			},
			// Step 2: import
			{
				ResourceName:            redisBRPResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccMDBRedisBRPImportStateIdFunc(redisBRPResource),
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			// Step 3: change retain_for_days — triggers replace (ForceNew), cluster stays
			{
				Config: testAccMDBRedisBRPConfig(clusterName, "keep-weekly", "Weekly backups for 60 days", 60, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisBRPExists(redisBRPResource),
					resource.TestCheckResourceAttr(redisBRPResource, "retain_for_days", "60"),
					resource.TestCheckResourceAttr(redisBRPResource, "description", "Weekly backups for 60 days"),
				),
			},
		},
	})
}

func testAccCheckMDBRedisBRPExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		_, err := redisv1sdk.NewBackupRetentionPolicyClient(config.SDKv2).Get(
			context.Background(),
			&redis.GetBackupRetentionPolicyRequest{
				ClusterId: rs.Primary.Attributes["cluster_id"],
				PolicyId:  rs.Primary.Attributes["policy_id"],
			},
		)
		return err
	}
}

func testAccMDBRedisBRPImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["policy_id"]), nil
	}
}

const redisBRPVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
`

func testAccMDBRedisBRPClusterConfig(clusterName string) string {
	return fmt.Sprintf(redisBRPVPCDependencies+`
resource "yandex_mdb_redis_cluster_v2" "foo" {
  name        = "%s"
  description = "Redis Backup Retention Policy Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config = {
    version  = "9.1-valkey"
    password = "passw0rdBRP!"
  }

  resources = {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
    disk_type_id       = "network-ssd"
  }

  hosts = {
    "aaa" = {
      zone      = "ru-central1-a"
	  subnet_id  = yandex_vpc_subnet.foo.id
	}
  }
}
`, clusterName)
}

func testAccMDBRedisBRPConfig(clusterName, policyName, description string, retainForDays int, dayOfMonth, dayOfWeek, month string) string {
	return testAccMDBRedisBRPClusterConfig(clusterName) + fmt.Sprintf(`
resource "yandex_mdb_redis_backup_retention_policy" "test" {
  cluster_id      = yandex_mdb_redis_cluster_v2.foo.id
  policy_name     = "%s"
  description     = "%s"
  retain_for_days = %d

  cron = {
    day_of_month = "%s"
    day_of_week  = "%s"
    month        = "%s"
  }
}
`, policyName, description, retainForDays, dayOfMonth, dayOfWeek, month)
}
