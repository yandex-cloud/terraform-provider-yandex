package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

//revive:disable:var-naming
func TestAccDataSourceYandexResourceManagerCloud_byIDNotFound(t *testing.T) {
	notExistCloudID := acctest.RandString(18)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceManagerCloud_byID(notExistCloudID),
				// "PermissionDenied" returned for non existed cloud id
				ExpectError: regexp.MustCompile("PermissionDenied"),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerCloud_byDefaultID(t *testing.T) {
	defaultCloudID := getExampleCloudID()
	defaultCloudName := getExampleCloudName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceManagerCloud_byID(defaultCloudID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.yandex_resourcemanager_cloud.acceptance", "id", defaultCloudID),
					resource.TestCheckResourceAttr("data.yandex_resourcemanager_cloud.acceptance", "name", defaultCloudName),
					testAccCheckCreatedAtAttr("data.yandex_resourcemanager_cloud.acceptance"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerCloud_byDefaultCloudName(t *testing.T) {
	defaultCloudName := getExampleCloudName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceManagerCloud_byName(defaultCloudName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.yandex_resourcemanager_cloud.acceptance", "id"),
					resource.TestCheckResourceAttrSet("data.yandex_resourcemanager_cloud.acceptance", "created_at"),
					resource.TestCheckResourceAttr("data.yandex_resourcemanager_cloud.acceptance", "name", defaultCloudName),
					testAccCheckCreatedAtAttr("data.yandex_resourcemanager_cloud.acceptance"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerCloud_byName(t *testing.T) {
	cloudName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckResourceManagerCloud_byName(cloudName),
				ExpectError: regexp.MustCompile("cloud not found: " + cloudName),
			},
		},
	})
}

func testAccCheckResourceManagerCloud_byID(cloudID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_cloud" "acceptance" {
  cloud_id = "%s"
}
`, cloudID)
}

func testAccCheckResourceManagerCloud_byName(name string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_cloud" "acceptance" {
  name = "%s"
}
`, name)
}
