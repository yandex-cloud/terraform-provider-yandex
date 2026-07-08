package mdb_mysql_backup_retention_policy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	mysqlv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	myBRPResource   = "yandex_mdb_mysql_backup_retention_policy.test"
	myBRPTestPrefix = "tf-mysql-brp"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccMDBMySQLBackupRetentionPolicy(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix(myBRPTestPrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create policy, verify attributes
			{
				Config: testAccMDBMySQLBRPConfig(clusterName, "keep-weekly", "Weekly backups for 30 days", 30, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLBRPExists(myBRPResource),
					resource.TestCheckResourceAttrSet(myBRPResource, "id"),
					resource.TestCheckResourceAttrSet(myBRPResource, "policy_id"),
					resource.TestCheckResourceAttrSet(myBRPResource, "cluster_id"),
					resource.TestCheckResourceAttr(myBRPResource, "policy_name", "keep-weekly"),
					resource.TestCheckResourceAttr(myBRPResource, "description", "Weekly backups for 30 days"),
					resource.TestCheckResourceAttr(myBRPResource, "retain_for_days", "30"),
					resource.TestCheckResourceAttr(myBRPResource, "cron.day_of_week", "1"),
					resource.TestCheckResourceAttr(myBRPResource, "cron.day_of_month", "*"),
					resource.TestCheckResourceAttr(myBRPResource, "cron.month", "*"),
				),
			},
			// Step 2: import
			{
				ResourceName:            myBRPResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccMDBMySQLBRPImportStateIdFunc(myBRPResource),
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			// Step 3: change retain_for_days — triggers replace (ForceNew), cluster stays
			{
				Config: testAccMDBMySQLBRPConfig(clusterName, "keep-weekly", "Weekly backups for 60 days", 60, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLBRPExists(myBRPResource),
					resource.TestCheckResourceAttr(myBRPResource, "retain_for_days", "60"),
					resource.TestCheckResourceAttr(myBRPResource, "description", "Weekly backups for 60 days"),
				),
			},
		},
	})
}

func testAccCheckMDBMySQLBRPExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		_, err := mysqlv1sdk.NewBackupRetentionPolicyClient(config.SDKv2).Get(
			context.Background(),
			&mysql.GetBackupRetentionPolicyRequest{
				ClusterId: rs.Primary.Attributes["cluster_id"],
				PolicyId:  rs.Primary.Attributes["policy_id"],
			},
		)
		return err
	}
}

func testAccMDBMySQLBRPImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["policy_id"]), nil
	}
}

const myBRPVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
`

func testAccMDBMySQLBRPClusterConfig(clusterName string) string {
	return fmt.Sprintf(myBRPVPCDependencies+`
resource "yandex_mdb_mysql_cluster_v2" "foo" {
  name        = "%s"
  description = "MySQL Backup Retention Policy Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  hosts = {
    "host" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
  }
}
`, clusterName)
}

func testAccMDBMySQLBRPConfig(clusterName, policyName, description string, retainForDays int, dayOfMonth, dayOfWeek, month string) string {
	return testAccMDBMySQLBRPClusterConfig(clusterName) + fmt.Sprintf(`
resource "yandex_mdb_mysql_backup_retention_policy" "test" {
  cluster_id      = yandex_mdb_mysql_cluster_v2.foo.id
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
