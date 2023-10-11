package yandex

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceYandexClientConfig(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceYandexClientConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.yandex_client_config.config", "cloud_id", os.Getenv("YC_CLOUD_ID")),
					resource.TestCheckResourceAttr("data.yandex_client_config.config", "folder_id", os.Getenv("YC_FOLDER_ID")),
					resource.TestCheckResourceAttr("data.yandex_client_config.config", "zone", os.Getenv("YC_ZONE")),
					resource.TestCheckResourceAttrSet("data.yandex_client_config.config", "iam_token"),
				),
			},
		},
	})
}

func testAccDataSourceYandexClientConfig() string {
	return `
data "yandex_client_config" "config" {}
`
}
