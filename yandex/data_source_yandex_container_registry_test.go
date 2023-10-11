package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceContainerRegistry_byID(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	label := acctest.RandomWithPrefix("label-value")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceContainerRegistryConfig(registryName, folderID, label, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_container_registry.source", "registry_id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"name", registryName),
					resource.TestCheckResourceAttrSet("data.yandex_container_registry.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"labels.test_label", label),
					testAccCheckCreatedAtAttr("data.yandex_container_registry.source"),
				),
			},
		},
	})
}

func TestAccDataSourceContainerRegistry_byName(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	label := acctest.RandomWithPrefix("label-value")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceContainerRegistryConfig(registryName, folderID, label, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_container_registry.source", "registry_id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"name", registryName),
					resource.TestCheckResourceAttrSet("data.yandex_container_registry.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"labels.test_label", label),
					testAccCheckCreatedAtAttr("data.yandex_container_registry.source"),
				),
			},
		},
	})
}

func testAccDataSourceContainerRegistryConfig(folderID, name, labelValue string, useID bool) string {
	if useID {
		return testAccDataSourceContainerRegistryResourceConfig(folderID, name, labelValue) + containerRegistryDataByIDConfig
	}

	return testAccDataSourceContainerRegistryResourceConfig(folderID, name, labelValue) + containerRegistryDataByNameConfig
}

func testAccDataSourceContainerRegistryResourceConfig(folderID, name, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "foobar" {
  name     = "%s"
  folder_id = "%s"

  labels = {
    test_label = "%s"
  }
}
`, folderID, name, labelValue)
}

const containerRegistryDataByIDConfig = `
data "yandex_container_registry" "source" {
  registry_id = "${yandex_container_registry.foobar.id}"
}
`

const containerRegistryDataByNameConfig = `
data "yandex_container_registry" "source" {
  name = "${yandex_container_registry.foobar.name}"
}
`
