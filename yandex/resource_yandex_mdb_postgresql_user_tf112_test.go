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

// Test that a PostgreSQL User can be created with password_wo and updated by incrementing password_wo_version
func TestAccMDBPostgreSQLUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-postgresql-user-pw-wo")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, "initialP@ssw0rd", 1, "USER_PASSWORD_ENCRYPTION_MD5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(pgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "user_password_encryption", "USER_PASSWORD_ENCRYPTION_MD5"),
				),
			},
			{
				Config:             testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, "initialP@ssw0rd", 1, "USER_PASSWORD_ENCRYPTION_MD5"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:      testAccMDBPostgreSQLUserConfigPasswordConflict(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`only one of .password. or .password_wo. can be specified`),
			},
			{
				Config:      testAccMDBPostgreSQLUserConfigPasswordWoWithoutVersion(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must\s+be\s+specified`),
			},
			{
				Config:      testAccMDBPostgreSQLUserConfigPasswordWoVersionWithoutPassword(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must\s+be\s+specified`),
			},
			{
				Config: testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, "rotatedP@ssw0rd", 2, "USER_PASSWORD_ENCRYPTION_SCRAM_SHA_256"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(pgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "user_password_encryption", "USER_PASSWORD_ENCRYPTION_SCRAM_SHA_256"),
				),
			},
		},
	})
}

func testAccMDBPostgreSQLUserConfigPasswordWo(name, passwordWo string, passwordWoVersion int, passwordEncryption string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id               = yandex_mdb_postgresql_cluster.foo.id
	name                     = "alice"
	password_wo              = "%s"
	password_wo_version      = %d
	user_password_encryption = "%s"
	login                    = true
	conn_limit               = 50
	}`, passwordWo, passwordWoVersion, passwordEncryption)
}

func testAccMDBPostgreSQLUserConfigPasswordConflict(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id          = yandex_mdb_postgresql_cluster.foo.id
	name                = "alice"
	password            = "mysecureP@ssw0rd"
	password_wo         = "mysecureP@ssw0rd"
	password_wo_version = 1
	login               = true
}`
}

func testAccMDBPostgreSQLUserConfigPasswordWoWithoutVersion(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id  = yandex_mdb_postgresql_cluster.foo.id
	name        = "alice"
	password_wo = "mysecureP@ssw0rd"
	login       = true
}`
}

func testAccMDBPostgreSQLUserConfigPasswordWoVersionWithoutPassword(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id          = yandex_mdb_postgresql_cluster.foo.id
	name                = "alice"
	password_wo_version = 1
	login               = true
}`
}
