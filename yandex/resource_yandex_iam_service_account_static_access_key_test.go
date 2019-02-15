package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

// Test that a service account key can be created and destroyed
func TestAccServiceAccountStaticAccessKey_basic(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountStaticAccessKey(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGoogleServiceAccountStaticAccessKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttrSet(resourceName, "access_key"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
		},
	})
}

func testAccCheckGoogleServiceAccountStaticAccessKeyExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		config := testAccProvider.Meta().(*Config)

		_, err := config.sdk.IAM().AWSCompatibility().AccessKey().Get(context.Background(), &awscompatibility.GetAccessKeyRequest{
			AccessKeyId: rs.Primary.ID,
		})

		return err
	}
}

func testAccServiceAccountStaticAccessKey(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_static_access_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
}
`, name, desc)
}
