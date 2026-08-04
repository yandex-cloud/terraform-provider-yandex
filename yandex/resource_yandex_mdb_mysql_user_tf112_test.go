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

// Test that a MySQL User can be created with password_wo and updated by incrementing password_wo_version
func TestAccMDBMySQLUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-mysql-user-pw-wo")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLUserConfigPasswordWo(clusterName, "initialP@ssw0rd", 1, "MYSQL_NATIVE_PASSWORD"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "name", "john"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(mysqlUserResourceJohn, "password_wo"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "authentication_plugin", "MYSQL_NATIVE_PASSWORD"),
				),
			},
			{
				Config:             testAccMDBMySQLUserConfigPasswordWo(clusterName, "initialP@ssw0rd", 1, "MYSQL_NATIVE_PASSWORD"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:      testAccMDBMySQLUserConfigPasswordConflict(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`only one of .password. or .password_wo. can be specified`),
			},
			{
				Config:      testAccMDBMySQLUserConfigPasswordWoWithoutVersion(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must be\s+specified`),
			},
			{
				Config:      testAccMDBMySQLUserConfigPasswordWoVersionWithoutPassword(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must be\s+specified`),
			},
			{
				Config: testAccMDBMySQLUserConfigPasswordWo(clusterName, "rotatedP@ssw0rd", 2, "CACHING_SHA2_PASSWORD"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "name", "john"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(mysqlUserResourceJohn, "password_wo"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "authentication_plugin", "CACHING_SHA2_PASSWORD"),
				),
			},
		},
	})
}

func testAccMDBMySQLUserConfigPasswordWo(name, passwordWo string, passwordWoVersion int, authenticationPlugin string) string {
	return testAccMDBMySQLUserConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_mysql_user" "john" {
	cluster_id            = yandex_mdb_mysql_cluster.foo.id
	name                  = "john"
	password_wo           = "%s"
	password_wo_version   = %d
	authentication_plugin = "%s"

	permission {
		database_name = yandex_mdb_mysql_database.testdb.name
		roles         = ["ALL"]
	}
	}`, passwordWo, passwordWoVersion, authenticationPlugin)
}

func testAccMDBMySQLUserConfigPasswordConflict(name string) string {
	return testAccMDBMySQLUserConfigStep0(name) + `
resource "yandex_mdb_mysql_user" "john" {
	cluster_id          = yandex_mdb_mysql_cluster.foo.id
	name                = "john"
	password            = "mysecureP@ssw0rd"
	password_wo         = "mysecureP@ssw0rd"
	password_wo_version = 1
}`
}

func testAccMDBMySQLUserConfigPasswordWoWithoutVersion(name string) string {
	return testAccMDBMySQLUserConfigStep0(name) + `
resource "yandex_mdb_mysql_user" "john" {
	cluster_id  = yandex_mdb_mysql_cluster.foo.id
	name        = "john"
	password_wo = "mysecureP@ssw0rd"
}`
}

func testAccMDBMySQLUserConfigPasswordWoVersionWithoutPassword(name string) string {
	return testAccMDBMySQLUserConfigStep0(name) + `
resource "yandex_mdb_mysql_user" "john" {
	cluster_id          = yandex_mdb_mysql_cluster.foo.id
	name                = "john"
	password_wo_version = 1
}`
}
