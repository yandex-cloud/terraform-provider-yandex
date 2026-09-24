//go:build tf1_12

package mdb_mongodb_user_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccMDBMongoDBUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-mongodb-user-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMongoDBUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password"),
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "password_wo_version", "1"),
				),
			},
			mdbMongoDBUserImportStep(mgUserResourceNameAlice),
			{
				Config:             testAccMDBMongoDBUserPasswordWoConfig(clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBMongoDBUserPasswordFieldsConfig(clusterName, `password = "legacyP@ssw0rd"
	password_wo = "writeOnlyP@ssw0rd"
	password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)(conflict|cannot be specified)`),
			},
			{
				Config:      testAccMDBMongoDBUserPasswordFieldsConfig(clusterName, `password_wo = "writeOnlyP@ssw0rd"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config:      testAccMDBMongoDBUserPasswordFieldsConfig(clusterName, `password_wo_version = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBMongoDBUserPasswordWoConfig(clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password"),
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "password_wo_version", "2"),
				),
			},
			{
				Config: testAccMDBMongoDBUserPasswordWoToIamConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "auth_type", "IAM"),
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password"),
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password_wo"),
					resource.TestCheckNoResourceAttr(mgUserResourceNameAlice, "password_wo_version"),
					testAccCheckMDBMongoDBUserHasAuthType(t, mgUserResourceNameAlice, mongodb.AuthType_AUTH_TYPE_IAM),
				),
			},
			{
				Config:             testAccMDBMongoDBUserPasswordWoToIamConfig(clusterName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccMDBMongoDBUserPasswordWoConfig(name, password string, version int) string {
	return testAccMDBMongoDBUserPasswordFieldsConfig(name, fmt.Sprintf("password_wo = %q\n\tpassword_wo_version = %d", password, version))
}

func testAccMDBMongoDBUserPasswordFieldsConfig(name, passwordFields string) string {
	config := testAccMDBMongoDBUserConfigStep1(name)
	return strings.Replace(config, `password   = "mysecureP@ssw0rd"`, passwordFields, 1)
}

func testAccMDBMongoDBUserPasswordWoToIamConfig(name string) string {
	// Keep the resource address so Terraform replaces the password user with an IAM user.
	return strings.Replace(testAccMDBMongoDBUserConfigIam(name), `"iam_user"`, `"alice"`, 1)
}
