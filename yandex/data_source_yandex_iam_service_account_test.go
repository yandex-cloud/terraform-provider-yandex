package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceYandexIAMServiceAccountById(t *testing.T) {
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Service Account desc"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataServiceAccountByName(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_iam_service_account.bar",
						"service_account_id"),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"name", accountName),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"description", accountDesc),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"folder_id", getExampleFolderID()),
				),
			},
		},
	})
}

func TestAccDataSourceYandexIAMServiceAccountByName(t *testing.T) {
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Service Account desc"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataServiceAccountById(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_iam_service_account.bar",
						"service_account_id"),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"name", accountName),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"description", accountDesc),
					resource.TestCheckResourceAttr("data.yandex_iam_service_account.bar",
						"folder_id", getExampleFolderID()),
				),
			},
		},
	})
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
