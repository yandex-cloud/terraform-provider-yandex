package acceptance

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const cdnDataSourceOriginGroup = "data.yandex_cdn_origin_group.test-dev-ds"

func TestAccDataSourceCDNOriginGroup_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	groupName := fmt.Sprintf("tf-test-cdn-origin-group-ds-%s", acctest.RandString(10))
	var originGroup cdn.OriginGroup

	folderID := acceptanceenv.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCDNOriginGroup_byID(groupName),
				Check: resource.ComposeTestCheckFunc(
					suite.testOriginGroupExists("yandex_cdn_origin_group.bar_group", &originGroup),
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
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	groupName := fmt.Sprintf("tf-test-cdn-origin-group-%s", acctest.RandString(10))
	var originGroup cdn.OriginGroup

	folderID := acceptanceenv.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCDNOriginGroup_byName(groupName),
				Check: resource.ComposeTestCheckFunc(
					suite.testOriginGroupExists("yandex_cdn_origin_group.bar_group", &originGroup),
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
