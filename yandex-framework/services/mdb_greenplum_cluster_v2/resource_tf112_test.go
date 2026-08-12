//go:build tf1_12

package mdb_greenplum_cluster_v2_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccMDBGreenplumClusterV2PasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-greenplum-v2-password-wo")
	resourceName := "yandex_mdb_greenplum_cluster_v2.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testhelpers.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGreenplumClusterV2PasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "user_password"),
					resource.TestCheckNoResourceAttr(resourceName, "user_password_wo"),
					resource.TestCheckResourceAttr(resourceName, "user_password_wo_version", "1"),
				),
			},
			{
				Config:             testAccGreenplumClusterV2PasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccGreenplumClusterV2PasswordFieldsConfig(clusterName, `user_password = "legacyP@ssw0rd"
  user_password_wo = "writeOnlyP@ssw0rd"
  user_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`one \(and only one\)`),
			},
			{
				Config:      testAccGreenplumClusterV2PasswordFieldsConfig(clusterName, `user_password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config:      testAccGreenplumClusterV2PasswordFieldsConfig(clusterName, `user_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccGreenplumClusterV2PasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "user_password"),
					resource.TestCheckNoResourceAttr(resourceName, "user_password_wo"),
					resource.TestCheckResourceAttr(resourceName, "user_password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccGreenplumClusterV2PasswordWoConfig(name, password string, version int) string {
	return testAccGreenplumClusterV2PasswordFieldsConfig(name, fmt.Sprintf("user_password_wo = %q\n  user_password_wo_version = %d", password, version))
}

func testAccGreenplumClusterV2PasswordFieldsConfig(name, passwordFields string) string {
	config := testAccResourceYandexMdbGreenplumClusterV2_basic(name, "Greenplum V2 write-only password test")
	return strings.Replace(config, `user_password = "test-user-password"`, passwordFields, 1)
}
