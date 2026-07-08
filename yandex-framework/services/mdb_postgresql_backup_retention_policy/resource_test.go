package mdb_postgresql_backup_retention_policy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	postgresql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	postgresqlv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/postgresql/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	pgBRPClusterResource = "yandex_mdb_postgresql_cluster_v2.foo"
	pgBRPResource        = "yandex_mdb_postgresql_backup_retention_policy.test"
	pgBRPTestPrefix      = "tf-pg-brp"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccMDBPostgreSQLBackupRetentionPolicy(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix(pgBRPTestPrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create policy, verify attributes
			{
				Config: testAccMDBPGBRPConfig(clusterName, "keep-weekly", "Weekly backups for 30 days", 30, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGBRPExists(pgBRPResource),
					resource.TestCheckResourceAttrSet(pgBRPResource, "id"),
					resource.TestCheckResourceAttrSet(pgBRPResource, "policy_id"),
					resource.TestCheckResourceAttrSet(pgBRPResource, "cluster_id"),
					resource.TestCheckResourceAttr(pgBRPResource, "policy_name", "keep-weekly"),
					resource.TestCheckResourceAttr(pgBRPResource, "description", "Weekly backups for 30 days"),
					resource.TestCheckResourceAttr(pgBRPResource, "retain_for_days", "30"),
					resource.TestCheckResourceAttr(pgBRPResource, "cron.day_of_week", "1"),
					resource.TestCheckResourceAttr(pgBRPResource, "cron.day_of_month", "*"),
					resource.TestCheckResourceAttr(pgBRPResource, "cron.month", "*"),
				),
			},
			// Step 2: import
			{
				ResourceName:            pgBRPResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccMDBPGBRPImportStateIdFunc(pgBRPResource),
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			// Step 3: change retain_for_days — triggers replace (ForceNew), cluster stays
			{
				Config: testAccMDBPGBRPConfig(clusterName, "keep-weekly", "Weekly backups for 60 days", 60, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGBRPExists(pgBRPResource),
					resource.TestCheckResourceAttr(pgBRPResource, "retain_for_days", "60"),
					resource.TestCheckResourceAttr(pgBRPResource, "description", "Weekly backups for 60 days"),
				),
			},
		},
	})
}

func testAccCheckMDBPGBRPExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		_, err := postgresqlv1sdk.NewBackupRetentionPolicyClient(config.SDKv2).Get(
			context.Background(),
			&postgresql.GetBackupRetentionPolicyRequest{
				ClusterId: rs.Primary.Attributes["cluster_id"],
				PolicyId:  rs.Primary.Attributes["policy_id"],
			},
		)
		return err
	}
}

func testAccMDBPGBRPImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["policy_id"]), nil
	}
}

const pgBRPVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
`

func testAccMDBPGBRPClusterConfig(clusterName string) string {
	return fmt.Sprintf(pgBRPVPCDependencies+`
resource "yandex_mdb_postgresql_cluster_v2" "foo" {
  name        = "%s"
  description = "PostgreSQL Backup Retention Policy Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = "18"

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 16
      disk_type_id       = "network-ssd"
    }
  }

  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
  }
}
`, clusterName)
}

func testAccMDBPGBRPConfig(clusterName, policyName, description string, retainForDays int, dayOfMonth, dayOfWeek, month string) string {
	return testAccMDBPGBRPClusterConfig(clusterName) + fmt.Sprintf(`
resource "yandex_mdb_postgresql_backup_retention_policy" "test" {
  cluster_id      = yandex_mdb_postgresql_cluster_v2.foo.id
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
