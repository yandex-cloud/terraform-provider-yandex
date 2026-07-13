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
				Config: testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(pgUserResourceNameAlice, "password"),
				),
			},
			{
				Config:             testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "password_wo_version", "2"),
				),
			},
		},
	})
}

// Test that specifying both password and password_wo produces an error
func TestAccMDBPostgreSQLUserPasswordConflict_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-postgresql-user-pw-conflict")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config:      testAccMDBPostgreSQLUserConfigPasswordConflict(clusterName),
				ExpectError: regexp.MustCompile(`only one of .password. or .password_wo. can be specified`),
			},
		},
	})
}

func testAccMDBPostgreSQLUserConfigPasswordWo(name string, passwordWoVersion int) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id          = yandex_mdb_postgresql_cluster.foo.id
	name                = "alice"
	password_wo         = "mysecureP@ssw0rd"
	password_wo_version = %d
	login               = true
	conn_limit          = 50
}`, passwordWoVersion)
}

// Create user with both password and password_wo — should fail
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
