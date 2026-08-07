//go:build tf1_12

package mdb_redis_user_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccMDBRedisUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-redis-user-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBRedisUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(redisUserResourceNameAlice, "passwords"),
					resource.TestCheckNoResourceAttr(redisUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(redisUserResourceNameAlice, "password_wo_version", "1"),
				),
			},
			{
				Config:             testAccMDBRedisUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBRedisUserPasswordFieldsConfig(clusterName, `passwords = ["legacyP@ssw0rd"]
	password_wo = "writeOnlyP@ssw0rd"
	password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`one \(and only one\)`),
			},
			{
				Config:      testAccMDBRedisUserPasswordFieldsConfig(clusterName, `password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config:      testAccMDBRedisUserPasswordFieldsConfig(clusterName, `password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBRedisUserPasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(redisUserResourceNameAlice, "passwords"),
					resource.TestCheckNoResourceAttr(redisUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(redisUserResourceNameAlice, "password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBRedisUserPasswordWoConfig(name, password string, version int) string {
	return testAccMDBRedisUserPasswordFieldsConfig(name, fmt.Sprintf("password_wo = %q\n\tpassword_wo_version = %d", password, version))
}

func testAccMDBRedisUserPasswordFieldsConfig(name, passwordFields string) string {
	return testAccMDBRedisUserConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_redis_user" "alice" {
	cluster_id = yandex_mdb_redis_cluster_v2.foo.id
	name       = "alice"
	%s
}`, passwordFields)
}
