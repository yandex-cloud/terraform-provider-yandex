package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccDataSourceYandexIAMServiceAccountById(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Service Account desc"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckIAMServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataServiceAccountByName(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_iam_service_account.bar",
						"service_account_id"),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"name", accountName),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"description", accountDesc),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"folder_id", acceptanceenv.FolderID()),
				),
			},
		},
	})
}

func TestAccDataSourceYandexIAMServiceAccountByName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Service Account desc"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckIAMServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataServiceAccountById(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_iam_service_account.bar",
						"service_account_id"),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"name", accountName),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"description", accountDesc),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"folder_id", acceptanceenv.FolderID()),
				),
			},
		},
	})
}

func (suite testSuite) testAccCheckIAMServiceAccountDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account" {
			continue
		}

		client := iamsdk.NewServiceAccountClient(config.SDK)

		_, err := client.Get(context.Background(), &iam.GetServiceAccountRequest{
			ServiceAccountId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Service account still exists")
		}
	}

	return nil
}

func testAccDataServiceAccountByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iam_service_account" "bar" {
  name = "${yandex_iam_service_account.foo.name}"
}

resource "yandex_iam_service_account" "foo" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testAccDataServiceAccountById(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iam_service_account" "bar" {
  service_account_id = "${yandex_iam_service_account.foo.id}"
}

resource "yandex_iam_service_account" "foo" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}
