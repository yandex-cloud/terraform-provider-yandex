package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

// Test that a service account resource can be created, updated, and destroyed
func TestAccServiceAccount_basic(t *testing.T) {
	t.Parallel()

	accountName := "a" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	accountDesc2 := "Terraform Test Update"
	folderID := getExampleFolderID()
	uniqueID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// The first step creates a basic service account
			{
				Config: testAccServiceAccountBasic(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountExists("yandex_iam_service_account.acceptance"),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "folder_id", folderID),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "name", accountName),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "description", accountDesc),
					testAccCheckCreatedAtAttr("yandex_iam_service_account.acceptance"),
				),
			},
			// The second step updates the service account
			{
				Config: testAccServiceAccountBasic(accountName, accountDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountNameModified("yandex_iam_service_account.acceptance", accountDesc2),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "folder_id", folderID),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "name", accountName),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "description", accountDesc2),
					testAccStoreServiceAccountUniqueID(&uniqueID),
				),
			},
			// The third step explicitly adds the same default folderID to the service account configuration
			// and ensure the service account is not recreated by comparing the value of its ID with the one from the previous step
			{
				Config: testAccServiceAccountWithFolderID(folderID, accountName, accountDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountNameModified("yandex_iam_service_account.acceptance", accountDesc2),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "folder_id", folderID),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "name", accountName),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "description", accountDesc2),
					resource.TestCheckResourceAttrPtr(
						"yandex_iam_service_account.acceptance", "id", &uniqueID),
				),
			},
		},
	})
}

func testAccStoreServiceAccountUniqueID(uniqueID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*uniqueID = s.RootModule().Resources["yandex_iam_service_account.acceptance"].Primary.ID
		return nil
	}
}

func testAccCheckYandexIAMServiceAccountExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		return nil
	}
}

func testAccCheckYandexIAMServiceAccountExistsWithID(n string, sa *iam.ServiceAccount) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.IAM().ServiceAccount().Get(context.Background(), &iam.GetServiceAccountRequest{
			ServiceAccountId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Service account not found")
		}

		*sa = *found

		return nil
	}
}

func testAccCheckYandexIAMServiceAccountNameModified(r, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.Attributes["description"] != n {
			return fmt.Errorf("description is %q expected %q", rs.Primary.Attributes["description"], n)
		}

		return nil
	}
}

func testAccCheckIAMServiceAccountDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account" {
			continue
		}

		_, err := config.sdk.IAM().ServiceAccount().Get(context.Background(), &iam.GetServiceAccountRequest{
			ServiceAccountId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Service account still exists")
		}
	}

	return nil
}

func testAccServiceAccountBasic(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%v"
  description = "%v"
}
`, name, desc)
}

func testAccServiceAccountWithFolderID(folderID, name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  folder_id   = "%v"
  name        = "%v"
  description = "%v"
}
`, folderID, name, desc)
}
