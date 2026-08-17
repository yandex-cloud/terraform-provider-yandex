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

var clickHouseAdminPasswordWoPairError = regexp.MustCompile(
	`(?s)all of.*admin_password_wo,admin_password_wo_version.*must\s+be\s+specified`,
)

func TestAccMDBClickHouseClusterAdminPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-clickhouse-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseClusterAdminPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(chResourceFoo, "admin_password"),
					resource.TestCheckNoResourceAttr(chResourceFoo, "admin_password_wo"),
					resource.TestCheckResourceAttr(chResourceFoo, "admin_password_wo_version", "1"),
				),
			},
			{
				Config:             testAccMDBClickHouseClusterAdminPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBClickHouseClusterAdminPasswordFieldsConfig(clusterName, `admin_password = "legacyP@ssw0rd"
  admin_password_wo = "writeOnlyP@ssw0rd"
  admin_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`only one of .admin_password. or .admin_password_wo. can be specified`),
			},
			{
				Config:      testAccMDBClickHouseClusterAdminPasswordFieldsConfig(clusterName, `admin_password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: clickHouseAdminPasswordWoPairError,
			},
			{
				Config:      testAccMDBClickHouseClusterAdminPasswordFieldsConfig(clusterName, `admin_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: clickHouseAdminPasswordWoPairError,
			},
			{
				Config: testAccMDBClickHouseClusterAdminPasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(chResourceFoo, "admin_password"),
					resource.TestCheckNoResourceAttr(chResourceFoo, "admin_password_wo"),
					resource.TestCheckResourceAttr(chResourceFoo, "admin_password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBClickHouseClusterAdminPasswordWoConfig(name, password string, version int) string {
	return testAccMDBClickHouseClusterAdminPasswordFieldsConfig(name, fmt.Sprintf("admin_password_wo = %q\n  admin_password_wo_version = %d", password, version))
}

func testAccMDBClickHouseClusterAdminPasswordFieldsConfig(name, passwordFields string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name                    = %q
  environment             = "PRESTABLE"
  version                 = %q
  network_id              = yandex_vpc_network.mdb-ch-test-net.id
  sql_user_management     = true
  sql_database_management = true
  %s

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.mdb-ch-test-subnet-a.id
  }
}
`, name, chVersion, passwordFields)
}
