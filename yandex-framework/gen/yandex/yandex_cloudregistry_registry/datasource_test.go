package yandex_cloudregistry_registry_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceCloudRegistry_byID(t *testing.T) {
	registryName := acctest.RandomWithPrefix("tf-registry")
	label := acctest.RandomWithPrefix("label-value")
	folderID := test.GetExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCloudRegistryConfig(registryName, folderID, label, true),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField("data.yandex_cloudregistry_registry.source", "registry_id"),
					resource.TestCheckResourceAttr("data.yandex_cloudregistry_registry.source",
						"name", registryName),
					resource.TestCheckResourceAttrSet("data.yandex_cloudregistry_registry.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_cloudregistry_registry.source",
						"folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_cloudregistry_registry.source",
						"labels.test_label", label),
					test.AccCheckCreatedAtAttr("data.yandex_cloudregistry_registry.source"),
				),
			},
		},
	})
}

func testAccDataSourceCloudRegistryConfig(name, folderID, labelValue string, useID bool) string {
	if useID {
		return testAccDataSourceCloudRegistryResourceConfig(name, folderID, "DOCKER", "LOCAL", labelValue) + cloudRegistryDataByIDConfig
	}

	return testAccDataSourceCloudRegistryResourceConfig(name, folderID, "DOCKER", "LOCAL", labelValue)
}

func testAccDataSourceCloudRegistryResourceConfig(name, folderID, kind, typeName, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "foobar" {
  name      = "%s"
  folder_id = "%s"
  kind      = "%s"
  type		= "%s"

  labels = {
    test_label = "%s"
  }
}
`, name, folderID, kind, typeName, labelValue)
}

const cloudRegistryDataByIDConfig = `
data "yandex_cloudregistry_registry" "source" {
  registry_id = "${yandex_cloudregistry_registry.foobar.id}"
}
`
