//go:build tf1_12

package mdb_clickhouse_user_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const clickHousePasswordWoUserResource = "yandex_mdb_clickhouse_user.alice"

func TestAccMDBClickHouseUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-clickhouse-user-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(clickHousePasswordWoUserResource, "password"),
					resource.TestCheckNoResourceAttr(clickHousePasswordWoUserResource, "password_wo"),
					resource.TestCheckResourceAttr(clickHousePasswordWoUserResource, "password_wo_version", "1"),
					resource.TestCheckResourceAttr(clickHousePasswordWoUserResource, "generate_password", "false"),
				),
			},
			{
				Config:             testAccMDBClickHouseUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBClickHouseUserPasswordFieldsConfig(clusterName, `password = "legacyP@ssw0rd"
	password_wo = "writeOnlyP@ssw0rd"
	password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)(conflict|cannot be specified)`),
			},
			{
				Config:      testAccMDBClickHouseUserPasswordFieldsConfig(clusterName, `password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config:      testAccMDBClickHouseUserPasswordFieldsConfig(clusterName, `password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBClickHouseUserPasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(clickHousePasswordWoUserResource, "password"),
					resource.TestCheckNoResourceAttr(clickHousePasswordWoUserResource, "password_wo"),
					resource.TestCheckResourceAttr(clickHousePasswordWoUserResource, "password_wo_version", "2"),
				),
			},
			{
				Config: testAccMDBClickHouseUserPasswordFieldsConfig(clusterName, `auth_method = "iam"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clickHousePasswordWoUserResource, "auth_method", "iam"),
					testAccCheckMDBClickHouseUserAuthMethod(clickHousePasswordWoUserResource, clickhouse.AuthMethod_AUTH_METHOD_IAM),
				),
			},
			{
				Config:             testAccMDBClickHouseUserPasswordFieldsConfig(clusterName, `auth_method = "iam"`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccMDBClickHouseUserPasswordWoConfig(name, password string, version int) string {
	return testAccMDBClickHouseUserPasswordFieldsConfig(name, fmt.Sprintf("password_wo = %q\n\tpassword_wo_version = %d", password, version))
}

func testAccMDBClickHouseUserPasswordFieldsConfig(name, passwordFields string) string {
	return testAccMDBClickHouseClusterConfigMain(name, "ClickHouse user write-only password test") + fmt.Sprintf(`
resource "yandex_mdb_clickhouse_user" "alice" {
	cluster_id = yandex_mdb_clickhouse_cluster.sewage.id
	name       = "alice"
	%s
}
`, passwordFields)
}
