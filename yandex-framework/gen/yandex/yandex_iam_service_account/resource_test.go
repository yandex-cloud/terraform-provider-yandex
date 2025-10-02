package yandex_iam_service_account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	iamv1sdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccServiceAccount_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	accountName := "a" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	folderID := test.GetExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccServiceAccountBasic(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountExists("yandex_iam_service_account.acceptance"),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "folder_id", folderID),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "name", accountName),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "description", accountDesc),
					test.AccCheckCreatedAtAttr("yandex_iam_service_account.acceptance"),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccServiceAccountBasic(accountName, accountDesc),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// Test that a service account resource can be created, updated, and destroyed
func TestAccServiceAccount_basic(t *testing.T) {
	t.Parallel()

	accountName := "a" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	accountDesc2 := "Terraform Test Update"
	folderID := test.GetExampleFolderID()
	uniqueID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
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
					test.AccCheckCreatedAtAttr("yandex_iam_service_account.acceptance"),
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
			{
				ResourceName:      "yandex_iam_service_account.acceptance",
				ImportState:       true,
				ImportStateVerify: true,
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

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := iamv1sdk.NewServiceAccountClient(config.SDKv2).Get(context.Background(), &iam.GetServiceAccountRequest{
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
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account" {
			continue
		}

		_, err := iamv1sdk.NewServiceAccountClient(config.SDKv2).Get(context.Background(), &iam.GetServiceAccountRequest{
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
