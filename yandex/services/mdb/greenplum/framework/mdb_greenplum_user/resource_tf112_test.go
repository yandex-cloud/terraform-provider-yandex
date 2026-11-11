//go:build tf1_12

package mdb_greenplum_user_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccMDBGreenplumUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-greenplum-user-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBGreenplumUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password"),
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "password_wo_version", "1"),
				),
			},
			{
				Config:             testAccMDBGreenplumUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBGreenplumUserPasswordFieldsConfig(clusterName, `password = "legacyP@ssw0rd"
	password_wo = "writeOnlyP@ssw0rd"
	password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`one \(and only one\)`),
			},
			{
				Config:      testAccMDBGreenplumUserPasswordFieldsConfig(clusterName, ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`one \(and only one\)`),
			},
			{
				Config:      testAccMDBGreenplumUserPasswordFieldsConfig(clusterName, `password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config:      testAccMDBGreenplumUserPasswordFieldsConfig(clusterName, `password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBGreenplumUserPasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password"),
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBGreenplumUserPasswordWoConfig(name, password string, version int) string {
	return testAccMDBGreenplumUserPasswordFieldsConfig(name, fmt.Sprintf("password_wo = %q\n\tpassword_wo_version = %d", password, version))
}

func testAccMDBGreenplumUserPasswordFieldsConfig(name, passwordFields string) string {
	config := testAccMDBGreenplumUserConfigStep1(name)
	return strings.Replace(config, `password       = "mysecureP@ssw0rd"`, passwordFields, 1)
}
