package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceCDNResource_byID(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf-test-ds-cdn-resource-%s", acctest.RandString(4))
	cname := fmt.Sprintf(
		"cdn.%s-yandex-test.net", acctest.RandomWithPrefix("tf-test-by-id"),
	)
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomCDNResourceConfig(groupName, cname, "", true),
				Check:  testAccCheckCDNResource(folderID, cname),
			},
		},
	})
}

func TestAccDataSourceCDNResource_byName(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf-test-ds-cdn-resource-%s", acctest.RandString(4))
	cname := fmt.Sprintf(
		"cdn.%s-yandex-test.net", acctest.RandomWithPrefix("tf-test-by-name"),
	)
	folderID := getExampleFolderID()

	sslCert := `
ssl_certificate {
	type = "not_used"
}
`
	testCheckFunc := resource.ComposeTestCheckFunc(
		testAccCheckCDNResource(folderID, cname),
		resource.TestCheckResourceAttr("data.yandex_cdn_resource.cdn_resource_ds", "ssl_certificate.0.type", "not_used"),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomCDNResourceConfig(groupName, cname, sslCert, false),
				Check:  testCheckFunc,
			},
		},
	})
}

func testAccCheckCDNResource(folderID, cname string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckResourceIDField("data.yandex_cdn_resource.cdn_resource_ds", "resource_id"),
		resource.TestCheckResourceAttr("data.yandex_cdn_resource.cdn_resource_ds", "cname", cname),
		resource.TestCheckResourceAttrSet("data.yandex_cdn_resource.cdn_resource_ds", "id"),
		resource.TestCheckResourceAttrSet("data.yandex_cdn_resource.cdn_resource_ds", "resource_id"),
		resource.TestCheckResourceAttr("data.yandex_cdn_resource.cdn_resource_ds", "folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_cdn_resource.cdn_resource_ds", "active", "false"),
		resource.TestCheckResourceAttr("data.yandex_cdn_resource.cdn_resource_ds", "secondary_hostnames.#", "0"),
		testAccCheckCreatedAtAttr("data.yandex_cdn_resource.cdn_resource_ds"),
	)
}

// TODO: resource options.
func testAccDataSourceCustomCDNResourceResourceConfig(groupName, cname, sslCert string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "cdn_group" {
	name     = "%s"

	origin {
		source = "ya.ru"
	}
}

resource "yandex_cdn_resource" "foo" {
	cname = "%s"

	active = false

	origin_protocol = "https"

	origin_group_name = yandex_cdn_origin_group.cdn_group.name

	%s
}

`, groupName, cname, sslCert)
}

func testAccDataSourceCustomCDNResourceConfig(groupName, cname, sslCert string, useID bool) string {
	if useID {
		return testAccDataSourceCustomCDNResourceResourceConfig(groupName, cname, sslCert) + cdnResourceDataByIDConfig
	}

	return testAccDataSourceCustomCDNResourceResourceConfig(groupName, cname, sslCert) + cdnResourceDataByNameConfig
}

const cdnResourceDataByIDConfig = `
data "yandex_cdn_resource" "cdn_resource_ds" {
  resource_id = "${yandex_cdn_resource.foo.id}"
}
`

const cdnResourceDataByNameConfig = `
data "yandex_cdn_resource" "cdn_resource_ds" {
  cname = "${yandex_cdn_resource.foo.cname}"
}
`
