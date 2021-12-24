package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

const cdnDataSourceOriginGroup = "data.yandex_cdn_origin_group.test-dev-ds"

func TestAccDataSourceCDNOriginGroup_byID(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf-test-cdn-origin-group-ds-%s", acctest.RandString(10))
	var originGroup cdn.OriginGroup

	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCDNOriginGroup_byID(groupName),
				Check: resource.ComposeTestCheckFunc(
					testOriginGroupExists("yandex_cdn_origin_group.bar_group", &originGroup),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "name", groupName),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "folder_id", folderID),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "use_next", "true"),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "origin.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceCDNOriginGroup_byName(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf-test-cdn-origin-group-%s", acctest.RandString(10))
	var originGroup cdn.OriginGroup

	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCDNOriginGroup_byName(groupName),
				Check: resource.ComposeTestCheckFunc(
					testOriginGroupExists("yandex_cdn_origin_group.bar_group", &originGroup),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "name", groupName),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "folder_id", folderID),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "use_next", "true"),
					resource.TestCheckResourceAttr(cdnDataSourceOriginGroup, "origin.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceCDNOriginGroup_byID(groupName string) string {
	return fmt.Sprintf(`
data "yandex_cdn_origin_group" "test-dev-ds" {
	origin_group_id = "${yandex_cdn_origin_group.bar_group.id}"
}

resource "yandex_cdn_origin_group" "bar_group" {
  name     = "%s"

  origin {
	source = "ya.ru"
  }

  origin {
	source = "yandex.ru"
  }
}
`, groupName)
}

func testAccDataSourceCDNOriginGroup_byName(groupName string) string {
	return fmt.Sprintf(`
data "yandex_cdn_origin_group" "test-dev-ds" {
	name = "${yandex_cdn_origin_group.bar_group.name}"
}

resource "yandex_cdn_origin_group" "bar_group" {
  name     = "%s"

  origin {
	source = "ya.ru"
  }

  origin {
	source = "yandex.ru"
  }
}
`, groupName)
}
