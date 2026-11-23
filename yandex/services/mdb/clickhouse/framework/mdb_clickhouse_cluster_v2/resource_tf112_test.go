//go:build tf1_12

package mdb_clickhouse_cluster_v2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccMDBClickHouseClusterV2ExternalReplicaPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	name := acctest.RandomWithPrefix("tf-clickhouse-external-wo")
	target := acctest.RandomWithPrefix("tf-clickhouse-external-target")
	const replicaPath = "shard_group.0.external_shard.0.replica.0."
	configuration := func(version, weight, priority int) string {
		return testAccMDBClickHouseCluster_remoteServers(name, target, fmt.Sprintf(`
    external_shard {
      name = "external"
      weight = %d
      replica {
        host = yandex_mdb_clickhouse_cluster_v2.remote_servers_target.hosts["eh1"].fqdn
        port = 9440
        secure = true
        user = "admin"
        password_wo = var.external_replica_password
        password_wo_version = %d
        priority = %d
      }
    }
`, weight, version, priority)) + `
variable "external_replica_password" {
  type = string
  sensitive = true
  ephemeral = true
}
`
	}
	variables := func(password string) config.Variables {
		return config.Variables{"external_replica_password": config.StringVariable(password)}
	}
	checkPasswordState := func(version string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr(chResourceRemoteServers, replicaPath+"password"),
			resource.TestCheckNoResourceAttr(chResourceRemoteServers, replicaPath+"password_wo"),
			resource.TestCheckResourceAttr(chResourceRemoteServers, replicaPath+"password_wo_version", version),
		)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: configuration(1, 100, 0), ConfigVariables: variables("strong_external_password"),
				Check: checkPasswordState("1"),
			},
			{
				Config: configuration(1, 100, 0), ConfigVariables: variables("ignored_without_version_change"),
				PlanOnly: true, ExpectNonEmptyPlan: false,
			},
			{
				Config: configuration(1, 200, 10), ConfigVariables: variables("ignored_without_version_change"),
				Check: resource.ComposeTestCheckFunc(
					checkPasswordState("1"),
					resource.TestCheckResourceAttr(chResourceRemoteServers, "shard_group.0.external_shard.0.weight", "200"),
					resource.TestCheckResourceAttr(chResourceRemoteServers, replicaPath+"priority", "10"),
				),
			},
			{
				Config: configuration(2, 200, 10), ConfigVariables: variables("rotated_external_password"),
				Check: checkPasswordState("2"),
			},
			{
				Config: configuration(2, 200, 10), ConfigVariables: variables("rotated_external_password"),
				PlanOnly: true, ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccMDBClickHouseClusterV2AdminPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-clickhouse-v2-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseClusterV2AdminPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(chResource, "admin_password"),
					resource.TestCheckNoResourceAttr(chResource, "admin_password_wo"),
					resource.TestCheckResourceAttr(chResource, "admin_password_wo_version", "1"),
				),
			},
			{
				Config:             testAccMDBClickHouseClusterV2AdminPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBClickHouseClusterV2AdminPasswordFieldsConfig(clusterName, `admin_password = "legacyP@ssw0rd"
  admin_password_wo = "writeOnlyP@ssw0rd"
  admin_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)(conflict|cannot be specified)`),
			},
			{
				Config:      testAccMDBClickHouseClusterV2AdminPasswordFieldsConfig(clusterName, `admin_password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config:      testAccMDBClickHouseClusterV2AdminPasswordFieldsConfig(clusterName, `admin_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBClickHouseClusterV2AdminPasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(chResource, "admin_password"),
					resource.TestCheckNoResourceAttr(chResource, "admin_password_wo"),
					resource.TestCheckResourceAttr(chResource, "admin_password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBClickHouseClusterV2AdminPasswordWoConfig(name, password string, version int) string {
	return testAccMDBClickHouseClusterV2AdminPasswordFieldsConfig(name, fmt.Sprintf("admin_password_wo = %q\n  admin_password_wo_version = %d", password, version))
}

func testAccMDBClickHouseClusterV2AdminPasswordFieldsConfig(name, passwordFields string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
resource "yandex_mdb_clickhouse_cluster_v2" "foo" {
  name                    = %q
  environment             = "PRESTABLE"
  version                 = %q
  network_id              = yandex_vpc_network.mdb-ch-test-net.id
  sql_user_management     = true
  sql_database_management = true
  %s

  hosts = {
    "clickhouse" = {
      type       = "CLICKHOUSE"
      zone       = "ru-central1-a"
      subnet_id  = yandex_vpc_subnet.mdb-ch-test-subnet-a.id
      shard_name = "shard1"
    }
  }

  shards = {
    shard1 = {}
  }

  maintenance_window {
    type = "ANYTIME"
  }
}
`, name, chVersion, passwordFields)
}
