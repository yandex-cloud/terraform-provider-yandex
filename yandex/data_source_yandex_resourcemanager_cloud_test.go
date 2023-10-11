package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//revive:disable:var-naming
func TestAccDataSourceYandexResourceManagerCloud_byIDNotFound(t *testing.T) {
	notExistCloudID := acctest.RandString(18)

	CustomProvidersTest(t, DefaultAndEmptyFolderProviders(), resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceManagerCloud_byID(notExistCloudID),
				// "PermissionDenied" returned for non existed cloud id
				ExpectError: regexp.MustCompile("NotFound"),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerCloud_byDefaultID(t *testing.T) {
	defaultCloudID := getExampleCloudID()
	defaultCloudName := getExampleCloudName()

	CustomProvidersTest(t, DefaultAndEmptyFolderProviders(), resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceManagerCloud_byID(defaultCloudID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_resourcemanager_cloud.acceptance", "cloud_id"),
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

	CustomProvidersTest(t, DefaultAndEmptyFolderProviders(), resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceManagerCloud_byName(defaultCloudName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_resourcemanager_cloud.acceptance", "cloud_id"),
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

	CustomProvidersTest(t, DefaultAndEmptyFolderProviders(), resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckResourceManagerCloud_byName(cloudName),
				ExpectError: regexp.MustCompile(`failed to resolve data source cloud by name: cloud with name "` + cloudName + `" not found`),
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
