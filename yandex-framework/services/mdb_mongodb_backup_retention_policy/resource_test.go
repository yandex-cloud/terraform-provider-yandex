package mdb_mongodb_backup_retention_policy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mongodb "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	mongodbv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/mongodb/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	mgBRPResource   = "yandex_mdb_mongodb_backup_retention_policy.test"
	mgBRPTestPrefix = "tf-mongodb-brp"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccMDBMongoDBBackupRetentionPolicy(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix(mgBRPTestPrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create policy, verify attributes
			{
				Config: testAccMDBMongoDBBRPConfig(clusterName, "keep-weekly", "Weekly backups for 30 days", 30, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBBRPExists(mgBRPResource),
					resource.TestCheckResourceAttrSet(mgBRPResource, "id"),
					resource.TestCheckResourceAttrSet(mgBRPResource, "policy_id"),
					resource.TestCheckResourceAttrSet(mgBRPResource, "cluster_id"),
					resource.TestCheckResourceAttr(mgBRPResource, "policy_name", "keep-weekly"),
					resource.TestCheckResourceAttr(mgBRPResource, "description", "Weekly backups for 30 days"),
					resource.TestCheckResourceAttr(mgBRPResource, "retain_for_days", "30"),
					resource.TestCheckResourceAttr(mgBRPResource, "cron.day_of_week", "1"),
					resource.TestCheckResourceAttr(mgBRPResource, "cron.day_of_month", "*"),
					resource.TestCheckResourceAttr(mgBRPResource, "cron.month", "*"),
				),
			},
			// Step 2: import
			{
				ResourceName:            mgBRPResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccMDBBRPImportStateIdFunc(mgBRPResource),
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			// Step 3: change retain_for_days — triggers replace (ForceNew), cluster stays
			{
				Config: testAccMDBMongoDBBRPConfig(clusterName, "keep-weekly", "Weekly backups for 60 days", 60, "*", "1", "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBBRPExists(mgBRPResource),
					resource.TestCheckResourceAttr(mgBRPResource, "retain_for_days", "60"),
					resource.TestCheckResourceAttr(mgBRPResource, "description", "Weekly backups for 60 days"),
				),
			},
		},
	})
}

func testAccCheckMDBMongoDBBRPExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		_, err := mongodbv1sdk.NewBackupRetentionPolicyClient(config.SDKv2).Get(
			context.Background(),
			&mongodb.GetBackupRetentionPolicyRequest{
				ClusterId: rs.Primary.Attributes["cluster_id"],
				PolicyId:  rs.Primary.Attributes["policy_id"],
			},
		)
		return err
	}
}

func testAccMDBBRPImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["policy_id"]), nil
	}
}

const mgBRPVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
`

func testAccMDBMongoDBBRPClusterConfig(clusterName string) string {
	return fmt.Sprintf(mgBRPVPCDependencies+`
resource "yandex_mdb_mongodb_cluster" "foo" {
  name        = "%s"
  description = "MongoDB Backup Retention Policy Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  cluster_config {
    version = "8.0"
  }

  host {
    zone_id   = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }

  resources_mongod {
    resource_preset_id = "s2.micro"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }
}
`, clusterName)
}

func testAccMDBMongoDBBRPConfig(clusterName, policyName, description string, retainForDays int, dayOfMonth, dayOfWeek, month string) string {
	return testAccMDBMongoDBBRPClusterConfig(clusterName) + fmt.Sprintf(`
resource "yandex_mdb_mongodb_backup_retention_policy" "test" {
  cluster_id      = yandex_mdb_mongodb_cluster.foo.id
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
