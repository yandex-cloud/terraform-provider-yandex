package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceContainerRegistry_byID(t *testing.T) {
	t.Parallel()

	containerRegistryName := acctest.RandomWithPrefix("tf-registry")
	label := acctest.RandomWithPrefix("label-value")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegisterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceContainerRegistryResourceConfig(containerRegistryName, folderID, label),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_container_registry.source", "registry_id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"name", containerRegistryName),
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

func testAccDataSourceContainerRegistryResourceConfig(folderID, name, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "foobar" {
  name     = "%s"
  folder_id = "%s"

  labels = {
    test_label = "%s"
  }
}
`, folderID, name, labelValue) + containerRegistryDataByIDConfig
}

const containerRegistryDataByIDConfig = `
data "yandex_container_registry" "source" {
  registry_id = "${yandex_container_registry.foobar.id}"
}
`
