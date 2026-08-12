//go:build tf1_12

package yandex

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var greenplumPasswordWoPairError = regexp.MustCompile(
	`(?s)all of.*user_password_wo,user_password_wo_version.*must\s+be\s+specified`,
)

func TestAccMDBGreenplumClusterPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-greenplum-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBGreenplumClusterPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(greenplumResourceFoo, "user_password"),
					resource.TestCheckNoResourceAttr(greenplumResourceFoo, "user_password_wo"),
					resource.TestCheckResourceAttr(greenplumResourceFoo, "user_password_wo_version", "1"),
				),
			},
			{
				Config:             testAccMDBGreenplumClusterPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBGreenplumClusterPasswordFieldsConfig(clusterName, `user_password = "legacyP@ssw0rd"
  user_password_wo = "writeOnlyP@ssw0rd"
  user_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`only one of .user_password. or .user_password_wo. can be specified`),
			},
			{
				Config:      testAccMDBGreenplumClusterPasswordFieldsConfig(clusterName, `user_password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: greenplumPasswordWoPairError,
			},
			{
				Config:      testAccMDBGreenplumClusterPasswordFieldsConfig(clusterName, `user_password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: greenplumPasswordWoPairError,
			},
			{
				Config: testAccMDBGreenplumClusterPasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(greenplumResourceFoo, "user_password"),
					resource.TestCheckNoResourceAttr(greenplumResourceFoo, "user_password_wo"),
					resource.TestCheckResourceAttr(greenplumResourceFoo, "user_password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBGreenplumClusterPasswordWoConfig(name, password string, version int) string {
	return testAccMDBGreenplumClusterPasswordFieldsConfig(name, fmt.Sprintf("user_password_wo = %q\n  user_password_wo_version = %d", password, version))
}

func testAccMDBGreenplumClusterPasswordFieldsConfig(name, passwordFields string) string {
	config := testAccMDBGreenplumClusterConfigStep0(name, "Greenplum write-only password test", "s2.micro") + "\n}"
	return strings.Replace(config, `user_password = "mysecurepassword"`, passwordFields, 1)
}
